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

package band_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRx1DataRate(t *testing.T) {
	for _, tc := range []struct {
		bandID string

		validIndexes []ttnpb.DataRateIndex
		validOffsets []uint32

		invalidIndexes []ttnpb.DataRateIndex
		invalidOffsets []uint32
	}{
		{
			bandID:       "AU_915_928",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6}, invalidIndexes: []ttnpb.DataRateIndex{7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "AS_923",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5, 6, 7}, invalidOffsets: []uint32{8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "CN_470_510",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5}, invalidIndexes: []ttnpb.DataRateIndex{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "CN_779_787",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7}, invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "EU_433",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7}, invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "EU_863_870",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7}, invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "IN_865_867",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5, 6, 7}, invalidOffsets: []uint32{8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "KR_920_923",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5}, invalidIndexes: []ttnpb.DataRateIndex{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "RU_864_870",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7}, invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3, 4, 5}, invalidOffsets: []uint32{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			bandID:       "US_902_928",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4}, invalidIndexes: []ttnpb.DataRateIndex{5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []uint32{0, 1, 2, 3}, invalidOffsets: []uint32{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
	} {
		t.Run(tc.bandID, func(t *testing.T) {
			a := assertions.New(t)

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
		})
	}
}

func TestParseChMaskBands(t *testing.T) {
	a := assertions.New(t)

	for _, b := range band.All {
		if !a.So(b.ParseChMask, should.NotBeNil) {
			t.Fatalf("Band %s should have a ParseChMask function defined", b.ID)
		}
	}
}

func TestGenerateChMasksBands(t *testing.T) {
	a := assertions.New(t)

	for _, b := range band.All {
		if !a.So(b.GenerateChMasks, should.NotBeNil) {
			t.Fatalf("Band %s should have a GenerateChMasks function defined", b.ID)
		}
	}
}

func TestFindSubBand(t *testing.T) {
	for _, b := range band.All {
		t.Run(b.ID, func(t *testing.T) {
			a := assertions.New(t)
			for _, ch := range b.UplinkChannels {
				sb, ok := b.FindSubBand(ch.Frequency)
				if !a.So(ok, should.BeTrue) {
					t.Fatalf("Frequency %d not found in sub-bands", ch.Frequency)
				}
				a.So(sb.MinFrequency, should.BeLessThanOrEqualTo, ch.Frequency)
				a.So(sb.MaxFrequency, should.BeGreaterThanOrEqualTo, ch.Frequency)
			}
		})
	}
}
