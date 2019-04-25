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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/pkg/errors"
)

// JsSDK namespace.
type JsSDK mg.Namespace

func (k JsSDK) yarn() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("yarn")); os.IsNotExist(err) {
		if err = installYarn(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.Run(nodeBin("yarn"), append([]string{fmt.Sprintf("--cwd=%s", filepath.Join("sdk", "js"))}, args...)...)
	}, nil
}

// Deps installs the javascript SDK dependencies.
func (k JsSDK) Deps() error {
	if mg.Verbose() {
		fmt.Println("Installing JS SDK dependencies")
	}
	yarn, err := k.yarn()
	if err != nil {
		return err
	}
	return yarn("install", "--no-progress")
}

// Build builds the source files and output into 'dist'.
func (k JsSDK) Build() error {
	mg.SerialDeps(JsSDK.Deps, JsSDK.Definitions)

	if mg.Verbose() {
		fmt.Println("Building JS SDK files…")
	}
	yarn, err := k.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "build")
}

// Watch builds the source files in watch mode.
func (k JsSDK) Watch() error {
	mg.SerialDeps(JsSDK.Deps, JsSDK.Definitions)

	if mg.Verbose() {
		fmt.Println("Building and watching JS SDK files…")
	}
	yarn, err := k.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "build:watch")
}

// Test runs jest unit tests.
func (k JsSDK) Test() error {
	if mg.Verbose() {
		fmt.Println("Running JS SDK tests…")
	}
	yarn, err := k.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "test")
}

// TestWatch runs jest unit tests in watch mode.
func (k JsSDK) TestWatch() error {
	if mg.Verbose() {
		fmt.Println("Running JS SDK tests in watch mode…")
	}
	yarn, err := k.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "test:watch")
}

// Clean clears all transpiled files.
func (k JsSDK) Clean() {
	mg.Deps(JsSDK.DefinitionsClean)
	sh.Rm(filepath.Join("sdk", "js", "dist"))
}

// Definitions extracts the api-definition.json from the proto generated api.json.
func (k JsSDK) Definitions() error {
	mg.Deps(Proto.JsSDK)
	changed, err := target.Path(filepath.Join("sdk", "js", "generated", "api-definition.json"), filepath.Join("sdk", "js", "generated", "api.json"))
	if err != nil {
		return errors.Wrap(err, "failed checking modtime")
	}
	if !changed {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Extracting api definitions from protos…")
	}
	yarn, err := k.yarn()
	if err != nil {
		return errors.Wrap(err, "failed constructing yarn command")
	}
	return yarn("run", "definitions")
}

// DefinitionsClean removes the generated api-definition.json.
func (k JsSDK) DefinitionsClean(context.Context) error {
	return sh.Rm(filepath.Join("sdk", "js", "generated", "api-definition.json"))
}

// Link links the local sdk package via `yarn link` to prevent caching issues.
func (k JsSDK) Link() error {
	if mg.Verbose() {
		fmt.Println("Linking sdk package…")
	}

	y, err := yarn()
	if err != nil {
		return err
	}

	err = y(fmt.Sprintf("--cwd=%s", filepath.Join("sdk", "js")), "link")
	if err != nil {
		return err
	}

	return y("link", "ttn-lw")
}
