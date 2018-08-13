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

func ExampleMemFetcher() {
	fetcher := fetch.NewMemFetcher(map[string][]byte{
		"myFile.yml": []byte("Content"),
	})
	content, err := fetcher.File("myFile.yml")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content of myFile.yml")
	fmt.Println(string(content))
}

func TestMemFetcher(t *testing.T) {
	a := assertions.New(t)
	fetcher := fetch.NewMemFetcher(map[string][]byte{
		"existingPath": []byte("testContent"),
	})

	// Read from an existing path and test content retrieval
	{
		content, err := fetcher.File("existingPath")
		a.So(err, should.BeNil)
		a.So(string(content), should.Equal, "testContent")
	}

	// Read from a non existing path
	{
		_, err := fetcher.File("nonExistingPath")
		a.So(err, should.NotBeNil)
	}
}
