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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestIsRepeatedUplink(t *testing.T) {
	for _, tc := range []struct {
		name     string
		this     *uplinkMessage
		that     *uplinkMessage
		repeated bool
	}{
		{
			name: "Repeated",
			this: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			that: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			repeated: true,
		},
		{
			name: "DifferentFrequency",
			this: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
			},
			that: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1100000,
			},
			repeated: false,
		},
		{
			name: "DifferentAntenna",
			this: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			that: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
				antennas:  []uint32{1, 2},
			},
			repeated: false,
		},
		{
			name: "DifferentPayload",
			this: &uplinkMessage{
				payload:   []byte{1, 2, 4},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			that: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			repeated: false,
		},
		{
			name: "DifferentPayloadSize",
			this: &uplinkMessage{
				payload:   []byte{1, 2},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			that: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
				antennas:  []uint32{1},
			},
			repeated: false,
		},
		{
			name: "NilMessage",
			this: nil,
			that: &uplinkMessage{
				payload:   []byte{1, 2, 3},
				frequency: 1000000,
			},
			repeated: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)

			a.So(isRepeatedUplink(tc.this, tc.that), should.Equal, tc.repeated)
			a.So(isRepeatedUplink(tc.that, tc.this), should.Equal, tc.repeated)
		})
	}
}

func TestUplinkMessageFromProto(t *testing.T) {
	a := assertions.New(t)
	a.So(uplinkMessageFromProto(&ttnpb.UplinkMessage{
		RawPayload: []byte{1, 2, 3},
		Settings:   ttnpb.TxSettings{Frequency: 100000},
		RxMetadata: []*ttnpb.RxMetadata{{AntennaIndex: 0}, {AntennaIndex: 3}},
	}), should.Resemble, &uplinkMessage{
		payload:   []byte{1, 2, 3},
		frequency: 100000,
		antennas:  []uint32{0, 3},
	})
}
