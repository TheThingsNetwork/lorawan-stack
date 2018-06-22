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

package component

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/web"
)

// RegisterWeb registers a web subsystem to the component
func (c *Component) RegisterWeb(s web.Registerer) {
	c.webSubsystems = append(c.webSubsystems, s)
}

func (c *Component) listenWeb() (err error) {
	if c.config.HTTP.Listen != "" {
		l, err := c.ListenTCP(c.config.HTTP.Listen)
		if err != nil {
			return errors.NewWithCause(err, "Could not listen on HTTP port")
		}
		lis, err := l.TCP()
		if err != nil {
			return errors.NewWithCause(err, "Could not create TCP HTTP listener")
		}
		c.logger.WithFields(log.Fields("namespace", "web", "address", c.config.HTTP.Listen)).Info("Listening for HTTP connections")
		go func() {
			if err := http.Serve(lis, c); err != nil {
				c.logger.WithError(err).Errorf("Error serving HTTP on %s", lis.Addr())
			}
		}()
	}

	if c.config.HTTP.ListenTLS != "" {
		l, err := c.ListenTCP(c.config.HTTP.ListenTLS)
		if err != nil {
			return errors.NewWithCause(err, "Could not listen on HTTP/tls port")
		}
		lis, err := l.TLS()
		if err != nil {
			return errors.NewWithCause(err, "Could not create TLS HTTP listener")
		}
		c.logger.WithFields(log.Fields("namespace", "web", "address", c.config.HTTP.ListenTLS)).Info("Listening for HTTPS connections")
		go func() {
			if err := http.Serve(lis, c); err != nil {
				c.logger.WithError(err).Errorf("Error serving HTTP on %s", lis.Addr())
			}
		}()
	}

	if c.config.HTTP.PProf {
		g := c.web.RootGroup("/debug/pprof") // TODO(#565): Add auth to pprof endpoints.
		g.GET("", func(c echo.Context) error { return c.Redirect(http.StatusFound, c.Path()+"/") })
		g.GET("/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
		g.GET("/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
		g.GET("/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	}

	if c.config.HTTP.Metrics {
		g := c.web.RootGroup("/metrics") // TODO(#565): Add auth to metrics endpoints.
		g.GET("/", func(c echo.Context) error { return c.Redirect(http.StatusFound, strings.TrimSuffix(c.Path(), "/")) })
		g.GET("", echo.WrapHandler(metrics.Exporter))
	}

	return nil
}

func (c *Component) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
		c.grpc.Server.ServeHTTP(w, r)
	} else {
		c.web.ServeHTTP(w, r)
	}
}
