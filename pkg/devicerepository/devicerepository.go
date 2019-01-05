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

// Package devicerepository allows to fetch device data from the Device Repository.
package devicerepository

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gopkg.in/yaml.v2"
)

// Client provides a device repository through a fetcher.
type Client struct {
	Fetcher fetch.Interface
}

type brand struct {
	id    string
	Name  string   `yaml:"name,omitempty"`
	URL   string   `yaml:"url,omitempty"`
	Logos []string `yaml:"logos,omitempty"`
}

var (
	errFetchFailed = errors.Define("fetch", "failed to fetch file `{filename}`")
	errParseFailed = errors.DefineInvalidArgument("parse", "parse failed")
)

const (
	brandsFile   = "brands.yml"
	devicesFile  = "devices.yml"
	versionsFile = "versions.yml"
)

// Brands fetches and parses the list of brands.
func (c Client) Brands() (map[string]ttnpb.EndDeviceBrand, error) {
	content, err := c.Fetcher.File(brandsFile)
	if err != nil {
		return nil, errFetchFailed.WithCause(err).WithAttributes("filename", brandsFile)
	}

	l := &struct {
		Version string           `yaml:"version"`
		Brands  map[string]brand `yaml:"brands,omitempty"`
	}{}
	if err = yaml.Unmarshal(content, l); err != nil {
		return nil, errParseFailed.WithCause(err)
	}

	brands := make(map[string]ttnpb.EndDeviceBrand)
	for id, brand := range l.Brands {
		brands[id] = ttnpb.EndDeviceBrand{
			ID:    id,
			Name:  brand.Name,
			URL:   brand.URL,
			Logos: brand.Logos,
		}
	}
	return brands, nil
}

type endDeviceModel struct {
	Name string `yaml:"name,omitempty"`
}

// DeviceModels fetches and parses the list of device models.
func (c Client) DeviceModels(brandID string) (map[string]ttnpb.EndDeviceModel, error) {
	content, err := c.Fetcher.File(brandID, devicesFile)
	if err != nil {
		return nil, errFetchFailed.WithCause(err).WithAttributes("filename", devicesFile)
	}

	l := &struct {
		Version string                    `yaml:"version"`
		Devices map[string]endDeviceModel `yaml:"devices,omitempty"`
	}{}
	if err = yaml.Unmarshal(content, l); err != nil {
		return nil, errParseFailed.WithCause(err)
	}

	devices := make(map[string]ttnpb.EndDeviceModel)
	for id, device := range l.Devices {
		devices[id] = ttnpb.EndDeviceModel{
			ID:      id,
			BrandID: brandID,
			Name:    device.Name,
		}
	}
	return devices, nil
}

type payloadFormat struct {
	Type      string `yaml:"type"`
	Parameter string `yaml:"parameter,omitempty"`
}

type payloadFormats struct {
	Up   *payloadFormat `yaml:"up,omitempty"`
	Down *payloadFormat `yaml:"down,omitempty"`
}

type endDeviceVersion struct {
	FirmwareVersion string         `yaml:"firmware_version"`
	Photos          []string       `yaml:"photos,omitempty"`
	PayloadFormats  payloadFormats `yaml:"payload_format,omitempty"`
}

var errInvalidPayloadFormatter = errors.DefineInvalidArgument("invalid_payload_formatter", "invalid payload formatter `{formatter}`")

// DeviceVersions fetches and parses the list of device versions.
func (c Client) DeviceVersions(brandID, modelID string) ([]ttnpb.EndDeviceVersion, error) {
	content, err := c.Fetcher.File(brandID, modelID, versionsFile)
	if err != nil {
		return nil, errFetchFailed.WithCause(err).WithAttributes("filename", versionsFile)
	}

	l := &struct {
		Version          string                        `yaml:"version"`
		HardwareVersions map[string][]endDeviceVersion `yaml:"hardware_versions,omitempty"`
	}{}
	if err = yaml.Unmarshal(content, l); err != nil {
		return nil, errParseFailed.WithCause(err)
	}

	var versions []ttnpb.EndDeviceVersion
	for hwVersion, fwVersions := range l.HardwareVersions {
		for _, version := range fwVersions {
			parseFormatter := func(pf payloadFormat) (ttnpb.PayloadFormatter, string, error) {
				switch pf.Type {
				case "cayennelpp":
					return ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP, "", nil
				case "grpc":
					return ttnpb.PayloadFormatter_FORMATTER_GRPC_SERVICE, pf.Parameter, nil
				case "javascript":
					content, err = c.Fetcher.File(brandID, modelID, hwVersion, pf.Parameter)
					if err != nil {
						return 0, "", errFetchFailed.WithCause(err).WithAttributes("filename", pf.Parameter)
					}
					return ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT, string(content), nil
				default:
					return 0, "", errInvalidPayloadFormatter.WithAttributes("formatter", pf.Type)
				}
			}

			formatters := ttnpb.MessagePayloadFormatters{}
			if version.PayloadFormats.Up != nil {
				formatters.UpFormatter, formatters.UpFormatterParameter, err = parseFormatter(*version.PayloadFormats.Up)
				if err != nil {
					return nil, err
				}
			} else {
				formatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_NONE
			}
			if version.PayloadFormats.Down != nil {
				formatters.DownFormatter, formatters.DownFormatterParameter, err = parseFormatter(*version.PayloadFormats.Down)
				if err != nil {
					return nil, err
				}
			} else {
				formatters.DownFormatter = ttnpb.PayloadFormatter_FORMATTER_NONE
			}

			versions = append(versions, ttnpb.EndDeviceVersion{
				EndDeviceVersionIdentifiers: ttnpb.EndDeviceVersionIdentifiers{
					BrandID:         brandID,
					ModelID:         modelID,
					HardwareVersion: hwVersion,
					FirmwareVersion: version.FirmwareVersion,
				},
				Photos:            version.Photos,
				DefaultFormatters: formatters,
			})
		}
	}

	return versions, nil
}
