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

package web

import (
	"net/http"
	"path"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/random"
	"github.com/TheThingsNetwork/ttn/pkg/web/cookie"
	"github.com/TheThingsNetwork/ttn/pkg/web/middleware"
	"github.com/labstack/echo"
)

// Registerer allows components to register their services to the web server
type Registerer interface {
	RegisterRoutes(s *Server)
}

type config struct {
	// Root is the root where the router will live.
	Root string

	// Prefix is the prefix for the request id's.
	Prefix string

	// NormalizationMode is the mode to use for request path normalization.
	NormalizationMode middleware.NormalizationMode

	// BlockKey is used to encrypt the cookie value.
	BlockKey []byte

	// HashKey is used to authenticate the cookie value using HMAC.
	HashKey []byte

	// Renderer is the renderer that will be used.
	Renderer echo.Renderer

	// ErrorTemplate is the name of the template to use for html errors.
	ErrorTemplate string
}

// Server is the server.
type Server struct {
	*echo.Group
	config *config
	server *echo.Echo
}

// RootGroup creates a new Echo router group with prefix and optional group-level middleware on the root Server.
func (s *Server) RootGroup(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group {
	return s.server.Group(prefix, middleware...)
}

// Option is an option for Server.
type Option func(*config)

// New builds a new server.
func New(logger log.Interface, opts ...Option) *Server {
	cfg := &config{
		Root:              "/",
		Prefix:            "",
		NormalizationMode: middleware.RedirectPermanent,
		ErrorTemplate:     "index.html",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.HashKey == nil {
		cfg.HashKey = random.Bytes(32)
		logger.WithField("hash_key", cfg.HashKey).Warn("Generated a random cookie hash key")
	}

	if cfg.BlockKey == nil {
		cfg.BlockKey = random.Bytes(32)
		logger.WithField("block_key", cfg.BlockKey).Warn("Generated a random cookie block key")
	}

	server := echo.New()

	server.Logger = &noopLogger{}
	server.HTTPErrorHandler = ErrorHandler(cfg.ErrorTemplate)
	server.Renderer = cfg.Renderer

	server.Use(
		middleware.Log(logger.WithField("namespace", "web")),
		middleware.ID(cfg.Prefix),
		cookie.Cookies(cfg.Root, cfg.BlockKey, cfg.HashKey),
	)

	group := server.Group(strings.TrimSuffix(cfg.Root, "/"), middleware.Normalize(cfg.NormalizationMode))

	return &Server{
		Group:  group,
		config: cfg,
		server: server,
	}
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

// onNormalize sets the normalization mode for the server.
func onNormalize(mode middleware.NormalizationMode) Option {
	return func(c *config) {
		c.NormalizationMode = mode
	}
}

var (
	// OnNormalizeIgnore does not normalize urls.
	OnNormalizeIgnore = onNormalize(middleware.Ignore)

	// OnNormalizeContinue normalizes urls but does not redirect clients.
	OnNormalizeContinue = onNormalize(middleware.Continue)

	// OnNormalizeRedirectTemporary redirects clients temporarily if they use denormalized urls.
	OnNormalizeRedirectTemporary = onNormalize(middleware.RedirectTemporary)

	// OnNormalizeRedirectPermanent redirects clients permanently if they use denormalized urls.
	OnNormalizeRedirectPermanent = onNormalize(middleware.RedirectPermanent)
)

// WithPrefix sets the prefix for request ID's.
func WithPrefix(prefix string) Option {
	return func(c *config) {
		c.Prefix = prefix
	}
}

// WithRoot sets the root path of the router.
func WithRoot(root string) Option {
	return func(c *config) {
		c.Root = root
	}
}

// WithCookieSecrets sets the secrets to be used for cookie encryption and validation.
// If not set, the server will use random values for these keys, which leads to cookies being
// invalid between across restarts.
func WithCookieSecrets(hash []byte, block []byte) Option {
	return func(c *config) {
		c.BlockKey = block
		c.HashKey = hash
	}
}

// WithRenderer sets the renderer.
func WithRenderer(renderer echo.Renderer) Option {
	return func(c *config) {
		c.Renderer = renderer
	}
}

// Static adds the http.FileSystem under the defined prefix.
func (s *Server) Static(prefix string, fs http.FileSystem) {
	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	path := path.Join(prefix, "*")
	handler := func(c echo.Context) error {
		fileServer.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}

	s.Group.GET(path, handler)
	s.Group.HEAD(path, handler)
}

// Routes returns the defined routes.
func (s *Server) Routes() []*echo.Route {
	return s.server.Routes()
}

// Render returns an echo HandlerFunc that renders the specified template
// without any data.
func (s *Server) Render(name string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, name, nil)
	}
}

// WithErrorTemplate sets the name of the error template to use for rendering html errors.
func WithErrorTemplate(name string) Option {
	return func(c *config) {
		c.ErrorTemplate = name
	}
}
