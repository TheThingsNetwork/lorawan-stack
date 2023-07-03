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

package webhandlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

func TestErrorHandler(t *testing.T) {
	ctx, getError := NewContextWithErrorValue(test.Context())

	err := errors.New("some_error")

	a := assertions.New(t)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(ctx)
	rec := httptest.NewRecorder()

	Error(rec, r, err)

	res := rec.Result()
	a.So(res.StatusCode, should.Equal, http.StatusInternalServerError)

	body, _ := io.ReadAll(res.Body)
	a.So(string(body), should.ContainSubstring, "some_error")

	a.So(getError(), should.EqualErrorOrDefinition, err)
}
