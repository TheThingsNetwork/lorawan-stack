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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"gopkg.in/yaml.v2"
)

// Headers namespace.
type Headers mg.Namespace

// HeaderRule in the header config file.
type HeaderRule struct {
	Include     []string `yaml:"include"`
	Exclude     []string `yaml:"exclude"`
	Header      string   `yaml:"header"`
	headerLines [][]byte
	Prefix      string `yaml:"prefix"`
}

func (r *HeaderRule) split() {
	r.headerLines = bytes.Split([]byte(strings.TrimSpace(r.Header)), []byte("\n"))
}

func (r *HeaderRule) match(path string) (match bool) {
	if len(r.Include) > 0 {
		for _, item := range r.Include {
			if strings.Contains(path, item) {
				match = true
			}
		}
	} else {
		match = true
	}
	for _, item := range r.Exclude {
		if strings.Contains(path, item) {
			return false
		}
	}
	return
}

// HeaderConfig is the format of the header configuration file.
type HeaderConfig struct {
	Rules []*HeaderRule `yaml:"rules"`
}

func (c *HeaderConfig) split() {
	for _, rule := range c.Rules {
		rule.split()
	}
}

func (c *HeaderConfig) get(filename string) (r *HeaderRule) {
	for _, rule := range c.Rules {
		if !rule.match(filename) {
			continue
		}
		if r == nil {
			r = &HeaderRule{}
		}
		if rule.Header != "" {
			r.Header, r.headerLines = rule.Header, rule.headerLines
		}
		if rule.Prefix != "" {
			r.Prefix = rule.Prefix
		}
	}
	return r
}

var (
	headerFile   string
	headerConfig HeaderConfig
)

func init() {
	headerFile = os.Getenv("HEADER_FILE")
	if headerFile == "" {
		headerFile = ".mage/header.yml"
	}
}

func (Headers) loadFile() error {
	headerBytes, err := ioutil.ReadFile(headerFile)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(headerBytes, &headerConfig); err != nil {
		return err
	}
	headerConfig.split()
	return nil
}

type checkErr struct {
	Path   string
	Reason string
}

func (err checkErr) Error() string {
	return fmt.Sprintf("%s %s", err.Path, err.Reason)
}

func (h Headers) check(path string) error {
	rule := headerConfig.get(path)
	if rule == nil {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for i, expected := range rule.headerLines {
		if !s.Scan() {
			return &checkErr{Path: path, Reason: "has less lines than expected header"}
		}
		line := s.Bytes()
		if i == 0 && bytes.Contains(line, []byte("generated")) {
			return nil // Skip generated files.
		}
		if !bytes.Equal(bytes.TrimPrefix(line, []byte(rule.Prefix)), expected) && !bytes.Equal(line, bytes.TrimSpace([]byte(rule.Prefix))) {
			return &checkErr{Path: path, Reason: fmt.Sprintf("did not contain expected header line: %s", string(expected))}
		}
	}
	if s.Scan() && len(s.Bytes()) != 0 {
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
		if selectedFiles != nil && !selectedFiles[path] {
			return nil
		}
		if checkErr := h.check(path); checkErr != nil {
			checkErrs = append(checkErrs, checkErr)
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

func init() {
	preCommitChecks = append(preCommitChecks, Headers.Check)
}
