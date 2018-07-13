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

// Package devicerepository allows to fetch device data from the Device Repository.
package devicerepository

import (
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gopkg.in/yaml.v2"
)

var formatters = map[string]ttnpb.PayloadFormatter{
	"cayennelpp": ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
	"grpc":       ttnpb.PayloadFormatter_FORMATTER_GRPC_SERVICE,
	"javascript": ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
}

// Client allows retrieval of device data.
type Client struct {
	Fetcher fetch.Interface
}

// Brand returns the general information related to a brand.
type Brand struct {
	id string

	Name  string   `yaml:"name,omitempty"`
	URL   string   `yaml:"url,omitempty"`
	Logos []string `yaml:"logos,omitempty"`
}

// Proto of the brand.
func (b Brand) Proto() *ttnpb.DeviceBrand {
	return &ttnpb.DeviceBrand{
		ID:    b.id,
		Name:  b.Name,
		URL:   b.URL,
		Logos: b.Logos,
	}
}

var (
	errFetchFailed = errors.DefineUnavailable("fetch_failed", "fetch failed of file `{filename}`")
	errParseFailed = errors.DefineInvalidArgument("parse_failed", "parse failed")
)

// Brands fetches and parses the list of brands.
func (c Client) Brands() (map[string]Brand, error) {
	filename := "brands.yml"
	content, err := c.Fetcher.File(filename)
	if err != nil {
		return nil, errFetchFailed.WithCause(err).WithAttributes("filename", filename)
	}

	l := &struct {
		Version string           `yaml:"version"`
		Brands  map[string]Brand `yaml:"brands,omitempty"`
	}{}
	if err = yaml.Unmarshal(content, l); err != nil {
		return nil, errParseFailed.WithCause(err)
	}

	brands := make(map[string]Brand)
	for id, brand := range l.Brands {
		brand.id = id
		brands[id] = brand
	}
	return brands, nil
}

var (
	errMissingFileExtension = errors.DefineInvalidArgument("missing_file_extension", "missing file extension in `{filename}`")
	errUnknownFileType      = errors.DefineInvalidArgument("unknown_file_type", "unknown file type `{filename}`")
)

// Device contains the general information related to a device.
type Device struct {
	brand string
	id    string

	Name string `yaml:"name,omitempty"`
}

// Proto returns the device in the protobuf format.
func (d Device) Proto() *ttnpb.EndDeviceModel {
	return &ttnpb.EndDeviceModel{BrandID: d.brand, ModelID: d.id, ModelName: d.Name}
}

// Devices returns the list of devices related to this brand, indexed by device ID.
func (c Client) Devices(brandID string) (map[string]Device, error) {
	filename := "devices.yml"
	content, err := c.Fetcher.File(brandID, filename)
	if err != nil {
		return nil, errFetchFailed.WithCause(err).WithAttributes("filename", filename)
	}

	l := &struct {
		Version string            `yaml:"version"`
		Devices map[string]Device `yaml:"devices,omitempty"`
	}{}
	if err = yaml.Unmarshal(content, l); err != nil {
		return nil, errParseFailed.WithCause(err)
	}

	devices := make(map[string]Device)
	for deviceID, device := range l.Devices {
		device.id = deviceID
		device.brand = brandID
		devices[deviceID] = device
	}

	return devices, nil
}

// DeviceVersions returns the versions of this device, along with this version's characteristics.
func (c Client) DeviceVersions(brandID, modelID string) (map[string]DeviceHardwareVersion, error) {
	filename := "versions.yml"
	content, err := c.Fetcher.File(brandID, modelID, filename)
	if err != nil {
		return nil, errFetchFailed.WithCause(err).WithAttributes("filename", filename)
	}

	l := &struct {
		Version          string                           `yaml:"version"`
		HardwareVersions map[string]DeviceHardwareVersion `yaml:"hardware_versions,omitempty"`
	}{}
	if err = yaml.Unmarshal(content, l); err != nil {
		return nil, errParseFailed.WithCause(err)
	}

	hardwareVersions := make(map[string]DeviceHardwareVersion)
	for version, versionDetails := range l.HardwareVersions {
		payloadFormats := []*PayloadFormat{
			versionDetails.PayloadFormats.Up,
			versionDetails.PayloadFormats.Down,
		}
		for _, payloadFormat := range payloadFormats {
			if payloadFormat == nil || payloadFormat.Type != "javascript" {
				continue
			}

			content, err = c.Fetcher.File(brandID, modelID, version, payloadFormat.Parameter)
			if err != nil {
				return nil, errFetchFailed.WithCause(err).WithAttributes("filename", payloadFormat.Parameter)
			}

			payloadFormat.Parameter = string(content)
		}

		versionDetails.brand = brandID
		versionDetails.device = modelID
		versionDetails.version = version
		hardwareVersions[version] = versionDetails
	}

	return hardwareVersions, nil
}

// PayloadFormat of a device.
type PayloadFormat struct {
	Type      string `yaml:"type"`
	Parameter string `yaml:"param,omitempty"`
}

// DevicePayloadFormats for a specific device hardware version.
type DevicePayloadFormats struct {
	Up   *PayloadFormat `yaml:"up,omitempty"`
	Down *PayloadFormat `yaml:"down,omitempty"`
}

// DeviceHardwareVersion with the characteristics of the specific version.
type DeviceHardwareVersion struct {
	brand, device, version string

	FirmwareVersions []string `yaml:"firmware_versions,omitempty"`
	Photos           []string `yaml:"photos,omitempty"`

	PayloadFormats DevicePayloadFormats `yaml:"payload_format,omitempty"`
}

// Protos returns the device's firmware versions for that specific hardware version.
func (d DeviceHardwareVersion) Protos() []*ttnpb.EndDeviceVersion {
	proto := ttnpb.EndDeviceVersion{
		EndDeviceModel: ttnpb.EndDeviceModel{
			ModelID: d.device,
			BrandID: d.brand,
		},
		HardwareVersion: d.version,

		DefaultFormatters: &ttnpb.DeviceFormatters{},
	}

	if d.PayloadFormats.Up != nil {
		if formatter, ok := formatters[d.PayloadFormats.Up.Type]; ok {
			proto.DefaultFormatters.UpFormatter = formatter
			proto.DefaultFormatters.UpFormatterParameter = d.PayloadFormats.Up.Parameter
		} else {
			proto.DefaultFormatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_NONE
			proto.DefaultFormatters.UpFormatterParameter = ""
		}
	}

	if d.PayloadFormats.Down != nil {
		if formatter, ok := formatters[d.PayloadFormats.Down.Type]; ok {
			proto.DefaultFormatters.DownFormatter = formatter
			proto.DefaultFormatters.DownFormatterParameter = d.PayloadFormats.Down.Parameter
		} else {
			proto.DefaultFormatters.DownFormatter = ttnpb.PayloadFormatter_FORMATTER_NONE
			proto.DefaultFormatters.DownFormatterParameter = ""
		}
	}

	protos := []*ttnpb.EndDeviceVersion{}
	switch len(d.FirmwareVersions) {
	case 0:
		protos = append(protos, &proto)
	case 1:
		proto.FirmwareVersion = d.FirmwareVersions[0]
		protos = append(protos, &proto)
	default:
		for _, firmwareVersion := range d.FirmwareVersions {
			firmwareProto := proto
			if proto.DefaultFormatters != nil {
				defaultFormatters := *proto.DefaultFormatters
				firmwareProto.DefaultFormatters = &defaultFormatters
			}
			firmwareProto.FirmwareVersion = firmwareVersion
			protos = append(protos, &firmwareProto)
		}
	}

	return protos
}
