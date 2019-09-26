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

package discover

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestDefaultPort(t *testing.T) {
	for input, expected := range map[string]string{
		"localhost:http": "localhost:http",
		"localhost:80":   "localhost:80",
		"localhost":      "localhost:8884",
		"[::1]:80":       "[::1]:80",
		"::1":            "[::1]:8884",
		"192.168.1.1:80": "192.168.1.1:80",
		"192.168.1.1":    "192.168.1.1:8884",
		":80":            ":80",
		"":               ":8884",
		"[::]:80":        "[::]:80",
		"::":             "[::]:8884",
		"[::]":           "", // Invalid address
		"[::":            "", // Invalid address
	} {
		t.Run(input, func(t *testing.T) {
			target, err := defaultPort(input, 8884)
			if err != nil {
				target = ""
			}
			assertions.New(t).So(target, should.Equal, expected)
		})
	}
}
