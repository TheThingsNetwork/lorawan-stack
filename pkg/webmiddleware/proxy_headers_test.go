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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestProxyHeaders(t *testing.T) {
	a := assertions.New(t)

	var config ProxyConfiguration
	err := config.ParseAndAddTrusted("192.0.2.0/24")
	a.So(err, should.BeNil)

	m := ProxyHeaders(config)

	t.Run("Trusted X-Forwarded-For", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(HeaderXForwardedFor, "12.34.56.78, 10.100.0.1")
		r.Header.Set(HeaderXForwardedHost, "thethings.network")
		r.Header.Set(HeaderXForwardedProto, SchemeHTTPS)
		r.Header.Set(HeaderXRealIP, "12.34.56.78")
		r.Header.Set(HeaderXForwardedClientCert, "Subject=\"...\"")
		r.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(r.Header.Get(HeaderXForwardedFor), should.Equal, "12.34.56.78, 10.100.0.1")
			a.So(r.Header.Get(HeaderXForwardedHost), should.Equal, "thethings.network")
			a.So(r.Header.Get(HeaderXForwardedProto), should.Equal, SchemeHTTPS)
			a.So(r.Header.Get(HeaderXRealIP), should.Equal, "12.34.56.78")
			a.So(r.Header.Get(HeaderXForwardedClientCert), should.Equal, "Subject=\"...\"")
			a.So(r.URL.String(), should.Equal, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("Trusted Forwarded", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(HeaderForwarded, "for=12.34.56.78, for=10.100.0.1;host=thethings.network;proto=https")
		r.Header.Set(HeaderXRealIP, "12.34.56.78")
		r.Header.Set(HeaderXForwardedTLSClientCert, "MIIDEDCC...")
		r.Header.Set(HeaderXForwardedTLSClientCertInfo, "Subject=\"...\"")
		r.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(r.Header.Get(HeaderForwarded), should.Equal, "for=12.34.56.78, for=10.100.0.1;host=thethings.network;proto=https")
			a.So(r.Header.Get(HeaderXRealIP), should.Equal, "12.34.56.78")
			a.So(r.Header.Get(HeaderXForwardedTLSClientCert), should.Equal, "MIIDEDCC...")
			a.So(r.Header.Get(HeaderXForwardedTLSClientCertInfo), should.Equal, "Subject=\"...\"")
			a.So(r.URL.String(), should.Equal, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("Untrusted X-Forwarded-For", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(HeaderXForwardedFor, "12.34.56.78")
		r.Header.Set(HeaderXForwardedHost, "thethings.network")
		r.Header.Set(HeaderXForwardedProto, SchemeHTTPS)
		r.Header.Set(HeaderXRealIP, "12.34.56.78")
		r.RemoteAddr = "12.34.56.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range []string{HeaderForwarded, HeaderXForwardedFor, HeaderXForwardedHost, HeaderXForwardedProto} {
				a.So(r.Header.Get(header), should.BeEmpty)
			}
			a.So(r.Header.Get(HeaderXRealIP), should.Equal, "12.34.56.1")
			a.So(r.URL.String(), should.NotEqual, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("Untrusted Forwarded", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(HeaderForwarded, "for=12.34.56.78;host=thethings.network;proto=https")
		r.Header.Set(HeaderXRealIP, "12.34.56.78")
		r.RemoteAddr = "12.34.56.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range []string{HeaderForwarded, HeaderXForwardedFor, HeaderXForwardedHost, HeaderXForwardedProto} {
				a.So(r.Header.Get(header), should.BeEmpty)
			}
			a.So(r.Header.Get(HeaderXRealIP), should.Equal, "12.34.56.1")
			a.So(r.URL.String(), should.NotEqual, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})
}
