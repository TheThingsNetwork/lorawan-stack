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
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"golang.org/x/net/html/charset"
)

// TTSCSV is the device template converter ID.
const TTSCSV = "the-things-stack-csv"

var errGenerateSessionKeyID = errors.DefineUnavailable("generate_session_key_id", "generate session key ID")

type ttsCSV struct{}

// Format implements the devicetemplates.Converter interface.
func (*ttsCSV) Format() *ttnpb.EndDeviceTemplateFormat {
	return &ttnpb.EndDeviceTemplateFormat{
		Name:           "The Things Stack CSV",
		Description:    "File containing end devices in The Things Stack CSV format.",
		FileExtensions: []string{".csv"},
	}
}

var (
	errParseCSV = errors.DefineInvalidArgument("parse_csv",
		"parse CSV at line `{line}` column `{column}`: {message}", "start_line",
	)
	errCSVHeader     = errors.DefineInvalidArgument("csv_header", "no known columns in CSV header")
	errParseCSVField = errors.DefineInvalidArgument("parse_csv_field",
		"parse CSV field at line `{line}` column `{column}`",
	)
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
		dst.Ids.DevEui = devEUI.Bytes()
		return []string{"ids.dev_eui"}, nil
	},
	"join_eui": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var joinEUI types.EUI64
		if err := joinEUI.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.Ids.JoinEui = joinEUI.Bytes()
		return []string{"ids.join_eui"}, nil
	},
	"app_eui": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var joinEUI types.EUI64
		if err := joinEUI.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.Ids.JoinEui = joinEUI.Bytes()
		return []string{"ids.join_eui"}, nil
	},
	"name": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		dst.Name = val
		return []string{"name"}, nil
	},
	"description": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		dst.Description = val
		return []string{"description"}, nil
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
			Key: key.Bytes(),
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
			Key: key.Bytes(),
		}
		dst.SupportsJoin = true
		return []string{"root_keys.nwk_key.key", "supports_join"}, nil
	},
	"rx1_delay": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var rx1Delay ttnpb.RxDelay
		if err := rx1Delay.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		if dst.MacSettings == nil {
			dst.MacSettings = &ttnpb.MACSettings{}
		}
		dst.MacSettings.Rx1Delay = &ttnpb.RxDelayValue{
			Value: rx1Delay,
		}
		return []string{"mac_settings.rx1_delay"}, nil
	},
	"supports_32_bit_f_cnt": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var supports32BitFCnt ttnpb.BoolValue
		if err := supports32BitFCnt.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		if dst.MacSettings == nil {
			dst.MacSettings = &ttnpb.MACSettings{}
		}
		dst.MacSettings.Supports_32BitFCnt = &supports32BitFCnt
		return []string{"mac_settings.supports_32_bit_f_cnt"}, nil
	},
	"dev_addr": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var devAddr types.DevAddr
		if dst.Session == nil {
			dst.Session = &ttnpb.Session{}
		}
		if err := devAddr.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		dst.Session.DevAddr = devAddr[:]
		return []string{"session.dev_addr"}, nil
	},
	"app_s_key": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		key := &types.AES128Key{}
		if err := key.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		if dst.Session == nil {
			dst.Session = &ttnpb.Session{}
		}
		if dst.Session.Keys == nil {
			skID, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
			if err != nil {
				return nil, errGenerateSessionKeyID.WithCause(err)
			}
			dst.Session.Keys = &ttnpb.SessionKeys{
				SessionKeyId: skID[:],
			}
		}
		dst.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{
			Key: key.Bytes(),
		}
		return []string{"session.keys.app_s_key.key"}, nil
	},
	"f_nwk_s_int_key": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		key := &types.AES128Key{}
		if err := key.UnmarshalText([]byte(val)); err != nil {
			return nil, err
		}
		if dst.Session == nil {
			dst.Session = &ttnpb.Session{}
		}
		if dst.Session.Keys == nil {
			skID, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
			if err != nil {
				return nil, errGenerateSessionKeyID.WithCause(err)
			}
			dst.Session.Keys = &ttnpb.SessionKeys{
				SessionKeyId: skID[:],
			}
		}
		dst.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
			Key: key.Bytes(),
		}
		return []string{"session.keys.f_nwk_s_int_key.key"}, nil
	},
	"last_f_cnt_up": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.Session == nil {
			dst.Session = &ttnpb.Session{}
		}
		s, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		dst.Session.LastFCntUp = uint32(s)
		return []string{"session.last_f_cnt_up"}, nil
	},
	"last_n_f_cnt_down": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.Session == nil {
			dst.Session = &ttnpb.Session{}
		}
		s, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		dst.Session.LastNFCntDown = uint32(s)
		return []string{"session.last_n_f_cnt_down"}, nil
	},
	"last_a_f_cnt_down": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.Session == nil {
			dst.Session = &ttnpb.Session{}
		}
		s, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		dst.Session.LastAFCntDown = uint32(s)
		return []string{"session.last_a_f_cnt_down"}, nil
	},
	"supports_class_c": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		var err error
		dst.SupportsClassC, err = strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return []string{"supports_class_c"}, nil
	},
	"vendor_id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		s, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		dst.VersionIds.VendorId = uint32(s)
		return []string{"version_ids.vendor_id"}, nil
	},
	"vendor_profile_id": func(dst *ttnpb.EndDevice, val string) ([]string, error) {
		if dst.VersionIds == nil {
			dst.VersionIds = &ttnpb.EndDeviceVersionIdentifiers{}
		}
		s, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		dst.VersionIds.VendorProfileId = uint32(s)
		return []string{"version_ids.vendor_profile_id"}, nil
	},
}

func determineComma(head []byte) (rune, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(head))
	if !scanner.Scan() {
		return 0, false
	}
	header := scanner.Text()
	for _, c := range []rune{';', ','} {
		if strings.ContainsRune(header, c) {
			return c, true
		}
	}
	return 0, false
}

// Convert implements the devicetemplates.Converter interface.
func (*ttsCSV) Convert(ctx context.Context, r io.Reader, ch chan<- *ttnpb.EndDeviceTemplate) error {
	defer close(ch)

	r, err := charset.NewReader(r, "text/csv")
	if err != nil {
		return err
	}

	// Best effort detection of the comma by peeking the first kilobyte and looking at a separator on the first line.
	comma := ';'
	const maxHeaderLength = 1024
	buf := bufio.NewReaderSize(r, maxHeaderLength)
	if head, _ := buf.Peek(maxHeaderLength); len(head) > 0 {
		if c, ok := determineComma(head); ok {
			comma = c
		}
	}

	dec := csv.NewReader(buf)
	dec.Comma = comma
	dec.TrimLeadingSpace = true

	// Populate the mapping of column index to a field setter function based on known header column names.
	fieldSetters := make(map[int]csvFieldSetterFunc)
	header, err := dec.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
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

	line := 1
	for {
		record, err := dec.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
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
				return errParseCSVField.WithCause(err).WithAttributes(
					"line", line,
					"column", i+1,
				)
			}
			paths = ttnpb.AddFields(paths, fieldPaths...)
		}
		if !ttnpb.HasAnyField(paths, "ids.device_id") && ttnpb.HasAnyField(paths, "ids.dev_eui") {
			dev.Ids.DeviceId = fmt.Sprintf("eui-%s", strings.ToLower(types.MustEUI64(dev.Ids.DevEui).String()))
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
		line++
	}
}

func init() {
	RegisterConverter(TTSCSV, &ttsCSV{})
}
