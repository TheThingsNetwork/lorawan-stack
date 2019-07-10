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
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Docs namespace
type Docs mg.Namespace

func execHugo(args ...string) error {
	return execGo("run", append([]string{"github.com/gohugoio/hugo", "-s", "doc"}, args...)...)
}

// Gen generates static website from the doc in doc/public.
func (Docs) Gen() error {
	baseUrl := getDocURL()
	var args []string
	if baseUrl != "" {
		args = []string{"--baseUrl", baseUrl + currentVersion}
	}
	return execHugo(args...)
}

// Gen generates static website from the doc in doc/public/$version.
func (Docs) GenVersion() error {
	mg.Deps(Version.getCurrent)
	baseUrl := getDocURL()
	var args []string
	if baseUrl != "" {
		args = []string{"--baseUrl", baseUrl + currentVersion}
	}
	return execHugo(args...)
}

// Docs install documentation dependencies.
func (Docs) Deps() error {
	return sh.RunV("git", "submodule", "update", "--init", "doc/themes/hugo-theme-techdoc")
}

// Server starts live documentation server.
func (Docs) Server() error {
	return execHugo("server")
}
