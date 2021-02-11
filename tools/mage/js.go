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
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

// Js namespace.
type Js mg.Namespace

var (
	devPort  = 8080
	prodPort = 1885
)

func yarnWorkingDirectoryArg(elem ...string) string {
	return fmt.Sprintf("--cwd=%s", filepath.Join(elem...))
}

func installYarn() error {
	ok, err := target.Path(
		filepath.Join("node_modules", "yarn"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if err := sh.RunV("npm", "install", "--no-package-lock", "--no-save", "--production=false", "yarn"); err != nil {
		return fmt.Errorf("failed to install yarn: %w", err)
	}
	return nil
}

func execYarn(stdout, stderr io.Writer, args ...string) error {
	_, err := sh.Exec(nil, stdout, stderr, "npx", append([]string{"yarn"}, args...)...)
	return err
}

func runYarn(args ...string) error {
	return sh.Run("npx", append([]string{"yarn"}, args...)...)
}

func runYarnV(args ...string) error {
	return sh.RunV("npx", append([]string{"yarn"}, args...)...)
}

func (Js) runYarnCommand(cmd string, args ...string) error {
	return runYarn(append([]string{"run", cmd}, args...)...)
}

func (Js) runYarnCommandV(cmd string, args ...string) error {
	return runYarnV(append([]string{"run", cmd}, args...)...)
}

func (js Js) runWebpack(config string, args ...string) error {
	return js.runYarnCommand("webpack", append([]string{fmt.Sprintf("--config=%s", config)}, args...)...)
}

func (js Js) runEslint(args ...string) error {
	return js.runYarnCommand("eslint", append([]string{"--color", "--no-ignore"}, args...)...)
}

func (js Js) waitOn() error {
	u, err := url.Parse(js.frontendURL())
	if err != nil {
		return err
	}
	return js.runYarnCommand("wait-on", []string{
		fmt.Sprintf("--timeout=%d", 120000),
		fmt.Sprintf("--interval=%d", 1000),
		fmt.Sprintf("http-get://%s/oauth", u.Host),
	}...)
}

func (js Js) runCypress(command string, args ...string) error {
	mg.Deps(js.waitOn)
	return js.runYarnCommand("cypress", append([]string{
		command,
		"--config-file", filepath.Join("config", "cypress.json"),
		"--config", fmt.Sprintf("baseUrl=%s", js.frontendURL())},
		args...)...)
}

func (js Js) frontendURL() string {
	if js.isProductionMode() {
		return fmt.Sprintf("http://localhost:%d", prodPort)
	}
	return fmt.Sprintf("http://localhost:%d", devPort)
}

func (Js) isProductionMode() bool {
	switch v := os.Getenv("NODE_ENV"); v {
	case "", "production":
		return true

	case "development":
		return false

	default:
		if mg.Verbose() {
			fmt.Printf("Unknown `NODE_ENV` value `%s`, assuming production mode\n", v)
		}
		return true
	}
}

func (js Js) deps() error {
	if mg.Verbose() {
		fmt.Println("Installing JS dependencies")
	}
	return runYarn("install", "--no-progress", "--production=false", "--check-files")
}

// Deps installs the javascript dependencies.
func (js Js) Deps() error {
	ok, err := target.Dir(
		"node_modules",
		"package.json",
		"yarn.lock",
		filepath.Join("sdk", "js", "src"),
		filepath.Join("sdk", "js", "generated"),
	)
	if err != nil {
		return targetError(err)
	}
	// installYarn updates modtime of node_modules, so we not only need to check that, but also the contents of node_modules.
	// NOTE: Getting rid of installYarn and installing both yarn and the dependencies here does not work, since JsSDK.Build
	// depends on yarn being available.
	if !ok {
		files, err := ioutil.ReadDir("node_modules")
		if err != nil {
			return fmt.Errorf("failed to read node_modules: %w", err)
		}
		if len(files) > 2 ||
			js.isProductionMode() && len(files) > 1 {
			// Check if it's only yarn and, in development mode, ttn-lw link installed in `node_modules`.
			// Additionally, check whether the SDK was installed correctly.
			// NOTE: There's no link in production mode.
			if _, err := os.Stat(filepath.Join("node_modules", "ttn-lw", "dist")); !os.IsNotExist(err) {
				return nil
			}
		}
	}

	mg.Deps(installYarn, JsSDK.Build)
	if !js.isProductionMode() {
		if mg.Verbose() {
			fmt.Println("Linking ttn-lw package")
		}
		if err := runYarn(yarnWorkingDirectoryArg("sdk", "js"), "link"); err != nil {
			return fmt.Errorf("failed to create JS SDK link: %w", err)
		}
		if err := runYarn("link", "ttn-lw"); err != nil {
			return fmt.Errorf("failed to link JS SDK: %w", err)
		}
	}
	return js.deps()
}

// BuildDll runs the webpack command to build the DLL bundle
func (js Js) BuildDll() error {
	if js.isProductionMode() {
		fmt.Println("Skipping DLL building (production mode)")
		return nil
	}

	ok, err := target.Path(
		filepath.Join("public", "libs.bundle.js"),
		"yarn.lock",
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	mg.Deps(js.Deps)
	if mg.Verbose() {
		fmt.Println("Running Webpack for DLL")
	}
	return js.runWebpack("config/webpack.dll.babel.js")
}

// Build runs the webpack command with the project config.
func (js Js) Build() error {
	mg.Deps(js.Deps, js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack")
	}
	return js.runWebpack("config/webpack.config.babel.js")
}

// Serve runs webpack-dev-server.
func (js Js) Serve() error {
	mg.Deps(js.Deps, js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack for Main Bundle in watch mode")
	}
	os.Setenv("DEV_SERVER_BUILD", "true")
	return js.runYarnCommandV("webpack-dev-server",
		"--config", "config/webpack.config.babel.js",
		"-w",
	)
}

// Messages extracts the frontend messages via babel.
func (js Js) Messages() error {
	mg.Deps(js.Deps)
	ok, err := target.Dir(
		filepath.Join(".cache", "messages"),
		filepath.Join("pkg", "webui", "console"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Extracting frontend messages")
	}
	if err = sh.Rm(filepath.Join(".cache", "messages")); err != nil {
		return fmt.Errorf("failed to delete existing messages: %w", err)
	}
	if err = os.MkdirAll(filepath.Join("pkg", "webui", "locales"), 0755); err != nil {
		return fmt.Errorf("failed to create locale directory: %w", err)
	}
	return execYarn(nil, os.Stderr, "babel", filepath.Join("pkg", "webui"))
}

// Translations builds the frontend locale files.
func (js Js) Translations() error {
	mg.Deps(js.Deps, js.Messages)
	ok, err := target.Dir(
		filepath.Join("pkg", "webui", "locales", "en.json"),
		filepath.Join(".cache", "messages"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Building frontend locale files")
	}
	return sh.Run("node", "tools/mage/translations.js")
}

// BackendTranslations builds the backend locale files.
func (js Js) BackendTranslations() error {
	mg.Deps(js.Deps)
	ok, err := target.Path(
		filepath.Join("pkg", "webui", "locales", ".backend", "en.json"),
		filepath.Join("config", "messages.json"),
	)
	if err != nil {
		return targetError(err)
	}
	if !ok {
		return nil
	}
	if mg.Verbose() {
		fmt.Println("Building backend locale files")
	}
	return sh.Run("node",
		"tools/mage/translations.js",
		"--backend-messages", "config/messages.json",
		"--locales", "pkg/webui/locales/.backend",
		"--backend-only",
	)
}

// Clean clears all generated files.
func (js Js) Clean() error {
	for _, p := range []string{
		".cache",
		"public",
		filepath.Join("pkg", "webui", "locales", ".backend"),
	} {
		if err := sh.Rm(p); err != nil {
			return fmt.Errorf("failed to delete %s: %w", p, err)
		}
	}
	return nil
}

// CleanDeps removes all installed node packages (rm -rf node_modules).
func (js Js) CleanDeps() error {
	if err := sh.Rm("node_modules"); err != nil {
		return fmt.Errorf("failed to delete node_modules: %w", err)
	}
	return nil
}

// Test runs frontend jest tests.
func (js Js) Test() error {
	mg.Deps(js.Deps)
	if mg.Verbose() {
		fmt.Println("Running tests")
	}
	return js.runYarnCommand("jest", filepath.Join("pkg", "webui"))
}

// Fmt formats all js files.
func (js Js) Fmt() error {
	mg.Deps(js.Deps)
	if mg.Verbose() {
		fmt.Println("Running prettier on .js files")
	}
	return js.runYarnCommand("prettier",
		"--config", "./config/.prettierrc.js",
		"--write",
		"./pkg/webui/**/*.js", "./config/**/*.js",
	)
}

// Lint runs eslint over frontend js files.
func (js Js) Lint() error {
	mg.Deps(js.Deps, Js.BackendTranslations)
	if mg.Verbose() {
		fmt.Println("Running eslint on .js files")
	}
	return js.runEslint("./pkg/webui/**/*.js", "./config/**/*.js")
}

// LintSnap runs eslint over frontend snap files.
func (js Js) LintSnap() error {
	mg.Deps(js.Deps)
	if mg.Verbose() {
		fmt.Println("Running eslint on .snap files")
	}
	return js.runEslint("./pkg/webui/**/*.snap")
}

// LintAll runs linters over js and snap files.
func (js Js) LintAll() {
	mg.Deps(js.Lint, js.LintSnap)
}

// Storybook runs a local server with storybook.
func (js Js) Storybook() error {
	mg.Deps(js.Deps)
	if mg.Verbose() {
		fmt.Println("Serving storybook")
	}
	return js.runYarnCommandV("start-storybook",
		"--config-dir", "./config/storybook",
		"--static-dir", "public",
		"--port", "9001",
	)
}

// Vulnerabilities runs yarn audit to check for vulnerable node packages.
func (js Js) Vulnerabilities() error {
	mg.Deps(installYarn)
	if mg.Verbose() {
		fmt.Println("Checking for vulnerabilities")
	}
	return runYarn("audit")
}

// CypressHeadless runs the Cypress end-to-end tests in the headless mode.
func (js Js) CypressHeadless() error {
	mg.Deps(Js.Deps)
	if mg.Verbose() {
		fmt.Println("Running Cypress E2E tests in headless mode")
	}
	return js.runCypress("run")
}

// CypressInteractive runs the Cypress end-to-end tests in interactive mode.
func (js Js) CypressInteractive() error {
	mg.Deps(Js.Deps)
	if mg.Verbose() {
		fmt.Println("Running Cypress E2E tests in interactive mode")
	}
	return js.runCypress("open")
}
