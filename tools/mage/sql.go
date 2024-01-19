// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	sqlfmtImage = "ghcr.io/tconbeer/sqlfmt:v0.21.0"
	lineLength  = "120"
)

// Sql namespace
type SQL mg.Namespace

// Fmt formats all .sql files
func (SQL) Fmt(context.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	err = sh.Run("docker", "run",
		"--rm",
		"-e", "SQLFMT_LINE_LENGTH="+lineLength,
		"-v", fmt.Sprintf("%s:/src", wd),
		sqlfmtImage)
	if err != nil {
		fmt.Println("returning here")
		return err
	}
	// Scan all files to replace '-- bun:split' with '--bun:split
	err = walkDir(wd, fixBunSplit)
	if err != nil {
		return err
	}
	return nil
}

func walkDir(root string, apply func(path string, filename string) error) error {
	files, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = walkDir(root+"/"+file.Name(), apply)
			if err != nil {
				return err
			}
		}
		err = apply(root, file.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func fixBunSplit(path string, filename string) error {
	if !strings.HasSuffix(filename, ".sql") {
		return nil
	}
	filePath := path + "/" + filename
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	file = []byte(strings.ReplaceAll(string(file), "-- bun:split", "--bun:split"))

	return os.WriteFile(filePath, file, 0o744)
}
