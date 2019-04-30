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
	"github.com/magefile/mage/target"
)

func nodeBin(cmd string) string { return filepath.Join("node_modules", ".bin", cmd) }

// DevDeps installs the javascript development dependencies.
func devDeps() error {
	_, err := yarn()
	if err != nil {
		return err
	}
	mg.Deps(Js.Deps)
	return nil
}

// Js namespace.
type Js mg.Namespace

func installYarn() error {
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

func yarn() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("yarn")); os.IsNotExist(err) {
		if err = installYarn(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.RunV(nodeBin("yarn"), args...)
	}, nil
}

func (js Js) node() (func(args ...string) error, error) {
	return func(args ...string) error {
		return sh.Run("node", args...)
	}, nil
}

func (js Js) execFromNodeBin(cmd string, verbose bool) (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin(cmd)); os.IsNotExist(err) {
		if err = js.DevDeps(); err != nil {
			return nil, err
		}
	}

	out := os.Stdout
	if !mg.Verbose() && !verbose {
		out = nil
	}
	return func(args ...string) error {
		_, err := sh.Exec(nil, out, os.Stderr, nodeBin(cmd), args...)
		return err
	}, nil
}

func (js Js) webpack() (func(args ...string) error, error) {
	return js.execFromNodeBin("webpack", false)
}

func (js Js) webpackServe() (func(args ...string) error, error) {
	return js.execFromNodeBin("webpack-dev-server", true)
}

func (js Js) babel() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("babel")); os.IsNotExist(err) {
		if err = js.DevDeps(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		_, err := sh.Exec(nil, nil, os.Stderr, nodeBin("babel"), args...)
		return err
	}, nil
}

func (js Js) jest() (func(args ...string) error, error) {
	return js.execFromNodeBin("jest", false)
}

func (js Js) eslint() (func(args ...string) error, error) {
	return js.execFromNodeBin("eslint", false)
}

func (js Js) stylint() (func(args ...string) error, error) {
	return js.execFromNodeBin("stylint", false)
}

func (js Js) storybook() (func(args ...string) error, error) {
	return js.execFromNodeBin("start-storybook", true)
}

// DevDeps installs the javascript development dependencies.
func (js Js) DevDeps() error {
	return devDeps()
}

// Deps installs the javascript dependencies.
func (js Js) Deps() error {
	files, readErr := ioutil.ReadDir("node_modules")
	changed, targetErr := target.Dir("node_modules", "./package.json", "./yarn.lock")
	// Check whether package.json/yarn.lock are newer than node_modules
	// and whether it is not only yarn that is installed via DevDeps()
	if readErr != nil || os.IsNotExist(targetErr) || (targetErr == nil && changed) || len(files) <= 4 {
		if mg.Verbose() {
			fmt.Println("Installing JS dependencies")
		}
		yarn, err := yarn()
		if err != nil {
			return err
		}
		err = yarn("install", "--no-progress", "--production=false")
		if err != nil {
			return err
		}
		mg.Deps(JsSDK.Link)
		return nil
	}
	return nil
}

// Build runs all necessary commands to build the console bundles and files.
func (js Js) Build() {
	mg.SerialDeps(js.Deps, JsSDK.Build, js.BuildDll, js.BuildMain)
}

// BuildMain runs the webpack command with the project config.
func (js Js) BuildMain() error {
	mg.Deps(js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack")
	}
	webpack, err := js.webpack()
	if err != nil {
		return err
	}
	return webpack("--config", "config/webpack.config.babel.js")
}

// BuildDll runs the webpack to build the DLL bundle
func (js Js) BuildDll() error {
	changed, err := target.Path("./public/libs.bundle.js", "./yarn.lock")
	if os.IsNotExist(err) || (err == nil && changed) {
		if mg.Verbose() {
			fmt.Println("Running Webpack for DLL…")
		}
		webpack, err := js.webpack()
		if err != nil {
			return err
		}
		return webpack("--config", "config/webpack.dll.babel.js")
	}
	return nil
}

// Serve builds necessary bundles and serves the console for development.
func (js Js) Serve() {
	mg.Deps(js.ServeMain)
}

// ServeMain runs webpack-dev-server
func (js Js) ServeMain() error {
	mg.Deps(js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack for Main Bundle in watch mode…")
	}
	webpackServe, err := js.webpackServe()
	if err != nil {
		return err
	}
	return webpackServe("--config", "config/webpack.config.babel.js", "-w")
}

// Messages extracts the frontend messages via babel.
func (js Js) Messages() error {
	changed, err := target.Dir("./.cache/messages", "./pkg/webui/console")
	if os.IsNotExist(err) || (err == nil && changed) {
		if mg.Verbose() {
			fmt.Println("Extracting frontend messages…")
		}
		babel, err := js.babel()
		if err != nil {
			return err
		}
		if err = sh.Rm(".cache/messages"); err != nil {
			return err
		}
		if err = os.MkdirAll("pkg/webui/locales", 0755); err != nil {
			return err
		}
		return babel("pkg/webui")
	}
	return nil
}

// Translations builds the frontend locale files.
func (js Js) Translations() error {
	changed, err := target.Dir("./pkg/webui/locales/en.json", "./.cache/messages")
	if os.IsNotExist(err) || (err == nil && changed) {
		mg.Deps(js.Messages)
		if mg.Verbose() {
			fmt.Println("Building frontend locale files…")
		}
		node, err := js.node()
		if err != nil {
			return err
		}
		return node(".mage/translations.js")
	}
	return nil
}

// BackendTranslations builds the backend locale files.
func (js Js) BackendTranslations() error {
	changed, err := target.Path("./pkg/webui/locales/.backend/en.json", "./config/messages.json")
	if os.IsNotExist(err) || (err == nil && changed) {

		if mg.Verbose() {
			fmt.Println("Building backend locale files…")
		}
		node, err := js.node()
		if err != nil {
			return err
		}

		return node(".mage/translations.js", "--backend-messages", "config/messages.json", "--locales", "pkg/webui/locales/.backend", "--backend-only")
	}
	return nil
}

// Clean clears all generated files.
func (js Js) Clean() {
	sh.Rm(".cache")
	sh.Rm("public")
	sh.Rm(filepath.Join("pkg", "webui", "locales", ".backend"))
}

// CleanDeps removes all installed node packages (rm -rf node_modules).
func (js Js) CleanDeps() {
	sh.Rm("node_modules")
}

// Test runs frontend jest tests.
func (js Js) Test() error {
	if mg.Verbose() {
		fmt.Println("Running Tests")
	}
	jest, err := js.jest()
	if err != nil {
		return err
	}
	return jest("./pkg/webui")
}

// Lint runs eslint over frontend js files.
func (js Js) Lint() error {
	if mg.Verbose() {
		fmt.Println("Running eslint on .js files")
	}
	eslint, err := js.eslint()
	if err != nil {
		return err
	}
	return eslint("./pkg/webui/**/*.js", "--no-ignore", "--color")
}

// LintSnap runs eslint over frontend snap files.
func (js Js) LintSnap() error {
	if mg.Verbose() {
		fmt.Println("Running eslint on .snap files")
	}
	eslint, err := js.eslint()
	if err != nil {
		return err
	}
	return eslint("./pkg/webui/**/*.snap", "--no-ignore", "--color")
}

// LintAll runs linters over js and snap files.
func (js Js) LintAll() {
	mg.Deps(js.Lint, js.LintSnap)
}

// Storybook runs a local server with storybook.
func (js Js) Storybook() error {
	if mg.Verbose() {
		fmt.Println("Serving storybook…")
	}
	storybook, err := js.storybook()
	if err != nil {
		return err
	}

	return storybook("--config-dir", "./config/storybook", "--static-dir", "public", "--port", "9001")
}

// Vulnerabilities runs yarn audit to check for vulnerable node packages.
func (js Js) Vulnerabilities() error {
	if mg.Verbose() {
		fmt.Println("Checking for vulnerabilities")
	}
	yarn, err := yarn()
	if err != nil {
		return err
	}
	return yarn("audit")
}
