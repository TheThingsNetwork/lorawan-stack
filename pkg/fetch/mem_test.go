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

package fetch_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func ExampleNewMemFetcher() {
	fetcher := fetch.NewMemFetcher(map[string][]byte{
		"file.txt":     []byte("content"),
		"dir/file.txt": []byte("content"),
	})
	content, err := fetcher.File("dir/file.txt")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content of myFile.yml")
	fmt.Println(string(content))
}

func TestMemFetcher(t *testing.T) {
	a := assertions.New(t)
	fetcher := fetch.NewMemFetcher(map[string][]byte{
		"file.txt":     []byte("content1"),
		"dir/file.txt": []byte("content2"),
	})

	// Read a file and test content retrieval.
	{
		content, err := fetcher.File("file.txt")
		a.So(err, should.BeNil)
		a.So(string(content), should.Equal, "content1")
	}

	// Read from a subdirectory and test content retrieval.
	{
		content, err := fetcher.File("dir", "file.txt")
		a.So(err, should.BeNil)
		a.So(string(content), should.Equal, "content2")
	}

	// Read from a non existing path.
	{
		_, err := fetcher.File("notfound.txt")
		a.So(err, should.NotBeNil)
	}
}
