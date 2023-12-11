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

package web

import (
	"context"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/fillcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
	"gopkg.in/yaml.v2"
)

var responseCompressionFeatureFlag = experimental.DefineFeature("http.server.transport.compression", true)

func compressionMiddleware(ctx context.Context) (func(http.Handler) http.Handler, error) {
	if !responseCompressionFeatureFlag.GetValue(ctx) {
		return func(next http.Handler) http.Handler { return next }, nil
	}
	m, err := gzhttp.NewWrapper()
	if err != nil {
		return nil, err
	}
	return func(h http.Handler) http.Handler { return m(h) }, nil
}

// Registerer allows components to register their services to the web server.
type Registerer interface {
	RegisterRoutes(s *Server)
}

// Server is the server.
type Server struct {
	// The root HTTP router.
	root *mux.Router

	// The main HTTP router.
	router *mux.Router

	// The HTTP router for API.
	apiRouter *mux.Router
}

type options struct {
	disableWarnings bool

	cookieHashKey  []byte
	cookieBlockKey []byte

	staticMount       string
	staticSearchPaths []string

	trustedProxies []string

	contextFillers []fillcontext.Filler

	redirectToHost  string
	redirectToHTTPS map[int]int

	logIgnorePaths []string
}

// Option for the web server
type Option func(*options)

// WithDisableWarnings configures if the webserver should emit misconfiguration warnings.
func WithDisableWarnings(disable bool) Option {
	return func(o *options) {
		o.disableWarnings = disable
	}
}

// WithContextFiller sets context fillers that are executed on every request context.
func WithContextFiller(contextFillers ...fillcontext.Filler) Option {
	return func(o *options) {
		o.contextFillers = append(o.contextFillers, contextFillers...)
	}
}

// WithTrustedProxies adds trusted proxies from which proxy headers are trusted.
func WithTrustedProxies(cidrs ...string) Option {
	return func(o *options) {
		o.trustedProxies = append(o.trustedProxies, cidrs...)
	}
}

// WithCookieKeys sets the cookie hash key and block key.
func WithCookieKeys(hashKey, blockKey []byte) Option {
	return func(o *options) {
		o.cookieHashKey, o.cookieBlockKey = hashKey, blockKey
	}
}

// WithStatic sets the mount and search paths for static assets.
func WithStatic(mount string, searchPaths ...string) Option {
	return func(o *options) {
		o.staticMount, o.staticSearchPaths = mount, searchPaths
	}
}

// WithRedirectToHost redirects all requests to this host.
func WithRedirectToHost(target string) Option {
	return func(o *options) {
		o.redirectToHost = target
	}
}

// WithRedirectToHTTPS redirects HTTP requests to HTTPS.
func WithRedirectToHTTPS(from, to int) Option {
	return func(o *options) {
		if o.redirectToHTTPS == nil {
			o.redirectToHTTPS = make(map[int]int)
		}
		o.redirectToHTTPS[from] = to
	}
}

// WithLogIgnorePaths silences log messages for a list of URLs.
func WithLogIgnorePaths(paths []string) Option {
	return func(o *options) {
		o.logIgnorePaths = paths
	}
}

// New builds a new server.
func New(ctx context.Context, opts ...Option) (*Server, error) {
	logger := log.FromContext(ctx).WithField("namespace", "web")

	options := new(options)
	for _, opt := range opts {
		opt(options)
	}

	hashKey, blockKey := options.cookieHashKey, options.cookieBlockKey

	if len(hashKey) == 0 || isZeros(hashKey) {
		hashKey = random.Bytes(64)
		if !options.disableWarnings {
			logger.Warn("No cookie hash key configured, generated a random one")
		}
	}

	if len(hashKey) != 32 && len(hashKey) != 64 {
		return nil, errors.New("Expected cookie hash key to be 32 or 64 bytes long")
	}

	if len(blockKey) == 0 || isZeros(blockKey) {
		blockKey = random.Bytes(32)
		if !options.disableWarnings {
			logger.Warn("No cookie block key configured, generated a random one")
		}
	}

	if len(blockKey) != 32 {
		return nil, errors.New("Expected cookie block key to be 32 bytes long")
	}

	var proxyConfiguration webmiddleware.ProxyConfiguration
	if err := proxyConfiguration.ParseAndAddTrusted(options.trustedProxies...); err != nil {
		return nil, err
	}
	compressor, err := compressionMiddleware(ctx)
	if err != nil {
		return nil, err
	}
	root := mux.NewRouter()
	root.NotFoundHandler = http.HandlerFunc(webhandlers.NotFound)
	root.Use(
		webhandlers.WithErrorHandlers(map[string]http.Handler{
			"text/html": webhandlers.Template,
		}),
		mux.MiddlewareFunc(webmiddleware.Recover()),
		compressor,
		otelmux.Middleware("ttn-lw-stack", otelmux.WithTracerProvider(tracing.FromContext(ctx))),
		mux.MiddlewareFunc(webmiddleware.FillContext(options.contextFillers...)),
		mux.MiddlewareFunc(webmiddleware.Peer()),
		mux.MiddlewareFunc(webmiddleware.RequestURL()),
		mux.MiddlewareFunc(webmiddleware.RequestID()),
		mux.MiddlewareFunc(webmiddleware.ProxyHeaders(proxyConfiguration)),
		mux.MiddlewareFunc(webmiddleware.Metadata("X-Forwarded-For", "User-Agent")),
		mux.MiddlewareFunc(webmiddleware.MaxBody(1024*1024*16)),
		mux.MiddlewareFunc(webmiddleware.SecurityHeaders()),
		mux.MiddlewareFunc(webmiddleware.Log(logger, options.logIgnorePaths)),
		mux.MiddlewareFunc(webmiddleware.Cookies(hashKey, blockKey)),
		mux.MiddlewareFunc(webmiddleware.NoCache),
	)

	var redirectConfig webmiddleware.RedirectConfiguration
	if options.redirectToHost != "" {
		if host, portStr, err := net.SplitHostPort(options.redirectToHost); err == nil {
			redirectConfig.HostName = func(string) string { return host }
			port, err := strconv.ParseUint(portStr, 10, 0)
			if err != nil {
				return nil, err
			}
			redirectConfig.Port = func(uint) uint { return uint(port) }
		} else {
			redirectConfig.HostName = func(string) string { return options.redirectToHost }
		}
	}
	if options.redirectToHTTPS != nil {
		redirectConfig.Scheme = func(string) string { return "https" }
		// Only redirect to HTTPS port if no port redirection has been configured
		if redirectConfig.Port == nil {
			redirectConfig.Port = func(current uint) uint {
				return uint(options.redirectToHTTPS[int(current)])
			}
		}
	}

	router := root.NewRoute().Subrouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Redirect(redirectConfig)),
	)

	apiRouter := mux.NewRouter()
	apiRouter.NotFoundHandler = http.HandlerFunc(webhandlers.NotFound)
	apiRouter.Use(
		webhandlers.WithErrorHandlers(map[string]http.Handler{
			"text/html": webhandlers.Template,
		}),
		mux.MiddlewareFunc(webmiddleware.CookieAuth("_session")),
		mux.MiddlewareFunc(webmiddleware.CSRF(
			hashKey,
			csrf.CookieName("_csrf"),
			csrf.FieldName("_csrf"),
			csrf.Path("/"),
		)),
		mux.MiddlewareFunc(
			webmiddleware.CORS(webmiddleware.CORSConfig{
				AllowedHeaders: []string{"Authorization", "Content-Type", "X-CSRF-Token"},
				AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
				AllowedOrigins: []string{"*"},
				ExposedHeaders: []string{
					"Date",
					"Content-Length",
					"X-Rate-Limit-Limit",
					"X-Rate-Limit-Available",
					"X-Rate-Limit-Reset",
					"X-Rate-Limit-Retry",
					"X-Request-Id",
					"X-Total-Count",
					"X-Warning",
				},
				MaxAge: 600,
			}),
		),
	)
	root.PathPrefix("/api/").Handler(apiRouter)

	s := &Server{
		root:      root,
		router:    router,
		apiRouter: apiRouter,
	}

	var staticPath string
	for _, path := range options.staticSearchPaths {
		if s, err := os.Stat(path); err == nil && s.IsDir() {
			staticPath = path
			break
		}
	}
	if staticPath != "" {
		staticDir := http.Dir(staticPath)
		logger := logger.WithFields(log.Fields("path", staticDir, "mount", options.staticMount))
		s.Static(options.staticMount, staticDir)

		// register hashed filenames
		manifest, err := os.ReadFile(filepath.Join(staticPath, "manifest.yaml"))
		if err != nil {
			logger.WithError(err).Warn("Failed to load manifest.yaml")
			return s, nil
		}
		hashedFiles := make(map[string]string)
		err = yaml.Unmarshal(manifest, &hashedFiles)
		if err != nil {
			return nil, errors.New("Corrupted manifest.yaml").WithCause(err)
		}
		for original, hashed := range hashedFiles {
			webui.RegisterHashedFile(original, hashed)
		}
		logger.Debug("Loaded manifest.yaml")
		logger.Debug("Serving static assets")
	} else if !options.disableWarnings {
		logger.WithField("search_paths", options.staticSearchPaths).Warn("No static assets found in any search path")
	}

	return s, nil
}

func isZeros(buf []byte) bool {
	for _, b := range buf {
		if b != 0x00 {
			return false
		}
	}

	return true
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.root.ServeHTTP(w, r)
}

// RootRouter returns the root router.
// In most cases the Router() should be used instead of the root router.
func (s *Server) RootRouter() *mux.Router {
	return s.root
}

// Router returns the main router.
func (s *Server) Router() *mux.Router {
	return s.router
}

// APIRouter returns the API router.
func (s *Server) APIRouter() *mux.Router {
	return s.apiRouter
}

func (s *Server) getRouter(path string) *mux.Router {
	if strings.HasPrefix(path, "/api/") {
		return s.apiRouter
	}
	return s.router
}

var hashRegex = regexp.MustCompile(`\.([a-f0-9]{20}|[a-f0-9]{32})(\.bundle)?\.(js|css|woff|woff2|ttf|eot|jpg|jpeg|png|svg)$`)

// Static adds the http.FileSystem under the defined prefix.
func (s *Server) Static(prefix string, fs http.FileSystem) {
	prefix = "/" + strings.Trim(prefix, "/") + "/"
	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	s.router.PathPrefix(prefix).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hashRegex.MatchString(path.Base(r.URL.String())) {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
			w.Header().Del("Pragma")
		}
		fileServer.ServeHTTP(w, r)
	})
}

// Prefix returns a route for the given path prefix.
func (s *Server) Prefix(prefix string) *mux.Route {
	return s.getRouter(prefix).PathPrefix(prefix)
}

// PrefixWithRedirect will create a route ending in slash.
// Paths which coincide with the route, but do not end with slash, will be
// redirect to the slash ending route.
func (s *Server) PrefixWithRedirect(prefix string) *mux.Route {
	prefix = "/" + strings.Trim(prefix, "/")
	prefixWithSlash := prefix
	if prefix != "/" {
		prefixWithSlash = prefix + "/"
		s.getRouter(prefix).Path(prefix).Handler(http.RedirectHandler(prefixWithSlash, http.StatusPermanentRedirect))
	}
	return s.getRouter(prefixWithSlash).PathPrefix(prefixWithSlash)
}
