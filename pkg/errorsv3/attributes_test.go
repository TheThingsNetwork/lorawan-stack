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

package errors_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestAttributes(t *testing.T) {
	a := assertions.New(t)

	var errInvalidFoo = errors.DefineInvalidArgument("test_attributes_invalid_foo", "Invalid Foo: {foo}", "foo")

	a.So(errInvalidFoo.Attributes(), should.BeEmpty)

	a.So(func() { errInvalidFoo.WithAttributes("only_key") }, should.Panic)

	err1 := errInvalidFoo.WithAttributes("foo", "bar")
	err2 := err1.WithAttributes("bar", "baz")

	a.So(err1, should.HaveSameErrorDefinitionAs, errInvalidFoo)
	a.So(errors.Attributes(err1), should.Resemble, map[string]interface{}{"foo": "bar"})
	a.So(errors.PublicAttributes(err1), should.Resemble, map[string]interface{}{"foo": "bar"})

	a.So(err2, should.HaveSameErrorDefinitionAs, err1)
	a.So(err2, should.HaveSameErrorDefinitionAs, errInvalidFoo)
	a.So(errors.Attributes(err2), should.Resemble, map[string]interface{}{"foo": "bar", "bar": "baz"})
	a.So(errors.PublicAttributes(err2), should.Resemble, map[string]interface{}{"foo": "bar"})
}

func TestSupportedAttributes(t *testing.T) {
	tt := []struct {
		Name   string
		V      interface{}
		Expect interface{}
	}{
		{"bool", false, false},
		{"int", int(42), int(42)},
		{"int8", int8(42), int8(42)},
		{"int16", int16(42), int16(42)},
		{"int32", int32(42), int32(42)},
		{"int64", int64(42), int64(42)},
		{"uint", uint(42), uint(42)},
		{"uint8", uint8(42), uint8(42)},
		{"uint16", uint16(42), uint16(42)},
		{"uint32", uint32(42), uint32(42)},
		{"uint64", uint64(42), uint64(42)},
		{"uintptr", uintptr(42), uintptr(42)},
		{"float32", float32(42), float32(42)},
		{"float64", float64(42), float64(42)},
		{"complex64", complex64(42), complex64(42)},
		{"complex128", complex128(42), complex128(42)},
		{"string", "foo", "foo"},
		{"eui", types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, "0102030405060708"},
	}

	for _, tt := range tt {
		t.Run(tt.Name, func(t *testing.T) {
			assertions.New(t).So(errors.Supported(tt.V), should.Equal, tt.Expect)
		})
	}

}
