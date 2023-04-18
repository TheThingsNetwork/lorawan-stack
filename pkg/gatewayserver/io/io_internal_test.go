// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package io

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestIsRepeatedUplink(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		this     *uplinkMessage
		that     *uplinkMessage
		repeated bool
	}{
		{
			name: "Repeated",
			this: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
				antennas:      []uint32{1},
			},
			that: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
				antennas:      []uint32{1},
			},
			repeated: true,
		},
		{
			name: "DifferentFrequency",
			this: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
			},
			that: &uplinkMessage{
				payloadHash:   123,
				frequency:     1100000,
				dataRateIndex: 234,
			},
			repeated: false,
		},
		{
			name: "DifferentDataRate",
			this: &uplinkMessage{
				payloadHash:   123,
				frequency:     100000,
				dataRateIndex: 234,
			},
			that: &uplinkMessage{
				payloadHash:   123,
				frequency:     100000,
				dataRateIndex: 235,
			},
			repeated: false,
		},
		{
			name: "DifferentAntenna",
			this: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
				antennas:      []uint32{1},
			},
			that: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
				antennas:      []uint32{1, 2},
			},
			repeated: false,
		},
		{
			name: "DifferentPayload",
			this: &uplinkMessage{
				payloadHash:   124,
				frequency:     1000000,
				dataRateIndex: 234,
				antennas:      []uint32{1},
			},
			that: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
				antennas:      []uint32{1},
			},
			repeated: false,
		},
		{
			name: "NilMessage",
			this: nil,
			that: &uplinkMessage{
				payloadHash:   123,
				frequency:     1000000,
				dataRateIndex: 234,
			},
			repeated: false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			a.So(isRepeatedUplink(tc.this, tc.that), should.Equal, tc.repeated)
			a.So(isRepeatedUplink(tc.that, tc.this), should.Equal, tc.repeated)
		})
	}
}

func TestUplinkMessageFromProto(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	phy, err := band.GetLatest(band.EU_863_870)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	a.So(uplinkMessageFromProto(&ttnpb.UplinkMessage{
		RawPayload: []byte{1, 2, 3},
		Settings:   &ttnpb.TxSettings{Frequency: 100000, DataRate: phy.DataRates[ttnpb.DataRateIndex_DATA_RATE_1].Rate},
		RxMetadata: []*ttnpb.RxMetadata{{AntennaIndex: 0}, {AntennaIndex: 3}},
	}, &phy), should.Resemble, &uplinkMessage{
		payloadHash:   15035938162879559083,
		frequency:     100000,
		dataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
		antennas:      []uint32{0, 3},
	})
}
