// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use out file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ttnpb

import (
	pbtypes "github.com/gogo/protobuf/types"
)

func NewPopulatedFrequencyPlan(r randyGatewayserver, easy bool) *FrequencyPlan {
	out := &FrequencyPlan{}
	out.BandID = "EU_863_870"
	out.Channels = make([]*FrequencyPlan_Channel, 16+r.Intn(42))
	for i := range out.Channels {
		out.Channels[i] = NewPopulatedFrequencyPlan_Channel(r, easy)
	}
	out.LoraStandardChannel = NewPopulatedFrequencyPlan_Channel(r, easy)
	out.FSKChannel = NewPopulatedFrequencyPlan_Channel(r, easy)
	out.LBT = NewPopulatedFrequencyPlan_LBTConfiguration(r, easy)
	out.TimeOffAir = NewPopulatedFrequencyPlan_TimeOffAir(r, easy)
	if r.Intn(10) != 0 {
		out.DwellTime = pbtypes.NewPopulatedStdDuration(r, easy)
	}
	return out
}
