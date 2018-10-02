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

package assets

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/assets/templates"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/fs"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/middleware"
)

// Config contains the configuration variables for the assets.
type Config struct {
	// Mount is the location where the assets are mounted from.
	Mount string `name:"mount" description:"Location where assets are mounted from"`

	// SearchPath is a list of paths for finding the directory to serve from,
	// falls back to the CDN.
	SearchPath []string `name:"search-path" description:"List of paths for finding the directory to serve from, falls back to the CDN"`

	// CDN is the public URL of a content delivery network to serve assets using Apps.
	CDN string `name:"url" description:"Public URL of a content delivery network to serve assets"`

	// Apps contains static HTML pages with applications that are loaded from the CDN.
	Apps map[string]templates.AppData `name:"-"`
}

// Assets serves through a web server either bundled assets in the binary or
// directly read from the file system.
type Assets struct {
	*component.Component
	logger log.Interface
	config Config
	fs     http.FileSystem
}

var (
	errNoLocation    = errors.DefineInvalidArgument("no_location", "no assets location specified; specify local search path or CDN")
	errLocalNotFound = errors.DefineNotFound("local_not_found", "assets not found in search path `{search_path}` and no CDN specified")
)

// New creates a new assets instance.
func New(c *component.Component, config Config) (*Assets, error) {
	assets := &Assets{
		Component: c,
		logger: log.FromContext(c.Context()).WithFields(log.Fields(
			"namespace", "assets",
			"mount", config.Mount,
		)),
		config: config,
	}

	for _, path := range config.SearchPath {
		if s, err := os.Stat(path); !os.IsNotExist(err) && s.IsDir() {
			assets.fs = http.Dir(path)
			assets.logger = assets.logger.WithField("path", path)
			break
		}
	}
	if assets.fs == nil {
		if config.CDN == "" {
			if len(config.SearchPath) > 0 {
				return nil, errLocalNotFound.WithAttributes("search_path", strings.Join(config.SearchPath, ", "))
			}
			return nil, errNoLocation
		}
		assets.logger = assets.logger.WithField("cdn", config.CDN)
	}

	assets.logger.Debug("Serving assets")
	c.RegisterWeb(assets)

	return assets, nil
}

// MustNew calls New and returns new assets or panics on an error.
// In most cases, you should just use New.
func MustNew(c *component.Component, config Config) *Assets {
	as, err := New(c, config)
	if err != nil {
		panic(err)
	}
	return as
}

// RegisterRoutes registers the assets to the web server, serving the assets.
func (a *Assets) RegisterRoutes(server *web.Server) {
	if a.fs != nil {
		server.Static(a.config.Mount, a.fs, middleware.Immutable)
	}
}

var errTemplateNotFound = errors.DefineNotFound("template_not_found", "template `{name}` not found")

// AppHandler returns an echo.HandlerFunc that renders an application.
func (a *Assets) AppHandler(name string, env interface{}) echo.HandlerFunc {
	var (
		logger       = a.logger.WithField("template", name)
		renderMu     sync.RWMutex
		renderErr    error
		renderResult []byte
		render       = func(t *template.Template, data templates.Data) {
			res := &bytes.Buffer{}
			err := t.Execute(res, data)
			renderMu.Lock()
			renderResult, renderErr = res.Bytes(), err
			renderMu.Unlock()
		}
	)
	if a.fs != nil {
		data := templates.Data{
			Root: a.config.Mount,
			Env:  env,
		}
		loadTemplate := func() {
			t, err := a.loadTemplate(name)
			if err != nil {
				logger.WithError(err).Error("Could not load template")
				return
			}
			render(t, data)
			logger.Debug("Loaded template")
		}
		loadTemplate()
		if httpFS, ok := a.fs.(http.Dir); ok {
			relName := filepath.Join(string(httpFS), name)
			fs.Watch(relName, events.HandlerFunc(func(evt events.Event) {
				if evt.Name() != "fs.write" {
					return
				}
				loadTemplate()
			}))
		}
	} else {
		appData, ok := a.config.Apps[name]
		if ok {
			render(templates.App, templates.Data{
				Root: a.config.CDN,
				Env:  env,
				Data: appData,
			})
		}
	}

	return func(c echo.Context) error {
		renderMu.RLock()
		res, err := renderResult, renderErr
		renderMu.RUnlock()
		if err != nil {
			return err
		}
		if res == nil {
			return errTemplateNotFound.WithAttributes("name", name)
		}
		c.Response().WriteHeader(http.StatusOK)
		c.Response().Writer.Write(res)
		return err
	}
}

func (a *Assets) loadTemplate(name string) (*template.Template, error) {
	f, err := a.fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errTemplateNotFound.WithCause(err)
		}
		return nil, err
	}
	defer f.Close()
	html, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return template.New(name).Parse(string(html))
}
