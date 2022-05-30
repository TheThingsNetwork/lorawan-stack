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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// EndDevice model.
type EndDevice struct {
	Model

	ApplicationID string `gorm:"unique_index:end_device_id_index;type:VARCHAR(36);not null;index:end_device_application_index"` //nolint:lll
	Application   *Application

	// BEGIN common fields
	DeviceID    string      `gorm:"unique_index:end_device_id_index;type:VARCHAR(36);not null"`
	Name        string      `gorm:"type:VARCHAR"`
	Description string      `gorm:"type:TEXT"`
	Attributes  []Attribute `gorm:"polymorphic:Entity;polymorphic_value:device"`
	// END common fields

	JoinEUI *EUI64 `gorm:"unique_index:end_device_eui_index;index:end_device_join_eui_index;type:VARCHAR(16);column:join_eui"` //nolint:lll
	DevEUI  *EUI64 `gorm:"unique_index:end_device_eui_index;index:end_device_dev_eui_index;type:VARCHAR(16);column:dev_eui"`   //nolint:lll

	BrandID         string `gorm:"type:VARCHAR"`
	ModelID         string `gorm:"type:VARCHAR"`
	HardwareVersion string `gorm:"type:VARCHAR"`
	FirmwareVersion string `gorm:"type:VARCHAR"`
	BandID          string `gorm:"type:VARCHAR"`

	NetworkServerAddress     string `gorm:"type:VARCHAR"`
	ApplicationServerAddress string `gorm:"type:VARCHAR"`
	JoinServerAddress        string `gorm:"type:VARCHAR"`

	ServiceProfileID string `gorm:"type:VARCHAR"`

	Locations []EndDeviceLocation

	Picture   *Picture
	PictureID *string `gorm:"type:UUID;index:end_device_picture_index"`

	ActivatedAt *time.Time `gorm:"default:null"`
	LastSeenAt  *time.Time `gorm:"default:null"`

	ClaimAuthenticationCodeSecret    []byte `gorm:"type:BYTEA"`
	ClaimAuthenticationCodeValidFrom *time.Time
	ClaimAuthenticationCodeValidTo   *time.Time
}

func init() {
	registerModel(&EndDevice{})
}

func mustEndDeviceVersionIDs(pb *ttnpb.EndDevice) *ttnpb.EndDeviceVersionIdentifiers {
	if pb.VersionIds == nil {
		pb.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
	}
	return pb.VersionIds
}

// functions to set fields from the device model into the device proto.
var devicePBSetters = map[string]func(*ttnpb.EndDevice, *EndDevice){
	"ids.join_eui": func(pb *ttnpb.EndDevice, dev *EndDevice) {
		if pb.Ids == nil {
			pb.Ids = &ttnpb.EndDeviceIdentifiers{}
		}
		pb.Ids.JoinEui = nil
		if joinEUI := dev.JoinEUI.toPB(); joinEUI != nil {
			pb.Ids.JoinEui = joinEUI.Bytes()
		}
	},
	"ids.dev_eui": func(pb *ttnpb.EndDevice, dev *EndDevice) {
		if pb.Ids == nil {
			pb.Ids = &ttnpb.EndDeviceIdentifiers{}
		}
		pb.Ids.DevEui = nil
		if devEUI := dev.DevEUI.toPB(); devEUI != nil {
			pb.Ids.DevEui = devEUI.Bytes()
		}
	},
	nameField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.Name = dev.Name
	},
	descriptionField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.Description = dev.Description
	},
	attributesField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.Attributes = attributes(dev.Attributes).toMap()
	},
	versionIDsField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{
			BrandId:         dev.BrandID,
			ModelId:         dev.ModelID,
			HardwareVersion: dev.HardwareVersion,
			FirmwareVersion: dev.FirmwareVersion,
			BandId:          dev.BandID,
		}
	},
	brandIDField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).BrandId = dev.BrandID
	},
	modelIDField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).ModelId = dev.ModelID
	},
	hardwareVersionField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).HardwareVersion = dev.HardwareVersion
	},
	firmwareVersionField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).FirmwareVersion = dev.FirmwareVersion
	},
	bandIDField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		mustEndDeviceVersionIDs(pb).BandId = dev.BandID
	},
	networkServerAddressField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.NetworkServerAddress = dev.NetworkServerAddress
	},
	applicationServerAddressField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.ApplicationServerAddress = dev.ApplicationServerAddress
	},
	joinServerAddressField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.JoinServerAddress = dev.JoinServerAddress
	},
	serviceProfileIDField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.ServiceProfileId = dev.ServiceProfileID
	},
	locationsField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.Locations = deviceLocations(dev.Locations).toMap()
	},
	pictureField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		if dev.Picture == nil {
			pb.Picture = nil
		} else {
			pb.Picture = dev.Picture.toPB()
		}
	},
	activatedAtField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.ActivatedAt = ttnpb.ProtoTime(dev.ActivatedAt)
	},
	lastSeenAtField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.LastSeenAt = ttnpb.ProtoTime(dev.LastSeenAt)
	},
	claimAuthenticationCodeField: func(pb *ttnpb.EndDevice, dev *EndDevice) {
		pb.ClaimAuthenticationCode = &ttnpb.EndDeviceAuthenticationCode{
			Value:     string(dev.ClaimAuthenticationCodeSecret),
			ValidFrom: ttnpb.ProtoTime(dev.ClaimAuthenticationCodeValidFrom),
			ValidTo:   ttnpb.ProtoTime(dev.ClaimAuthenticationCodeValidTo),
		}
	},
}

// functions to set fields from the device proto into the device model.
var deviceModelSetters = map[string]func(*EndDevice, *ttnpb.EndDevice){
	"ids.join_eui": func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.JoinEUI = eui(types.MustEUI64(pb.GetIds().GetJoinEui()))
	},
	"ids.dev_eui": func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.DevEUI = eui(types.MustEUI64(pb.GetIds().GetDevEui()))
	},
	nameField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.Name = pb.Name
	},
	descriptionField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.Description = pb.Description
	},
	attributesField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.Attributes = attributes(dev.Attributes).updateFromMap(pb.Attributes)
	},
	versionIDsField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.BrandID = pb.GetVersionIds().GetBrandId()
		dev.ModelID = pb.GetVersionIds().GetModelId()
		dev.HardwareVersion = pb.GetVersionIds().GetHardwareVersion()
		dev.FirmwareVersion = pb.GetVersionIds().GetFirmwareVersion()
		dev.BandID = pb.GetVersionIds().GetBandId()
	},
	brandIDField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.BrandID = pb.GetVersionIds().GetBrandId()
	},
	modelIDField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.ModelID = pb.GetVersionIds().GetModelId()
	},
	hardwareVersionField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.HardwareVersion = pb.GetVersionIds().GetHardwareVersion()
	},
	firmwareVersionField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.FirmwareVersion = pb.GetVersionIds().GetFirmwareVersion()
	},
	bandIDField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.BandID = pb.GetVersionIds().GetBandId()
	},
	networkServerAddressField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.NetworkServerAddress = pb.NetworkServerAddress
	},
	applicationServerAddressField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.ApplicationServerAddress = pb.ApplicationServerAddress
	},
	joinServerAddressField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.JoinServerAddress = pb.JoinServerAddress
	},
	serviceProfileIDField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.ServiceProfileID = pb.ServiceProfileId
	},
	locationsField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.Locations = deviceLocations(dev.Locations).updateFromMap(pb.Locations)
	},
	pictureField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.PictureID, dev.Picture = nil, nil
		if pb.Picture != nil {
			dev.Picture = &Picture{}
			dev.Picture.fromPB(pb.Picture)
		}
	},
	activatedAtField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.ActivatedAt = ttnpb.StdTime(pb.ActivatedAt)
	},
	lastSeenAtField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		dev.LastSeenAt = ttnpb.StdTime(pb.LastSeenAt)
	},
	claimAuthenticationCodeField: func(dev *EndDevice, pb *ttnpb.EndDevice) {
		if pb.ClaimAuthenticationCode != nil {
			dev.ClaimAuthenticationCodeSecret = []byte(pb.ClaimAuthenticationCode.Value)
			if pb.ClaimAuthenticationCode.ValidFrom != nil {
				dev.ClaimAuthenticationCodeValidFrom = ttnpb.StdTime(pb.ClaimAuthenticationCode.ValidFrom)
			}
			if pb.ClaimAuthenticationCode.ValidTo != nil {
				dev.ClaimAuthenticationCodeValidTo = ttnpb.StdTime(pb.ClaimAuthenticationCode.ValidTo)
			}
		}
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultEndDeviceFieldMask store.FieldMask

func init() {
	paths := make([]string, 0, len(devicePBSetters))
	for _, path := range ttnpb.EndDeviceFieldPathsNested {
		if _, ok := devicePBSetters[path]; ok {
			paths = append(paths, path)
		}
	}
	defaultEndDeviceFieldMask = paths
}

// fieldmask path to column name in devices table.
var deviceColumnNames = map[string][]string{
	"ids.join_eui":                {"join_eui"},
	"ids.dev_eui":                 {"dev_eui"},
	attributesField:               {},
	nameField:                     {nameField},
	descriptionField:              {descriptionField},
	versionIDsField:               {"brand_id", "model_id", "hardware_version", "firmware_version", "band_id"},
	brandIDField:                  {"brand_id"},
	modelIDField:                  {"model_id"},
	bandIDField:                   {"band_id"},
	hardwareVersionField:          {"hardware_version"},
	firmwareVersionField:          {"firmware_version"},
	networkServerAddressField:     {networkServerAddressField},
	applicationServerAddressField: {applicationServerAddressField},
	joinServerAddressField:        {joinServerAddressField},
	serviceProfileIDField:         {serviceProfileIDField},
	locationsField:                {},
	activatedAtField:              {activatedAtField},
	lastSeenAtField:               {lastSeenAtField},
	claimAuthenticationCodeField: {
		"claim_authentication_code_secret",
		"claim_authentication_code_valid_from",
		"claim_authentication_code_valid_to",
	},
}

func (dev EndDevice) toPB(pb *ttnpb.EndDevice, fieldMask store.FieldMask) {
	if pb.Ids == nil {
		pb.Ids = &ttnpb.EndDeviceIdentifiers{}
	}
	pb.Ids.ApplicationIds = &ttnpb.ApplicationIdentifiers{
		ApplicationId: dev.ApplicationID,
	}
	pb.Ids.DeviceId = dev.DeviceID
	pb.Ids.JoinEui = nil
	pb.Ids.DevEui = nil

	if joinEui := dev.JoinEUI.toPB(); joinEui != nil {
		pb.Ids.JoinEui = joinEui.Bytes()
	}
	if devEui := dev.DevEUI.toPB(); devEui != nil {
		pb.Ids.DevEui = devEui.Bytes()
	}
	pb.CreatedAt = ttnpb.ProtoTimePtr(cleanTime(dev.CreatedAt))
	pb.UpdatedAt = ttnpb.ProtoTimePtr(cleanTime(dev.UpdatedAt))
	if len(fieldMask) == 0 {
		fieldMask = defaultEndDeviceFieldMask
	}
	for _, path := range fieldMask {
		if setter, ok := devicePBSetters[path]; ok {
			setter(pb, &dev)
		}
	}
}

func (dev *EndDevice) fromPB(pb *ttnpb.EndDevice, fieldMask store.FieldMask) (columns []string) {
	if len(fieldMask) == 0 {
		fieldMask = defaultEndDeviceFieldMask
	}
	for _, path := range fieldMask {
		if setter, ok := deviceModelSetters[path]; ok {
			setter(dev, pb)
			if columnNames, ok := deviceColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			}
			continue
		}
	}
	return columns
}
