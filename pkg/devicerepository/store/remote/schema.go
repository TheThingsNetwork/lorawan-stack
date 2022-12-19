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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

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
			pb.Dimensions.Width = &wrapperspb.FloatValue{Value: w}
		}
		if h := d.Dimensions.Height; h > 0 {
			pb.Dimensions.Height = &wrapperspb.FloatValue{Value: h}
		}
		if d := d.Dimensions.Diameter; d > 0 {
			pb.Dimensions.Diameter = &wrapperspb.FloatValue{Value: d}
		}
		if l := d.Dimensions.Length; l > 0 {
			pb.Dimensions.Length = &wrapperspb.FloatValue{Value: l}
		}
	}

	if w := d.Weight; w > 0 {
		pb.Weight = &wrapperspb.FloatValue{Value: w}
	}

	if battery := d.Battery; battery != nil {
		pb.Battery = &ttnpb.EndDeviceModel_Battery{
			Replaceable: &wrapperspb.BoolValue{Value: d.Battery.Replaceable},
			Type:        d.Battery.Type,
		}
	}

	if oc := d.OperatingConditions; oc != nil {
		pb.OperatingConditions = &ttnpb.EndDeviceModel_OperatingConditions{}

		if rh := oc.RelativeHumidity; rh != nil {
			pb.OperatingConditions.RelativeHumidity = &ttnpb.EndDeviceModel_OperatingConditions_Limits{
				Min: &wrapperspb.FloatValue{Value: rh.Min},
				Max: &wrapperspb.FloatValue{Value: rh.Max},
			}
		}

		if temp := oc.Temperature; temp != nil {
			pb.OperatingConditions.Temperature = &ttnpb.EndDeviceModel_OperatingConditions_Limits{
				Min: &wrapperspb.FloatValue{Value: temp.Min},
				Max: &wrapperspb.FloatValue{Value: temp.Max},
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
