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

package ttnpb_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMACVersion_Version(t *testing.T) {
	a := assertions.New(t)
	for v := range MACVersion_name {
		if v == int32(MAC_UNKNOWN) {
			continue
		}
		a.So(func() { MACVersion(v).Version() }, should.NotPanic)
	}
}

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

func TestDataRateIndex(t *testing.T) {
	a := assertions.New(t)
	a.So(DataRateIndex_DATA_RATE_4.String(), should.Equal, "DATA_RATE_4")

	b, err := DataRateIndex_DATA_RATE_4.MarshalText()
	a.So(err, should.BeNil)
	a.So(string(b), should.Resemble, "4")

	for _, str := range []string{"4", "DATA_RATE_4"} {
		var idx DataRateIndex
		err = idx.UnmarshalText([]byte(str))
		a.So(idx, should.Equal, DataRateIndex_DATA_RATE_4)
	}
}

func TestDeviceEIRP(t *testing.T) {
	a := assertions.New(t)
	a.So(DeviceEIRP_DEVICE_EIRP_10.String(), should.Equal, "DEVICE_EIRP_10")

	b, err := DeviceEIRP_DEVICE_EIRP_10.MarshalText()
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte("DEVICE_EIRP_10"))

	var v DeviceEIRP
	err = v.UnmarshalText([]byte("DEVICE_EIRP_10"))
	a.So(v, should.Equal, DeviceEIRP_DEVICE_EIRP_10)
	a.So(err, should.BeNil)
}
