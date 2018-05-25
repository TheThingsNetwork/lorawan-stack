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

package oauth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestLoginFlow(t *testing.T) {
	a := assertions.New(t)

	srv := testServer(t)

	// Can not call the `me` endpoint as not logged in.
	{
		req := httptest.NewRequest(echo.GET, "/api/me", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		resp := rec.Result()
		a.So(resp.StatusCode, should.Equal, http.StatusUnauthorized)
	}

	// Log in.
	cookie := new(http.Cookie)
	{
		body := strings.NewReader(`{"user_id":"` + userID + `","password":"` + password + `"}`)
		req := httptest.NewRequest(echo.POST, "/api/auth/login", body)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		resp := rec.Result()
		a.So(resp.StatusCode, should.Equal, http.StatusOK)
		if a.So(resp.Cookies(), should.NotBeEmpty) {
			// Set in cookie the session cookie.
			cookie = resp.Cookies()[0]
		}
	}

	// Can call the `me` endpoint as logged in.
	{
		req := httptest.NewRequest(echo.GET, "/api/me", nil)
		req.AddCookie(cookie)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		resp := rec.Result()
		a.So(resp.StatusCode, should.Equal, http.StatusOK)
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		me := new(user)
		err = json.Unmarshal(buf, me)
		a.So(err, should.BeNil)
		a.So(me.UserID, should.Equal, userID)
	}

	// Log out.
	{
		req := httptest.NewRequest(echo.POST, "/api/auth/logout", nil)
		req.AddCookie(cookie)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		resp := rec.Result()
		a.So(resp.StatusCode, should.Equal, http.StatusOK)
		a.So(resp.Cookies(), should.NotBeEmpty)
		a.So(resp.Cookies()[0].Value, should.Equal, "<deleted>")
	}
}
