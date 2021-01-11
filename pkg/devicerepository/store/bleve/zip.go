// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package bleve

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"path"
)

// unarchive extracts a zip archive (passed as a byte array) at a destination directory.
// The directory will be created if it does not already exist. A filter can be used
// to decide whether each file should be extracted, based on its path in the archive.
func unarchive(b []byte, destinationDirectory string, filter func(path string) (string, bool)) error {
	rd := bytes.NewReader(b)
	archive, err := zip.NewReader(rd, rd.Size())
	if err != nil {
		return err
	}

	for _, file := range archive.File {
		fileName, skip := filter(file.Name)
		if skip {
			continue
		}
		destination := path.Join(destinationDirectory, fileName)
		if err := os.MkdirAll(path.Dir(destination), 0755); err != nil {
			return err
		}
		r, err := file.Open()
		if err != nil {
			return err
		}
		defer r.Close()
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(destinationDirectory, fileName), b, file.Mode()); err != nil {
			return err
		}
	}
	return nil
}
