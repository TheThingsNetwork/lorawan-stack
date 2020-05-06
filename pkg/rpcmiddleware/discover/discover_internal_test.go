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
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
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
			target, err := DefaultPort(input, 8884)
			if err != nil {
				target = ""
			}
			assertions.New(t).So(target, should.Equal, expected)
		})
	}
}

func TestDefaultURL(t *testing.T) {
	for _, tc := range []struct {
		target   string
		port     int
		tls      bool
		expected string
	}{
		{
			target:   "localhost",
			port:     80,
			tls:      false,
			expected: "http://localhost:80",
		},
		{
			target:   "host.with.port:http",
			port:     8000,
			tls:      false,
			expected: "http://host.with.port:http",
		},
		{
			target:   "hostname:433",
			port:     4000,
			tls:      true,
			expected: "https://hostname:433",
		},
		{
			target:   "hostname",
			port:     8443,
			tls:      true,
			expected: "https://hostname:8443",
		},
	} {
		t.Run(tc.expected, func(t *testing.T) {
			target, err := DefaultURL(tc.target, tc.port, tc.tls)
			if err != nil {
				target = ""
			}
			assertions.New(t).So(target, should.Equal, tc.expected)
		})
	}
}
