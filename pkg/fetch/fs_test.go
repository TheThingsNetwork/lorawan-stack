// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/fetch"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
