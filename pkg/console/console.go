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

	"go.thethings.network/lorawan-stack/pkg/assets"
	"go.thethings.network/lorawan-stack/pkg/component"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/web"
	"golang.org/x/oauth2"
)

var (
	errOAuthNotConfigured = errors.DefineInvalidArgument("oauth_not_configured", "no OAuth client ID and/or secret configured for the Console")
	errExtractPath        = errors.DefineInvalidArgument("extract_path_from_public_url", "could not extract from public URL `{url}`")
)

// Config is the configuration for the Console.
type Config struct {
	// PublicURL is the public URL of the Console.
	PublicURL string `name:"public-url" description:"Public URL of the Console"`

	// DefaultLanguage is the default language of the Console.
	DefaultLanguage string `name:"language" description:"Default language of the Console"`

	// IdentityServerURL is the location of the Identity Server.
	IdentityServerURL string `name:"identity-server-url" description:"URL of the Identity Server"`

	// OAuth is the OAuth config for the Console.
	OAuth OAuth `name:"oauth"`

	// mount is the location where the Console is mounted.
	mount string `name:"-"`
}

// OAuth is the OAuth config for the Console.
type OAuth struct {
	// ID is the client ID for the Console.
	ID string `name:"client-id" description:"The OAuth client ID for the Console"`

	// Secret is the client secret for the Console.
	Secret string `name:"client-secret" description:"The OAuth client secret for the Console" json:"-"`
}

func (o OAuth) isZero() bool {
	return o.ID == "" && o.Secret == ""
}

// Console is the Console component.
type Console struct {
	*component.Component
	assets *assets.Assets
	config Config
	oauth  *oauth2.Config
}

// New returns a new Console.
func New(c *component.Component, assets *assets.Assets, config Config) (*Console, error) {
	if config.OAuth.isZero() {
		return nil, errOAuthNotConfigured
	}

	console := &Console{
		Component: c,
		assets:    assets,
		config:    config,
	}

	mount, err := path(console.config.PublicURL)
	if err != nil {
		return nil, errExtractPath.WithAttributes("url", console.config.PublicURL).WithCause(err)
	}
	console.config.mount = mount

	console.oauth = &oauth2.Config{
		ClientID:     console.config.OAuth.ID,
		ClientSecret: console.config.OAuth.Secret,
		RedirectURL:  fmt.Sprintf("%s/oauth/callback", strings.TrimSuffix(console.config.PublicURL, "/")),
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("%s/oauth/token", strings.TrimSuffix(console.config.IdentityServerURL, "/")),
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", strings.TrimSuffix(console.config.IdentityServerURL, "/")),
		},
	}

	c.RegisterWeb(console)

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
	env := map[string]interface{}{
		"console":          true,
		"mount":            console.config.mount,
		"default_language": console.config.DefaultLanguage,
	}

	group := server.Group(console.config.mount)
	group.Use(console.assets.Errors("error.html", env))

	group.GET("/oauth/callback", console.Callback)

	api := group.Group("/api")
	api.GET("/auth/token", console.Token)
	api.PUT("/auth/refresh", console.RefreshToken)
	api.GET("/auth/login", console.Login)
	api.POST("/auth/logout", console.Logout)

	// Set up HTML routes.
	index := console.assets.AppHandler("console.html", env)
	group.GET("/", index)
	group.GET("/*", index)
}
