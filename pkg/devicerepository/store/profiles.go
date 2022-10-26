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

package store

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const mhz = 1000000

// regionalParametersToPB maps LoRaWAN schema regional parameters to ttnpb.PHYVersion enum values.
var regionalParametersToPB = map[string]ttnpb.PHYVersion{
	"TS001-1.0":        ttnpb.PHYVersion_TS001_V1_0,
	"TS001-1.0.1":      ttnpb.PHYVersion_TS001_V1_0_1,
	"RP001-1.0.2":      ttnpb.PHYVersion_RP001_V1_0_2,
	"RP001-1.0.2-RevB": ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
	"RP001-1.0.3-RevA": ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
	"RP001-1.1-RevA":   ttnpb.PHYVersion_RP001_V1_1_REV_A,
	"RP001-1.1-RevB":   ttnpb.PHYVersion_RP001_V1_1_REV_B,
	"RP002-1.0.0":      ttnpb.PHYVersion_RP002_V1_0_0,
	"RP002-1.0.1":      ttnpb.PHYVersion_RP002_V1_0_1,
	"RP002-1.0.2":      ttnpb.PHYVersion_RP002_V1_0_2,
}

// EndDeviceProfile is the profile of a LoRaWAN end device as defined in the LoRaWAN backend interfaces.
type EndDeviceProfile struct {
	VendorProfileID uint32 `yaml:"vendorProfileID" json:"vendor_profile_id"`
	SupportsClassB  bool   `yaml:"supportsClassB" json:"supports_class_b"`
	ClassBTimeout   uint32 `yaml:"classBTimeout" json:"class_b_timeout"`
	PingSlotPeriod  uint32 `yaml:"pingSlotPeriod" json:"ping_slot_period"`

	PingSlotDataRateIndex     *ttnpb.DataRateIndex  `yaml:"pingSlotDataRateIndex" json:"ping_slot_data_rate_index"`
	PingSlotFrequency         float64               `yaml:"pingSlotFrequency" json:"ping_slot_frequency"`
	SupportsClassC            bool                  `yaml:"supportsClassC" json:"supports_class_c"`
	ClassCTimeout             uint32                `yaml:"classCTimeout" json:"class_c_timeout"`
	MACVersion                ttnpb.MACVersion      `yaml:"macVersion" json:"mac_version"`
	RegionalParametersVersion string                `yaml:"regionalParametersVersion" json:"regional_parameters_version"`
	SupportsJoin              bool                  `yaml:"supportsJoin" json:"supports_join"`
	Rx1Delay                  *ttnpb.RxDelay        `yaml:"rx1Delay" json:"rx1_delay"`
	Rx1DataRateOffset         *ttnpb.DataRateOffset `yaml:"rx1DataRateOffset" json:"rx1_data_rate_offset"`
	Rx2DataRateIndex          *ttnpb.DataRateIndex  `yaml:"rx2DataRateIndex" json:"rx2_data_rate_index"`
	Rx2Frequency              float64               `yaml:"rx2Frequency" json:"rx2_frequency"`
	FactoryPresetFrequencies  []float64             `yaml:"factoryPresetFrequencies" json:"factory_preset_frequencies"`
	MaxEIRP                   float32               `yaml:"maxEIRP" json:"max_eirp"`
	MaxDutyCycle              float64               `yaml:"maxDutyCycle" json:"max_duty_cycle"`
	Supports32BitFCnt         bool                  `yaml:"supports32bitFCnt" json:"supports_32_bit_f_cnt"`
}

// pingSlotPeriodToPB maps LoRaWAN schema ping slot period to ttnpb.PingSlotPeriod enum values.
var pingSlotPeriodToPB = map[uint32]ttnpb.PingSlotPeriod{
	1:   ttnpb.PingSlotPeriod_PING_EVERY_1S,
	2:   ttnpb.PingSlotPeriod_PING_EVERY_2S,
	4:   ttnpb.PingSlotPeriod_PING_EVERY_4S,
	8:   ttnpb.PingSlotPeriod_PING_EVERY_8S,
	16:  ttnpb.PingSlotPeriod_PING_EVERY_16S,
	32:  ttnpb.PingSlotPeriod_PING_EVERY_32S,
	64:  ttnpb.PingSlotPeriod_PING_EVERY_64S,
	128: ttnpb.PingSlotPeriod_PING_EVERY_128S,
}

var errRegionalParametersVersion = errors.DefineNotFound("regional_parameters_version", "unknown Regional Parameters version `{phy_version}`")

// ToTemplatePB returns a ttnpb.EndDeviceTemplate from an end device profile.
func (p EndDeviceProfile) ToTemplatePB(ids *ttnpb.EndDeviceVersionIdentifiers, info *ttnpb.EndDeviceModel_FirmwareVersion_Profile) (*ttnpb.EndDeviceTemplate, error) {
	phyVersion, ok := regionalParametersToPB[p.RegionalParametersVersion]
	if !ok {
		return nil, errRegionalParametersVersion.WithAttributes("phy_version", p.RegionalParametersVersion)
	}

	paths := []string{
		"version_ids",
		"supports_join",
		"supports_class_b",
		"supports_class_c",
		"lorawan_version",
		"lorawan_phy_version",
	}
	dev := &ttnpb.EndDevice{
		VersionIds:        ids,
		SupportsJoin:      p.SupportsJoin,
		SupportsClassB:    p.SupportsClassB,
		SupportsClassC:    p.SupportsClassC,
		LorawanVersion:    p.MACVersion,
		LorawanPhyVersion: phyVersion,
	}

	if info != nil && info.CodecId != "" {
		dev.Formatters = &ttnpb.MessagePayloadFormatters{
			DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
			UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
		}
		paths = append(paths, "formatters")
	}

	dev.MacSettings = &ttnpb.MACSettings{}
	if p.ClassBTimeout > 0 {
		t := time.Duration(p.ClassBTimeout) * time.Second
		dev.MacSettings.ClassBTimeout = ttnpb.ProtoDurationPtr(t)
		paths = append(paths, "mac_settings.class_b_timeout")
	}
	if p.ClassCTimeout > 0 {
		t := time.Duration(p.ClassCTimeout) * time.Second
		dev.MacSettings.ClassCTimeout = ttnpb.ProtoDurationPtr(t)
		paths = append(paths, "mac_settings.class_c_timeout")
	}
	if v := p.PingSlotDataRateIndex; v != nil {
		dev.MacSettings.PingSlotDataRateIndex = &ttnpb.DataRateIndexValue{
			Value: *v,
		}
		paths = append(paths, "mac_settings.ping_slot_data_rate_index")
	}
	if p.PingSlotFrequency > 0 {
		dev.MacSettings.PingSlotFrequency = &ttnpb.ZeroableFrequencyValue{
			Value: uint64(p.PingSlotFrequency * mhz),
		}
		paths = append(paths, "mac_settings.ping_slot_frequency")
	}
	if p.PingSlotPeriod > 0 {
		dev.MacSettings.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
			Value: pingSlotPeriodToPB[p.PingSlotPeriod],
		}
		paths = append(paths, "mac_settings.ping_slot_periodicity")
	}
	if v := p.Rx1Delay; v != nil {
		dev.MacSettings.Rx1Delay = &ttnpb.RxDelayValue{
			Value: *v,
		}
		paths = append(paths, "mac_settings.rx1_delay")
	}
	if v := p.Rx1DataRateOffset; v != nil {
		dev.MacSettings.Rx1DataRateOffset = &ttnpb.DataRateOffsetValue{
			Value: *v,
		}
		paths = append(paths, "mac_settings.rx1_data_rate_offset")
	}
	if v := p.Rx2DataRateIndex; v != nil {
		dev.MacSettings.Rx2DataRateIndex = &ttnpb.DataRateIndexValue{
			Value: *v,
		}
		paths = append(paths, "mac_settings.rx2_data_rate_index")
	}
	if p.Rx2Frequency > 0 {
		dev.MacSettings.Rx2Frequency = &ttnpb.FrequencyValue{
			Value: uint64(p.Rx2Frequency * mhz),
		}
		paths = append(paths, "mac_settings.rx2_frequency")
	}
	if p.Supports32BitFCnt {
		dev.MacSettings.Supports_32BitFCnt = &ttnpb.BoolValue{
			Value: true,
		}
		paths = append(paths, "mac_settings.supports_32_bit_f_cnt")
	}
	if fs := p.FactoryPresetFrequencies; len(fs) > 0 {
		dev.MacSettings.FactoryPresetFrequencies = make([]uint64, 0, len(fs))
		for _, freq := range fs {
			dev.MacSettings.FactoryPresetFrequencies = append(dev.MacSettings.FactoryPresetFrequencies, uint64(freq*mhz))
		}
		paths = append(paths, "mac_settings.factory_preset_frequencies")
	}
	if dc := p.MaxDutyCycle; dc > 0 {
		dev.MacSettings.MaxDutyCycle = &ttnpb.AggregatedDutyCycleValue{
			Value: dutyCycleFromFloat(dc),
		}
		paths = append(paths, "mac_settings.max_duty_cycle")
	}

	if !p.SupportsJoin && p.MaxEIRP > 0 {
		dev.MacState = &ttnpb.MACState{
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp: p.MaxEIRP,
			},
		}
		paths = append(paths, "mac_state.desired_parameters.max_eirp")
	}
	return &ttnpb.EndDeviceTemplate{
		EndDevice: dev,
		FieldMask: ttnpb.FieldMask(paths...),
	}, nil
}

// dutyCycleFromFloat converts a float value (0 < dc < 1) to a ttnpb.AggregatedDutyCycle
// enum value. The enum value is rounded-down to the closest value, which means
// that dc == 0.3 will return ttnpb.DUTY_CYCLE_4 (== 0.25).
func dutyCycleFromFloat(dc float64) ttnpb.AggregatedDutyCycle {
	counts := 0
	for counts = 0; dc < 1 && counts < 15; counts++ {
		dc *= 2
	}
	return ttnpb.AggregatedDutyCycle(counts)
}
