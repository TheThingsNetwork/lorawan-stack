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
			Name:           "InvalidBandID",
			BandID:         "EU",
			DataRates:      DataRates{},
			ErrorAssertion: errors.IsNotFound,
		},
		{
			Name:   "ValidBAndID",
			BandID: "EU_433",
			DataRates: DataRates{
				[3]int{12, 125, 0},
				[3]int{11, 125, 0},
				[3]int{10, 125, 0},
				[3]int{9, 125, 0},
				[3]int{8, 125, 0},
				[3]int{7, 125, 0},
				[3]int{7, 250, 0},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			drs, err := getDataRatesFromBandID(tc.BandID)
			if err != nil && (tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !a.So(drs, should.Resemble, tc.DataRates) {
				t.Fatalf("Invalid datarates: %v", drs)
			}
		})
	}
}
func TestGetUint32IntegerAsByteSlice(t *testing.T) {
	a := assertions.New(t)

	b, err := getInt32AsByteSlice(0x12)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte{0x12, 0, 0, 0})

	b, err = getInt32AsByteSlice(0x12345678)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte{0x78, 0x56, 0x34, 0x12})

	b, err = getInt32AsByteSlice(0x7FFFFFFF)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte{0xFF, 0xFF, 0xFF, 0x7F})
}

func TestGetDataRateFromDataRateIndex(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		BandID           string
		DataRateIndex    int
		ExpectedDataRate ttnpb.DataRate
		ErrorAssertion   func(error) bool
	}{
		{
			Name:   "Valid_EU",
			BandID: "EU_863_870",
			ExpectedDataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 12,
				Bandwidth:       125000,
			}}},
		},
		{
			Name:          "Valid_EU_FSK",
			BandID:        "EU_863_870",
			DataRateIndex: 7,
			ExpectedDataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_FSK{FSK: &ttnpb.FSKDataRate{
				BitRate: 50000,
			}}},
		},
		{
			Name:             "Invalid_EU",
			BandID:           "EU_863_870",
			DataRateIndex:    16,
			ExpectedDataRate: ttnpb.DataRate{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errDataRateIndex)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			dr, err := getDataRateFromIndex(tc.BandID, tc.DataRateIndex)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				if !a.So(dr, should.Resemble, tc.ExpectedDataRate) {
					t.Fatalf("Invalid datarate: %v", dr)
				}
			}
		})

	}
}
