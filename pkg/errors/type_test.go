// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
)

var types = map[string]Type{
	"Unknown":                 Unknown,
	"Internal":                Internal,
	"Invalid argument":        InvalidArgument,
	"Out of range":            OutOfRange,
	"Not found":               NotFound,
	"Conflict":                Conflict,
	"Already exists":          AlreadyExists,
	"Unauthorized":            Unauthorized,
	"Permission denied":       PermissionDenied,
	"Timeout":                 Timeout,
	"Not implemented":         NotImplemented,
	"Temporarily unavailable": TemporarilyUnavailable,
	"Permanently unavailable": PermanentlyUnavailable,
	"Canceled":                Canceled,
	"Resource exhausted":      ResourceExhausted,
}

func TestTypeString(t *testing.T) {
	a := assertions.New(t)

	for str, typ := range types {
		a.So(typ.String(), assertions.ShouldEqual, str)
	}
}

func TestTypeMarshal(t *testing.T) {
	a := assertions.New(t)

	for str, typ := range types {
		text, err := typ.MarshalText()
		a.So(err, assertions.ShouldBeNil)
		a.So(text, assertions.ShouldResemble, []byte(str))
	}
}

func TestTypeUnmarshal(t *testing.T) {
	a := assertions.New(t)

	for str, typ := range types {
		var res Type
		err := res.UnmarshalText([]byte(str))
		a.So(err, assertions.ShouldBeNil)
		a.So(res, assertions.ShouldEqual, typ)

		err = res.UnmarshalText([]byte(strings.ToLower(str)))
		a.So(err, assertions.ShouldBeNil)
		a.So(res, assertions.ShouldEqual, typ)
	}

}
