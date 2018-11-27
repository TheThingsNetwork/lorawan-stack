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
	"context"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// Server is the interface for the OAuth server.
type Server interface {
	web.Registerer

	Login(c echo.Context) error
	CurrentUser(c echo.Context) error
	Logout(c echo.Context) error
}

type server struct {
	ctx    context.Context
	config Config
	store  Store
}

// Store used by the OAuth server.
type Store interface {
	// UserStore and UserSessionStore are needed for user login/logout.
	store.UserStore
	store.UserSessionStore
}

// UIConfig is the combined configuration for the OAuth UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// FrontendConfig is the configuration for the OAuth frontend.
type FrontendConfig struct {
	Language string `json:"language" name:"-"`
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
	return s
}

func (s *server) RegisterRoutes(server *web.Server) {
	group := server.Group(s.config.Mount, webui.RenderErrors, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("template_data", s.config.UI.TemplateData)
			frontendConfig := s.config.UI.FrontendConfig
			frontendConfig.Language = s.config.UI.TemplateData.Language
			c.Set("app_config", struct {
				OAuth bool `json:"oauth"`
				FrontendConfig
			}{
				OAuth:          true,
				FrontendConfig: frontendConfig,
			})
			return next(c)
		}
	})

	csrf := middleware.CSRF()

	group.GET("/login", webui.Render, csrf, s.redirectToNext)

	group.POST("/api/auth/login", s.Login, csrf)
	group.POST("/api/auth/logout", s.Logout, csrf, s.requireLogin)
	group.GET("/api/me", s.CurrentUser, csrf, s.requireLogin)

	if s.config.Mount != "" && s.config.Mount != "/" {
		group.GET("", webui.Render, csrf)
	}
	group.GET("/*", webui.Render, csrf)
}
