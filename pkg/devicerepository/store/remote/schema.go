// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package remote

import (
	"time"

	pbtypes "github.com/gogo/protobuf/types"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

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

// Vendor is an end device vendor.
type Vendor struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	VendorID uint32   `yaml:"vendorID"`
	Draft    bool     `yaml:"draft,omitempty"`
	Email    string   `yaml:"email"`
	Website  string   `yaml:"website"`
	PEN      uint32   `yaml:"pen"`
	OUIs     []string `yaml:"ouis"`
	Logo     string   `yaml:"logo"`
}

// ToPB creates a ttnpb.EndDeviceBrand protocol buffer from a Vendor.
func (v Vendor) ToPB(paths ...string) (*ttnpb.EndDeviceBrand, error) {
	pb := &ttnpb.EndDeviceBrand{
		BrandId:                       v.ID,
		Name:                          v.Name,
		LoraAllianceVendorId:          v.VendorID,
		Email:                         v.Email,
		Website:                       v.Website,
		Logo:                          v.Logo,
		PrivateEnterpriseNumber:       v.PEN,
		OrganizationUniqueIdentifiers: v.OUIs,
	}

	res := &ttnpb.EndDeviceBrand{}
	if err := res.SetFields(pb, paths...); err != nil {
		return nil, err
	}
	return res, nil
}

// VendorsIndex is the format for the vendor/index.yaml file.
type VendorsIndex struct {
	Vendors []Vendor `yaml:"vendors"`
}

// VendorEndDevicesIndex is the format of the `vendor/<vendor-id>/index.yaml` file.
type VendorEndDevicesIndex struct {
	EndDevices []string `yaml:"endDevices"`
}

// EndDeviceModel is the format of the `vendor/<vendor-id>/<model-id>.yaml` file.
type EndDeviceModel struct {
	Name             string `yaml:"name"`
	Description      string `yaml:"description"`
	HardwareVersions []struct {
		Version    string `yaml:"version"`
		Numeric    uint32 `yaml:"numeric"`
		PartNumber string `yaml:"partNumber"`
	} `yaml:"hardwareVersions"`
	FirmwareVersions []struct {
		Version          string   `yaml:"version"`
		Numeric          uint32   `yaml:"numeric"`
		HardwareVersions []string `yaml:"hardwareVersions"`
		Profiles         map[string]struct {
			VendorID         string `yaml:"vendorID"`
			ID               string `yaml:"id"`
			Codec            string `yaml:"codec"`
			LoRaWANCertified bool   `yaml:"lorawanCertified"`
		} `yaml:"profiles"`
	} `yaml:"firmwareVersions"`
	Sensors    []string `yaml:"sensors"`
	Dimensions *struct {
		Width    float32 `yaml:"width"`
		Height   float32 `yaml:"height"`
		Diameter float32 `yaml:"diameter"`
		Length   float32 `yaml:"length"`
	} `yaml:"dimensions"`
	Weight  float32 `yaml:"weight"`
	Battery *struct {
		Replaceable bool   `yaml:"replaceable"`
		Type        string `yaml:"type"`
	} `yaml:"battery"`
	OperatingConditions *struct {
		Temperature *struct {
			Min float32 `yaml:"min"`
			Max float32 `yaml:"max"`
		} `yaml:"temperature"`
		RelativeHumidity *struct {
			Min float32 `yaml:"min"`
			Max float32 `yaml:"max"`
		} `yaml:"relativeHumidity"`
	} `yaml:"operatingConditions"`
	IPCode          string                  `yaml:"ipCode"`
	KeyProvisioning []ttnpb.KeyProvisioning `yaml:"keyProvisioning"`
	KeySecurity     ttnpb.KeySecurity       `yaml:"keySecurity"`
	Photos          *struct {
		Main  string   `yaml:"main"`
		Other []string `yaml:"other"`
	} `yaml:"photos"`
	Videos *struct {
		Main  string   `yaml:"main"`
		Other []string `yaml:"other"`
	} `yaml:"videos"`
	ProductURL   string `yaml:"productURL"`
	DataSheetURL string `yaml:"dataSheetURL"`
	ResellerURLs []struct {
		Name   string   `yaml:"name"`
		Region []string `yaml:"region"`
		URL    string   `yaml:"url"`
	} `yaml:"resellerURLs"`
	Compliances *struct {
		Safety []struct {
			Body     string `yaml:"body"`
			Norm     string `yaml:"norm"`
			Standard string `yaml:"standard"`
			Version  string `yaml:"version"`
		} `yaml:"safety"`
		RadioEquipment []struct {
			Body     string `yaml:"body"`
			Norm     string `yaml:"norm"`
			Standard string `yaml:"standard"`
			Version  string `yaml:"version"`
		} `yaml:"radioEquipment"`
	} `yaml:"compliances"`
	AdditionalRadios []string `yaml:"additionalRadios"`
}

// ToPB converts an EndDefinitionDefinition to a Protocol Buffer.
func (d EndDeviceModel) ToPB(brandID, modelID string, paths ...string) (*ttnpb.EndDeviceModel, error) {
	pb := &ttnpb.EndDeviceModel{
		BrandId:          brandID,
		ModelId:          modelID,
		Name:             d.Name,
		Description:      d.Description,
		FirmwareVersions: make([]*ttnpb.EndDeviceModel_FirmwareVersion, 0, len(d.FirmwareVersions)),
		Sensors:          d.Sensors,
		IpCode:           d.IPCode,
		KeyProvisioning:  d.KeyProvisioning,
		KeySecurity:      d.KeySecurity,
		ProductUrl:       d.ProductURL,
		DatasheetUrl:     d.DataSheetURL,
		AdditionalRadios: d.AdditionalRadios,
	}

	if hwVersions := d.HardwareVersions; hwVersions != nil {
		pb.HardwareVersions = make([]*ttnpb.EndDeviceModel_HardwareVersion, 0, len(hwVersions))
		for _, ver := range hwVersions {
			pb.HardwareVersions = append(pb.HardwareVersions, &ttnpb.EndDeviceModel_HardwareVersion{
				Version:    ver.Version,
				Numeric:    ver.Numeric,
				PartNumber: ver.PartNumber,
			})
		}
	}
	for _, ver := range d.FirmwareVersions {
		pbver := &ttnpb.EndDeviceModel_FirmwareVersion{
			Version:                   ver.Version,
			Numeric:                   ver.Numeric,
			SupportedHardwareVersions: ver.HardwareVersions,
		}
		pbver.Profiles = make(map[string]*ttnpb.EndDeviceModel_FirmwareVersion_Profile, len(ver.Profiles))
		for region, profile := range ver.Profiles {
			pbver.Profiles[regionToBandID[region]] = &ttnpb.EndDeviceModel_FirmwareVersion_Profile{
				VendorId:         profile.VendorID,
				ProfileId:        profile.ID,
				LorawanCertified: profile.LoRaWANCertified,
				CodecId:          profile.Codec,
			}
		}
		pb.FirmwareVersions = append(pb.FirmwareVersions, pbver)
	}

	if dim := d.Dimensions; dim != nil {
		pb.Dimensions = &ttnpb.EndDeviceModel_Dimensions{}
		if w := d.Dimensions.Width; w > 0 {
			pb.Dimensions.Width = &pbtypes.FloatValue{Value: w}
		}
		if h := d.Dimensions.Height; h > 0 {
			pb.Dimensions.Height = &pbtypes.FloatValue{Value: h}
		}
		if d := d.Dimensions.Diameter; d > 0 {
			pb.Dimensions.Diameter = &pbtypes.FloatValue{Value: d}
		}
		if l := d.Dimensions.Length; l > 0 {
			pb.Dimensions.Length = &pbtypes.FloatValue{Value: l}
		}
	}

	if w := d.Weight; w > 0 {
		pb.Weight = &pbtypes.FloatValue{Value: w}
	}

	if battery := d.Battery; battery != nil {
		pb.Battery = &ttnpb.EndDeviceModel_Battery{
			Replaceable: &pbtypes.BoolValue{Value: d.Battery.Replaceable},
			Type:        d.Battery.Type,
		}
	}

	if oc := d.OperatingConditions; oc != nil {
		pb.OperatingConditions = &ttnpb.EndDeviceModel_OperatingConditions{}

		if rh := oc.RelativeHumidity; rh != nil {
			pb.OperatingConditions.RelativeHumidity = &ttnpb.EndDeviceModel_OperatingConditions_Limits{
				Min: &pbtypes.FloatValue{Value: rh.Min},
				Max: &pbtypes.FloatValue{Value: rh.Max},
			}
		}

		if temp := oc.Temperature; temp != nil {
			pb.OperatingConditions.Temperature = &ttnpb.EndDeviceModel_OperatingConditions_Limits{
				Min: &pbtypes.FloatValue{Value: temp.Min},
				Max: &pbtypes.FloatValue{Value: temp.Max},
			}
		}
	}

	if p := d.Photos; p != nil {
		pb.Photos = &ttnpb.EndDeviceModel_Photos{
			Main:  p.Main,
			Other: p.Other,
		}
	}

	if v := d.Videos; v != nil {
		pb.Videos = &ttnpb.EndDeviceModel_Videos{
			Main:  v.Main,
			Other: v.Other,
		}
	}

	if rs := d.ResellerURLs; len(rs) > 0 {
		pb.Resellers = make([]*ttnpb.EndDeviceModel_Reseller, 0, len(rs))
		for _, reseller := range rs {
			pb.Resellers = append(pb.Resellers, &ttnpb.EndDeviceModel_Reseller{
				Name:   reseller.Name,
				Url:    reseller.URL,
				Region: reseller.Region,
			})
		}
	}

	if cs := d.Compliances; cs != nil {
		pb.Compliances = &ttnpb.EndDeviceModel_Compliances{
			Safety:         make([]*ttnpb.EndDeviceModel_Compliances_Compliance, 0, len(cs.Safety)),
			RadioEquipment: make([]*ttnpb.EndDeviceModel_Compliances_Compliance, 0, len(cs.RadioEquipment)),
		}

		for _, c := range cs.Safety {
			pb.Compliances.Safety = append(pb.Compliances.Safety, &ttnpb.EndDeviceModel_Compliances_Compliance{
				Version:  c.Version,
				Body:     c.Body,
				Standard: c.Standard,
				Norm:     c.Norm,
			})
		}
		for _, c := range cs.RadioEquipment {
			pb.Compliances.RadioEquipment = append(pb.Compliances.RadioEquipment, &ttnpb.EndDeviceModel_Compliances_Compliance{
				Version:  c.Version,
				Body:     c.Body,
				Standard: c.Standard,
				Norm:     c.Norm,
			})
		}
	}

	res := &ttnpb.EndDeviceModel{}
	if err := res.SetFields(pb, paths...); err != nil {
		return nil, err
	}
	return res, nil
}

// EndDeviceProfile is the format of the `vendor/<vendor-id>/<profile-id>.yaml` file.
type EndDeviceProfile struct {
	VendorProfileID uint32 `yaml:"vendorProfileID"`
	SupportsClassB  bool   `yaml:"supportsClassB"`
	ClassBTimeout   uint32 `yaml:"classBTimeout"`
	PingSlotPeriod  uint32 `yaml:"pingSlotPeriod"`

	PingSlotDataRateIndex     *ttnpb.DataRateIndex  `yaml:"pingSlotDataRateIndex"`
	PingSlotFrequency         float64               `yaml:"pingSlotFrequency"`
	SupportsClassC            bool                  `yaml:"supportsClassC"`
	ClassCTimeout             uint32                `yaml:"classCTimeout"`
	MACVersion                ttnpb.MACVersion      `yaml:"macVersion"`
	RegionalParametersVersion string                `yaml:"regionalParametersVersion"`
	SupportsJoin              bool                  `yaml:"supportsJoin"`
	Rx1Delay                  *ttnpb.RxDelay        `yaml:"rx1Delay"`
	Rx1DataRateOffset         *ttnpb.DataRateOffset `yaml:"rx1DataRateOffset"`
	Rx2DataRateIndex          *ttnpb.DataRateIndex  `yaml:"rx2DataRateIndex"`
	Rx2Frequency              float64               `yaml:"rx2Frequency"`
	FactoryPresetFrequencies  []float64             `yaml:"factoryPresetFrequencies"`
	MaxEIRP                   float32               `yaml:"maxEIRP"`
	MaxDutyCycle              float64               `yaml:"maxDutyCycle"`
	Supports32BitFCnt         bool                  `yaml:"supports32bitFCnt"`
}

var errRegionalParametersVersion = errors.DefineNotFound("regional_parameters_version", "unknown Regional Parameters version `{phy_version}`")

const mhz = 1000000

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

	if info.CodecId != "" {
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
		dev.MacSettings.PingSlotFrequency = &ttnpb.FrequencyValue{
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
		FieldMask: &pbtypes.FieldMask{
			Paths: paths,
		},
	}, nil
}

type EncodedCodecData struct {
	FPort    uint32   `yaml:"fPort"`
	Bytes    []byte   `yaml:"bytes"`
	Warnings []string `yaml:"warnings"`
	Errors   []string `yaml:"errors"`
}

type DecodedCodecData struct {
	Data     map[string]interface{} `yaml:"data"`
	Warnings []string               `yaml:"warnings"`
	Errors   []string               `yaml:"errors"`
}

type DecoderCodecExample struct {
	Description string           `yaml:"description"`
	Input       EncodedCodecData `yaml:"input"`
	Output      DecodedCodecData `yaml:"output"`
}

type EncoderCodecExample struct {
	Description string           `yaml:"description"`
	Input       DecodedCodecData `yaml:"input"`
	Output      EncodedCodecData `yaml:"output"`
}

type EndDeviceEncoderCodec struct {
	FileName string                `yaml:"fileName"`
	Examples []EncoderCodecExample `yaml:"examples"`
}

type EndDeviceDecoderCodec struct {
	FileName string                `yaml:"fileName"`
	Examples []DecoderCodecExample `yaml:"examples"`
}

// EndDeviceCodecs is the format of the `vendor/<vendor>/<codec-id>.yaml` files.
type EndDeviceCodecs struct {
	CodecID         string
	UplinkDecoder   EndDeviceDecoderCodec `yaml:"uplinkDecoder"`
	DownlinkDecoder EndDeviceDecoderCodec `yaml:"downlinkDecoder"`
	DownlinkEncoder EndDeviceEncoderCodec `yaml:"downlinkEncoder"`
}
