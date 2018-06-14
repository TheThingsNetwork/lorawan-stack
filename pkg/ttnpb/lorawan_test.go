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

package ttnpb_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestMACVersionCompare(t *testing.T) {
	for _, tc := range []struct {
		A, B     MACVersion
		Expected int
		Panics   bool
	}{
		{
			A:        MAC_V1_0,
			B:        MAC_V1_0_1,
			Expected: -1,
		},
		{
			A:        MAC_V1_1,
			B:        MAC_V1_0,
			Expected: 1,
		},
		{
			A:        MAC_V1_1,
			B:        MAC_V1_1,
			Expected: 0,
		},
		{
			A:        MAC_V1_0_2,
			B:        MAC_V1_1,
			Expected: -1,
		},
		{
			A:      MAC_UNKNOWN,
			B:      MAC_V1_1,
			Panics: true,
		},
		{
			A:      MAC_UNKNOWN,
			B:      MAC_UNKNOWN,
			Panics: true,
		},
		{
			A:      MAC_V1_0,
			B:      MAC_UNKNOWN,
			Panics: true,
		},
	} {
		a := assertions.New(t)

		if tc.Panics {
			a.So(func() { tc.A.Compare(tc.B) }, should.Panic)
			return
		}

		a.So(tc.A.Compare(tc.B), should.Equal, tc.Expected)
		if tc.A != tc.B {
			a.So(tc.B.Compare(tc.A), should.Equal, -tc.Expected)
		}
	}
}

func TestStringers(t *testing.T) {
	for _, tc := range []struct {
		Stringer fmt.Stringer
		String   string
	}{
		{
			Stringer: MAC_V1_0,
			String:   "1.0.0",
		},
		{
			Stringer: MAC_V1_0_1,
			String:   "1.0.1",
		},
		{
			Stringer: MAC_V1_0_2,
			String:   "1.0.2",
		},
		{
			Stringer: MAC_V1_1,
			String:   "1.1.0",
		},
		{
			Stringer: PHY_V1_0,
			String:   "1.0.0",
		},
		{
			Stringer: PHY_V1_0_1,
			String:   "1.0.1",
		},
		{
			Stringer: PHY_V1_0_2_REV_A,
			String:   "1.0.2-a",
		},
		{
			Stringer: PHY_V1_0_2_REV_B,
			String:   "1.0.2-b",
		},
		{
			Stringer: PHY_V1_1_REV_A,
			String:   "1.1.0-a",
		},
		{
			Stringer: PHY_V1_1_REV_B,
			String:   "1.1.0-b",
		},
	} {
		assertions.New(t).So(tc.Stringer.String(), should.Equal, tc.String)
	}
}
