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
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/TheThingsIndustries/magepkg/git"
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

var goModule = "go.thethings.network/lorawan-stack"

func goBuildVersionFlags() (string, error) {
	commit, branch, tag, err := git.Info()
	if err != nil {
		return "", err
	}
	var flags []string
	for k, v := range map[string]string{
		"GitCommit": commit,
		"GitBranch": branch,
		"TTN":       tag,
		"BuildDate": now.Format("2006-01-02T15:04:05Z"),
	} {
		flags = append(flags, fmt.Sprintf("-X %s/pkg/version.%s=%s", goModule, k, v))
	}
	return strings.Join(flags, " "), nil
}

func goTags() string {
	tagMap := map[string]bool{
		"mage": true,
	}
	if env := os.Getenv("GO_TAGS"); env != "" {
		for _, tag := range strings.Split(env, ",") {
			tagMap[tag] = true
		}
	}
	var tags []string
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	return strings.Join(tags, ",")
}

type goBuildConfig struct {
	GOOS   string
	GOARCH string
	GOARM  string
}

var goBuildConfigs = []goBuildConfig{
	{GOOS: "linux", GOARCH: "386"},
	{GOOS: "linux", GOARCH: "amd64"},
	{GOOS: "linux", GOARCH: "arm", GOARM: "6"},
	{GOOS: "linux", GOARCH: "arm", GOARM: "7"},
	{GOOS: "linux", GOARCH: "arm64"},
	{GOOS: "windows", GOARCH: "386"},
	{GOOS: "windows", GOARCH: "amd64"},
	{GOOS: "darwin", GOARCH: "amd64"},
}

var releaseDir = "dist"

func init() {
	if releaseDirEnv := os.Getenv("RELEASE_DIR"); releaseDirEnv != "" {
		releaseDir = releaseDirEnv
	}
}

func goBuild(ctx context.Context, binary string, config goBuildConfig) error {
	goEnv := make(map[string]string)
	goEnv["CGO_ENABLED"] = "0"
	goEnv["GOOS"] = config.GOOS
	goEnv["GOARCH"] = config.GOARCH
	goEnv["GOARM"] = config.GOARM
	env, err := sh.OutputWith(goEnv, "go", "env", "-json")
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(env), &goEnv)
	if err != nil {
		return err
	}
	versionFlags, err := goBuildVersionFlags()
	if err != nil {
		return err
	}
	fmt.Printf("Building ttn-lw-%s-%s-%s%s%s...\n", binary, goEnv["GOOS"], goEnv["GOARCH"], goEnv["GOARM"], goEnv["GOEXE"])
	return sh.RunWith(goEnv,
		"go", "build",
		"-v",
		"-tags", goTags(),
		"-ldflags", "-w -s "+versionFlags, // NOTE: this needs to remain a single string.
		"-o", fmt.Sprintf("%s/ttn-lw-%s-$GOOS-$GOARCH$GOARM$GOEXE", releaseDir, binary),
		fmt.Sprintf("./cmd/ttn-lw-%s", binary),
	)
}

var goBinaries = []string{
	"stack",
	"cli",
}

func init() {
	if goBinariesEnv := os.Getenv("GO_BINARIES"); goBinariesEnv != "" {
		goBinaries = strings.Split(goBinariesEnv, ",")
	}
}

// Build builds all Go binaries.
func (Go) Build(ctx context.Context) (err error) {
	for _, binary := range goBinaries {
		err = goBuild(ctx, binary, goBuildConfig{
			GOOS:   os.Getenv("GOOS"),
			GOARCH: os.Getenv("GOARCH"),
			GOARM:  os.Getenv("GOARM"),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildCrossPlatform builds all Go binaries for all configured platforms.
func (Go) BuildCrossPlatform(ctx context.Context) (err error) {
	for _, config := range goBuildConfigs {
		err = goBuild(ctx, "stack", config)
		if err != nil {
			return err
		}
		for _, binary := range goBinaries {
			err = goBuild(ctx, binary, config)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
