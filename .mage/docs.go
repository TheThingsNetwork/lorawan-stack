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
	"net/url"
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Docs namespace
type Docs mg.Namespace

func execHugo(args ...string) error {
	return execGo("run", append([]string{"github.com/gohugoio/hugo", "-s", "./doc"}, args...)...)
}

// Deps installs documentation dependencies.
func (Docs) Deps() error {
	return sh.RunV("git", "submodule", "update", "--init", "doc/themes/hugo-theme-techdoc")
}

// Build builds a static website from the documentation into doc/public.
func (Docs) Build() error {
	mg.Deps(Version.getCurrent)
	var args []string
	if baseURL := os.Getenv("HUGO_BASE_URL"); baseURL != "" {
		url, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		url.Path = path.Join(url.Path, currentVersion)
		destination := path.Join("public", currentVersion)
		args = append(args, "-b", url.String(), "-d", destination)
	}
	return execHugo(args...)
}

// Server starts a documentation server.
func (Docs) Server() error {
	return execHugo("server")
}
