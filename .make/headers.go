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

//+build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var help = `headers.go can be used to control, fix or remove header comments in source code files.

Possibles usages:

1. Using command-line arguments:

    $ go run headers.go {check,fix,remove} file1.go file2.go Makefile .make/build.make [...]

2. Using environment variables:

    $ FILES="file1.go file2.go Makefile" go run headers.go {check,fix,remove}

Optional environment variables:

- HEADER_FILE: File to read the license from. Default: .make/header.txt.
- RULES_FILE: Override the license file to use for files matching a certain regular expression.
  RULES_FILE must point to a YAML file, that matches the following format:

  - match: ".*ttn.*" # Regular expression
    headers-file: .make/APACHE_2_LICENSE # Path to the license to use
  - match: "^api"
    headers-file: .make/MIT_LICENSE

  Files that do not match any of the regular expressions will use the HEADER_FILE license.`

var (
	makeRegex      = regexp.MustCompile(".*\\.make$")
	makefileRegex  = regexp.MustCompile(".*Makefile$")
	shRegex        = regexp.MustCompile(".*\\.sh$")
	generatedRegex = regexp.MustCompile("generated")
)

func prefixFunction(filename string) func(string) string {
	byteFilename := []byte(filename)
	commentPrefix := "//"
	if makeRegex.Match(byteFilename) || makefileRegex.Match(byteFilename) || shRegex.Match(byteFilename) {
		commentPrefix = "#"
	}
	return func(line string) string {
		if line == "" || line == "\n" {
			return fmt.Sprintf("%s%s", commentPrefix, line)
		}
		return fmt.Sprintf("%s %s", commentPrefix, line)
	}
}

func hasHeaders(licenseContent []byte, filename string) (valid, generated bool, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, false, err
	}
	defer file.Close()

	withPrefix := prefixFunction(filename)

	checkedFileReader := bufio.NewReader(file)
	licenseReader := bufio.NewReader(bytes.NewBuffer(licenseContent))
	for {
		licenseLine, err := licenseReader.ReadString('\n')
		if err != nil && err != io.EOF {
			return false, false, fmt.Errorf("could not read license content from bytes buffer (%s)", err)
		} else if err == io.EOF {
			return true, false, nil
		}
		checkedFileLine, err := checkedFileReader.ReadString('\n')
		if err != nil && err != io.EOF {
			return false, false, fmt.Errorf("could not read file (%s)", err)
		}

		expected := withPrefix(licenseLine)
		if checkedFileLine != expected {
			if generatedRegex.Match([]byte(checkedFileLine)) {
				return false, true, nil
			}
			return false, false, nil
		}
	}
}

func addHeader(licenseContent []byte, filename string) error {
	tempFilename := fmt.Sprintf("%s-fixed-headers", filename)
	newFile, err := os.Create(tempFilename)
	if err != nil {
		return err
	}

	withPrefix := prefixFunction(filename)

	licenseReader := bufio.NewReader(bytes.NewBuffer(licenseContent))
	for {
		licenseLine, err := licenseReader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("could not read license content from bytes buffer (%s)", err)
		} else if err == io.EOF {
			break
		}

		if _, err := newFile.WriteString(withPrefix(licenseLine)); err != nil {
			return err
		}
	}

	if _, err := newFile.Write([]byte("\n")); err != nil {
		return err
	}

	originalFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	if _, err := io.Copy(newFile, originalFile); err != nil {
		return err
	}

	if err := originalFile.Close(); err != nil {
		return err
	}

	if err := newFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempFilename, filename)
}

func nbLines(str []byte) int {
	nb := 1
	for _, char := range str {
		if char == '\n' {
			nb++
		}
	}
	return nb
}

func removeHeaders(nbLines int, filename string) error {
	tempFilename := fmt.Sprintf("%s-without-headers", filename)
	newFile, err := os.Create(tempFilename)
	if err != nil {
		return err
	}

	originalFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	originalFileReader := bufio.NewReader(originalFile)
	for line := 0; line < nbLines; line++ {
		_, err := originalFileReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("could not read original file content (%s)", err)
		}
	}

	if _, err := io.Copy(newFile, originalFileReader); err != nil {
		return err
	}

	if err := originalFile.Close(); err != nil {
		return err
	}

	if err := newFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempFilename, filename)
}

type headersOperation struct {
	licenseContent []byte
	filenames      []string
}

func (o headersOperation) check() bool {
	allFilesValid := true
	for _, file := range o.filenames {
		if valid, generated, err := hasHeaders(o.licenseContent, file); err != nil {
			log.Printf("Could not check headers in %s: %s\n", file, err)
			allFilesValid = false
		} else if !valid && !generated {
			log.Printf("Invalid headers in %s.\n", file)
			allFilesValid = false
		}
	}
	return allFilesValid
}

func (o headersOperation) remove() bool {
	var wasError error
	for _, file := range o.filenames {
		if valid, generated, err := hasHeaders(o.licenseContent, file); err != nil {
			log.Printf("Could not check headers in %s: %s\n", file, err)
			wasError = err
		} else if !generated {
			if !valid {
				log.Printf("No headers in %s.\n", file)
			} else {
				if err := removeHeaders(nbLines(o.licenseContent), file); err != nil {
					log.Printf("Could not remove headers in %s: %s\n", file, err)
					wasError = err
				}
			}
		}
	}
	return wasError == nil
}

func (o headersOperation) fix() bool {
	var wasError error
	for _, file := range o.filenames {
		if valid, generated, err := hasHeaders(o.licenseContent, file); err != nil {
			log.Printf("Could not remove headers in %s: %s\n", file, err)
			wasError = err
		} else if !valid && !generated {
			if err := addHeader(o.licenseContent, file); err != nil {
				log.Printf("Could not fix %s: %s\n", file, err)
			} else {
				log.Printf("Fixed headers in %s.\n", file)
			}
		}
	}
	return wasError == nil
}

func executeOperation(command, licenseFilePath string, files []string) (success bool) {
	licenseContent, err := ioutil.ReadFile(licenseFilePath)
	if err != nil {
		log.Fatalf("Could not read license content in %s: %s\n", licenseFilePath, err)
	}

	operation := headersOperation{
		filenames:      files,
		licenseContent: licenseContent,
	}

	switch command {
	case "remove":
		success = operation.remove()
	case "fix":
		success = operation.fix()
	case "check":
		success = operation.check()
	default:
		log.Printf("Unknown command %s.\n", command)
	}
	return
}

type ruleParameters struct {
	Match       string `yaml:"match"`
	HeadersFile string `yaml:"headers-file"`

	matchedFiles []string `yaml:",omitempty"`
}

func processRulesList(rulesFilePath, defaultLicenseFilePath string, files []string) ([]*ruleParameters, error) {
	rulesBytes, err := ioutil.ReadFile(rulesFilePath)
	if err != nil {
		return nil, fmt.Errorf("Could not read rules file specified in $RULES_FILE: %s", err)
	}

	rules := []*ruleParameters{}
	if err := yaml.Unmarshal(rulesBytes, &rules); err != nil {
		log.Fatalf("Could not unmarshal %s: %s\n", rulesFilePath, err)
	}

	for _, rule := range rules {
		if rule.Match == "" {
			return nil, fmt.Errorf("One of the rules has no `match` regular expression")
		}
		regex, err := regexp.Compile(rule.Match)
		if err != nil {
			return nil, fmt.Errorf("Could not compile %s to regular expression: %s", rule.Match, err)
		}
		var unprocessedFiles []string
		for _, file := range files {
			if regex.MatchString(file) {
				rule.matchedFiles = append(rule.matchedFiles, file)
			} else {
				unprocessedFiles = append(unprocessedFiles, file)
			}
		}
		files = unprocessedFiles
	}

	// Add rule for default license
	rules = append(rules, &ruleParameters{HeadersFile: defaultLicenseFilePath, matchedFiles: files})

	return rules, nil
}

func main() {
	files := []string{}
	if filenames := os.Getenv("FILES"); filenames != "" {
		files = strings.Split(filenames, "\n")
	}
	if len(os.Args) <= 1 {
		fmt.Println(help)
		os.Exit(1)
	}

	command := os.Args[1]
	if len(files) == 0 && len(os.Args) >= 3 {
		files = os.Args[2:]
	}

	licenseFilePath := os.Getenv("HEADER_FILE")
	if licenseFilePath == "" {
		licenseFilePath = ".make/header.txt"
	}

	successful := true
	var rules []*ruleParameters

	rulesFilePath := os.Getenv("RULES_FILE")
	if rulesFilePath != "" {
		var err error
		rules, err = processRulesList(rulesFilePath, licenseFilePath, files)
		if err != nil {
			log.Fatalf("Could not process rules file: %s\n", err)
		}
	} else {
		rules = []*ruleParameters{{
			HeadersFile:  licenseFilePath,
			matchedFiles: files,
		}}
	}

	for _, rule := range rules {
		successful = successful && executeOperation(command, rule.HeadersFile, rule.matchedFiles)
	}
	if !successful {
		os.Exit(1)
	}
}
