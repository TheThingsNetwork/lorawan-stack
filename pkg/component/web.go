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

package component

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/heptiolabs/healthcheck"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/web"
)

const (
	metricsUsername = "metrics"
	pprofUsername   = "pprof"
	healthUsername  = "health"
)

func (c *Component) initWeb() error {
	webOptions := []web.Option{
		web.WithContextFiller(c.FillContext),
		web.WithCookieKeys(c.config.HTTP.Cookie.HashKey, c.config.HTTP.Cookie.BlockKey),
		web.WithStatic(c.config.HTTP.Static.Mount, c.config.HTTP.Static.SearchPath...),
	}
	if c.config.HTTP.RedirectToHost != "" {
		webOptions = append(webOptions, web.WithRedirectToHost(c.config.HTTP.RedirectToHost))
	}
	if c.config.HTTP.RedirectToHTTPS {
		httpAddr, err := net.ResolveTCPAddr("tcp", c.config.HTTP.Listen)
		if err != nil {
			return err
		}
		httpsAddr, err := net.ResolveTCPAddr("tcp", c.config.HTTP.ListenTLS)
		if err != nil {
			return err
		}
		if httpsAddr.Port == 0 {
			httpsAddr.Port = 443
		}
		webOptions = append(webOptions, web.WithRedirectToHTTPS(httpAddr.Port, httpsAddr.Port))
		if httpAddr.Port != 80 && httpsAddr.Port != 443 {
			webOptions = append(webOptions, web.WithRedirectToHTTPS(80, 443))
		}
	}
	web, err := web.New(c.ctx, webOptions...)
	if err != nil {
		return err
	}
	c.web = web
	return nil
}

// RegisterWeb registers a web subsystem to the component.
func (c *Component) RegisterWeb(s web.Registerer) {
	c.webSubsystems = append(c.webSubsystems, s)
}

// RegisterLivenessCheck registers a liveness check for the component.
func (c *Component) RegisterLivenessCheck(name string, check healthcheck.Check) {
	c.healthHandler.AddLivenessCheck(name, check)
}

// RegisterReadinessCheck registers a readiness check for the component.
func (c *Component) RegisterReadinessCheck(name string, check healthcheck.Check) {
	c.healthHandler.AddReadinessCheck(name, check)
}

func (c *Component) serveWeb(lis net.Listener) error {
	return http.Serve(lis, c)
}

func (c *Component) webEndpoints() []Endpoint {
	return []Endpoint{
		NewTCPEndpoint(c.config.HTTP.Listen, "Web"),
		NewTLSEndpoint(c.config.HTTP.ListenTLS, "Web"),
	}
}

// listenWeb starts the web listeners on the addresses and endpoints configured in the HTTP section.
func (c *Component) listenWeb() (err error) {
	err = c.serveOnEndpoints(c.webEndpoints(), (*Component).serveWeb, "web")
	if err != nil {
		return
	}

	if c.config.HTTP.PProf.Enable {
		var middleware []echo.MiddlewareFunc
		if c.config.HTTP.PProf.Password != "" {
			middleware = append(middleware, c.basicAuth(pprofUsername, c.config.HTTP.PProf.Password))
		}
		g := c.web.RootGroup("/debug/pprof", middleware...)
		g.GET("", func(c echo.Context) error { return c.Redirect(http.StatusFound, c.Path()+"/") })
		g.GET("/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
		g.GET("/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
		g.GET("/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	}

	if c.config.HTTP.Metrics.Enable {
		var middleware []echo.MiddlewareFunc
		if c.config.HTTP.Metrics.Password != "" {
			middleware = append(middleware, c.basicAuth(metricsUsername, c.config.HTTP.Metrics.Password))
		}
		g := c.web.RootGroup("/metrics", middleware...)
		g.GET("/", func(c echo.Context) error { return c.Redirect(http.StatusFound, strings.TrimSuffix(c.Path(), "/")) })
		g.GET("", echo.WrapHandler(metrics.Exporter), func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Request().Header.Del("Accept-Encoding")
				return next(c)
			}
		})
	}

	if c.config.HTTP.Health.Enable {
		var middleware []echo.MiddlewareFunc
		if c.config.HTTP.Health.Password != "" {
			middleware = append(middleware, c.basicAuth(healthUsername, c.config.HTTP.Health.Password))
		}
		g := c.web.RootGroup("/healthz", middleware...)
		g.GET("/live", echo.WrapHandler(http.HandlerFunc(c.healthHandler.LiveEndpoint)))
		g.GET("/ready", echo.WrapHandler(http.HandlerFunc(c.healthHandler.ReadyEndpoint)))
	}

	return nil
}

func (c *Component) basicAuth(username, password string) echo.MiddlewareFunc {
	usernameBytes, passwordBytes := []byte(username), []byte(password)
	return middleware.BasicAuth(func(username string, password string, ctx echo.Context) (bool, error) {
		usernameCompare := subtle.ConstantTimeCompare([]byte(username), usernameBytes)
		passwordCompare := subtle.ConstantTimeCompare([]byte(password), passwordBytes)
		if usernameCompare != 1 || passwordCompare != 1 {
			c.Logger().WithFields(log.Fields(
				"namespace", "web",
				"url", ctx.Path(),
				"remote_addr", ctx.RealIP(),
			)).Warn("Basic auth failed")
			return false, nil
		}
		return true, nil
	})
}
