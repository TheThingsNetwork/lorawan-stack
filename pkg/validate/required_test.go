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

package validate

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type isZeroer struct {
	isZero bool
}

var izCalled int

func (iz isZeroer) IsZero() bool {
	izCalled++
	return iz.isZero
}

type stringType string
type boolType bool
type int64Type int64
type uint64Type uint64
type float64Type float64
type structType struct {
	field bool
}

var isZeroCases = []struct {
	v      interface{}
	isZero bool
}{
	{nil, true},
	{(interface{})(nil), true},

	{false, true},
	{int(0), true},
	{int64(0), true},
	{int32(0), true},
	{int16(0), true},
	{int8(0), true},
	{uint(0), true},
	{uint64(0), true},
	{uint32(0), true},
	{uint16(0), true},
	{uint8(0), true},
	{float64(0), true},
	{float32(0), true},
	{float32(0), true},
	{"", true},

	{[]bool{}, true},
	{[]int{}, true},
	{[]int64{}, true},
	{[]int32{}, true},
	{[]int16{}, true},
	{[]int8{}, true},
	{[]uint{}, true},
	{[]uint64{}, true},
	{[]uint32{}, true},
	{[]uint16{}, true},
	{[]uint8{}, true},
	{[]float64{}, true},
	{[]float32{}, true},
	{[]float32{}, true},
	{[]string{}, true},

	{(*time.Time)(nil), true},
	{time.Time{}, true},
	{&time.Time{}, true},

	{map[string]interface{}{}, true},
	{map[string]interface{}{"foo": "bar"}, false},

	{map[string]string{}, true},
	{map[string]string{"foo": "bar"}, false},

	{types.AES128Key{}, true},
	{types.EUI64{}, true},
	{types.NetID{}, true},
	{types.DevAddr{}, true},
	{types.DevNonce{}, true},
	{types.JoinNonce{}, true},

	{(*types.AES128Key)(nil), true},
	{(*types.EUI64)(nil), true},
	{(*types.NetID)(nil), true},
	{(*types.DevAddr)(nil), true},
	{(*types.DevNonce)(nil), true},
	{(*types.JoinNonce)(nil), true},

	{isZeroer{isZero: false}, false},
	{isZeroer{isZero: true}, true},

	{(*isZeroer)(nil), true},
	{&isZeroer{isZero: false}, false},
	{&isZeroer{isZero: true}, true},

	{stringType(""), true},
	{stringType("foo"), false},

	{boolType(false), true},
	{boolType(true), false},

	{int64Type(0), true},
	{int64Type(42), false},

	{uint64Type(0), true},
	{uint64Type(42), false},

	{float64Type(0), true},
	{float64Type(42), false},

	{structType{true}, false},
	{structType{false}, true},

	{unsafe.Pointer(nil), true},
	{unsafe.Pointer(&([]byte{42})[0]), false},
}

func TestIsZero(t *testing.T) {
	for i, tc := range isZeroCases {
		t.Run(fmt.Sprintf("%d %T(%v)", i, tc.v, tc.v), func(t *testing.T) {
			assertions.New(t).So(isZero(tc.v), should.Equal, tc.isZero)
		})
	}
	assertions.New(t).So(izCalled, should.Equal, 4)
}

func TestRequired(t *testing.T) {
	a := assertions.New(t)

	a.So(Field("", Required), should.NotBeNil)
	a.So(Field("f", Required), should.BeNil)

	a.So(Field("", NotRequired), should.BeNil)
	a.So(Field("f", NotRequired), should.BeNil)
}

func BenchmarkIsZero(b *testing.B) {
	for i, tc := range isZeroCases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				isZero(tc.v)
			}
		})
	}
}

func BenchmarkIsZeroValue(b *testing.B) {
	for i, tc := range isZeroCases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				isZeroValue(reflect.ValueOf(tc.v))
			}
		})
	}
}
