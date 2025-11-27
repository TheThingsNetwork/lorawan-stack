// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

// Package band implements LoRaWAN frequency band definitions.
package band

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

const (
	// UZ_923_DRAFT is the ID of the draft Uzbekistan 923Mhz band.
	UZ_923_DRAFT = "UZ_923_DRAFT" // nolint: revive,staticcheck
)

var (
	uz923DraftBeaconFrequencies = []uint64{926600000}
	uz923DraftDefaultChannels   = []Channel{
		{
			Frequency:   926400000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Frequency:   926600000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Frequency:   926800000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
	}

	uz923DraftDownlinkDRTable = [6][6]ttnpb.DataRateIndex{
		{0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0},
		{2, 1, 0, 0, 0, 0},
		{3, 2, 1, 0, 0, 0},
		{4, 3, 2, 1, 0, 0},
		{5, 4, 3, 2, 1, 0},
	}
)
