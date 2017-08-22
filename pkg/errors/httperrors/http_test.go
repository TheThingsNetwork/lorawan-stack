// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	go_errors "errors"
	"net/http/httptest"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/smartystreets/assertions"
)

// code creates Codes for testing
var code = errors.Range(10000, 11000)

func TestHTTP(t *testing.T) {
	a := assertions.New(t)

	d := &errors.ErrDescriptor{
		MessageFormat: "You do not have access to app with id {app_id}",
		Code:          code(77),
		Type:          errors.PermissionDenied,
	}
	d.Register()

	attributes := errors.Attributes{
		"app_id": "foo",
		"count":  42,
	}

	err := d.New(attributes)

	w := httptest.NewRecorder()
	e := ToHTTP(err, w)
	a.So(e, assertions.ShouldBeNil)

	resp := w.Result()

	got := FromHTTP(resp)
	a.So(got.Code(), assertions.ShouldEqual, err.Code())
	a.So(got.Type(), assertions.ShouldEqual, err.Type())
	a.So(got.Error(), assertions.ShouldEqual, err.Error())
	a.So(got.Attributes()["app_id"], assertions.ShouldResemble, attributes["app_id"])
	a.So(got.Attributes()["count"], assertions.ShouldAlmostEqual, attributes["count"])
}

func TestToUnspecifiedHTTP(t *testing.T) {
	a := assertions.New(t)

	err := go_errors.New("A random error")

	w := httptest.NewRecorder()
	e := ToHTTP(err, w)
	a.So(e, assertions.ShouldBeNil)

	resp := w.Result()

	got := FromHTTP(resp)
	a.So(got.Code(), assertions.ShouldEqual, errors.NoCode)
	a.So(got.Type(), assertions.ShouldEqual, errors.Unknown)
	a.So(got.Error(), assertions.ShouldEqual, err.Error())
	a.So(got.Attributes(), assertions.ShouldBeNil)
}

func TestHTTPResponse(t *testing.T) {
	a := assertions.New(t)

	w := httptest.NewRecorder()
	resp := w.Result()
	resp.StatusCode = 404
	resp.Status = "404 Not found"

	got := FromHTTP(resp)
	a.So(got.Code(), assertions.ShouldEqual, errors.NoCode)
	a.So(got.Type(), assertions.ShouldEqual, errors.NotFound)
	a.So(got.Error(), assertions.ShouldEqual, "Not found")
	a.So(got.Attributes(), assertions.ShouldBeNil)
}
