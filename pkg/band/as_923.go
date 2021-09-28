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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	// AS_923 is the ID of the Asian 923Mhz band
	AS_923 = "AS_923"

	as923BeaconFrequency = 923400000
)

var as923DefaultChannels = []Channel{
	{
		Frequency:   923200000,
		MaxDataRate: ttnpb.DATA_RATE_5,
	},
	{
		Frequency:   923400000,
		MaxDataRate: ttnpb.DATA_RATE_5,
	},
}
