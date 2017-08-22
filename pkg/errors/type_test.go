// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
		a.So(typ.String(), should.Equal, str)
	}
}

func TestTypeMarshal(t *testing.T) {
	a := assertions.New(t)

	for str, typ := range types {
		text, err := typ.MarshalText()
		a.So(err, should.BeNil)
		a.So(text, should.Resemble, []byte(str))
	}
}

func TestTypeUnmarshal(t *testing.T) {
	a := assertions.New(t)

	for str, typ := range types {
		var res Type
		err := res.UnmarshalText([]byte(str))
		a.So(err, should.BeNil)
		a.So(res, should.Equal, typ)

		err = res.UnmarshalText([]byte(strings.ToLower(str)))
		a.So(err, should.BeNil)
		a.So(res, should.Equal, typ)
	}

}
