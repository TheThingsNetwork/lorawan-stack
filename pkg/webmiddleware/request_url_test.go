// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package webmiddleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestRequestURL(t *testing.T) {
	m := RequestURL()

	t.Run("HTTP", func(t *testing.T) {
		a := assertions.New(t)

		r := httptest.NewRequest(http.MethodGet, "/path", nil)

		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(r.URL.String(), should.Equal, "http://example.com/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("HTTPS", func(t *testing.T) {
		a := assertions.New(t)

		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.TLS = &tls.ConnectionState{
			Version:           tls.VersionTLS12,
			HandshakeComplete: true,
			ServerName:        r.Host,
		}

		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(r.URL.String(), should.Equal, "https://example.com/path")
		})).ServeHTTP(rec, r)
	})
}
