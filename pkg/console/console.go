// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package console

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/web/oauthclient"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// Console is the Console component.
type Console struct {
	*component.Component
	oc     *oauthclient.OAuthClient
	config Config
}

// New returns a new Console.
func New(c *component.Component, config Config) (*Console, error) {
	config.OAuth.StateCookieName = "_console_state"
	config.OAuth.AuthCookieName = "_console_auth"
	config.OAuth.RootURL = config.UI.CanonicalURL
	oc, err := oauthclient.New(c, config.OAuth)
	if err != nil {
		return nil, err
	}

	console := &Console{
		Component: c,
		oc:        oc,
		config:    config,
	}

	if console.config.Mount == "" {
		console.config.Mount = console.config.UI.MountPath()
	}

	c.RegisterWeb(console)
	c.RegisterWeb(events.New(c))

	return console, nil
}

type ctxKeyType struct{}

var ctxKey ctxKeyType

func (console *Console) configFromContext(ctx context.Context) *Config {
	if config, ok := ctx.Value(ctxKey).(*Config); ok {
		return config
	}
	return &console.config
}

// path extracts the mounted location from the public Console URL.
func path(u string) (string, error) {
	p, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	if p.Path == "" {
		return "/", nil
	}

	return p.Path, nil
}

func generateConsoleCSPString(config *Config, nonce string, others ...webui.ContentSecurityPolicy) string {
	baseURLs := webui.RewriteSchemes(
		webui.WebsocketSchemeRewrites,
		config.UI.StackConfig.GS.BaseURL,
		config.UI.StackConfig.IS.BaseURL,
		config.UI.StackConfig.JS.BaseURL,
		config.UI.StackConfig.NS.BaseURL,
		config.UI.StackConfig.AS.BaseURL,
		config.UI.StackConfig.EDTC.BaseURL,
		config.UI.StackConfig.QRG.BaseURL,
		config.UI.StackConfig.GCS.BaseURL,
		config.UI.StackConfig.DCS.BaseURL,
	)
	return webui.ContentSecurityPolicy{
		ConnectionSource: append([]string{
			"'self'",
			config.UI.SentryDSN,
			"gravatar.com",
			"www.gravatar.com",
		}, baseURLs...),
		StyleSource: []string{
			"'self'",
			config.UI.AssetsBaseURL,
			config.UI.BrandingBaseURL,
			"'unsafe-inline'",
		},
		ScriptSource: []string{
			"'self'",
			config.UI.AssetsBaseURL,
			config.UI.BrandingBaseURL,
			"'unsafe-eval'",
			"'strict-dynamic'",
			fmt.Sprintf("'nonce-%s'", nonce),
		},
		BaseURI: []string{
			"'self'",
		},
		FrameAncestors: []string{
			"'none'",
		},
	}.Merge(others...).Clean().String()
}

// RegisterRoutes implements web.Registerer. It registers the Console to the web server.
func (console *Console) RegisterRoutes(server *web.Server) {
	router := server.PrefixWithRedirect(console.config.Mount).Subrouter()
	router.Use(
		mux.MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r, nonce := webui.WithNonce(r)
				cspString := generateConsoleCSPString(console.configFromContext(r.Context()), nonce)
				w.Header().Set("Content-Security-Policy", cspString)
				next.ServeHTTP(w, r)
			})
		}),
		mux.MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				config := console.configFromContext(r.Context())
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
		}),
		webhandlers.WithErrorHandlers(map[string]http.Handler{
			"text/html": webui.Template,
		}),
		mux.MiddlewareFunc(
			webmiddleware.CSRF(
				console.GetBaseConfig(console.Context()).HTTP.Cookie.HashKey,
				csrf.CookieName("_console_csrf"),
				csrf.FieldName("_console_csrf"),
				csrf.Path(console.config.Mount),
			),
		),
	)
	api := router.NewRoute().PathPrefix("/api/auth/").Subrouter()
	api.Path("/token").HandlerFunc(console.oc.HandleToken).Methods(http.MethodGet)
	api.Path("/logout").HandlerFunc(console.oc.HandleLogout).Methods(http.MethodPost)

	router.Path("/login/ttn-stack").HandlerFunc(console.oc.HandleLogin).Methods(http.MethodGet)
	router.Path("/oauth/callback").HandlerFunc(console.oc.HandleCallback).Methods(http.MethodGet)
	router.NewRoute().Handler(webui.Template)
}
