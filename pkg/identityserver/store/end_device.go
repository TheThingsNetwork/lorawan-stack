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

import (
	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// EndDevice model.
type EndDevice struct {
	Model

	ApplicationID string `gorm:"unique_index:end_device_id_index;type:VARCHAR(36);not null;index:end_device_application_index"`
	Application   *Application

	// BEGIN common fields
	DeviceID    string      `gorm:"unique_index:end_device_id_index;type:VARCHAR(36);not null"`
	Name        string      `gorm:"type:VARCHAR"`
	Description string      `gorm:"type:TEXT"`
	Attributes  []Attribute `gorm:"polymorphic:Entity;polymorphic_value:device"`
	// END common fields

	JoinEUI *EUI64 `gorm:"unique_index:end_device_eui_index;index:end_device_join_eui_index;type:VARCHAR(16);column:join_eui"`
	DevEUI  *EUI64 `gorm:"unique_index:end_device_eui_index;index:end_device_dev_eui_index;type:VARCHAR(16);column:dev_eui"`

	BrandID         string `gorm:"type:VARCHAR"`
	ModelID         string `gorm:"type:VARCHAR"`
	HardwareVersion string `gorm:"type:VARCHAR"`
	FirmwareVersion string `gorm:"type:VARCHAR"`

	NetworkServerAddress     string `gorm:"type:VARCHAR"`
	ApplicationServerAddress string `gorm:"type:VARCHAR"`
	JoinServerAddress        string `gorm:"type:VARCHAR"`

	ServiceProfileID string `gorm:"type:VARCHAR"`

	Locations []EndDeviceLocation
}

func init() {
	registerModel(&EndDevice{})
}

func mustEndDeviceVersionIDs(pb *ttnpb.EndDevice) *ttnpb.EndDeviceVersionIdentifiers {
	if pb.VersionIDs == nil {
		pb.VersionIDs = &ttnpb.EndDeviceVersionIdentifiers{}
	}
	return pb.VersionIDs
}

// functions to set fields from the device model into the device proto.
var devicePBSetters = map[string]func(*ttnpb.EndDevice, *EndDevice){
	"ids.join_eui":   func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.JoinEUI = dev.JoinEUI.toPB() },
	"ids.dev_eui":    func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.DevEUI = dev.DevEUI.toPB() },
	nameField:        func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.Name = dev.Name },
	descriptionField: func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.Description = dev.Description },
	attributesField:  func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.Attributes = attributes(dev.Attributes).toMap() },
	versionIDsField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.VersionIDs = &ttnpb.EndDeviceVersionIdentifiers{
			BrandID:         dev.BrandID,
			ModelID:         dev.ModelID,
			HardwareVersion: dev.HardwareVersion,
			FirmwareVersion: dev.FirmwareVersion,
		}
	},
	brandIDField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).BrandID = dev.BrandID
	},
	modelIDField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).ModelID = dev.ModelID
	},
	hardwareVersionField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).HardwareVersion = dev.HardwareVersion
	},
	firmwareVersionField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).FirmwareVersion = dev.FirmwareVersion
	},
	networkServerAddressField:     func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.NetworkServerAddress = dev.NetworkServerAddress },
	applicationServerAddressField: func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.ApplicationServerAddress = dev.ApplicationServerAddress },
	joinServerAddressField:        func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.JoinServerAddress = dev.JoinServerAddress },
	serviceProfileIDField:         func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.ServiceProfileID = dev.ServiceProfileID },
	locationsField:                func(pb *ttnpb.EndDevice, dev *EndDevice) { pb.Locations = deviceLocations(dev.Locations).toMap() },
}

// functions to set fields from the device proto into the device model.
var deviceModelSetters = map[string]func(*EndDevice, *ttnpb.EndDevice){
	"ids.join_eui":   func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.JoinEUI = eui(pb.JoinEUI) },
	"ids.dev_eui":    func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.DevEUI = eui(pb.DevEUI) },
	nameField:        func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.Name = pb.Name },
	descriptionField: func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.Description = pb.Description },
	attributesField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.Attributes = attributes(dev.Attributes).updateFromMap(pb.Attributes)
	},
	versionIDsField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.BrandID = pb.GetVersionIDs().GetBrandID()
		dev.ModelID = pb.GetVersionIDs().GetModelID()
		dev.HardwareVersion = pb.GetVersionIDs().GetHardwareVersion()
		dev.FirmwareVersion = pb.GetVersionIDs().GetFirmwareVersion()
	},
	brandIDField: func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.BrandID = pb.GetVersionIDs().GetBrandID() },
	modelIDField: func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.ModelID = pb.GetVersionIDs().GetModelID() },
	hardwareVersionField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.HardwareVersion = pb.GetVersionIDs().GetHardwareVersion()
	},
	firmwareVersionField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.FirmwareVersion = pb.GetVersionIDs().GetFirmwareVersion()
	},
	networkServerAddressField:     func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.NetworkServerAddress = pb.NetworkServerAddress },
	applicationServerAddressField: func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.ApplicationServerAddress = pb.ApplicationServerAddress },
	joinServerAddressField:        func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.JoinServerAddress = pb.JoinServerAddress },
	serviceProfileIDField:         func(dev *EndDevice, pb *ttnpb.EndDevice) { dev.ServiceProfileID = pb.ServiceProfileID },
	locationsField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.Locations = deviceLocations(dev.Locations).updateFromMap(pb.Locations)
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultEndDeviceFieldMask = &types.FieldMask{}

func init() {
	paths := make([]string, 0, len(devicePBSetters))
	for _, path := range ttnpb.EndDeviceFieldPathsNested {
		if _, ok := devicePBSetters[path]; ok {
			paths = append(paths, path)
		}
	}
	defaultEndDeviceFieldMask.Paths = paths
}

// fieldmask path to column name in devices table.
var deviceColumnNames = map[string][]string{
	"ids.join_eui":                {"join_eui"},
	"ids.dev_eui":                 {"dev_eui"},
	attributesField:               {},
	nameField:                     {nameField},
	descriptionField:              {descriptionField},
	versionIDsField:               {"brand_id", "model_id", "hardware_version", "firmware_version"},
	brandIDField:                  {"brand_id"},
	modelIDField:                  {"model_id"},
	hardwareVersionField:          {"hardware_version"},
	firmwareVersionField:          {"firmware_version"},
	networkServerAddressField:     {networkServerAddressField},
	applicationServerAddressField: {applicationServerAddressField},
	joinServerAddressField:        {joinServerAddressField},
	serviceProfileIDField:         {serviceProfileIDField},
	locationsField:                {},
}

func (dev EndDevice) toPB(pb *ttnpb.EndDevice, fieldMask *types.FieldMask) {
	pb.EndDeviceIdentifiers.ApplicationID = dev.ApplicationID
	pb.EndDeviceIdentifiers.DeviceID = dev.DeviceID
	pb.EndDeviceIdentifiers.JoinEUI = dev.JoinEUI.toPB() // Always present.
	pb.EndDeviceIdentifiers.DevEUI = dev.DevEUI.toPB()   // Always present.
	pb.CreatedAt = cleanTime(dev.CreatedAt)
	pb.UpdatedAt = cleanTime(dev.UpdatedAt)
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultEndDeviceFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := devicePBSetters[path]; ok {
			setter(pb, &dev)
		}
	}
}

func (dev *EndDevice) fromPB(pb *ttnpb.EndDevice, fieldMask *types.FieldMask) (columns []string) {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultEndDeviceFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := deviceModelSetters[path]; ok {
			setter(dev, pb)
			if columnNames, ok := deviceColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			}
			continue
		}
	}
	return
}
