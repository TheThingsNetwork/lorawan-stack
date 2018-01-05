// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"net/http"
	"net/http/pprof"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	"github.com/labstack/echo"
	"github.com/soheilhy/cmux"
)

// RegisterWeb registers a web subsystem to the component
func (c *Component) RegisterWeb(s web.Registerer) {
	c.webSubsystems = append(c.webSubsystems, s)
}

func (c *Component) listenWeb() (err error) {
	serve := func(mux cmux.CMux) {
		h2 := mux.Match(cmux.HTTP2())
		go func() {
			if err := http.Serve(h2, c.web); err != nil {
				c.logger.WithError(err).Errorf("Error serving HTTP on %s", h2.Addr())
			}
		}()
		h1 := mux.Match(cmux.HTTP1Fast())
		go func() {
			if err := http.Serve(h1, c.web); err != nil {
				c.logger.WithError(err).Errorf("Error serving HTTP on %s", h1.Addr())
			}
		}()
	}

	if c.config.HTTP.Listen != "" {
		l, err := c.Listen(c.config.HTTP.Listen)
		if err != nil {
			return errors.NewWithCause("Could not listen on HTTP port", err)
		}
		mux, err := l.TCP()
		if err != nil {
			return errors.NewWithCause("Could not create TCP mux on top of HTTP listener", err)
		}
		serve(mux)
	}

	if c.config.HTTP.ListenTLS != "" {
		l, err := c.Listen(c.config.HTTP.ListenTLS)
		if err != nil {
			return errors.NewWithCause("Could not listen on HTTP/tls port", err)
		}
		mux, err := l.TLS()
		if err != nil {
			return errors.NewWithCause("Could not create TLS mux on top of HTTP/tls listener", err)
		}
		serve(mux)
	}

	if c.config.HTTP.PProf {
		var pprofMiddleware []echo.MiddlewareFunc

		// TODO: Add auth to pprof endpoints

		c.web.GET("/debug/pprof/", echo.WrapHandler(http.HandlerFunc(pprof.Index)), pprofMiddleware...)
		c.web.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)), pprofMiddleware...)
	}

	return nil
}
