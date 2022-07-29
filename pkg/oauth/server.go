// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/v3/pkg/account/session"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	oauth_store "go.thethings.network/lorawan-stack/v3/pkg/oauth/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// Server is the interface for the OAuth server.
type Server interface {
	web.Registerer

	Authorize(authorizePage http.Handler) http.HandlerFunc
	Token(w http.ResponseWriter, r *http.Request)
}

type server struct {
	c             *component.Component
	config        Config
	osinConfig    *osin.ServerConfig
	store         oauth_store.TransactionalInterface
	session       session.Session
	generateCSP   func(config *Config, nonce string) string
	schemaDecoder *schema.Decoder
}

type sessionStore struct {
	oauth_store.TransactionalInterface
}

// Transact implements oauth_store.Interface.
func (s *sessionStore) Transact(ctx context.Context, f func(ctx context.Context, st session.Store) error) error {
	return s.TransactionalInterface.Transact(ctx, func(ctx context.Context, st oauth_store.Interface) error { return f(ctx, st) })
}

// NewServer returns a new OAuth server on top of the given store.
func NewServer(c *component.Component, store oauth_store.TransactionalInterface, config Config, cspFunc func(config *Config, nonce string) string) (Server, error) {
	s := &server{
		c:             c,
		config:        config,
		store:         store,
		session:       session.Session{Store: &sessionStore{store}},
		generateCSP:   cspFunc,
		schemaDecoder: schema.NewDecoder(),
	}
	s.schemaDecoder.IgnoreUnknownKeys(true)

	if s.config.Mount == "" {
		s.config.Mount = s.config.UI.MountPath()
	}

	s.osinConfig = &osin.ServerConfig{
		AuthorizationExpiration: int32((5 * time.Minute).Seconds()),
		AccessExpiration:        int32(time.Hour.Seconds()),
		TokenType:               "bearer",
		AllowedAuthorizeTypes: osin.AllowedAuthorizeType{
			osin.CODE,
		},
		AllowedAccessTypes: osin.AllowedAccessType{
			osin.AUTHORIZATION_CODE,
			osin.REFRESH_TOKEN,
			osin.PASSWORD,
		},
		ErrorStatusCode:           http.StatusBadRequest,
		AllowClientSecretInParams: true,
		RedirectUriSeparator:      redirectURISeparator,
		RetainTokenAfterRefresh:   false,
	}

	return s, nil
}

type ctxKeyType struct{}

var ctxKey ctxKeyType

func (s *server) configFromContext(ctx context.Context) *Config {
	if config, ok := ctx.Value(ctxKey).(*Config); ok {
		return config
	}
	return &s.config
}

func (s *server) now() time.Time { return time.Now().UTC() }

func (s *server) oauth2(ctx context.Context) *osin.Server {
	oauth2 := osin.NewServer(s.osinConfig, &storage{
		ctx:   ctx,
		store: s.store,
	})
	oauth2.AuthorizeTokenGen = s
	oauth2.AccessTokenGen = s
	oauth2.Now = s.now
	oauth2.Logger = &osinLogger{ctx: ctx}
	return oauth2
}

const (
	osinErrorFormat             = "error=%v, internal_error=%#v "
	osinAuthCodeErrorFormat     = "auth_code_request=%s"
	osinRefreshTokenErrorFormat = "refresh_token=%s"
)

type osinLogger struct {
	ctx context.Context
}

func (l *osinLogger) Printf(format string, v ...interface{}) {
	logger := log.FromContext(l.ctx)
	if strings.HasPrefix(format, osinErrorFormat) && len(v) >= 2 {
		format = strings.TrimPrefix(format, osinErrorFormat)
		logger = logger.WithField("oauth_error", v[0])
		if err, ok := v[1].(error); ok {
			logger = logger.WithField("oauth_error_cause", err)
		}
		v = v[2:]
		if len(v) >= 1 {
			switch format {
			case osinAuthCodeErrorFormat:
				logger.WithField("oauth_error_message", v[0]).Warn("OAuth authorization_code error")
				return
			case osinRefreshTokenErrorFormat:
				logger.WithField("oauth_error_message", v[0]).Warn("OAuth refresh_token error")
				return
			}
		}
	}
	logger.Warnf("OAuth internal error: "+format, v...)
}

// These errors map to errors in the osin library.
var (
	errInvalidRequest          = errors.DefineInvalidArgument("invalid_request", "invalid or missing request parameter")
	errUnauthorizedClient      = errors.DefinePermissionDenied("unauthorized_client", "client is not authorized to request a token using this method")
	errAccessDenied            = errors.DefinePermissionDenied("access_denied", "access denied")
	errUnsupportedResponseType = errors.DefineUnimplemented("unsupported_response_type", "unsupported response type")
	errInvalidScope            = errors.DefineInvalidArgument("invalid_scope", "invalid scope")
	errUnsupportedGrantType    = errors.DefineUnimplemented("unsupported_grant_type", "unsupported grant type")
	errInvalidGrant            = errors.DefinePermissionDenied("invalid_grant", "invalid, expired or revoked authorization code")
	errInvalidClient           = errors.DefinePermissionDenied("invalid client", "invalid or unauthenticated client")
	errInternal                = errors.Define("internal", "internal error {id}")
	errInvalidRedirectURI      = errors.DefinePermissionDenied("invalid_redirect_uri", "invalid redirect URI")
)

func (s *server) output(w http.ResponseWriter, r *http.Request, resp *osin.Response) {
	headers := w.Header()
	for i, k := range resp.Headers {
		for _, v := range k {
			headers.Add(i, v)
		}
	}

	var osinErr error
	if resp.IsError {
		switch resp.ErrorId {
		case osin.E_INVALID_REQUEST:
			osinErr = errInvalidRequest
		case osin.E_UNAUTHORIZED_CLIENT:
			osinErr = errUnauthorizedClient
		case osin.E_ACCESS_DENIED:
			osinErr = errAccessDenied
		case osin.E_UNSUPPORTED_RESPONSE_TYPE:
			osinErr = errUnsupportedResponseType
		case osin.E_INVALID_SCOPE:
			osinErr = errInvalidScope
		case osin.E_UNSUPPORTED_GRANT_TYPE:
			osinErr = errUnsupportedGrantType
		case osin.E_INVALID_GRANT:
			osinErr = errInvalidGrant
		case osin.E_INVALID_CLIENT:
			osinErr = errInvalidClient
		default:
			osinErr = errInternal
		}
		if resp.InternalError != nil {
			if ttnErr, ok := errors.From(resp.InternalError); ok {
				osinErr = ttnErr
			} else if _, isURIValidationError := resp.InternalError.(osin.UriValidationError); isURIValidationError {
				osinErr = errInvalidRedirectURI.WithCause(resp.InternalError)
			} else {
				osinErr = osinErr.(*errors.Definition).WithCause(resp.InternalError)
			}
		}
		log.FromContext(r.Context()).WithError(osinErr).Warn("OAuth error")
	}

	if resp.Type == osin.REDIRECT {
		location, err := resp.GetRedirectUrl()
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		uiMount := strings.TrimSuffix(s.config.UI.MountPath(), "/")
		if strings.HasPrefix(location, "/code") || strings.HasPrefix(location, "/local-callback") {
			location = uiMount + location
		}
		http.Redirect(w, r, location, http.StatusFound)
		return
	}

	if osinErr != nil {
		webhandlers.Error(w, r, osinErr)
		return
	}

	webhandlers.JSON(w, r, resp.Output)
}

func (s *server) RegisterRoutes(server *web.Server) {
	router := server.PrefixWithRedirect(s.config.Mount).Subrouter()
	router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r, nonce := webui.WithNonce(r)
				cspString := s.generateCSP(s.configFromContext(r.Context()), nonce)
				w.Header().Set("Content-Security-Policy", cspString)
				next.ServeHTTP(w, r)
			})
		},
		ratelimit.HTTPMiddleware(s.c.RateLimiter(), "http:oauth"),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				config := s.configFromContext(r.Context())
				r = webui.WithTemplateData(r, config.UI.TemplateData)
				frontendConfig := config.UI.FrontendConfig
				frontendConfig.Language = config.UI.TemplateData.Language
				r = webui.WithAppConfig(r, struct {
					FrontendConfig
				}{
					FrontendConfig: frontendConfig,
				})
				next.ServeHTTP(w, r)
			})
		},
		webhandlers.WithErrorHandlers(map[string]http.Handler{
			"text/html": webui.Template,
		}),
	)

	csrfMiddleware := webmiddleware.CSRF(
		s.config.CSRFAuthKey,
		csrf.CookieName("_csrf"),
		csrf.FieldName("_csrf"),
		csrf.Path("/"),
	)

	page := router.NewRoute().Subrouter()
	page.Use(mux.MiddlewareFunc(csrfMiddleware))

	// The logout route is currently in use by existing OAuth clients. As part of
	// the public API it should not be removed in this major.
	page.Path("/logout").HandlerFunc(s.ClientLogout).Methods(http.MethodGet)

	authorizeHandler := s.redirectToLogin(s.Authorize(webui.Template))
	page.Path("/authorize").Handler(authorizeHandler).Methods(http.MethodGet, http.MethodPost)

	router.Path("/local-callback").HandlerFunc(s.redirectToLocal).Methods(http.MethodGet)

	// No CSRF here:
	router.Path("/token").HandlerFunc(s.Token).Methods(http.MethodPost)
}
