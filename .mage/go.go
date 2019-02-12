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
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Go namespace.
type Go mg.Namespace

var minGoVersion = "1.11.4"

// CheckVersion checks the installed Go version against the minimum version we support.
func (Go) CheckVersion() error {
	versionStr, err := sh.Output("go", "version")
	if err != nil {
		return err
	}
	version := strings.Split(strings.TrimPrefix(strings.Fields(versionStr)[2], "go"), ".")
	major, _ := strconv.Atoi(version[0])
	minor, _ := strconv.Atoi(version[1])
	var patch int
	if len(version) > 2 {
		patch, _ = strconv.Atoi(version[2])
	}
	current := semver.Version{Major: uint64(major), Minor: uint64(minor), Patch: uint64(patch)}
	min, _ := semver.Parse(minGoVersion)
	if current.LT(min) {
		return fmt.Errorf("Your version of Go (%s) is not supported. Please install Go %s or later", versionStr, minGoVersion)
	}
	return nil
}
