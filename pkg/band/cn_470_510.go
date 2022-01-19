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

package band

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

const (
	// CN_470_510 is the ID of the Chinese 470-510Mhz band
	CN_470_510 = "CN_470_510"
)

var (
	cn470510DownlinkDRTable = [6][6]ttnpb.DataRateIndex{
		{0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0},
		{2, 1, 0, 0, 0, 0},
		{3, 2, 1, 0, 0, 0},
		{4, 3, 2, 1, 0, 0},
		{5, 4, 3, 2, 1, 0},
	}

	cn470510UplinkChannels = func() []Channel {
		uplinkChannels := make([]Channel, 0, 96)
		for i := 0; i < 96; i++ {
			uplinkChannels = append(uplinkChannels, Channel{
				Frequency:   uint64(470300000 + 200000*i),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
			})
		}
		return uplinkChannels
	}()

	cn470510DownlinkChannels = func() []Channel {
		downlinkChannels := make([]Channel, 0, 48)
		for i := 0; i < 48; i++ {
			downlinkChannels = append(downlinkChannels, Channel{
				Frequency:   uint64(500300000 + 200000*i),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
			})
		}
		return downlinkChannels
	}()

	cn470510BeaconFrequencies = func() [8]uint64 {
		var beaconFrequencies [8]uint64
		for i := 0; i < 8; i++ {
			beaconFrequencies[i] = 508300000 + uint64(i*200000)
		}
		return beaconFrequencies
	}()
)
