// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"net/url"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	events_grpc "go.thethings.network/lorawan-stack/pkg/events/grpc"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/webui"
	"golang.org/x/oauth2"
)

// APIConfig for upstream APIs.
type APIConfig struct {
	Enabled bool   `json:"enabled" name:"enabled" description:"Enable this API"`
	BaseURL string `json:"base_url" name:"base-url" description:"Base URL to the HTTP API"`
}

// UIConfig is the combined configuration for the Console UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// FrontendConfig is the configuration for the Console frontend.
type FrontendConfig struct {
	Language string    `json:"language" name:"-"`
	IS       APIConfig `json:"is" name:"is"`
	GS       APIConfig `json:"gs" name:"gs"`
	NS       APIConfig `json:"ns" name:"ns"`
	AS       APIConfig `json:"as" name:"as"`
	JS       APIConfig `json:"js" name:"js"`
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
	oauth  *oauth2.Config
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

	console.oauth = &oauth2.Config{
		ClientID:     console.config.OAuth.ClientID,
		ClientSecret: console.config.OAuth.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/oauth/callback", strings.TrimSuffix(console.config.UI.CanonicalURL, "/")),
		Endpoint: oauth2.Endpoint{
			TokenURL: console.config.OAuth.TokenURL,
			AuthURL:  console.config.OAuth.AuthorizeURL,
		},
	}

	c.RegisterWeb(console)

	c.RegisterGRPC(events_grpc.NewEventsServer(c.Context(), events.DefaultPubSub))

	return console, nil
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
	group := server.Group(console.config.Mount, webui.RenderErrors, func(next echo.HandlerFunc) echo.HandlerFunc {
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
	})

	if console.config.Mount != "" && console.config.Mount != "/" {
		group.GET("", webui.Render)
	}
	group.GET("/*", webui.Render)

	api := group.Group("/api", middleware.CSRF())
	api.GET("/auth/token", console.Token)
	api.PUT("/auth/refresh", console.RefreshToken)
	api.GET("/auth/login", console.Login)
	api.POST("/auth/logout", console.Logout)

	page := group.Group("", middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
	page.GET("/oauth/callback", console.Callback)
}
