// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package fs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type memory map[string]string
type memfile strings.Reader

func (m memory) Open(name string) (http.File, error) {
	str, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("File %s does not exist", name)
	}

	return (*memfile)(strings.NewReader(str)), nil
}

func (m *memfile) Read(b []byte) (int, error) {
	return (*strings.Reader)(m).Read(b)
}

func (m *memfile) Seek(offset int64, whence int) (int64, error) {
	return (*strings.Reader)(m).Seek(offset, whence)
}

func (m *memfile) Close() error {
	return nil
}

func (m *memfile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (m *memfile) Stat() (os.FileInfo, error) {
	return nil, nil
}

func TestCombined(t *testing.T) {
	a := assertions.New(t)

	first := memory{
		"foo": "first",
		"bar": "first",
	}

	second := memory{
		"bar": "second",
		"baz": "second",
	}

	combined := Combine(first, second)

	{
		file, err := combined.Open("foo")
		a.So(err, should.BeNil)

		content, err := ioutil.ReadAll(file)
		a.So(err, should.BeNil)

		a.So(string(content), should.Equal, first["foo"])
	}

	{
		file, err := combined.Open("bar")
		a.So(err, should.BeNil)

		content, err := ioutil.ReadAll(file)
		a.So(err, should.BeNil)

		a.So(string(content), should.Equal, first["bar"])
	}

	{
		file, err := combined.Open("baz")
		a.So(err, should.BeNil)

		content, err := ioutil.ReadAll(file)
		a.So(err, should.BeNil)

		a.So(string(content), should.Equal, second["baz"])
	}

	{
		_, err := combined.Open("quu")
		a.So(err, should.NotBeNil)
	}
}

func TestSubDir(t *testing.T) {
	a := assertions.New(t)

	fs := memory{
		"foo/bar/baz": "content",
	}

	{
		sub := Subdirectory(fs, "foo")
		file, err := sub.Open("bar/baz")
		a.So(err, should.BeNil)

		content, err := ioutil.ReadAll(file)
		a.So(err, should.BeNil)

		a.So(string(content), should.Equal, fs["foo/bar/baz"])
	}

	{
		sub := Subdirectory(fs, "foo/bar")
		file, err := sub.Open("baz")
		a.So(err, should.BeNil)

		content, err := ioutil.ReadAll(file)
		a.So(err, should.BeNil)

		a.So(string(content), should.Equal, fs["foo/bar/baz"])
	}
}

func TestHide(t *testing.T) {
	a := assertions.New(t)

	fs := memory{
		"/foo": "content",
		"/bar": "content",
	}

	hidden := Hide(fs, "/foo")

	file, err := hidden.Open("/bar")
	a.So(err, should.BeNil)

	content, err := ioutil.ReadAll(file)
	a.So(err, should.BeNil)

	a.So(string(content), should.Equal, fs["/bar"])

	file, err = hidden.Open("/foo")
	a.So(err, should.NotBeNil)
}
