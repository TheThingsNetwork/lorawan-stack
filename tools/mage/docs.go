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
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/target"
)

// Docs namespace
type Docs mg.Namespace

func runHugo(args ...string) error {
	return runGoTool(append([]string{"-tags", "extended", "github.com/gohugoio/hugo", "-s", "./doc"}, args...)...)
}

func downloadFile(targetpath string, url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(targetpath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	_, err = io.Copy(out, resp.Body)
	return err
}

const defaultFrequencyPlanUrl = "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master/frequency-plans.yml"

// Deps installs the documentation dependencies.
func (d Docs) Deps() (err error) {
	fileUrl := os.Getenv("FREQUENCY_PLAN_URL")
	fileTarget := filepath.Join("doc", "data", "frequency-plans.yml")
	ok, err := target.Path(fileTarget)
	if err != nil {
		return targetError(err)
	}
	if ok {
		if fileUrl == "" {
			fileUrl = defaultFrequencyPlanUrl
		}
		if err = downloadFile(fileTarget, fileUrl); err != nil {
			return err
		}
		if mg.Verbose() {
			fmt.Printf("Downloaded %q to %q\n", fileUrl, fileTarget)
		}
	}
	ok, err = target.Dir(
		filepath.Join("doc", "themes", "the-things-stack", "node_modules"),
		filepath.Join("doc", "themes", "the-things-stack", "package.json"),
		filepath.Join("doc", "themes", "the-things-stack", "yarn.lock"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Installing documentation dependencies")
	}
	mg.Deps(installYarn)
	return runYarnV(
		yarnWorkingDirectoryArg("doc", "themes", "the-things-stack"),
		"install",
		"--no-progress",
		"--production=false",
	)
}

var (
	docRedirectTemplateFilePath = filepath.Join("doc", "redirect.html.tmpl")
	docRedirectFilePath         = filepath.Join("doc", "public", "index.html")
)

// Build builds a static website from the documentation into public/doc.
// If the HUGO_BASE_URL environment variable is set, it also builds a public website into doc/public.
func (d Docs) Build() (err error) {
	mg.Deps(d.Deps)
	if err = runHugo("-b", "/assets/doc", "-d", "../public/doc"); err != nil {
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
	return runHugo("-b", url.String(), "-d", destination, "--environment", "gh-pages")
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
	return runHugo("server", "--environment", "gh-pages")
}
