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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestRedirect(t *testing.T) {
	m := Redirect(RedirectConfiguration{
		Scheme: func(s string) string { return SchemeHTTPS },
		HostName: func(h string) string {
			if strings.HasPrefix(h, "dev.") {
				return h
			}
			return "dev." + h
		},
		Port: func(p uint) uint {
			switch p {
			case 1885:
				return 8885
			default:
				return 0
			}
		},
		Path: strings.ToLower,
	})

	t.Run("None", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "https://dev.example.com/path?query=true", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusOK)
		body, _ := io.ReadAll(res.Body)
		a.So(string(body), should.Equal, "OK")
	})

	for _, tc := range []struct {
		Name     string
		URL      string
		Redirect string
	}{
		{
			Name:     "HostName",
			URL:      "https://example.com/",
			Redirect: "https://dev.example.com/",
		},
		{
			Name:     "Scheme",
			URL:      "http://dev.example.com/",
			Redirect: "https://dev.example.com/",
		},
		{
			Name:     "Port",
			URL:      "http://dev.example.com:1885/",
			Redirect: "https://dev.example.com:8885/",
		},
		{
			Name:     "Path",
			URL:      "https://dev.example.com/PATH",
			Redirect: "https://dev.example.com/path",
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			r := httptest.NewRequest(http.MethodGet, tc.URL, nil)
			rec := httptest.NewRecorder()
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler was called when it shouldn't have")
			})).ServeHTTP(rec, r)

			res := rec.Result()

			a.So(res.StatusCode, should.Equal, http.StatusFound)
			a.So(res.Header.Get("Location"), should.Equal, tc.Redirect)
		})
	}
}
