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
	this := &Gateway{}
	v1 := NewPopulatedGatewayIdentifiers(r, easy)
	this.GatewayIdentifiers = *v1
	this.Description = randStringGateway(r)
	this.FrequencyPlanID = NewPopulatedID(r)
	this.ClusterAddress = randStringGateway(r)
	if r.Intn(10) != 0 {
		v2 := r.Intn(5)
		this.Antennas = make([]GatewayAntenna, v2)
		for i := 0; i < v2; i++ {
			v3 := NewPopulatedGatewayAntenna(r, easy)
			this.Antennas[i] = *v3
		}
	}
	if r.Intn(10) != 0 {
		v4 := r.Intn(5)
		this.Radios = make([]GatewayRadio, v4)
		for i := 0; i < v4; i++ {
			v5 := NewPopulatedGatewayRadio(r, easy)
			this.Radios[i] = *v5
		}
	}
	if r.Intn(10) != 0 {
		this.ActivatedAt = pbtypes.NewPopulatedStdTime(r, easy)
	}
	v6 := NewPopulatedGatewayPrivacySettings(r, easy)
	this.PrivacySettings = *v6
	this.AutoUpdate = bool(r.Intn(2) == 0)
	this.Platform = randStringGateway(r)
	if r.Intn(10) != 0 {
		v7 := r.Intn(10)
		this.Attributes = make(map[string]string)
		for i := 0; i < v7; i++ {
			this.Attributes[randStringGateway(r)] = randStringGateway(r)
		}
	}
	if r.Intn(10) != 0 {
		this.ContactAccountIDs = NewPopulatedOrganizationOrUserIdentifiers(r, easy)
	}
	v8 := pbtypes.NewPopulatedStdTime(r, easy)
	this.CreatedAt = *v8
	v9 := pbtypes.NewPopulatedStdTime(r, easy)
	this.UpdatedAt = *v9
	this.DisableTxDelay = bool(r.Intn(2) == 0)
	if !easy && r.Intn(10) != 0 {
	}
	return this
}
