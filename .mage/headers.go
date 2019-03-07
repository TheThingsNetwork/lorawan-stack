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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
)

// Headers namespace.
type Headers mg.Namespace

var fileCommentPrefix = map[string]string{
	".go":    "// ",
	".js":    "// ",
	".make":  "# ",
	".styl":  "// ",
	".proto": "// ",
}

type fileFilter struct {
	exclude []string
	include []string
}

func (f fileFilter) match(search string) (match bool) {
	if len(f.include) > 0 {
		for _, item := range f.include {
			if strings.Contains(search, item) {
				match = true
			}
		}
	} else {
		match = true
	}
	for _, item := range f.exclude {
		if strings.Contains(search, item) {
			return false
		}
	}
	return
}

var (
	headerFilter fileFilter
	headerFile   string
	headerLines  []string
)

func init() {
	headerFilter = fileFilter{
		exclude: strings.Fields(os.Getenv("HEADER_EXCLUDE")),
		include: strings.Fields(os.Getenv("HEADER_INCLUDE")),
	}
	headerFile = os.Getenv("HEADER_FILE")
	if headerFile == "" {
		headerFile = ".mage/header.txt"
	}
}

func (Headers) loadFile() error {
	headerBytes, err := ioutil.ReadFile(headerFile)
	if err != nil {
		return err
	}
	headerLines = strings.Split(strings.TrimSpace(string(headerBytes)), "\n")
	return nil
}

type checkErr struct {
	Path   string
	Reason string
}

func (err checkErr) Error() string {
	return fmt.Sprintf("%s %s", err.Path, err.Reason)
}

func (h Headers) check(path, commentPrefix string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for i, expected := range headerLines {
		expected = strings.TrimSpace(commentPrefix + expected)
		if !s.Scan() {
			return &checkErr{Path: path, Reason: "has less lines than expected header"}
		}
		line := s.Text()
		if i == 0 && strings.Contains(line, "generated") {
			return nil // Skip generated files.
		}
		if s.Text() != expected {
			return &checkErr{Path: path, Reason: fmt.Sprintf("did not contain expected line: %s", expected)}
		}
	}
	if s.Scan() && s.Text() != "" {
		return &checkErr{Path: path, Reason: "does not have empty line after header"}
	}
	return nil
}

type errorSlice []error

func (errs errorSlice) Error() string {
	switch len(errs) {
	case 0:
		return ""
	case 1:
		return errs[0].Error()
	default:
		var b strings.Builder
		b.WriteString("multiple errors:\n")
		for _, err := range errs {
			b.WriteString("  - " + err.Error() + "\n")
		}
		return b.String()
	}
}

// Check checks that all files contain the required file header.
func (h Headers) Check() error {
	mg.Deps(Headers.loadFile)
	var checkErrs errorSlice
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			switch path {
			case ".cache", ".dev", ".env", ".git", "dist", "node_modules", "public", "sdk/js/dist", "sdk/js/node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !headerFilter.match(path) {
			return nil
		}
		if prefix, ok := fileCommentPrefix[filepath.Ext(path)]; ok {
			if checkErr := h.check(path, prefix); checkErr != nil {
				checkErrs = append(checkErrs, checkErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(checkErrs) > 0 {
		return checkErrs
	}
	return nil
}
