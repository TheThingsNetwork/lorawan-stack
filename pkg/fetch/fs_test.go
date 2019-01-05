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

package fetch_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type frequencyPlansFileSystem string

func createMockFileSystem() (frequencyPlansFileSystem, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	return frequencyPlansFileSystem(dir), nil
}

func (fs frequencyPlansFileSystem) Destroy() error {
	return os.RemoveAll(string(fs))
}

func (fs frequencyPlansFileSystem) Dir() string {
	return string(fs)
}

func TestFilesystem(t *testing.T) {
	a := assertions.New(t)
	filename := "file"
	content := []byte("Hello world")

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	// Creating working file
	{
		f, err := os.Create(filepath.Join(fs.Dir(), filename))
		a.So(err, should.BeNil)

		_, err = f.Write(content)
		a.So(err, should.BeNil)
		err = f.Close()
		a.So(err, should.BeNil)
	}

	fetcher := fetch.FromFilesystem(fs.Dir())

	// Reading working file
	{
		fileContent, err := fetcher.File(filename)
		a.So(err, should.BeNil)
		a.So(string(fileContent), should.Equal, string(content))
	}

	// Reading non-working file
	{
		_, err = fetcher.File("non-existing file")
		a.So(err, should.NotBeNil)
	}
}
