// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package types

import (
	"reflect"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestTypes(t *testing.T) {
	a := assertions.New(t)

	zeroSubjects := func() []Interface {
		return []Interface{
			&DevNonce{},
			&JoinNonce{},
			&NetID{},
			&DevAddr{},
			&DevAddrPrefix{},
			&EUI64{},
			&EUI64Prefix{},
			&AES128Key{},
		}
	}

	for _, sub := range zeroSubjects() {
		a.So(sub.IsZero(), should.BeTrue)
	}

	subjects := func() []Interface {
		return append(zeroSubjects(),
			&DevNonce{12, 34},
			&JoinNonce{12, 34, 56},
			&NetID{12, 34, 56},
			&DevAddr{12, 34, 56, 78},
			&DevAddrPrefix{DevAddr{12, 34, 56, 78}, 8},
			&DevAddrPrefix{DevAddr{12, 34, 56, 78}, 12},
			&EUI64{12, 34, 56, 78, 12, 34, 56, 78},
			&EUI64Prefix{EUI64{12, 34, 56, 78, 12, 34, 56, 78}, 2},
			&EUI64Prefix{EUI64{12, 34, 56, 78, 12, 34, 56, 78}, 16},
			&AES128Key{12, 34, 56, 78, 12, 34, 56, 78, 12, 34, 56, 78, 12, 34, 56, 78},
		)
	}

	for _, sub := range subjects() {
		t.Run(reflect.TypeOf(sub).String(), func(t *testing.T) {
			a = assertions.New(t)

			// MarshalText, String, GoString
			text, err := sub.MarshalText()
			a.So(err, should.BeNil)
			a.So(string(text), should.Equal, sub.String())
			a.So(string(text), should.Equal, sub.GoString())

			// MarshalBinary, Size
			bytes, err := sub.MarshalBinary()
			a.So(err, should.BeNil)
			a.So(bytes, should.HaveLength, sub.Size())

			// MarshalJSON
			json, err := sub.MarshalJSON()
			a.So(err, should.BeNil)

			// Marshal, MarshalTo
			marshaled := make([]byte, sub.Size())
			i, err := sub.MarshalTo(marshaled)
			a.So(err, should.BeNil)
			a.So(i, should.Resemble, sub.Size())
			a.So(marshaled, should.Resemble, bytes)
			marshaled, err = sub.Marshal()
			a.So(err, should.BeNil)
			a.So(marshaled, should.Resemble, bytes)

			// Unmarshal
			err = sub.Unmarshal(bytes)
			a.So(err, should.BeNil)
			a.So(sub.String(), should.Equal, string(text))

			// UnmarshalJSON
			err = sub.UnmarshalJSON(json)
			a.So(err, should.BeNil)
			a.So(sub.String(), should.Equal, string(text))

			// UnmarshalBinary
			err = sub.UnmarshalBinary(marshaled)
			a.So(err, should.BeNil)
			a.So(sub.String(), should.Equal, string(text))

			// UnmarshalText
			err = sub.UnmarshalText(text)
			a.So(err, should.BeNil)
			a.So(sub.String(), should.Equal, string(text))
		})
	}

	for _, sub := range zeroSubjects() {
		t.Run(reflect.TypeOf(sub).String(), func(t *testing.T) {
			a = assertions.New(t)

			// Empty should not error
			err := sub.UnmarshalBinary([]byte{})
			a.So(err, should.BeNil)

			// Too short
			err = sub.UnmarshalBinary([]byte{1})
			a.So(err, should.NotBeNil)

			// Too long
			err = sub.UnmarshalBinary([]byte(strings.Repeat("foo", 32)))
			a.So(err, should.NotBeNil)

			// Empty should not error
			err = sub.UnmarshalText([]byte{})
			a.So(err, should.BeNil)

			// Too short
			err = sub.UnmarshalText([]byte{1})
			a.So(err, should.NotBeNil)

			// Too long
			err = sub.UnmarshalText([]byte(strings.Repeat("foo", 32)))
			a.So(err, should.NotBeNil)

			// Invalid hesx
			err = sub.UnmarshalText([]byte(strings.Repeat("zz", 2)))
			a.So(err, should.NotBeNil)
			err = sub.UnmarshalText([]byte(strings.Repeat("zz", 3)))
			a.So(err, should.NotBeNil)
			err = sub.UnmarshalText([]byte(strings.Repeat("zz", 4)))
			a.So(err, should.NotBeNil)
			err = sub.UnmarshalText([]byte(strings.Repeat("zz", 8)))
			a.So(err, should.NotBeNil)
			err = sub.UnmarshalText([]byte(strings.Repeat("zz", 16)))
			a.So(err, should.NotBeNil)

			// Invalid prefixes
			err = sub.UnmarshalText([]byte("f00f00f0/"))
			a.So(err, should.NotBeNil)
			err = sub.UnmarshalText([]byte("f00f00f0/fail"))
			a.So(err, should.NotBeNil)
			err = sub.UnmarshalText([]byte("f00f00f00/fail"))
			a.So(err, should.NotBeNil)
		})
	}
}
