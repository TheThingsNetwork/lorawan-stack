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
	"encoding/json"
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

func TestNotFoundXSSSanitization(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	xssPath := `/api/v3/<script>alert("xss")</script>`

	r := httptest.NewRequest(http.MethodGet, xssPath, nil)
	rec := httptest.NewRecorder()

	NotFound(rec, r)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)

	a.So(res.StatusCode, should.Equal, http.StatusNotFound)

	// Decode the JSON to inspect the attribute value after JSON decoding.
	var resp map[string]any
	a.So(json.Unmarshal(body, &resp), should.BeNil)

	// Extract the route attribute from the error details.
	details, _ := resp["details"].([]any) //nolint:revive
	a.So(len(details), should.BeGreaterThan, 0)
	detail, _ := details[0].(map[string]any)          //nolint:revive
	attrs, _ := detail["attributes"].(map[string]any) //nolint:revive
	route, _ := attrs["route"].(string)               //nolint:revive

	// The attribute value must be HTML-escaped, not the raw XSS payload.
	a.So(route, should.NotContainSubstring, "<script>")
	a.So(route, should.ContainSubstring, "&lt;script&gt;")
}

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
