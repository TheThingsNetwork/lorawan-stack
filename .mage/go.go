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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Go namespace.
type Go mg.Namespace

var minGoVersion = "1.11.4"

var goModuleEnv = map[string]string{
	"GO111MODULE": "on",
}

func execGo(cmd string, args ...string) error {
	_, err := sh.Exec(goModuleEnv, os.Stdout, os.Stderr, "go", append([]string{cmd}, args...)...)
	return err
}

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

var goPackageDirs []string

func (Go) packageDirs() (packageDirs []string, err error) {
	if goPackageDirs != nil {
		return goPackageDirs, nil
	}
	defer func() {
		goPackageDirs = packageDirs
	}()

	dirs, err := sh.OutputWith(goModuleEnv, "go", "list", "-f", "{{.Dir}}", "./...")
	if err != nil {
		return nil, err
	}
	all := strings.Split(strings.TrimSpace(dirs), "\n")
	if selectedDirs == nil {
		return all, nil
	}
	selected := make([]string, 0, len(all))
	for _, dir := range all {
		if selectedDirs[dir] {
			selected = append(selected, dir)
		}
	}
	return selected, nil
}

// Fmt formats and simplifies all Go files.
func (g Go) Fmt() error {
	dirs, err := g.packageDirs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return nil
	}
	return sh.RunCmd("gofmt", "-w", "-s")(dirs...)
}

// Lint lints all Go files.
func (g Go) Lint() error {
	dirs, err := g.packageDirs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return nil
	}
	return execGo("run", append([]string{"github.com/mgechev/revive", "-config=.revive.toml", "-formatter=stylish"}, dirs...)...)
}

// Misspell fixes common spelling mistakes in Go files.
func (g Go) Misspell() error {
	dirs, err := g.packageDirs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return nil
	}
	return execGo("run", append([]string{"github.com/client9/misspell/cmd/misspell", "-w"}, dirs...)...)
}

// Unconvert removes unnecessary type conversions from Go files.
func (g Go) Unconvert() error {
	dirs, err := g.packageDirs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return nil
	}
	return execGo("run", append([]string{"github.com/mdempsky/unconvert", "-safe", "-apply"}, dirs...)...)
}

// Quality runs code quality checks on Go files.
func (g Go) Quality() {
	mg.Deps(g.Fmt, g.Misspell, g.Unconvert)
	g.Lint() // Errors are allowed.
}

func init() {
	preCommitChecks = append(preCommitChecks, Go.Quality)
}

// Test tests all Go packages.
func (Go) Test() error {
	return execGo("test", "./...")
}

const goCoverageFile = "coverage.out"

// Cover tests all Go packages and writes test coverage into the coverage file.
func (Go) Cover() error {
	return execGo("test", "-cover", "-covermode=atomic", "-coverprofile="+goCoverageFile, "-timeout=5m", "./...")
}

var coverallsIgnored = []string{".pb.go:", ".pb.gw.go:", ".fm.go:"}

// Coveralls sends the test coverage to Coveralls.
func (Go) Coveralls() error {
	mg.Deps(Go.Cover)
	inFile, err := os.Open(goCoverageFile)
	if err != nil {
		return err
	}
	outFile, err := os.OpenFile("coveralls_"+goCoverageFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		outFile.Close()
		os.Remove("coveralls_" + goCoverageFile)
	}()
	s := bufio.NewScanner(inFile)
nextLine:
	for s.Scan() {
		line := s.Text()
		for _, suffix := range coverallsIgnored {
			if strings.Contains(line, suffix) {
				continue nextLine
			}
		}
		if _, err = fmt.Fprintln(outFile, line); err != nil {
			return err
		}
	}
	if err = outFile.Close(); err != nil {
		return err
	}
	service := os.Getenv("COVERALLS_SERVICE")
	if service == "" {
		service = "travis-ci"
	}
	return execGo("run", "github.com/mattn/goveralls", "-coverprofile=coveralls_"+goCoverageFile, "-service="+service, "-repotoken="+os.Getenv("COVERALLS_TOKEN"))
}
