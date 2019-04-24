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
	"os"
	"os/user"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

// DevDeps installs the javascript SDK development dependencies.
func (sdkJs SdkJs) devDeps() error {
	_, err := sdkJs.yarn()
	return err
}

// JsSDK namespace.
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

func (sdkJs SdkJs) docker() (func(args ...string) error, error) {
	return func(args ...string) error {
		return sh.Run("docker", args...)
	}, nil
}

// DevDeps installs the javascript development dependencies.
func (sdkJs SdkJs) DevDeps() error {
	return sdkJs.devDeps()
}

// Deps installs the javascript dependencies.
func (sdkJs SdkJs) Deps() error {
	if mg.Verbose() {
		fmt.Println("Installing JS dependencies")
	}
	yarn, err := sdkJs.yarn()
	if err != nil {
		return err
	}
	return yarn("add", "--no-progress")
}

// Build builds the source files and output into 'dist'.
func (sdkJs SdkJs) Build() error {
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
	sh.Rm("./sdk/js/dist")
}

// Protos generates the api.json for the JS SDK
func (sdkJs SdkJs) Protos() error {
	if mg.Verbose() {
		fmt.Println("Extracting api definitions from protos…")
	}

	docker, err := sdkJs.docker()
	if err != nil {
		return err
	}
	PWD, err := os.Getwd()
	if err != nil {
		return err
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	PWD_PARENT := filepath.Dir(PWD)

	PROTOC_DOCKER_IMAGE := "thethingsindustries/protoc:3.1.3"
	SDK_PROTOC_FLAGS := "--doc_opt=json,api.json --doc_out=" + PWD + "/sdk/js/generated"
	API_PROTO_FILES := PWD + "/api/*.proto"

	return docker(
		"run", "--user", u.Uid, "--rm",
		"--mount", "type=bind,src="+PWD+"/api,dst="+PWD+"/api",
		"--mount", "type=bind,src="+PWD+"/pkg/ttnpb,dst=/out/go.thethings.network/lorawan-stack/pkg/ttnpb",
		"--mount", "type=bind,src="+PWD+"/sdk/js,dst="+PWD+"/sdk/js",
		"-w", PWD, PROTOC_DOCKER_IMAGE, "-I"+PWD_PARENT, SDK_PROTOC_FLAGS, API_PROTO_FILES,
	)
}

// Definitions extracts the api-definition.json from the proto generated api.json
func (sdkJs SdkJs) Definitions() error {
	mg.Deps(sdkJs.Protos)
	changed, err := target.Path("./sdk/js/generated/api-definition.json", "./sdk/js/generated/api.json")
	if os.IsNotExist(err) || (err == nil && changed) {
		if mg.Verbose() {
			fmt.Println("Extracting api definitions from protos…")
		}
		yarn, err := sdkJs.yarn()
		if err != nil {
			return err
		}

		return yarn("run", "definitions")
	}
	return nil
}

// CleanProtos clears all generated proto files.
func (sdkJs SdkJs) CleanProtos() {
	sh.Rm("./sdk/js/generated/api.json")
	sh.Rm("./sdk/js/generated/api-definition.json")
}
