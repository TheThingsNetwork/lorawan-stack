// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package store

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

// GatewayAntenna model.
type GatewayAntenna struct {
	Model

	Gateway   *Gateway
	GatewayID string `gorm:"type:UUID;unique_index:gateway_antenna_id_index;index:gateway_antenna_gateway_index;not null"`
	Index     int    `gorm:"unique_index:gateway_antenna_id_index;not null"`

	Attributes []Attribute `gorm:"polymorphic:Entity;polymorphic_value:gateway"`

	Gain float32

	Location
}

func init() {
	registerModel(&GatewayAntenna{})
}

func (a GatewayAntenna) toPB() ttnpb.GatewayAntenna {
	return ttnpb.GatewayAntenna{
		Gain: a.Gain,
		Location: ttnpb.Location{
			Latitude:  a.Latitude,
			Longitude: a.Longitude,
			Altitude:  a.Altitude,
			Accuracy:  a.Accuracy,
			Source:    ttnpb.SOURCE_REGISTRY,
		},
		Attributes: attributes(a.Attributes).toMap(),
	}
}

func (a *GatewayAntenna) fromPB(pb ttnpb.GatewayAntenna) {
	a.Gain = pb.Gain
	a.Location = Location{
		Latitude:  pb.Location.Latitude,
		Longitude: pb.Location.Longitude,
		Altitude:  pb.Location.Altitude,
		Accuracy:  pb.Location.Accuracy,
	}
	a.Attributes = attributes(a.Attributes).updateFromMap(pb.Attributes)
}
