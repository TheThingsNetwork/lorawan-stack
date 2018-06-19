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

package ttnpb

import pbtypes "github.com/gogo/protobuf/types"

func NewPopulatedGateway(r randyGateway, easy bool) *Gateway {
	out := &Gateway{}
	v1 := NewPopulatedGatewayIdentifiers(r, easy)
	out.GatewayIdentifiers = *v1
	out.Description = randStringGateway(r)
	out.FrequencyPlanID = NewPopulatedID(r)
	out.ClusterAddress = randStringGateway(r)
	if r.Intn(10) != 0 {
		v2 := r.Intn(5)
		out.Antennas = make([]GatewayAntenna, v2)
		for i := 0; i < v2; i++ {
			v3 := NewPopulatedGatewayAntenna(r, easy)
			out.Antennas[i] = *v3
		}
	}
	if r.Intn(10) != 0 {
		v4 := r.Intn(5)
		out.Radios = make([]GatewayRadio, v4)
		for i := 0; i < v4; i++ {
			v5 := NewPopulatedGatewayRadio(r, easy)
			out.Radios[i] = *v5
		}
	}
	if r.Intn(10) != 0 {
		out.ActivatedAt = pbtypes.NewPopulatedStdTime(r, easy)
	}
	v6 := NewPopulatedGatewayPrivacySettings(r, easy)
	out.PrivacySettings = *v6
	out.AutoUpdate = bool(r.Intn(2) == 0)
	out.Platform = randStringGateway(r)
	if r.Intn(10) != 0 {
		v7 := r.Intn(10)
		out.Attributes = make(map[string]string)
		for i := 0; i < v7; i++ {
			out.Attributes[randStringGateway(r)] = randStringGateway(r)
		}
	}
	if r.Intn(10) != 0 {
		out.ContactAccountIDs = NewPopulatedOrganizationOrUserIdentifiers(r, easy)
	}
	v8 := pbtypes.NewPopulatedStdTime(r, easy)
	out.CreatedAt = *v8
	v9 := pbtypes.NewPopulatedStdTime(r, easy)
	out.UpdatedAt = *v9
	out.DisableTxDelay = bool(r.Intn(2) == 0)
	if !easy && r.Intn(10) != 0 {
	}
	return out
}
