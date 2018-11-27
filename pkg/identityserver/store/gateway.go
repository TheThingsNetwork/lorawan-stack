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

package store

import (
	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Gateway model.
type Gateway struct {
	Model
	SoftDelete

	GatewayEUI *EUI64 `gorm:"unique_index:eui;type:VARCHAR(16);column:gateway_eui"`

	// BEGIN common fields
	GatewayID   string `gorm:"unique_index:id;type:VARCHAR(36)"`
	Name        string `gorm:"type:VARCHAR"`
	Description string `gorm:"type:TEXT"`
	// END common fields

	GatewayServerAddress string `gorm:"type:VARCHAR"`
	AutoUpdate           bool
	UpdateChannel        string `gorm:"type:VARCHAR"`
	FrequencyPlanID      string `gorm:"type:VARCHAR"`
	ScheduleDownlinkLate bool
	StatusPublic         bool
	LocationPublic       bool
}

func init() {
	registerModel(&Gateway{})
}

// functions to set fields from the gateway model into the gateway proto.
var gatewayPBSetters = map[string]func(*ttnpb.Gateway, *Gateway){
	"ids.eui":                 func(pb *ttnpb.Gateway, gtw *Gateway) { pb.EUI = (*types.EUI64)(gtw.GatewayEUI) }, // can we do this?
	nameField:                 func(pb *ttnpb.Gateway, gtw *Gateway) { pb.Name = gtw.Name },
	descriptionField:          func(pb *ttnpb.Gateway, gtw *Gateway) { pb.Description = gtw.Description },
	gatewayServerAddressField: func(pb *ttnpb.Gateway, gtw *Gateway) { pb.GatewayServerAddress = gtw.GatewayServerAddress },
	autoUpdateField:           func(pb *ttnpb.Gateway, gtw *Gateway) { pb.AutoUpdate = gtw.AutoUpdate },
	updateChannelField:        func(pb *ttnpb.Gateway, gtw *Gateway) { pb.UpdateChannel = gtw.UpdateChannel },
	frequencyPlanIDField:      func(pb *ttnpb.Gateway, gtw *Gateway) { pb.FrequencyPlanID = gtw.FrequencyPlanID },
	scheduleDownlinkLateField: func(pb *ttnpb.Gateway, gtw *Gateway) { pb.ScheduleDownlinkLate = gtw.ScheduleDownlinkLate },
	statusPublicField:         func(pb *ttnpb.Gateway, gtw *Gateway) { pb.StatusPublic = gtw.StatusPublic },
	locationPublicField:       func(pb *ttnpb.Gateway, gtw *Gateway) { pb.LocationPublic = gtw.LocationPublic },
}

// functions to set fields from the gateway proto into the gateway model.
var gatewayModelSetters = map[string]func(*Gateway, *ttnpb.Gateway){
	"ids.eui":                 func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.GatewayEUI = (*EUI64)(pb.EUI) }, // can we do this?
	nameField:                 func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.Name = pb.Name },
	descriptionField:          func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.Description = pb.Description },
	gatewayServerAddressField: func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.GatewayServerAddress = pb.GatewayServerAddress },
	autoUpdateField:           func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.AutoUpdate = pb.AutoUpdate },
	updateChannelField:        func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.UpdateChannel = pb.UpdateChannel },
	frequencyPlanIDField:      func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.FrequencyPlanID = pb.FrequencyPlanID },
	scheduleDownlinkLateField: func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.ScheduleDownlinkLate = pb.ScheduleDownlinkLate },
	statusPublicField:         func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.StatusPublic = pb.StatusPublic },
	locationPublicField:       func(gtw *Gateway, pb *ttnpb.Gateway) { gtw.LocationPublic = pb.LocationPublic },
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultGatewayFieldMask = &pbtypes.FieldMask{}

func init() {
	paths := make([]string, 0, len(gatewayPBSetters))
	for path := range gatewayPBSetters {
		paths = append(paths, path)
	}
	defaultGatewayFieldMask.Paths = paths
}

// fieldmask path to column name in gateways table, if other than proto field.
var gatewayColumnNames = map[string]string{
	"ids.gateway_id": "gateway_id",
	"ids.eui":        "gateway_eui",
}

func (gtw Gateway) toPB(pb *ttnpb.Gateway, fieldMask *pbtypes.FieldMask) {
	pb.GatewayIdentifiers.GatewayID = gtw.GatewayID
	pb.CreatedAt = cleanTime(gtw.CreatedAt)
	pb.UpdatedAt = cleanTime(gtw.UpdatedAt)
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultGatewayFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := gatewayPBSetters[path]; ok {
			setter(pb, &gtw)
		}
	}
}

func (gtw *Gateway) fromPB(pb *ttnpb.Gateway, fieldMask *pbtypes.FieldMask) (columns []string) {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultGatewayFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := gatewayModelSetters[path]; ok {
			setter(gtw, pb)
			columnName, ok := gatewayColumnNames[path]
			if !ok {
				columnName = path
			}
			columns = append(columns, columnName)
			continue
		}
	}
	return
}
