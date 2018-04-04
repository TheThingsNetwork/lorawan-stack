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
	"External":                External,
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
