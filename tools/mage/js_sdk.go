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
)

// JsSDK namespace.
type JsSDK mg.Namespace

func (k JsSDK) runYarnV(args ...string) error {
	return runYarnV(append([]string{yarnWorkingDirectoryArg("sdk", "js")}, args...)...)
}

func (k JsSDK) runYarnCommandV(cmd string, args ...string) error {
	return k.runYarnV(append([]string{"run", cmd}, args...)...)
}

// Deps installs the javascript SDK dependencies.
func (k JsSDK) Deps() error {
	ok, err := target.Dir(
		filepath.Join("sdk", "js", "node_modules"),
		filepath.Join("sdk", "js", "package.json"),
		filepath.Join("sdk", "js", "yarn.lock"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Installing JS SDK dependencies")
	}
	mg.Deps(Js.deps)

	// On initial installs, the dependency installation will cause the SDK itself
	// to be installed in an unbuilt state. In that case we need to remove the
	// module altogether so it can be properly reinstalled later.
	if _, err := os.Stat(filepath.Join("node_modules", "ttn-lw", "dist")); os.IsNotExist(err) {
		sh.Rm(filepath.Join("node_modules", "ttn-lw"))
	}
	return k.runYarnV("install", "--no-progress", "--production=false")
}

// Build builds the source files and output into 'dist'.
func (k JsSDK) Build() error {
	ok, err := target.Dir(
		filepath.Join("sdk", "js", "dist"),
		filepath.Join("sdk", "js", "src"),
		filepath.Join("sdk", "js", "generated"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	mg.Deps(k.Deps, k.Definitions)
	if mg.Verbose() {
		fmt.Println("Building JS SDK files")
	}
	return k.runYarnCommandV("build")
}

// Watch builds the source files in watch mode.
func (k JsSDK) Watch() error {
	mg.Deps(JsSDK.Deps, JsSDK.Definitions)
	if mg.Verbose() {
		fmt.Println("Building and watching JS SDK files")
	}
	return k.runYarnCommandV("build:watch")
}

// Fmt formats all js files.
func (k JsSDK) Fmt() error {
	mg.Deps(JsSDK.Deps)
	if mg.Verbose() {
		fmt.Println("Running prettier on JS SDK .js files")
	}
	return k.runYarnCommandV("fmt")
}

// Lint runs eslint over sdk js files.
func (k JsSDK) Lint() error {
	mg.Deps(JsSDK.Deps, JsSDK.Definitions)
	if mg.Verbose() {
		fmt.Println("Running eslint on JS SDK .js files")
	}
	return k.runYarnCommandV("lint")
}

// Test runs jest unit tests.
func (k JsSDK) Test() error {
	mg.Deps(JsSDK.Deps)
	if mg.Verbose() {
		fmt.Println("Running JS SDK tests")
	}
	return k.runYarnCommandV("test")
}

// TestWatch runs jest unit tests in watch mode.
func (k JsSDK) TestWatch() error {
	mg.Deps(JsSDK.Deps)
	if mg.Verbose() {
		fmt.Println("Running JS SDK tests in watch mode")
	}
	return k.runYarnCommandV("test:watch")
}

// Clean clears all transpiled files.
func (k JsSDK) Clean() {
	mg.Deps(JsSDK.DefinitionsClean)
	sh.Rm(filepath.Join("sdk", "js", "dist"))
}

// CleanDeps removes all installed node packages (rm -rf node_modules).
func (JsSDK) CleanDeps() {
	sh.Rm(filepath.Join("sdk", "js", "node_modules"))
}

// Definitions extracts the api-definition.json from the proto generated api.json.
func (k JsSDK) Definitions() error {
	mg.Deps(Proto.JsSDK, JsSDK.AllowedFieldMaskPaths)
	ok, err := target.Path(
		filepath.Join("sdk", "js", "generated", "api-definition.json"),
		filepath.Join("sdk", "js", "generated", "api.json"),
		filepath.Join("sdk", "js", "generated", "allowed-field-mask-paths.json"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Extracting api definitions from protos")
	}
	return k.runYarnCommandV("definitions")
}

// DefinitionsClean removes the generated api-definition.json.
func (k JsSDK) DefinitionsClean(context.Context) error {
	err := sh.Rm(filepath.Join("sdk", "js", "generated", "allowed-field-mask-paths.json"))
	if err != nil {
		return err
	}
	return sh.Rm(filepath.Join("sdk", "js", "generated", "api-definition.json"))
}

// AllowedFieldMaskPaths builds the allowed field masks file based on the ttnpb package.
func (k JsSDK) AllowedFieldMaskPaths() error {
	ok, err := target.Path(
		filepath.Join("sdk", "js", "generated", "allowed-field-mask-paths.json"),
		filepath.Join("pkg", "ttnpb", "field_mask_validation.go"),
		filepath.Join("tools", "generate_allowed_field_mask_paths.go"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	return runGoTool("generate_allowed_field_mask_paths.go")
}

// DeviceFieldMasks generates end device entity map.
func (k JsSDK) DeviceFieldMasks() error {
	ok, err := target.Path(
		filepath.Join("sdk", "js", "generated", "device-entity-map.json"),
		filepath.Join("sdk", "js", "generated", "device-field-masks.json"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	return sh.Run("node", "sdk/js/util/device-field-mask-mapper.js")
}
