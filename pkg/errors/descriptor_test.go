// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
)

// code creates Codes for testing
var code = Range(10000, 11000)

func TestDescriptor(t *testing.T) {
	a := assertions.New(t)

	d := &ErrDescriptor{
		MessageFormat: "You do not have access to app with id {app_id}",
		Code:          code(77),
		Type:          PermissionDenied,
		registered:    true,
	}

	attributes := Attributes{
		"app_id": "foo",
	}
	err := d.New(attributes)

	a.So(err.Error(), assertions.ShouldEqual, "You do not have access to app with id foo")
	a.So(err.Code(), assertions.ShouldEqual, d.Code)
	a.So(err.Type(), assertions.ShouldEqual, d.Type)
	a.So(err.Attributes(), assertions.ShouldResemble, attributes)
}

func TestDescriptorCause(t *testing.T) {
	a := assertions.New(t)

	d := &ErrDescriptor{
		MessageFormat: "You do not have access to app with id {app_id}",
		Code:          code(77),
		Type:          PermissionDenied,
		registered:    true,
	}

	attributes := Attributes{
		"app_id": "foo",
	}
	cause := fmt.Errorf("This is an error")
	err := d.NewWithCause(attributes, cause)

	a.So(err.Error(), assertions.ShouldEqual, "You do not have access to app with id foo")
	a.So(err.Code(), assertions.ShouldEqual, d.Code)
	a.So(err.Type(), assertions.ShouldEqual, d.Type)
	a.So(err.Attributes()["app_id"], assertions.ShouldResemble, attributes["app_id"])
	a.So(err.Attributes()[causeKey], assertions.ShouldResemble, cause)

	a.So(Cause(err), assertions.ShouldEqual, cause)
}
