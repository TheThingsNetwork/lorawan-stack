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

package oauth

import (
	"net/http"
	"net/url"
	"time"

	"github.com/RangelReale/osin"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/assets"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
)

// Config is the configuration of the OAuth server.
type Config struct {
	// AuthorizationCodeTTL is the duration issued authorization codes are valid.
	AuthorizationCodeTTL time.Duration `name:"authorization-code-ttl" description:"Validity of issued authorization codes."`

	// AccessTokenTTL is the duration issued access tokens are valid.
	AccessTokenTTL time.Duration `name:"access-token-ttl" description:"Validity of issued access tokens."`

	// PublicURL is the public URL of the OAuth server.
	PublicURL string `name:"public-url" description:"Public URL of the OAuth provider."`

	// Assets are the assets to be used.
	Assets *assets.Assets `name:"-"`

	// Store is the store to be used.
	Store *store.Store `name:"-"`

	// Specializers are the specializers to be used.
	Specializers SpecializersConfig `name:"-"`

	// hostname is the hostname used by the OAuth provider.
	// This field is derived from the `PublicURL` config value.
	hostname string `name:"-"`

	// mount the URL path where the OAuth provider will be mounted.
	// This field is derived from the `PublicURL` config value.
	mount string `name:"-"`
}

// SpecializersConfig is the type that contains the specializers used by the
// OAuth server.
type SpecializersConfig struct {
	User   store.UserSpecializer
	Client store.ClientSpecializer
}

// IsValid returns true if and only if all specializers are set.
func (s *SpecializersConfig) IsValid() bool {
	return s.User != nil && s.Client != nil
}

// Server represents an OAuth 2.0 Server.
type Server struct {
	*component.Component
	config Config
	oauth  *osin.Server
}

// New returns a new OAuth server that renders the OAuth provider and relevant routes.
func New(c *component.Component, config Config) (*Server, error) {
	if config.Store == nil {
		return nil, errors.New("No store configured for OAuth")
	}

	if !config.Specializers.IsValid() {
		return nil, errors.New("No specializers configured for OAuth")
	}

	if config.Assets == nil {
		return nil, errors.New("No assets configured for OAuth")
	}

	var err error
	config.mount, err = path(config.PublicURL)
	if err != nil {
		return nil, errors.NewWithCausef(err, "Invalid public URL `%s` passed to OAuth provider", config.PublicURL)
	}

	s := &Server{
		Component: c,
		config:    config,
		oauth: osin.NewServer(
			&osin.ServerConfig{
				AuthorizationExpiration:     int32(config.AuthorizationCodeTTL.Seconds()),
				AccessExpiration:            int32(config.AccessTokenTTL.Seconds()),
				ErrorStatusCode:             http.StatusUnauthorized,
				RequirePKCEForPublicClients: false,
				RedirectUriSeparator:        "",
				RetainTokenAfterRefresh:     false,
				AllowClientSecretInParams:   false,
				TokenType:                   "bearer",
				AllowedAuthorizeTypes: osin.AllowedAuthorizeType{
					osin.CODE,
				},
				AllowedAccessTypes: osin.AllowedAccessType{
					osin.AUTHORIZATION_CODE,
					osin.REFRESH_TOKEN,
					osin.PASSWORD,
				},
			}, &storage{
				Store:             config.Store,
				clientSpecializer: config.Specializers.Client,
			}),
	}

	s.oauth.AuthorizeTokenGen = s
	s.oauth.AccessTokenGen = s
	s.oauth.Now = func() time.Time {
		return time.Now()
	}

	c.RegisterWeb(s)

	return s, nil
}

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

// RegisterRoutes registers the server to the web server.
func (s *Server) RegisterRoutes(server *web.Server) {
	env := map[string]interface{}{
		"oauth": true,
		"mount": s.config.mount,
	}

	// Handler that serve the HTML page.
	index := s.config.Assets.Render("oauth.html", env)

	group := server.Group(s.config.mount)
	group.Use(s.config.Assets.Errors("oauth.html", env))

	group.POST("/oauth/token", s.token)
	group.Any("/oauth/authorize", s.authorize(index), s.RedirectToLogin)

	group.POST("/api/auth/login", s.login)
	group.GET("/api/me", s.me, s.RequireLogin)
	group.POST("/api/auth/logout", s.logout, s.RequireLogin)

	group.GET("/register", index, s.RedirectToAccount)
	group.GET("/login", index, s.RedirectToNext)
	group.GET("/*", index)
}

// output outputs an osin response.
func (s *Server) output(c echo.Context, resp *osin.Response) error {
	if resp.IsError && resp.InternalError != nil {
		log.FromContext(s.Context()).WithError(resp.InternalError).WithFields(log.Fields(
			"status_code", resp.StatusCode,
			"status_text", resp.StatusText,
			"error_status_code", resp.ErrorStatusCode,
			"url", resp.URL,
			"error_id", resp.ErrorId,
			"output", resp.Output,
		)).Error("OAuth provider error when handling a request")

		return errors.NewWithCause(ErrInternal.New(nil), resp.InternalError.Error())
	}

	headers := c.Response().Header()

	for i, k := range resp.Headers {
		for _, v := range k {
			headers.Add(i, v)
		}
	}

	if resp.Type == osin.REDIRECT {
		location, err := resp.GetRedirectUrl()
		if err != nil {
			return err
		}
		headers.Add("Location", location)

		return c.NoContent(http.StatusFound)
	}

	return c.JSON(resp.StatusCode, resp.Output)
}
