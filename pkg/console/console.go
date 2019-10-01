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
	"github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/component"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/oauthclient"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// UIConfig is the combined configuration for the Console UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// StackConfig is the configuration of the stack components.
type StackConfig struct {
	IS webui.APIConfig `json:"is" name:"is"`
	GS webui.APIConfig `json:"gs" name:"gs"`
	NS webui.APIConfig `json:"ns" name:"ns"`
	AS webui.APIConfig `json:"as" name:"as"`
	JS webui.APIConfig `json:"js" name:"js"`
}

// FrontendConfig is the configuration for the Console frontend.
type FrontendConfig struct {
	Language    string `json:"language" name:"-"`
	SupportLink string `json:"support_link" name:"support-link" description:"The URI that the support button will point to"`
	StackConfig `json:"stack_config" name:",squash"`
}

// Config is the configuration for the Console.
type Config struct {
	OAuth oauthclient.Config `name:"oauth"`
	Mount string             `name:"mount" description:"Path on the server where the Console will be served"`
	UI    UIConfig           `name:"ui"`
}

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
	api.GET("/auth/token", console.oc.HandleToken)
	api.POST("/auth/logout", console.oc.HandleLogout)

	page := group.Group("", middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
	page.GET("/oauth/callback", console.oc.HandleCallback)

	group.GET("/login/ttn-stack", console.oc.HandleLogin)

	if console.config.Mount != "" && console.config.Mount != "/" {
		group.GET("", webui.Template.Handler, middleware.CSRF())
	}
	group.GET("/*", webui.Template.Handler, middleware.CSRF())
}
