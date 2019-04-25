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
	"github.com/pkg/errors"
)

// DevDeps installs the javascript SDK development dependencies.
func (sdkJs SdkJs) devDeps() error {
	_, err := sdkJs.yarn()
	return err
}

// SdkJs namespace.
type SdkJs mg.Namespace

func (sdkJs SdkJs) yarn() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("yarn")); os.IsNotExist(err) {
		if err = installYarn(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.Run(nodeBin("yarn"), append([]string{"--cwd=sdk/js"}, args...)...)
	}, nil
}

// DevDeps installs the javascript development dependencies.
func (sdkJs SdkJs) DevDeps() error {
	return sdkJs.devDeps()
}

// Deps installs the javascript dependencies.
func (sdkJs SdkJs) Deps() error {
	if mg.Verbose() {
		fmt.Println("Installing JS SDK dependencies")
	}
	yarn, err := sdkJs.yarn()
	if err != nil {
		return err
	}
	return yarn("install", "--no-progress")
}

// Build builds the source files and output into 'dist'.
func (sdkJs SdkJs) Build() error {
	mg.SerialDeps(sdkJs.Deps, sdkJs.Definitions)

	if mg.Verbose() {
		fmt.Println("Building JS SDK files…")
	}
	yarn, err := sdkJs.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "build")
}

// Watch builds the source files in watch mode.
func (sdkJs SdkJs) Watch() error {
	mg.SerialDeps(sdkJs.Deps, sdkJs.Definitions)

	if mg.Verbose() {
		fmt.Println("Building and watching JS SDK files…")
	}
	yarn, err := sdkJs.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "build:watch")
}

// Test runs jest unit tests.
func (sdkJs SdkJs) Test() error {
	if mg.Verbose() {
		fmt.Println("Running JS SDK tests…")
	}
	yarn, err := sdkJs.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "test")
}

// TestWatch runs jest unit tests in watch mode.
func (sdkJs SdkJs) TestWatch() error {
	if mg.Verbose() {
		fmt.Println("Running JS SDK tests in watch mode…")
	}
	yarn, err := sdkJs.yarn()
	if err != nil {
		return err
	}

	return yarn("run", "test:watch")
}

// Clean clears all transpiled files.
func (sdkJs SdkJs) Clean() {
	mg.Deps(SdkJs.DefinitionsClean)
	sh.Rm("./sdk/js/dist")
}

// Definitions extracts the api-definition.json from the proto generated api.json.
func (sdkJs SdkJs) Definitions(context.Context) error {
	mg.Deps(Proto.SdkJs)
	yarn, err := yarn()
	if err != nil {
		return errors.Wrap(err, "failed to construct yarn command")
	}
	if err := yarn("--cwd=./sdk/js", "run", "definitions"); err != nil {
		return errors.Wrap(err, "failed to generate definitions")
	}
	return nil
}

// DefinitionsClean removes the generated api-definition.json.
func (sdkJs SdkJs) DefinitionsClean(context.Context) error {
	return sh.Rm(filepath.Join("sdk", "js", "generated", "api-definition.json"))
}

// Link links the local sdk package via `yarn link` to prevent caching issues.
func (sdkJs SdkJs) Link() error {
	if mg.Verbose() {
		fmt.Println("Linking sdk package…")
	}

	y, err := yarn()
	if err != nil {
		return err
	}

	err = y("--cwd=./sdk/js", "link")
	if err != nil {
		return err
	}

	return y("link", "ttn-lw")
}
