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
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/magefile/mage/mg"
)

// Go namespace.
type Go mg.Namespace

var minGoVersion = "1.21.0"

var goTags = os.Getenv("GO_TAGS")

const (
	gofumpt      = "mvdan.cc/gofumpt@v0.6.0"
	golangciLint = "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2"
	goveralls    = "github.com/mattn/goveralls@v0.0.12"
	bufCLI       = "github.com/bufbuild/buf/cmd/buf@v1.31.0"
)

func buildGoArgs(cmd string, args ...string) []string {
	if goTags != "" {
		args = append([]string{fmt.Sprintf("-tags=%s", goTags)}, args...)
	}
	return append([]string{cmd}, args...)
}

func execGoFrom(dir string, stdout, stderr io.Writer, cmd string, args ...string) error {
	return execFrom(dir, nil, stdout, stderr, "go", buildGoArgs(cmd, args...)...)
}

func execGo(stdout, stderr io.Writer, cmd string, args ...string) error {
	return execGoFrom("", stdout, stderr, cmd, args...)
}

func runGo(args ...string) error {
	return execGo(os.Stdout, os.Stderr, "run", args...)
}

func runGoFrom(dir string, args ...string) error {
	return execGoFrom(dir, os.Stdout, os.Stderr, "run", args...)
}

func writeToFile(filename string, value []byte) error {
	return os.WriteFile(filename, value, 0o644)
}

func outputGo(cmd string, args ...string) (string, error) {
	var buf bytes.Buffer
	if err := execGo(&buf, os.Stderr, cmd, args...); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func outputJSONGo(cmd string, args ...string) ([]byte, error) {
	var buf bytes.Buffer
	if err := execGo(&buf, os.Stderr, cmd, args...); err != nil {
		return nil, err
	}
	raw := buf.String()
	jsonStartIdx := strings.Index(raw, "{")
	if jsonStartIdx == -1 {
		return nil, fmt.Errorf("No JSON found in output")
	}
	start := raw[jsonStartIdx:]
	jsonEndIdx := strings.Index(start, "}")
	if jsonEndIdx == -1 {
		return nil, fmt.Errorf("No JSON found in output")
	}
	return []byte(start[:jsonEndIdx+1]), nil
}

func runGoTool(args ...string) error {
	return runGoFrom("tools", append([]string{"-exec", "go run exec_from.go -dir .."}, args...)...)
}

// CheckVersion checks the installed Go version against the minimum version we support.
func (Go) CheckVersion() error {
	if mg.Verbose() {
		fmt.Println("Checking Go version")
	}
	versionStr, err := outputGo("version")
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
		return fmt.Errorf("Your version of Go (%s) is not supported. Please install Go %s or later",
			versionStr, minGoVersion)
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

	dirs, err := outputGo("list", "-f", "{{.Dir}}", "./...")
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
	if mg.Verbose() {
		fmt.Printf("Formatting and simplifying %d Go packages\n", len(dirs))
	}
	return runGoTool(append([]string{gofumpt, "-w"}, dirs...)...)
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
	if mg.Verbose() {
		fmt.Printf("Linting %d Go packages\n", len(dirs))
	}
	return runGoTool(append([]string{golangciLint, "run"}, dirs...)...)
}

// Quality runs code quality checks on Go files.
func (g Go) Quality() {
	mg.Deps(g.Fmt)
	_ = g.Lint() // Errors are allowed.
}

func init() {
	preCommitChecks = append(preCommitChecks, Go.Quality)
}

func runGoTest(args ...string) error {
	return execGo(os.Stdout, os.Stderr, "test", append([]string{"-timeout=5m", "-failfast"}, args...)...)
}

// Test tests all Go packages.
func (Go) Test() error {
	if mg.Verbose() {
		fmt.Println("Testing all Go packages")
	}
	return runGoTest("./...")
}

var goBinaries = []string{"./cmd/ttn-lw-cli", "./cmd/ttn-lw-stack"}

// TestBinaries tests the Go binaries by executing them with the --help flag.
func (Go) TestBinaries() error {
	if mg.Verbose() {
		fmt.Println("Testing Go binaries")
	}
	for _, binary := range goBinaries {
		_, err := outputGo("run", binary, "config")
		if err != nil {
			return err
		}
	}
	return nil
}

const goCoverageFile = "coverage.out"

// Cover tests all Go packages and writes test coverage into the coverage file.
func (Go) Cover() error {
	if mg.Verbose() {
		fmt.Println("Testing all Go packages with coverage")
	}
	return runGoTest("-cover", "-covermode=atomic", "-coverprofile="+goCoverageFile, "./...")
}

var coverallsIgnored = []string{
	".fm.go:",
	".pb.go:",
	".pb.gw.go:",
	".pb.validate.go",
}

// Coveralls sends the test coverage to Coveralls.
func (g Go) Coveralls() error {
	mg.Deps(g.Cover)
	if mg.Verbose() {
		fmt.Println("Filtering Go coverage output")
	}
	inFile, err := os.Open(goCoverageFile)
	if err != nil {
		return err
	}
	outFile, err := os.OpenFile("coveralls_"+goCoverageFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
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
		service = "github"
	}
	if mg.Verbose() {
		fmt.Println("Sending Go coverage to Coveralls")
	}
	return runGoTool(goveralls, "-coverprofile=coveralls_"+goCoverageFile, "-service="+service)
}

// Generate runs go generate.
func (Go) Generate() error {
	return execGo(os.Stdout, os.Stderr, "generate", "./...")
}

// Messages builds the file with translatable messages in Go code.
func (Go) Messages() error {
	return runGoTool("generate_i18n.go")
}

// EventData builds the file with event data.
func (Go) EventData() error {
	return runGoTool("generate_event_data.go")
}
