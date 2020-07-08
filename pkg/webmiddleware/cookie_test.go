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

	"github.com/gorilla/securecookie"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestCookies(t *testing.T) {
	a := assertions.New(t)

	m := Cookies(
		[]byte("1234123412341234123412341234123412341234123412341234123412341234"),
		[]byte("12341234123412341234123412341234"),
	)

	t.Run("Set and Get Cookie", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()

		var sc *securecookie.SecureCookie

		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sc, _ = GetSecureCookie(r.Context())
			value := map[string]string{
				"foo": "bar",
			}
			if encoded, err := sc.Encode("cookie-name", value); err == nil {
				cookie := &http.Cookie{
					Name:     "cookie-name",
					Value:    encoded,
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				http.SetCookie(rec, cookie)
			}
		})).ServeHTTP(rec, r)
		res := rec.Result()

		a.So(res.StatusCode, should.Equal, http.StatusOK)
		a.So(res.Header.Get("Set-Cookie"), should.ContainSubstring, "cookie-name")

		cookies := res.Cookies()

		a.So(cookies, should.HaveLength, 1)
		a.So(cookies[0].Name, should.Equal, "cookie-name")
		a.So(cookies[0].Value, should.NotBeEmpty)

		for _, cookie := range res.Cookies() {
			if cookie.Name == "cookie-name" {
				value := make(map[string]string)
				err := sc.Decode("cookie-name", cookie.Value, &value)
				a.So(err, should.Equal, nil)
				a.So(value["foo"], should.Equal, "bar")
			}
		}
	})
}
