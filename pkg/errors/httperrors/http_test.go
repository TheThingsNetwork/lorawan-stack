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

package httperrors

import (
	go_errors "errors"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

func TestHTTP(t *testing.T) {
	a := assertions.New(t)

	d := &errors.ErrDescriptor{
		MessageFormat: "You do not have access to application `{app_id}`",
		Code:          77,
		Type:          errors.PermissionDenied,
		Namespace:     "pkg/foo",
		SafeAttributes: []string{
			"app_id",
			"count",
		},
	}
	d.Register()

	attributes := errors.Attributes{
		"app_id": "foo",
		"count":  42,
		"unsafe": "secret",
	}

	err := d.New(attributes)

	w := httptest.NewRecorder()
	e := ToHTTP(err, w)
	a.So(e, should.BeNil)

	resp := w.Result()

	got := FromHTTP(resp)
	a.So(got.Code(), should.Equal, err.Code())
	a.So(got.Type(), should.Equal, err.Type())
	a.So(got.Message(), should.Equal, err.Message())
	a.So(got.ID(), should.Equal, err.ID())
	a.So(got.Error(), should.Equal, "pkg/foo[77]: You do not have access to application `foo`")
	a.So(got.Attributes()["app_id"], should.Resemble, attributes["app_id"])
	a.So(got.Attributes()["count"], should.AlmostEqual, attributes["count"])
	a.So(got.Attributes(), should.NotContainKey, "unsafe")
	a.So(got.Namespace(), should.Equal, d.Namespace)
}

func TestToUnspecifiedHTTP(t *testing.T) {
	a := assertions.New(t)

	err := go_errors.New("A random error")

	w := httptest.NewRecorder()
	e := ToHTTP(err, w)
	a.So(e, should.BeNil)

	resp := w.Result()

	got := FromHTTP(resp)
	a.So(got.Code(), should.Equal, errors.NoCode)
	a.So(got.Type(), should.Equal, errors.Unknown)
	a.So(got.Error(), should.Equal, err.Error())
	a.So(got.ID(), should.NotBeEmpty)
	a.So(got.Attributes(), should.BeNil)
	a.So(got.Namespace(), should.BeEmpty)
}

func TestHTTPResponse(t *testing.T) {
	a := assertions.New(t)

	w := httptest.NewRecorder()
	resp := w.Result()
	resp.StatusCode = 404
	resp.Status = "Not found"

	got := FromHTTP(resp)
	a.So(got.Code(), should.Equal, errors.NoCode)
	a.So(got.Type(), should.Equal, errors.NotFound)
	a.So(got.Error(), should.Equal, "Not found")
	a.So(got.ID(), should.NotBeEmpty)
	a.So(got.Attributes(), should.BeNil)
	a.So(got.Namespace(), should.BeEmpty)
}
