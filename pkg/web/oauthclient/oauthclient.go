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

package oauthclient

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"golang.org/x/oauth2"
)

// Config is the configuration for the OAuth client.
type Config struct {
	AuthorizeURL string `name:"authorize-url" description:"The OAuth Authorize URL"`
	TokenURL     string `name:"token-url" description:"The OAuth Token Exchange URL"`
	RootURL      string `name:"-"`

	ClientID     string `name:"client-id" description:"The OAuth client ID"`
	ClientSecret string `name:"client-secret" description:"The OAuth client secret" json:"-"`

	StateCookieName string `name:"-"`
	AuthCookieName  string `name:"-"`
}

// OAuthClient is the OAuth client component.
type OAuthClient struct {
	component *component.Component
	rootURL   string
	config    Config
}

var errNoOAuthConfig = errors.DefineInvalidArgument("no_oauth_config", "no OAuth configuration found for the OAuth client")

func (c Config) isZero() bool {
	return c.AuthorizeURL == "" || c.TokenURL == "" || c.ClientID == "" || c.ClientSecret == ""
}

func (oc *OAuthClient) getMountPath() string {
	var path string
	u, err := url.Parse(oc.config.RootURL)
	if err != nil || u.Path == "" {
		path = "/"
	} else {
		path = u.Path
	}

	return path
}

// New returns a new OAuth client instance.
func New(c *component.Component, config Config) (*OAuthClient, error) {
	if config.isZero() {
		return nil, errNoOAuthConfig
	}

	oc := &OAuthClient{
		component: c,
		config:    config,
	}

	return oc, nil
}

type ctxKeyType struct{}

var ctxKey ctxKeyType

func (oc *OAuthClient) configFromContext(ctx context.Context) *Config {
	if config, ok := ctx.Value(ctxKey).(*Config); ok {
		return config
	}
	return &oc.config
}

func (oc *OAuthClient) oauth(c echo.Context) *oauth2.Config {
	config := oc.configFromContext(c.Request().Context())

	authorizeURL := config.AuthorizeURL
	redirectURL := fmt.Sprintf("%s/oauth/callback", strings.TrimSuffix(oc.config.RootURL, "/"))
	if oauthRootURL, err := url.Parse(oc.config.RootURL); err == nil {
		rootURL := (&url.URL{Scheme: oauthRootURL.Scheme, Host: oauthRootURL.Host}).String()
		if strings.HasPrefix(authorizeURL, rootURL) {
			authorizeURL = strings.TrimPrefix(authorizeURL, rootURL)
			redirectURL = strings.TrimPrefix(redirectURL, rootURL)
		}
	}

	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			TokenURL:  config.TokenURL,
			AuthURL:   authorizeURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}
