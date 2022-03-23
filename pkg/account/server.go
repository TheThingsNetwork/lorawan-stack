// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package account

import (
	"context"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	sess "go.thethings.network/lorawan-stack/v3/pkg/account/session"
	account_store "go.thethings.network/lorawan-stack/v3/pkg/account/store"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// Server is the interface for the account app server.
type Server interface {
	web.Registerer

	Login(w http.ResponseWriter, r *http.Request)
	CurrentUser(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

// Component represents the Component to the Account app.
type Component interface {
	Context() context.Context
	RateLimiter() ratelimit.Interface
}

type server struct {
	c             Component
	config        Config
	store         account_store.TransactionalInterface
	session       sess.Session
	generateCSP   func(config *oauth.Config, nonce string) string
	schemaDecoder *schema.Decoder
}

// NewServer returns a new account app on top of the given store.
func NewServer(
	c *component.Component,
	store account_store.TransactionalInterface,
	config Config,
	cspFunc func(config *oauth.Config, nonce string) string,
) (Server, error) {
	s := &server{
		c:             c,
		config:        config,
		store:         store,
		session:       sess.Session{Store: store, Config: config.Session},
		generateCSP:   cspFunc,
		schemaDecoder: schema.NewDecoder(),
	}
	s.schemaDecoder.IgnoreUnknownKeys(true)

	if s.config.OAuth.Mount == "" {
		s.config.OAuth.Mount = s.config.OAuth.UI.MountPath()
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

func (s *server) Printf(format string, v ...interface{}) {
	log.FromContext(s.c.Context()).Warnf(format, v...)
}

func (s *server) RegisterRoutes(server *web.Server) {
	csrfMiddleware := webmiddleware.CSRF(
		s.config.OAuth.CSRFAuthKey,
		csrf.CookieName("_csrf"),
		csrf.FieldName("_csrf"),
		csrf.Path("/"),
	)
	router := server.PrefixWithRedirect(s.config.OAuth.Mount).Subrouter()
	router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r, nonce := webui.WithNonce(r)
				cspString := s.generateCSP(&s.configFromContext(r.Context()).OAuth, nonce)
				w.Header().Set("Content-Security-Policy", cspString)
				next.ServeHTTP(w, r)
			})
		},
		ratelimit.HTTPMiddleware(s.c.RateLimiter(), "http:account"),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				config := s.configFromContext(r.Context()).OAuth
				r = webui.WithTemplateData(r, config.UI.TemplateData)
				frontendConfig := config.UI.FrontendConfig
				frontendConfig.Language = config.UI.TemplateData.Language
				r = webui.WithAppConfig(r, struct {
					oauth.FrontendConfig
				}{
					FrontendConfig: frontendConfig,
				})
				next.ServeHTTP(w, r)
			})
		},
		webhandlers.WithErrorHandlers(map[string]http.Handler{
			"text/html": webui.Template,
		}),
		mux.MiddlewareFunc(csrfMiddleware),
	)

	logoutHandler := s.requireLogin(http.HandlerFunc(s.Logout))
	currentUserHandler := s.requireLogin(http.HandlerFunc(s.CurrentUser))
	api := router.NewRoute().PathPrefix("/api").Subrouter()
	api.Path("/auth/login").HandlerFunc(s.Login).Methods(http.MethodPost)
	api.Path("/auth/token-login").HandlerFunc(s.TokenLogin).Methods(http.MethodPost)
	api.Path("/auth/logout").Handler(logoutHandler).Methods(http.MethodPost)
	api.Path("/me").Handler(currentUserHandler).Methods(http.MethodGet)

	loginHandler := s.redirectToNext(webui.Template)
	page := router.NewRoute().Subrouter()
	page.Path("/login").Handler(loginHandler).Methods(http.MethodGet)
	page.Path("/token-login").Handler(loginHandler).Methods(http.MethodGet)
	page.NewRoute().Handler(webui.Template)
}
