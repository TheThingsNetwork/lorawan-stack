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
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/webhandlers"
	. "go.thethings.network/lorawan-stack/pkg/webmiddleware"
)

func TestMaxBody(t *testing.T) {
	m := MaxBody(16)

	t.Run("Normal Request", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusOK)
	})

	t.Run("Request Too Big", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(
			http.MethodPost, "/",
			bytes.NewBuffer([]byte("this is a little to much")),
		)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := ioutil.ReadAll(r.Body)
			a.So(err, should.HaveSameErrorDefinitionAs, ErrRequestBodyTooLarge)
			webhandlers.Error(w, r, err)
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusBadRequest)

		body, _ := ioutil.ReadAll(res.Body)
		a.So(string(body), should.ContainSubstring, "request_body_too_large")
	})
}
