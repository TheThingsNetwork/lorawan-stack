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

package interop

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	echomiddleware "github.com/labstack/echo/middleware"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/middleware"
)

// Registerer allows components to register their interop services to the web server.
type Registerer interface {
	RegisterInterop(s *Server)
}

// Server is the server.
type Server struct {
	*rootGroup
	config config.Interop
	server *echo.Echo
}

type rootGroup struct {
	*echo.Group
}

// New builds a new server.
func New(ctx context.Context, config config.Interop) (*Server, error) {
	logger := log.FromContext(ctx).WithField("namespace", "interop")

	server := echo.New()

	server.Logger = web.NewNoopLogger()
	server.HTTPErrorHandler = ErrorHandler

	server.Use(
		middleware.ID("interop"),
		echomiddleware.BodyLimit("16M"),
		echomiddleware.Secure(),
		middleware.Recover(),
	)

	return &Server{
		rootGroup: &rootGroup{
			Group: server.Group(
				"",
				middleware.Log(logger),
				middleware.Normalize(middleware.RedirectPermanent),
			),
		},
		config: config,
		server: server,
	}, nil
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

// Group creates a sub group.
func (s *Server) Group(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group {
	t := strings.TrimSuffix(prefix, "/")
	return s.rootGroup.Group.Group(t, middleware...)
}

// RootGroup creates a new Echo router group with prefix and optional group-level middleware on the root Server.
func (s *Server) RootGroup(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group {
	t := strings.TrimSuffix(prefix, "/")
	return s.server.Group(t, middleware...)
}
