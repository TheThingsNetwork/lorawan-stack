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

package console

import (
	"context"
	"net/url"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	web_errors "go.thethings.network/lorawan-stack/v3/pkg/errors/web"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/web/middleware"
	"go.thethings.network/lorawan-stack/v3/pkg/web/oauthclient"
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

// RegisterRoutes implements web.Registerer. It registers the Console to the web server.
func (console *Console) RegisterRoutes(server *web.Server) {
	group := server.Group(
		console.config.Mount,
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				config := console.configFromContext(c.Request().Context())
				c.Set("template_data", config.UI.TemplateData)
				frontendConfig := config.UI.FrontendConfig
				frontendConfig.Language = config.UI.TemplateData.Language
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
		middleware.CSRF("_console_csrf", console.config.Mount, console.GetBaseConfig(console.Context()).HTTP.Cookie.HashKey),
	)

	api := group.Group("/api/auth")
	api.GET("/token", console.oc.HandleToken)
	api.POST("/logout", console.oc.HandleLogout)

	group.GET("/oauth/callback", console.oc.HandleCallback)

	group.GET("/login/ttn-stack", console.oc.HandleLogin)

	group.GET("/*", webui.Template.Handler)
}
