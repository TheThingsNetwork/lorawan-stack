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

// Channel abstracts a band's channel properties.
type Channel struct {
	// Frequency indicates the frequency of the channel.
	Frequency uint64
	// MinDataRate indicates the index of the minimal data rates accepted on this channel.
	MinDataRate ttnpb.DataRateIndex
	// MinDataRate indicates the index of the maximal data rates accepted on this channel.
	MaxDataRate ttnpb.DataRateIndex
}

func channelIndexIdentity(idx uint8) (uint8, error) {
	return idx, nil
}

func channelIndexModulo(n uint8) func(uint8) (uint8, error) {
	return func(idx uint8) (uint8, error) {
		return idx % n, nil
	}
}
