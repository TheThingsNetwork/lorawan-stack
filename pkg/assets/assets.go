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
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/assets/templates"
	"go.thethings.network/lorawan-stack/pkg/component"
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
	config Config
	fs     http.FileSystem
}

// New creates a new assets instance.
func New(c *component.Component, config Config) (*Assets, error) {
	assets := &Assets{
		Component: c,
		config:    config,
	}
	logger := log.FromContext(assets.Context()).WithFields(log.Fields(
		"namespace", "assets",
		"mount", config.Mount,
	))

	for _, path := range config.SearchPath {
		if s, err := os.Stat(path); !os.IsNotExist(err) && s.IsDir() {
			assets.fs = http.Dir(path)
			logger = logger.WithField("path", path)
			break
		}
	}
	if assets.fs == nil {
		if config.CDN == "" {
			return nil, errInvalidConfiguration
		}
		logger = logger.WithField("cdn", config.CDN)
	}

	logger.Debug("Serving assets")
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

// AppHandler returns an echo.HandlerFunc that renders an application.
func (a *Assets) AppHandler(name string, env interface{}) echo.HandlerFunc {
	var (
		t    *template.Template
		err  error
		data = templates.Data{
			Env: env,
		}
	)
	if a.fs != nil {
		t, err = a.loadTemplate(name)
		if err == nil {
			data.Root = a.config.Mount
		}
	} else {
		t = templates.App
		appData, ok := a.config.Apps[name]
		if !ok {
			err = errTemplateNotFound
		} else {
			data.Data = appData
			data.Root = a.config.CDN
		}
	}

	return func(c echo.Context) error {
		if err != nil {
			return err
		}
		c.Response().WriteHeader(http.StatusOK)
		return t.Execute(c.Response().Writer, data)
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
