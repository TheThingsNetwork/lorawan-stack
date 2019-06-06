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
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"io/ioutil"
	"os"
	"text/template"
)

type HugoConfig struct {
	CurrentVersion string
}

// Docs namespace
type Docs mg.Namespace

var hugoArgs = []string{"run", "github.com/gohugoio/hugo", "-s", "doc"}

// Gen generates static website from the doc in doc/public.
func (Docs) Gen() error {
	baseUrl := getDocURL()
	args := append(hugoArgs, "-d", "public/")
	if baseUrl != "" {
		args = append(args, "--baseUrl", baseUrl)
	}
	return sh.RunWith(map[string]string{"GO111MODULE": "on"}, "go", args...)
}

// Gen generates static website from the doc in doc/public/$version.
func (Docs) GenVersion() error {
	mg.Deps(Version.getCurrent)
	baseUrl := getDocURL()
	args := append(hugoArgs, "-d", "public/"+currentVersion)
	if baseUrl != "" {
		args = append(args, "--baseUrl", baseUrl+currentVersion)
	}
	return sh.RunWith(map[string]string{"GO111MODULE": "on"}, "go", args...)
}

// Docs install documentation dependencies.
func (Docs) Deps() error {
	return sh.RunV("git", "submodule", "update", "--init", "doc/themes/hugo-theme-techdoc")
}

// Server starts live documentation server.
func (Docs) Server() error {
	return sh.RunWith(map[string]string{"GO111MODULE": "on"}, "go", append(hugoArgs, "server")...)
}

// Config generate hugo documentation with git metadata.
func (Docs) Config() error {

	mg.Deps(Version.getCurrent)
	tmpl, err := ioutil.ReadFile("./doc/config.tmpl")
	if err != nil {
		return err
	}
	cfg := HugoConfig{
		CurrentVersion: currentVersion,
	}
	t := template.Must(template.New("config").Parse(string(tmpl)))
	file, err := os.OpenFile("doc/config.toml", os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		return nil
	}
	err = t.Execute(file, cfg)
	return err
}

func getDocURL() string {
	return os.Getenv("HUGO_DOC_URL")
}
