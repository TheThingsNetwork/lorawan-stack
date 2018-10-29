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

package band_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func GetUplink() ttnpb.UplinkMessage                   { return ttnpb.UplinkMessage{} }
func SendDownlink(ttnpb.DownlinkMessage)               {}
func ParseSpreadingFactor(types.DataRate) uint32       { return 7 }
func ParseBandwidth(types.DataRate) uint32             { return 125000 }
func ParseBitRate(types.DataRate) uint32               { return 0 }
func ParseModulation(types.DataRate) ttnpb.Modulation  { return ttnpb.Modulation_LORA }
func GetReceptionTimestamp(ttnpb.UplinkMessage) uint64 { return 0 }

func Example() {
	euBand, err := band.GetByID(band.EU_863_870)
	if err != nil {
		panic(err)
	}

	up := GetUplink()
	sets := up.GetSettings()
	drIdx, err := euBand.Rx1DataRate(sets.GetDataRateIndex(), 0, false)
	if err != nil {
		panic(err)
	}

	chIdx, err := euBand.Rx1Channel(sets.GetChannelIndex())
	if err != nil {
		panic(err)
	}

	dr := euBand.DataRates[drIdx]

	downlink := ttnpb.DownlinkMessage{
		Settings: ttnpb.TxSettings{
			DataRateIndex:   ttnpb.DataRateIndex(drIdx),
			Frequency:       euBand.DownlinkChannels[chIdx].Frequency,
			ChannelIndex:    chIdx,
			Modulation:      ParseModulation(dr.Rate),
			SpreadingFactor: ParseSpreadingFactor(dr.Rate),
			BitRate:         ParseBitRate(dr.Rate),
			Bandwidth:       ParseBandwidth(dr.Rate),
		},
		TxMetadata: ttnpb.TxMetadata{
			Timestamp: GetReceptionTimestamp(up) + 1000000000*uint64(euBand.ReceiveDelay1),
		},
	}
	SendDownlink(downlink)
}

func TestRx1DataRate(t *testing.T) {
	a := assertions.New(t)

	for _, tc := range []struct {
		bandID string

		validIndexes []ttnpb.DataRateIndex
		validOffsets []uint32

		invalidIndexes []ttnpb.DataRateIndex
		invalidOffsets []uint32
	}{
		{
			bandID:       "AU_915_928",
			validIndexes: []ttnpb.DataRateIndex{0, 3, 5}, invalidIndexes: []ttnpb.DataRateIndex{8, 10},
			validOffsets: []uint32{0, 3, 4}, invalidOffsets: []uint32{10},
		},
		{
			bandID:       "AS_923",
			validIndexes: []ttnpb.DataRateIndex{0, 5, 11, 13},
			validOffsets: []uint32{0, 2, 7}, invalidOffsets: []uint32{8},
		},
		{
			bandID:       "CN_470_510",
			validIndexes: []ttnpb.DataRateIndex{0, 5}, invalidIndexes: []ttnpb.DataRateIndex{7, 10},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{8},
		},
		{
			bandID:       "CN_779_787",
			validIndexes: []ttnpb.DataRateIndex{0, 5}, invalidIndexes: []ttnpb.DataRateIndex{10},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{8},
		},
		{
			bandID:       "EU_433",
			validIndexes: []ttnpb.DataRateIndex{0, 5}, invalidIndexes: []ttnpb.DataRateIndex{10},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{8},
		},
		{
			bandID:       "EU_863_870",
			validIndexes: []ttnpb.DataRateIndex{0, 5, 7}, invalidIndexes: []ttnpb.DataRateIndex{8},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{6, 8},
		},
		{
			bandID:       "IN_865_867",
			validIndexes: []ttnpb.DataRateIndex{0, 5, 11, 13},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{6, 8},
		},
		{
			bandID:       "KR_920_923",
			validIndexes: []ttnpb.DataRateIndex{0, 5, 11, 13},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{6, 8},
		},
		{
			bandID:       "RU_864_870",
			validIndexes: []ttnpb.DataRateIndex{0, 5, 11, 13},
			validOffsets: []uint32{0, 2, 5}, invalidOffsets: []uint32{6, 8},
		},
		{
			bandID:       "US_902_928",
			validIndexes: []ttnpb.DataRateIndex{0, 4}, invalidIndexes: []ttnpb.DataRateIndex{5, 10},
			validOffsets: []uint32{0, 2, 3}, invalidOffsets: []uint32{4, 6, 8},
		},
	} {
		b, err := band.GetByID(tc.bandID)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Error when getting band %s: %s", tc.bandID, err)
		}

		for _, validIndex := range tc.validIndexes {
			for _, validOffset := range tc.validOffsets {
				_, err := b.Rx1DataRate(validIndex, validOffset, true)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Computing Rx1 data rate should have succeeded with index %d and offset %d", validIndex, validOffset)
				}
			}
		}
		for _, invalidIndex := range tc.invalidIndexes {
			for _, offset := range append(tc.validOffsets, tc.invalidOffsets...) {
				_, err := b.Rx1DataRate(invalidIndex, offset, true)
				if !a.So(err, should.NotBeNil) {
					t.Fatalf("Computing Rx1 data rate should not have succeeded with index %d and offset %d", invalidIndex, offset)
				}
			}
		}
		for _, index := range append(tc.validIndexes, tc.invalidIndexes...) {
			for _, invalidOffset := range tc.invalidOffsets {
				_, err := b.Rx1DataRate(index, invalidOffset, true)
				if !a.So(err, should.NotBeNil) {
					t.Fatalf("Computing Rx1 data rate should not have succeeded with index %d and offset %d", index, invalidOffset)
				}
			}
		}
	}
}

func TestChannelMasksBands(t *testing.T) {
	a := assertions.New(t)

	for _, b := range band.All {
		if !a.So(b.ParseChMask, should.NotBeNil) {
			t.Fatalf("Band %s should have a ChannelMask function defined", b.ID)
		}
	}
}
