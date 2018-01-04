// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"net/http"
	"net/http/pprof"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	"github.com/labstack/echo"
)

// RegisterWeb registers a web subsystem to the component
func (c *Component) RegisterWeb(s web.Registerer) {
	c.webSubsystems = append(c.webSubsystems, s)
}

func (c *Component) listenWeb() (err error) {
	if c.config.HTTP.Listen != "" {
		l, err := c.Listen(c.config.HTTP.Listen)
		if err != nil {
			return errors.NewWithCause("Could not listen on HTTP port", err)
		}
		lis, err := l.TCP()
		if err != nil {
			return errors.NewWithCause("Could not create TCP HTTP listener", err)
		}
		go func() {
			if err := http.Serve(lis, c.web); err != nil {
				c.logger.WithError(err).Errorf("Error serving HTTP on %s", lis.Addr())
			}
		}()
	}

	if c.config.HTTP.ListenTLS != "" {
		l, err := c.Listen(c.config.HTTP.ListenTLS)
		if err != nil {
			return errors.NewWithCause("Could not listen on HTTP/tls port", err)
		}
		lis, err := l.TLS()
		if err != nil {
			return errors.NewWithCause("Could not create TLS HTTP listener", err)
		}
		go func() {
			if err := http.Serve(lis, c.web); err != nil {
				c.logger.WithError(err).Errorf("Error serving HTTP on %s", lis.Addr())
			}
		}()
	}

	if c.config.HTTP.PProf {
		var pprofMiddleware []echo.MiddlewareFunc

		// TODO: Add auth to pprof endpoints

		c.web.GET("/debug/pprof/", echo.WrapHandler(http.HandlerFunc(pprof.Index)), pprofMiddleware...)
		c.web.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)), pprofMiddleware...)
	}

	return nil
}
