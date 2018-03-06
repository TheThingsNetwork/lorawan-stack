// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDescriptor(t *testing.T) {
	a := assertions.New(t)

	d := &ErrDescriptor{
		MessageFormat: "You do not have access to application `{app_id}`",
		Code:          77,
		Type:          PermissionDenied,
		registered:    true,
	}

	attributes := Attributes{
		"app_id": "foo",
	}
	err := d.New(attributes)

	a.So(d.Describes(err), should.BeTrue)
	a.So(err.Error(), should.Equal, "[77]: You do not have access to application `foo`")
	a.So(err.Code(), should.Equal, d.Code)
	a.So(err.Type(), should.Equal, d.Type)
	a.So(err.Attributes(), should.Resemble, attributes)

	a.So(d.Describes(err), should.BeTrue)
	a.So(d.Describes(errors.New("Something else")), should.BeFalse)
}

func TestDescriptorCause(t *testing.T) {
	a := assertions.New(t)

	d := &ErrDescriptor{
		MessageFormat: "You do not have access to application `{app_id}`",
		Code:          77,
		Type:          PermissionDenied,
		registered:    true,
	}

	attributes := Attributes{
		"app_id": "foo",
	}
	cause := fmt.Errorf("This is an error")
	err := d.NewWithCause(attributes, cause)

	a.So(d.Describes(err), should.BeTrue)
	a.So(d.Caused(err), should.BeTrue)
	a.So(err.Error(), should.Equal, "[77]: You do not have access to application `foo`")
	a.So(err.Code(), should.Equal, d.Code)
	a.So(err.Type(), should.Equal, d.Type)
	a.So(err.Attributes()["app_id"], should.Resemble, attributes["app_id"])
	a.So(err.Attributes()[causeKey], should.Resemble, cause)

	a.So(d.Describes(err), should.BeTrue)
	a.So(d.Describes(errors.New("Something else")), should.BeFalse)

	a.So(Cause(err), should.Equal, cause)
}

func TestInheritedErrors(t *testing.T) {
	a := assertions.New(t)

	d := &ErrDescriptor{
		MessageFormat: "You do not have access to application `{app_id}`",
		Code:          77,
		Type:          PermissionDenied,
		registered:    true,
	}
	d2 := &ErrDescriptor{
		MessageFormat: "You do not have access to application `{app_id}`",
		Code:          790,
		Type:          PermissionDenied,
		registered:    true,
	}

	attributes := Attributes{
		"app_id": "foo",
	}
	err := d.New(attributes)

	err2 := d2.NewWithCause(attributes, err)
	a.So(d.Caused(err2), should.BeTrue)

	err3 := errors.New("Undefined error")
	a.So(d.Caused(err3), should.BeFalse)

	err4 := d2.New(attributes)
	a.So(d.Caused(err4), should.BeFalse)

	err5 := d2.NewWithCause(attributes, err3)
	a.So(d.Caused(err5), should.BeFalse)

	err6 := NewWithCause(err2, "Inherited error")
	a.So(d.Caused(err6), should.BeTrue)
}
