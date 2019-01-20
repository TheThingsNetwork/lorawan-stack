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

package messages

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestGetDataRatesFromFrequencyPlan(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name           string
		BandID         string
		DataRates      DataRates
		ErrorAssertion func(error) bool
	}{
		{
			"InvalidBandID",
			"EU",
			DataRates{},
			errors.IsNotFound,
		},
		{
			"ValidBAndID",
			"EU_433",
			DataRates{
				[3]int{12, 125, 0},
				[3]int{11, 125, 0},
				[3]int{10, 125, 0},
				[3]int{9, 125, 0},
				[3]int{8, 125, 0},
				[3]int{7, 125, 0},
				[3]int{7, 250, 0},
			},
			nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			drs, err := getDataRatesFromBandID(tc.BandID)
			if err != nil && (tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(drs, should.Resemble, tc.DataRates)) {
				t.Fatalf("Invalid datarates: %v", drs)
			}
		})
	}
}

func TestGetUint16IntegerAsByteSlice(t *testing.T) {
	a := assertions.New(t)

	a.So(getUint16IntegerAsByteSlice(0x12), should.Resemble, [2]byte{0x12, 0})
	a.So(getUint16IntegerAsByteSlice(0x1234), should.Equal, [2]byte{0x34, 0x12})
}

func TestGetUint32IntegerAsByteSlice(t *testing.T) {
	a := assertions.New(t)

	b, err := getInt32IntegerAsByteSlice(0x12)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte{0x12, 0, 0, 0})

	b, err = getInt32IntegerAsByteSlice(0x12345678)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte{0x78, 0x56, 0x34, 0x12})

	b, err = getInt32IntegerAsByteSlice(0x7FFFFFFF)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte{0xFF, 0xFF, 0xFF, 0x7F})
}

func TestGetMHDR(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name          string
		MHDR          uint
		ParsedMHDR    ttnpb.MHDR
		ExpectedError error
	}{
		{
			"Valid/1",
			0x00,
			ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
			nil,
		},
		{
			"Valid/2",
			224,
			ttnpb.MHDR{MType: ttnpb.MType_PROPRIETARY, Major: ttnpb.Major_LORAWAN_R1},
			nil,
		},
		{
			"Valid/3",
			255,
			ttnpb.MHDR{MType: ttnpb.MType_PROPRIETARY, Major: 3},
			nil,
		},
		{
			"Invalid/1",
			256,
			ttnpb.MHDR{},
			errJoinRequestMessage,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			mhdr, err := getMHDRFromInt(tc.MHDR)
			if !(a.So(err, should.Resemble, tc.ExpectedError)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(mhdr, should.Resemble, tc.ParsedMHDR)) {
				t.Fatalf("Invalid mhdr: %v", mhdr)
			}
		})
	}
}

func TestGetDataRateFromDataRateIndex(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name             string
		BandID           string
		DataRateIndex    int
		ExpectedDataRate ttnpb.DataRate
		ExpectedError    error
	}{
		{
			"Valid_EU",
			"EU_863_870",
			0,
			ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 12,
				Bandwidth:       125000,
			}}},
			nil,
		},
		{
			"Valid_EU_FSK",
			"EU_863_870",
			7,
			ttnpb.DataRate{Modulation: &ttnpb.DataRate_FSK{FSK: &ttnpb.FSKDataRate{
				BitRate: 50000,
			}}},
			nil,
		},
		{
			"Invalid_EU",
			"EU_863_870",
			8,
			ttnpb.DataRate{},
			errDataRateIndex,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			dr, err := getDataRateFromDataRateIndex(tc.BandID, tc.DataRateIndex)
			if !(a.So(err, should.Resemble, tc.ExpectedError)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(dr, should.Resemble, tc.ExpectedDataRate)) {
				t.Fatalf("Invalid datarate: %v", dr)
			}
		})

	}
}
