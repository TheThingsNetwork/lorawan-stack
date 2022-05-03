// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package devicetemplates

import (
	"context"
	"encoding/csv"
	"io"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"golang.org/x/net/html/charset"
)

// TTSCSV is the device template converter ID.
const TTSCSV = "the-things-stack-csv"

type ttsCSV struct{}

// Format implements the devicetemplates.Converter interface.
func (t *ttsCSV) Format() *ttnpb.EndDeviceTemplateFormat {
	return &ttnpb.EndDeviceTemplateFormat{
		Name:           "The Things Stack CSV",
		Description:    "File containing end devices in The Things Stack CSV format.",
		FileExtensions: []string{".csv"},
	}
}

var (
	errParseCSV  = errors.DefineInvalidArgument("parse_csv", "parse CSV at line `{line}` column `{column}`: {message}", "start_line")
	errCSVHeader = errors.DefineInvalidArgument("csv_header", "no known columns in CSV header")
)

func convertCSVErr(err error) error {
	var parseErr *csv.ParseError
	if errors.As(err, &parseErr) {
		return errParseCSV.WithCause(err).WithAttributes(
			"start_line", parseErr.StartLine,
			"line", parseErr.Line,
			"column", parseErr.Column,
			"message", parseErr.Err.Error(),
		)
	}
	return err
}

type csvFieldSetterFunc func(dst *ttnpb.EndDevice, field string) (paths []string, err error)

var csvFieldSetters = map[string]csvFieldSetterFunc{
	"id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		dst.Ids.DeviceId = val
		return []string{"ids.device_id"}, nil
	},
	"dev_eui": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var devEUI types.EUI64
		if err := devEUI.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.Ids.DevEui = &devEUI
		return []string{"ids.dev_eui"}, nil
	},
	"join_eui": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var joinEUI types.EUI64
		if err := joinEUI.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.Ids.JoinEui = &joinEUI
		return []string{"ids.join_eui"}, nil
	},
	"name": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		dst.Name = val
		return []string{"name"}, nil
	},
	"frequency_plan_id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		dst.FrequencyPlanId = val
		return []string{"frequency_plan_id"}, nil
	},
	"lorawan_version": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var version ttnpb.MACVersion
		if err := version.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.LorawanVersion = version
		return []string{"lorawan_version"}, nil
	},
	"lorawan_phy_version": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var version ttnpb.PHYVersion
		if err := version.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.LorawanPhyVersion = version
		return []string{"lorawan_phy_version"}, nil
	},
	"brand_id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		dst.VersionIds.BrandId = val
		return []string{"version_ids.brand_id"}, nil
	},
	"model_id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		dst.VersionIds.ModelId = val
		return []string{"version_ids.model_id"}, nil
	},
	"hardware_version": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		dst.VersionIds.HardwareVersion = val
		return []string{"version_ids.hardware_version"}, nil
	},
	"firmware_version": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		dst.VersionIds.FirmwareVersion = val
		return []string{"version_ids.firmware_version"}, nil
	},
	"band_id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		dst.VersionIds.BandId = val
		return []string{"version_ids.band_id"}, nil
	},
	"app_key": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var key types.AES128Key
		if err := key.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		if dst.RootKeys == nil {
			dst.RootKeys = &ttnpb.RootKeys{}
		}
		dst.RootKeys.AppKey = &ttnpb.KeyEnvelope{
			Key: &key,
		}
		dst.SupportsJoin = true
		return []string{"root_keys.app_key.key", "supports_join"}, nil
	},
	"nwk_key": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var key types.AES128Key
		if err := key.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		if dst.RootKeys == nil {
			dst.RootKeys = &ttnpb.RootKeys{}
		}
		dst.RootKeys.NwkKey = &ttnpb.KeyEnvelope{
			Key: &key,
		}
		dst.SupportsJoin = true
		return []string{"root_keys.nwk_key.key", "supports_join"}, nil
	},
}

// Convert implements the devicetemplates.Converter interface.
func (t *ttsCSV) Convert(ctx context.Context, r io.Reader, ch chan<- *ttnpb.EndDeviceTemplate) error {
	defer close(ch)

	r, err := charset.NewReader(r, "text/csv")
	if err != nil {
		return err
	}

	dec := csv.NewReader(r)
	dec.Comma = ';'
	dec.TrimLeadingSpace = true

	// Populate the mapping of column index to a field setter function based on known header column names.
	fieldSetters := make(map[int]csvFieldSetterFunc)
	header, err := dec.Read()
	if err != nil {
		if err == io.EOF {
			return errCSVHeader.WithCause(err)
		}
		return convertCSVErr(err)
	}
	for i, column := range header {
		fieldSetter, ok := csvFieldSetters[column]
		if !ok {
			continue
		}
		fieldSetters[i] = fieldSetter
	}
	if len(fieldSetters) == 0 {
		return errCSVHeader.New()
	}

	for {
		record, err := dec.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return convertCSVErr(err)
		}

		dev := &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{},
		}
		paths := make([]string, 0, len(record))
		for i, val := range record {
			fieldSetter, ok := fieldSetters[i]
			if !ok {
				continue
			}
			fieldPaths, err := fieldSetter(dev, val)
			if err != nil {
				return err
			}
			paths = ttnpb.AddFields(paths, fieldPaths...)
		}
		if !ttnpb.HasAnyField(paths, "ids.device_id") && ttnpb.HasAnyField(paths, "ids.dev_eui") {
			dev.Ids.DeviceId = strings.ToLower(dev.Ids.DevEui.String())
			paths = ttnpb.AddFields(paths, "ids.device_id")
		}

		tmpl := &ttnpb.EndDeviceTemplate{
			EndDevice: dev,
			FieldMask: ttnpb.FieldMask(paths...),
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- tmpl:
		}
	}
}

func init() {
	RegisterConverter(TTSCSV, &ttsCSV{})
}
