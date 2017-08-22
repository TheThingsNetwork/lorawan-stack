// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/smartystreets/assertions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// code creates Codes for testing
var code = errors.Range(10000, 11000)

func TestGRPC(t *testing.T) {
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

	code := GRPCCode(err)
	a.So(code, assertions.ShouldEqual, codes.PermissionDenied)

	// other errors should be unknown
	other := fmt.Errorf("Foo")
	code = GRPCCode(other)
	a.So(code, assertions.ShouldEqual, codes.Unknown)

	grpcErr := ToGRPC(err)

	got := FromGRPC(grpcErr)
	a.So(got.Code(), assertions.ShouldEqual, d.Code)
	a.So(got.Type(), assertions.ShouldEqual, d.Type)
	a.So(got.Error(), assertions.ShouldEqual, "You do not have access to app with id foo")

	a.So(got.Attributes(), assertions.ShouldNotBeEmpty)
	a.So(got.Attributes()["app_id"], assertions.ShouldResemble, attributes["app_id"])
	a.So(got.Attributes()["count"], assertions.ShouldAlmostEqual, attributes["count"])
}

func TestFromUnspecifiedGRPC(t *testing.T) {
	a := assertions.New(t)

	err := grpc.Errorf(codes.DeadlineExceeded, "This is an error")

	got := FromGRPC(err)
	a.So(got.Code(), assertions.ShouldEqual, errors.NoCode)
	a.So(got.Type(), assertions.ShouldEqual, errors.Timeout)
	a.So(got.Error(), assertions.ShouldEqual, "This is an error")
	a.So(got.Attributes(), assertions.ShouldBeNil)
}
