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
	"fmt"
	"net/url"
	"strings"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/webui"
	"golang.org/x/oauth2"
)

// UIConfig is the combined configuration for the Console UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// FrontendConfig is the configuration for the Console frontend.
type FrontendConfig struct {
	Language string          `json:"language" name:"-"`
	IS       webui.APIConfig `json:"is" name:"is"`
	GS       webui.APIConfig `json:"gs" name:"gs"`
	NS       webui.APIConfig `json:"ns" name:"ns"`
	AS       webui.APIConfig `json:"as" name:"as"`
	JS       webui.APIConfig `json:"js" name:"js"`
}

// Config is the configuration for the Console.
type Config struct {
	OAuth OAuth    `name:"oauth"`
	Mount string   `name:"mount" description:"Path on the server where the Console will be served"`
	UI    UIConfig `name:"ui"`
}

// OAuth is the OAuth config for the Console.
type OAuth struct {
	AuthorizeURL string `name:"authorize-url" description:"The OAuth Authorize URL"`
	TokenURL     string `name:"token-url" description:"The OAuth Token Exchange URL"`

	ClientID     string `name:"client-id" description:"The OAuth client ID for the Console"`
	ClientSecret string `name:"client-secret" description:"The OAuth client secret for the Console" json:"-"`
}

var errNoOAuthConfig = errors.DefineInvalidArgument("no_oauth_config", "no OAuth configuration found for the Console")

func (o OAuth) isZero() bool {
	return o.AuthorizeURL == "" || o.TokenURL == "" || o.ClientID == "" || o.ClientSecret == ""
}

// Console is the Console component.
type Console struct {
	*component.Component
	config Config
}

// New returns a new Console.
func New(c *component.Component, config Config) (*Console, error) {
	if config.OAuth.isZero() {
		return nil, errNoOAuthConfig
	}

	console := &Console{
		Component: c,
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

func (console *Console) oauth(c echo.Context) *oauth2.Config {
	config := console.configFromContext(c.Request().Context())
	return &oauth2.Config{
		ClientID:     config.OAuth.ClientID,
		ClientSecret: config.OAuth.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/oauth/callback", strings.TrimSuffix(config.UI.CanonicalURL, "/")),
		Endpoint: oauth2.Endpoint{
			TokenURL:  config.OAuth.TokenURL,
			AuthURL:   config.OAuth.AuthorizeURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
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
				c.Set("template_data", console.config.UI.TemplateData)
				frontendConfig := console.config.UI.FrontendConfig
				frontendConfig.Language = console.config.UI.TemplateData.Language
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
	api.GET("/auth/token", console.Token)
	api.POST("/auth/logout", console.Logout)

	page := group.Group("", middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
	page.GET("/oauth/callback", console.Callback)

	group.GET("/login/ttn-stack", console.Login)

	if console.config.Mount != "" && console.config.Mount != "/" {
		group.GET("", webui.Template.Handler, middleware.CSRF())
	}
	group.GET("/*", webui.Template.Handler, middleware.CSRF())
}
