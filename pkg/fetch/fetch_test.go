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
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/fetch"
)

func Example() {
	fetcher := fetch.FromHTTP("http://webserver.thethings.network/repository", true)
	content, err := fetcher.File("README.md")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content of http://webserver.thethings.network/repository/README.md:")
	fmt.Println(string(content))
}
