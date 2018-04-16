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

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/assets/fs"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/middleware"
)

// Config contains the configuration variables for the assets.
type Config struct {
	// Mount is the root path where the assets will be served from.
	Mount string `name:"mount" description:"The path where the assets will be mounted on the web server"`

	// CDN is the root public url where the assets will be served from.
	CDN string `name:"cdn" description:"The URL that will be used to serve the assets, falls back to assets.mount"`

	// Directory forces the assets to be read from the file system (instead of from the bindata in the binary).
	Directory string `name:"dir" description:"Force assets to be read from this directory instead of bundled ones."`
}

// Assets serves through a web server either bundled assets in the binary or
// directly read from the file system.
type Assets struct {
	*component.Component
	assetsFS http.FileSystem
	config   Config
}

// New creates a new assets instance.
func New(c *component.Component, config Config) *Assets {
	assets := &Assets{
		Component: c,
		assetsFS:  assetFS(),
		config:    config,
	}

	logger := log.FromContext(assets.Context()).WithField("namespace", "assets")

	if assets.config.CDN == "" {
		assets.config.CDN = assets.config.Mount
	}

	logger.WithFields(log.Fields(
		"mount", assets.config.Mount,
		"cdn", assets.config.CDN,
		"directory", assets.config.Directory,
	)).Debug("Serving assets")

	c.RegisterWeb(assets)

	return assets
}

// RegisterRoutes registers the assets to the web server, serving the assets.
func (a *Assets) RegisterRoutes(server *web.Server) {
	fs := fs.Hide(a.FileSystem(), "/console.html", "/oauth.html")
	server.Static(a.config.Mount, fs, middleware.Immutable)
}

// FileSystem returns a http.FileSystem that contains the assets.
func (a *Assets) FileSystem() http.FileSystem {
	if a.config.Directory != "" {
		return http.Dir(a.config.Directory)
	}

	return a.assetsFS
}

type data struct {
	// Root is the root where the assets will be served from.
	Root string

	// Env is the custom environment.
	Env interface{}

	// Error is a possible error that occurred.
	Error *errors.Error
}

// Render creates an echo.HandlerFunc that renders the selected template html file
// from the assets filesystem.
func (a *Assets) Render(name string, env interface{}) echo.HandlerFunc {
	template := a.template(name)

	return func(c echo.Context) error {
		t := a.fresh(name, template)

		data := data{
			Root:  a.config.CDN,
			Env:   env,
			Error: nil,
		}

		c.Response().WriteHeader(http.StatusOK)
		return t.Execute(c.Response().Writer, data)
	}
}

// template reads the template file from the filesystem and parses it.
// Panics if anything goes wrong.
func (a *Assets) template(name string) *template.Template {
	index, err := a.FileSystem().Open(name)
	if err != nil {
		panic(err)
	}

	html, err := ioutil.ReadAll(index)
	if err != nil {
		panic(err)
	}

	t, err := template.New(name).Parse(string(html))
	if err != nil {
		panic(err)
	}

	return t
}

func (a *Assets) fresh(name string, t *template.Template) *template.Template {
	if a.config.Directory == "" {
		return t
	}

	return a.template(name)
}
