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
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/pkg/errors"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// Server is the interface for the OAuth server.
type Server interface {
	web.Registerer

	Login(c echo.Context) error
	CurrentUser(c echo.Context) error
	Logout(c echo.Context) error
	Authorize(authorizePage echo.HandlerFunc) echo.HandlerFunc
	Token(c echo.Context) error
}

type server struct {
	ctx        context.Context
	config     Config
	osinConfig *osin.ServerConfig
	store      Store
}

// Store used by the OAuth server.
type Store interface {
	// UserStore and UserSessionStore are needed for user login/logout.
	store.UserStore
	store.UserSessionStore
	// ClientStore is needed for getting the OAuth client.
	store.ClientStore
	// OAuth is needed for OAuth authorizations.
	store.OAuthStore
}

// UIConfig is the combined configuration for the OAuth UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// FrontendConfig is the configuration for the OAuth frontend.
type FrontendConfig struct {
	Language string          `json:"language" name:"-"`
	IS       webui.APIConfig `json:"is" name:"is"`
}

// Config is the configuration for the OAuth server.
type Config struct {
	Mount string   `name:"mount" description:"Path on the server where the OAuth server will be served"`
	UI    UIConfig `name:"ui"`
}

// NewServer returns a new OAuth server on top of the given store.
func NewServer(ctx context.Context, store Store, config Config) Server {
	s := &server{
		ctx:    ctx,
		config: config,
		store:  store,
	}

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

	return s
}

func (s *server) now() time.Time { return time.Now().UTC() }

func (s *server) oauth2(ctx context.Context) *osin.Server {
	oauth2 := osin.NewServer(s.osinConfig, &storage{
		ctx:     ctx,
		clients: s.store,
		oauth:   s.store,
	})
	oauth2.AuthorizeTokenGen = s
	oauth2.AccessTokenGen = s
	oauth2.Now = s.now
	oauth2.Logger = s
	return oauth2
}

func (s *server) Printf(format string, v ...interface{}) {
	log.FromContext(s.ctx).Warnf(format, v...)
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

func (s *server) output(c echo.Context, resp *osin.Response) error {
	headers := c.Response().Header()
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
				osinErr = osinErr.(errors.Definition).WithCause(resp.InternalError)
			}
		}
		log.FromContext(c.Request().Context()).WithError(osinErr).Warn("OAuth error")
	}

	if resp.Type == osin.REDIRECT {
		location, err := resp.GetRedirectUrl()
		if err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, location)
	}

	if osinErr != nil {
		return osinErr
	}

	return c.JSON(resp.StatusCode, resp.Output)
}

func (s *server) RegisterRoutes(server *web.Server) {
	group := server.Group(
		s.config.Mount,
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set("template_data", s.config.UI.TemplateData)
				frontendConfig := s.config.UI.FrontendConfig
				frontendConfig.Language = s.config.UI.TemplateData.Language
				c.Set("app_config", struct {
					FrontendConfig
				}{
					FrontendConfig: frontendConfig,
				})
				return next(c)
			}
		},
		web_errors.ErrorMiddleware(map[string]web_errors.ErrorRenderer{
			"text/html": webui.Template,
		}),
	)

	api := group.Group("/api", middleware.CSRF())
	api.POST("/auth/login", s.Login)
	api.POST("/auth/logout", s.Logout, s.requireLogin)
	api.GET("/me", s.CurrentUser, s.requireLogin)

	page := group.Group("", middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
	page.GET("/login", webui.Template.Handler, s.redirectToNext)
	page.GET("/authorize", s.Authorize(webui.Template.Handler), s.redirectToLogin)
	page.POST("/authorize", s.Authorize(webui.Template.Handler), s.redirectToLogin)

	if s.config.Mount != "" && s.config.Mount != "/" {
		group.GET("", webui.Template.Handler, middleware.CSRF())
	}
	group.GET("/*", webui.Template.Handler, middleware.CSRF())

	// No CSRF here:
	group.GET("/code", webui.Template.Handler)
	group.GET("/local-callback", s.redirectToLocal)
	group.POST("/token", s.Token)
}
