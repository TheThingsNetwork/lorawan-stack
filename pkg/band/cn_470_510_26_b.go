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
	// CN_470_510_26_B is the ID of the Chinese 470-510Mhz band which uses a 26 MHz antenna, type B
	CN_470_510_26_B = "CN_470_510_26_B"
)

var (
	cn47051026BUplinkChannels = func() []Channel {
		uplinkChannels := make([]Channel, 0, 48)
		// 26 MHz Type B
		for i := 0; i < 48; i++ {
			uplinkChannels = append(uplinkChannels, Channel{
				Frequency:   uint64(480300000 + 200000*i),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
			})
		}
		return uplinkChannels
	}()

	cn47051026BDownlinkChannels = func() []Channel {
		downlinkChannels := make([]Channel, 0, 24)
		// 26 MHz Type B
		for i := 0; i < 24; i++ {
			downlinkChannels = append(downlinkChannels, Channel{
				Frequency:   uint64(500100000 + 200000*i),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
			})
		}
		return downlinkChannels
	}()
)
