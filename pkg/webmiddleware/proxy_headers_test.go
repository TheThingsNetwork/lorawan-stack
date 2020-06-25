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

package webmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
		r.Header.Set(headerXForwardedFor, "12.34.56.78, 10.100.0.1")
		r.Header.Set(headerXForwardedHost, "thethings.network")
		r.Header.Set(headerXForwardedProto, schemeHTTPS)
		r.Header.Set(headerXRealIP, "12.34.56.78")
		r.Header.Set(headerXForwardedClientCert, "Subject=\"...\"")
		r.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(r.Header.Get(headerXForwardedFor), should.Equal, "12.34.56.78, 10.100.0.1")
			a.So(r.Header.Get(headerXForwardedHost), should.Equal, "thethings.network")
			a.So(r.Header.Get(headerXForwardedProto), should.Equal, schemeHTTPS)
			a.So(r.Header.Get(headerXRealIP), should.Equal, "12.34.56.78")
			a.So(r.Header.Get(headerXForwardedClientCert), should.Equal, "Subject=\"...\"")
			a.So(r.URL.String(), should.Equal, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("Trusted Forwarded", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(headerForwarded, "for=12.34.56.78, for=10.100.0.1;host=thethings.network;proto=https")
		r.Header.Set(headerXRealIP, "12.34.56.78")
		r.Header.Set(headerXForwardedTLSClientCert, "MIIDEDCC...")
		r.Header.Set(headerXForwardedTLSClientCertInfo, "Subject=\"...\"")
		r.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(r.Header.Get(headerForwarded), should.Equal, "for=12.34.56.78, for=10.100.0.1;host=thethings.network;proto=https")
			a.So(r.Header.Get(headerXRealIP), should.Equal, "12.34.56.78")
			a.So(r.Header.Get(headerXForwardedTLSClientCert), should.Equal, "MIIDEDCC...")
			a.So(r.Header.Get(headerXForwardedTLSClientCertInfo), should.Equal, "Subject=\"...\"")
			a.So(r.URL.String(), should.Equal, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("Untrusted X-Forwarded-For", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(headerXForwardedFor, "12.34.56.78")
		r.Header.Set(headerXForwardedHost, "thethings.network")
		r.Header.Set(headerXForwardedProto, schemeHTTPS)
		r.Header.Set(headerXRealIP, "12.34.56.78")
		r.RemoteAddr = "12.34.56.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range []string{headerForwarded, headerXForwardedFor, headerXForwardedHost, headerXForwardedProto} {
				a.So(r.Header.Get(header), should.BeEmpty)
			}
			a.So(r.Header.Get(headerXRealIP), should.Equal, "12.34.56.1")
			a.So(r.URL.String(), should.NotEqual, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})

	t.Run("Untrusted Forwarded", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(headerForwarded, "for=12.34.56.78;host=thethings.network;proto=https")
		r.Header.Set(headerXRealIP, "12.34.56.78")
		r.RemoteAddr = "12.34.56.1:1234"
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range []string{headerForwarded, headerXForwardedFor, headerXForwardedHost, headerXForwardedProto} {
				a.So(r.Header.Get(header), should.BeEmpty)
			}
			a.So(r.Header.Get(headerXRealIP), should.Equal, "12.34.56.1")
			a.So(r.URL.String(), should.NotEqual, "https://thethings.network/path")
		})).ServeHTTP(rec, r)
	})
}
