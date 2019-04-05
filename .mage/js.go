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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func nodeBin(cmd string) string { return filepath.Join("node_modules", ".bin", cmd) }

// Js namespace.
type Js mg.Namespace

func (Js) installYarn() error {
	packageJSONBytes, err := ioutil.ReadFile("package.json")
	if err != nil {
		return err
	}
	var packageJSON struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err = json.Unmarshal(packageJSONBytes, &packageJSON); err != nil {
		return err
	}
	yarn, ok := packageJSON.DevDependencies["yarn"]
	if ok {
		yarn = "yarn@" + yarn
	} else {
		yarn = "yarn"
	}
	if mg.Verbose() {
		fmt.Printf("Installing Yarn %s\n", yarn)
	}
	return sh.RunV("npm", "install", "--no-package-lock", "--no-save", "--production=false", yarn)
}

func (js Js) yarn() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("yarn")); os.IsNotExist(err) {
		if err = js.installYarn(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.RunV(nodeBin("yarn"), args...)
	}, nil
}

// DevDeps installs the javascript development dependencies.
func (js Js) DevDeps() error {
	_, err := js.yarn()
	return err
}

// Deps installs the javascript dependencies.
func (js Js) Deps() error {
	if mg.Verbose() {
		fmt.Println("Installing JS dependencies")
	}
	yarn, err := js.yarn()
	if err != nil {
		return err
	}
	return yarn("install", "--no-progress", "--production=false")
}
