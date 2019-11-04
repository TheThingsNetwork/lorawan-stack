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

package ttnmage

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

// Docs namespace
type Docs mg.Namespace

func execHugo(args ...string) error {
	return execGo("run", append([]string{"-tags", "extended", "github.com/gohugoio/hugo", "-s", "./doc"}, args...)...)
}

func (d Docs) yarn() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("yarn")); os.IsNotExist(err) {
		if err = installYarn(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.Run(nodeBin("yarn"), append([]string{fmt.Sprintf("--cwd=%s", filepath.Join("doc", "themes", "the-things-stack"))}, args...)...)
	}, nil
}

// Deps installs the documentation dependencies.
func (d Docs) Deps() error {
	changed, err := target.Path("./doc/themes/the-things-stack/node_modules", "./doc/themes/the-things-stack/package.json", "./doc/themes/the-things-stack/yarn.lock")
	if os.IsNotExist(err) || (err == nil && changed) {
		if mg.Verbose() {
			fmt.Println("Installing JS SDK dependencies")
		}
		yarn, err := d.yarn()
		if err != nil {
			return err
		}
		return yarn("install", "--no-progress", "--production=false")
	}
	return nil
}

const (
	docRedirectTemplateFilePath = "doc/redirect.html.tmpl"
	docRedirectFilePath         = "doc/public/index.html"
)

// Build builds a static website from the documentation into public/doc.
// If the HUGO_BASE_URL environment variable is set, it also builds a public website into doc/public.
func (d Docs) Build() (err error) {
	if err = execHugo("-b", "/assets/doc", "-d", "../public/doc"); err != nil {
		return err
	}
	baseURL := os.Getenv("HUGO_BASE_URL")
	if baseURL == "" {
		return nil
	}
	mg.Deps(Version.getCurrent)
	url, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, currentVersion)
	destination := path.Join("public", currentVersion)
	defer func() {
		genErr := d.generateRedirect()
		if err == nil {
			err = genErr
		}
	}()
	return execHugo("-b", url.String(), "-d", destination)
}

func (Docs) generateRedirect() error {
	docTmpl, err := template.New(filepath.Base(docRedirectTemplateFilePath)).ParseFiles(docRedirectTemplateFilePath)
	if err != nil {
		return err
	}
	target, err := os.OpenFile(docRedirectFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil
	}
	defer target.Close()
	return docTmpl.Execute(target, struct {
		CurrentVersion string
	}{
		CurrentVersion: currentVersion,
	})
}

// Server starts a documentation server.
func (Docs) Server() error {
	return execHugo("server")
}
