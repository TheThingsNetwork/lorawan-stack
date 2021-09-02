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

// Beacon parameters of a specific band.
type Beacon struct {
	DataRateIndex    ttnpb.DataRateIndex
	CodingRate       string
	InvertedPolarity bool
	// Channel returns in Hz on which beaconing is performed.
	//
	// beaconTime is the integer value, converted in float64, of the 4 bytes “Time” field of the beacon frame.
	ComputeFrequency func(beaconTime float64) uint64
}
