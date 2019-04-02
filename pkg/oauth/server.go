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

package oauth

import (
	"context"
	"net/http"
	"strings"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openshift/osin"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// Server is the interface for the OAuth server.
type Server interface {
	web.Registerer

	Login(c echo.Context) error
	CurrentUser(c echo.Context) error
	Logout(c echo.Context) error
	Authorize(authorizePage echo.HandlerFunc) echo.HandlerFunc
	Token(c echo.Context) error
}

type server struct {
	ctx        context.Context
	config     Config
	osinConfig *osin.ServerConfig
	store      Store
}

// Store used by the OAuth server.
type Store interface {
	// UserStore and UserSessionStore are needed for user login/logout.
	store.UserStore
	store.UserSessionStore
	// ClientStore is needed for getting the OAuth client.
	store.ClientStore
	// OAuth is needed for OAuth authorizations.
	store.OAuthStore
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

	s.osinConfig = &osin.ServerConfig{
		AuthorizationExpiration: int32((5 * time.Minute).Seconds()),
		AccessExpiration:        int32(time.Hour.Seconds()),
		TokenType:               "bearer",
		AllowedAuthorizeTypes: osin.AllowedAuthorizeType{
			osin.CODE,
		},
		AllowedAccessTypes: osin.AllowedAccessType{
			osin.AUTHORIZATION_CODE,
			osin.REFRESH_TOKEN,
			osin.PASSWORD,
		},
		ErrorStatusCode:           http.StatusBadRequest,
		AllowClientSecretInParams: true,
		RedirectUriSeparator:      redirectURISeparator,
		RetainTokenAfterRefresh:   false,
	}

	return s
}

func (s *server) now() time.Time { return time.Now().UTC() }

func (s *server) oauth2(ctx context.Context) *osin.Server {
	oauth2 := osin.NewServer(s.osinConfig, &storage{
		ctx:     ctx,
		clients: s.store,
		oauth:   s.store,
	})
	oauth2.AuthorizeTokenGen = s
	oauth2.AccessTokenGen = s
	oauth2.Now = s.now
	oauth2.Logger = s
	return oauth2
}

func (s *server) Printf(format string, v ...interface{}) {
	log.FromContext(s.ctx).Warnf(format, v...)
}

func (s *server) output(c echo.Context, resp *osin.Response) error {
	if resp.IsError && resp.InternalError != nil {
		return resp.InternalError
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
		uiMount := strings.TrimSuffix(s.config.UI.MountPath(), "/") + "/"
		if strings.HasPrefix(location, "/") && !strings.HasPrefix(location, uiMount) {
			location = uiMount + location
		}
		return c.Redirect(http.StatusFound, location)
	}
	return c.JSON(resp.StatusCode, resp.Output)
}

func (s *server) RegisterRoutes(server *web.Server) {
	group := server.Group(
		s.config.Mount,
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set("template_data", s.config.UI.TemplateData)
				frontendConfig := s.config.UI.FrontendConfig
				frontendConfig.Language = s.config.UI.TemplateData.Language
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
	api.POST("/auth/login", s.Login)
	api.POST("/auth/logout", s.Logout, s.requireLogin)
	api.GET("/me", s.CurrentUser, s.requireLogin)

	page := group.Group("", middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
	page.GET("/login", webui.Template.Handler, s.redirectToNext)
	page.GET("/authorize", s.Authorize(webui.Template.Handler), s.redirectToLogin)
	page.POST("/authorize", s.Authorize(webui.Template.Handler), s.redirectToLogin)

	if s.config.Mount != "" && s.config.Mount != "/" {
		group.GET("", webui.Template.Handler, middleware.CSRF())
	}
	group.GET("/*", webui.Template.Handler, middleware.CSRF())

	group.POST("/token", s.Token) // No CSRF here.
}
