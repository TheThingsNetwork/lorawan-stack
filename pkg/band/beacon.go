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

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Beacon parameters of a specific band.
type Beacon struct {
	DataRateIndex ttnpb.DataRateIndex
	CodingRate    string
	Frequencies   []uint64
}

// ComputeFrequency computes the frequency of the beacon at time t.
func (b Beacon) ComputeFrequency(t time.Time) uint64 {
	switch n := len(b.Frequencies); n {
	case 0:
		panic("no frequencies available")
	case 1:
		return b.Frequencies[0]
	default:
		beacon := gpstime.ToGPS(t) / 128
		return b.Frequencies[int(beacon)%n]
	}
}

var usAuBeaconFrequencies = func() []uint64 {
	freqs := make([]uint64, 8)
	for i := 0; i < 8; i++ {
		freqs[i] = 923300000 + uint64(i*600000)
	}
	return freqs
}()
