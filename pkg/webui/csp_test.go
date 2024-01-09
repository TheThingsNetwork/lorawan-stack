// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package webui_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

func TestRewriteScheme(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		baseURL  string
		rewrites map[string]string
		expected []string
	}{
		{
			name:    "no match",
			baseURL: "https://example.com",
			rewrites: map[string]string{
				"http": "ws",
			},
			expected: []string{"https://example.com"},
		},
		{
			name:    "match",
			baseURL: "https://example.com",
			rewrites: map[string]string{
				"https": "wss",
			},
			expected: []string{"https://example.com", "wss://example.com"},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)
			actual := webui.RewriteScheme(tc.rewrites, tc.baseURL)
			a.So(actual, should.Resemble, tc.expected)
		})
	}
}

func TestRewriteSchemes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		baseURLs []string
		rewrites map[string]string
		expected []string
	}{
		{
			name:     "no match",
			baseURLs: []string{"https://foo.example.com", "https://bar.example.com"},
			rewrites: map[string]string{
				"http": "ws",
			},
			expected: []string{"https://foo.example.com", "https://bar.example.com"},
		},
		{
			name:     "match",
			baseURLs: []string{"https://foo.example.com", "https://bar.example.com"},
			rewrites: map[string]string{
				"https": "wss",
			},
			expected: []string{
				"https://foo.example.com", "wss://foo.example.com", "https://bar.example.com", "wss://bar.example.com",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)
			actual := webui.RewriteSchemes(tc.rewrites, tc.baseURLs...)
			a.So(actual, should.Resemble, tc.expected)
		})
	}
}
