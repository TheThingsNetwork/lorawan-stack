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
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type MockInterface struct {
	FileFunc func(...string) ([]byte, error)
}

func (m MockInterface) File(pathElements ...string) ([]byte, error) {
	if m.FileFunc == nil {
		panic("File called, but not set")
	}
	return m.FileFunc(pathElements...)
}

func TestWithBasePath(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		BasePath, Path []string
		Error          error
		Bytes          []byte
		AssertPath     func(*testing.T, ...string) bool
		AssertError    func(*testing.T, error) bool
		AssertBytes    func(*testing.T, []byte) bool
	}{
		{
			Name: "empty base path/empty path",
			AssertPath: func(t *testing.T, pathElements ...string) bool {
				return assertions.New(t).So(pathElements, should.BeEmpty)
			},
			AssertError: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeError)
			},
			AssertBytes: func(t *testing.T, b []byte) bool {
				return assertions.New(t).So(b, should.BeNil)
			},
		},
		{
			Name:  "empty base path/path [foo bar baz]",
			Path:  []string{"foo", "bar", "baz"},
			Bytes: []byte{0x42, 0x41},
			AssertPath: func(t *testing.T, pathElements ...string) bool {
				return assertions.New(t).So(pathElements, should.Resemble, []string{"foo", "bar", "baz"})
			},
			AssertError: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
			AssertBytes: func(t *testing.T, b []byte) bool {
				return assertions.New(t).So(b, should.Resemble, []byte{0x42, 0x41})
			},
		},
		{
			Name:  "empty base path/path [/foo bar baz]",
			Path:  []string{"/foo", "bar", "baz"},
			Bytes: []byte{0x42, 0x41},
			AssertPath: func(t *testing.T, pathElements ...string) bool {
				return assertions.New(t).So(pathElements, should.Resemble, []string{"/foo", "bar", "baz"})
			},
			AssertError: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
			AssertBytes: func(t *testing.T, b []byte) bool {
				return assertions.New(t).So(b, should.Resemble, []byte{0x42, 0x41})
			},
		},
		{
			Name:     "base [foo bar baz]/empty path",
			BasePath: []string{"foo", "bar", "baz"},
			AssertPath: func(t *testing.T, pathElements ...string) bool {
				return assertions.New(t).So(pathElements, should.Resemble, []string{"foo", "bar", "baz"})
			},
			AssertError: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeError)
			},
			AssertBytes: func(t *testing.T, b []byte) bool {
				return assertions.New(t).So(b, should.BeNil)
			},
		},
		{
			Name:     "base [foo bar baz]/path [42 baz bar foo]",
			BasePath: []string{"foo", "bar", "baz"},
			Path:     []string{"42", "baz", "bar", "foo"},
			Bytes:    []byte{0x42, 0x41},
			AssertPath: func(t *testing.T, pathElements ...string) bool {
				return assertions.New(t).So(pathElements, should.Resemble, []string{"foo", "bar", "baz", "42", "baz", "bar", "foo"})
			},
			AssertError: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
			AssertBytes: func(t *testing.T, b []byte) bool {
				return assertions.New(t).So(b, should.Resemble, []byte{0x42, 0x41})
			},
		},
		{
			Name:     "base [foo bar baz]/path [/42 baz bar foo]",
			BasePath: []string{"foo", "bar", "baz"},
			Path:     []string{"/42", "baz", "bar", "foo"},
			Bytes:    []byte{0x42, 0x41},
			AssertPath: func(t *testing.T, pathElements ...string) bool {
				return assertions.New(t).So(pathElements, should.Resemble, []string{"/42", "baz", "bar", "foo"})
			},
			AssertError: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
			AssertBytes: func(t *testing.T, b []byte) bool {
				return assertions.New(t).So(b, should.Resemble, []byte{0x42, 0x41})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			m := MockInterface{
				FileFunc: func(pathElements ...string) ([]byte, error) {
					a.So(tc.AssertPath(t, pathElements...), should.BeTrue)
					return tc.Bytes, tc.Error
				},
			}

			basePath := deepcopy.Copy(tc.BasePath).([]string)
			path := deepcopy.Copy(tc.Path).([]string)
			b, err := fetch.WithBasePath(m, basePath...).File(path...)
			if a.So(tc.AssertError(t, err), should.BeTrue) {
				a.So(tc.AssertBytes(t, b), should.BeTrue)
			}
			a.So(basePath, should.Resemble, tc.BasePath)
			a.So(path, should.Resemble, tc.Path)
		})
	}
}
