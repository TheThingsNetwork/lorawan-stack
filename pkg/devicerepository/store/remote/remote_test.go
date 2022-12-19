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

package remote_test

import (
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store/remote"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestRemoteStore(t *testing.T) {
	a := assertions.New(t)

	s := remote.NewRemoteStore(fetch.FromFilesystem("testdata"))
	t.Run("TestGetBrands", func(t *testing.T) {
		t.Run("Limit", func(t *testing.T) {
			list, err := s.GetBrands(store.GetBrandsRequest{
				Paths: []string{
					"brand_id",
					"name",
				},
				Limit: 1,
			})
			a.So(err, should.BeNil)
			a.So(list.Brands, should.Resemble, []*ttnpb.EndDeviceBrand{
				{
					BrandId: "foo-vendor",
					Name:    "Foo Vendor",
				},
			})
		})

		t.Run("SecondPage", func(t *testing.T) {
			list, err := s.GetBrands(store.GetBrandsRequest{
				Paths: []string{
					"brand_id",
					"name",
				},
				Limit: 1,
				Page:  2,
			})
			a.So(err, should.BeNil)
			a.So(list.Brands, should.Resemble, []*ttnpb.EndDeviceBrand{
				{
					BrandId: "full-vendor",
					Name:    "Full Vendor",
				},
			})
		})

		t.Run("Paths", func(t *testing.T) {
			list, err := s.GetBrands(store.GetBrandsRequest{
				Paths: ttnpb.EndDeviceBrandFieldPathsNested,
			})
			a.So(err, should.BeNil)
			a.So(list.Brands, should.Resemble, []*ttnpb.EndDeviceBrand{
				{
					BrandId:              "foo-vendor",
					Name:                 "Foo Vendor",
					LoraAllianceVendorId: 42,
				},
				{
					BrandId:                       "full-vendor",
					Name:                          "Full Vendor",
					LoraAllianceVendorId:          44,
					Email:                         "mail@example.com",
					Website:                       "example.org",
					PrivateEnterpriseNumber:       42,
					OrganizationUniqueIdentifiers: []string{"010203", "030405"},
					Logo:                          "logo.svg",
				},
			})
		})
	})

	t.Run("TestGetModels", func(t *testing.T) {
		t.Run("AllBrands", func(t *testing.T) {
			list, err := s.GetModels(store.GetModelsRequest{
				Paths: []string{
					"brand_id",
					"model_id",
					"name",
				},
			})
			a.So(err, should.BeNil)
			a.So(list.Models, should.Resemble, []*ttnpb.EndDeviceModel{
				{
					BrandId: "foo-vendor",
					ModelId: "dev1",
					Name:    "Device 1",
				},
				{
					BrandId: "foo-vendor",
					ModelId: "dev2",
					Name:    "Device 2",
				},
				{
					BrandId: "full-vendor",
					ModelId: "full-device",
					Name:    "Full Device",
				},
			})
		})

		t.Run("Limit", func(t *testing.T) {
			list, err := s.GetModels(store.GetModelsRequest{
				BrandID: "foo-vendor",
				Limit:   1,
				Paths: []string{
					"brand_id",
					"model_id",
					"name",
				},
			})
			a.So(err, should.BeNil)
			a.So(list.Models, should.Resemble, []*ttnpb.EndDeviceModel{
				{
					BrandId: "foo-vendor",
					ModelId: "dev1",
					Name:    "Device 1",
				},
			})
		})

		t.Run("Offset", func(t *testing.T) {
			list, err := s.GetModels(store.GetModelsRequest{
				BrandID: "foo-vendor",
				Limit:   1,
				Page:    2,
				Paths: []string{
					"brand_id",
					"model_id",
					"name",
				},
			})
			a.So(err, should.BeNil)
			a.So(list.Models, should.Resemble, []*ttnpb.EndDeviceModel{
				{
					BrandId: "foo-vendor",
					ModelId: "dev2",
					Name:    "Device 2",
				},
			})
		})

		t.Run("Paths", func(t *testing.T) {
			list, err := s.GetModels(store.GetModelsRequest{
				BrandID: "foo-vendor",
				Paths:   ttnpb.EndDeviceModelFieldPathsNested,
			})
			a.So(err, should.BeNil)
			a.So(list.Models, should.Resemble, []*ttnpb.EndDeviceModel{
				{
					BrandId:     "foo-vendor",
					ModelId:     "dev1",
					Name:        "Device 1",
					Description: "My Description",
					HardwareVersions: []*ttnpb.EndDeviceModel_HardwareVersion{
						{
							Version:    "1.0",
							Numeric:    1,
							PartNumber: "P4RTN0",
						},
					},
					FirmwareVersions: []*ttnpb.EndDeviceModel_FirmwareVersion{
						{
							Version:                   "1.0",
							SupportedHardwareVersions: []string{"1.0"},
							Profiles: map[string]*ttnpb.EndDeviceModel_FirmwareVersion_Profile{
								"EU_863_870": {
									ProfileId:        "profile1",
									LorawanCertified: true,
								},
								"US_902_928": {
									CodecId:          "foo-codec",
									ProfileId:        "profile2",
									LorawanCertified: true,
								},
							},
						},
					},
				},
				{
					BrandId:     "foo-vendor",
					ModelId:     "dev2",
					Name:        "Device 2",
					Description: "My Description 2",
					HardwareVersions: []*ttnpb.EndDeviceModel_HardwareVersion{
						{
							Version:    "2.0",
							Numeric:    2,
							PartNumber: "P4RTN02",
						},
					},
					FirmwareVersions: []*ttnpb.EndDeviceModel_FirmwareVersion{
						{
							Version:                   "1.1",
							SupportedHardwareVersions: []string{"2.0"},
							Profiles: map[string]*ttnpb.EndDeviceModel_FirmwareVersion_Profile{
								"EU_433": {
									CodecId:          "foo-codec",
									ProfileId:        "profile2",
									LorawanCertified: true,
								},
							},
						},
					},
					Sensors: []string{"temperature"},
				},
			})
		})

		t.Run("Full", func(t *testing.T) {
			a := assertions.New(t)
			list, err := s.GetModels(store.GetModelsRequest{
				BrandID: "full-vendor",
				Paths:   ttnpb.EndDeviceModelFieldPathsNested,
			})
			a.So(err, should.BeNil)
			a.So(list.Models[0], should.Resemble, &ttnpb.EndDeviceModel{
				BrandId:     "full-vendor",
				ModelId:     "full-device",
				Name:        "Full Device",
				Description: "A description",
				HardwareVersions: []*ttnpb.EndDeviceModel_HardwareVersion{
					{
						Version:    "0.1",
						Numeric:    1,
						PartNumber: "0A0B",
					},
					{
						Version:    "0.2",
						Numeric:    2,
						PartNumber: "0A0C",
					},
				},
				FirmwareVersions: []*ttnpb.EndDeviceModel_FirmwareVersion{
					{
						Version:                   "1.0",
						SupportedHardwareVersions: []string{"0.1", "0.2"},
						Profiles: map[string]*ttnpb.EndDeviceModel_FirmwareVersion_Profile{
							"EU_863_870": {
								VendorId:  "module-maker",
								ProfileId: "module-profile",
								CodecId:   "",
							},
							"US_902_928": {
								ProfileId: "full-profile",
								CodecId:   "codec",
							},
						},
					},
				},
				Sensors: []string{"temperature", "gas"},
				Dimensions: &ttnpb.EndDeviceModel_Dimensions{
					Width:    &wrapperspb.FloatValue{Value: 1},
					Height:   &wrapperspb.FloatValue{Value: 2},
					Diameter: &wrapperspb.FloatValue{Value: 3},
					Length:   &wrapperspb.FloatValue{Value: 4},
				},
				Weight: &wrapperspb.FloatValue{Value: 5},
				Battery: &ttnpb.EndDeviceModel_Battery{
					Replaceable: &wrapperspb.BoolValue{Value: true},
					Type:        "AAA",
				},
				OperatingConditions: &ttnpb.EndDeviceModel_OperatingConditions{
					Temperature: &ttnpb.EndDeviceModel_OperatingConditions_Limits{
						Min: &wrapperspb.FloatValue{Value: 1},
						Max: &wrapperspb.FloatValue{Value: 2},
					},
					RelativeHumidity: &ttnpb.EndDeviceModel_OperatingConditions_Limits{
						Min: &wrapperspb.FloatValue{Value: 3},
						Max: &wrapperspb.FloatValue{Value: 4},
					},
				},
				IpCode:          "IP67",
				KeyProvisioning: []ttnpb.KeyProvisioning{ttnpb.KeyProvisioning_KEY_PROVISIONING_CUSTOM},
				KeySecurity:     ttnpb.KeySecurity_KEY_SECURITY_READ_PROTECTED,
				Photos: &ttnpb.EndDeviceModel_Photos{
					Main:  "a.jpg",
					Other: []string{"b.jpg", "c.jpg"},
				},
				Videos: &ttnpb.EndDeviceModel_Videos{
					Main:  "a.mp4",
					Other: []string{"b.mp4", "https://youtube.com/watch?v=c.mp4"},
				},
				ProductUrl:   "https://product.vendor.io",
				DatasheetUrl: "https://production.vendor.io/datasheet.pdf",
				Resellers: []*ttnpb.EndDeviceModel_Reseller{
					{
						Name:   "Reseller 1",
						Region: []string{"European Union"},
						Url:    "https://example.com/eu",
					},
					{
						Name:   "Reseller 2",
						Region: []string{"United States", "Canada"},
						Url:    "https://example.com/na",
					},
				},
				Compliances: &ttnpb.EndDeviceModel_Compliances{
					Safety: []*ttnpb.EndDeviceModel_Compliances_Compliance{
						{
							Body:     "IEC",
							Norm:     "EN",
							Standard: "62368-1",
						},
						{
							Body:     "IEC",
							Norm:     "EN",
							Standard: "60950-22",
						},
					},
					RadioEquipment: []*ttnpb.EndDeviceModel_Compliances_Compliance{
						{
							Body:     "ETSI",
							Norm:     "EN",
							Standard: "301 489-1",
							Version:  "2.2.0",
						},
						{
							Body:     "ETSI",
							Norm:     "EN",
							Standard: "301 489-3",
							Version:  "2.1.0",
						},
					},
				},
				AdditionalRadios: []string{"nfc", "wifi"},
			})
		})
	})

	t.Run("TestGetTemplate", func(t *testing.T) {
		t.Run("ByEndDeviceVersionIdentifiers", func(t *testing.T) {
			template, err := s.GetTemplate(&ttnpb.GetTemplateRequest{
				VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
					BrandId:         "foo-vendor",
					ModelId:         "dev1",
					FirmwareVersion: "1.0",
					BandId:          "EU_863_870",
				},
			}, nil)
			a.So(err, should.BeNil)
			a.So(template, should.Resemble, &ttnpb.EndDeviceTemplate{
				EndDevice: &ttnpb.EndDevice{
					VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
						BrandId:         "foo-vendor",
						ModelId:         "dev1",
						FirmwareVersion: "1.0",
						BandId:          "EU_863_870",
					},
					LorawanPhyVersion: ttnpb.PHYVersion_PHY_V1_0_3_REV_A,
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					SupportsJoin:      true,
					MacSettings: &ttnpb.MACSettings{
						Supports_32BitFCnt: &ttnpb.BoolValue{
							Value: true,
						},
					},
				},
				FieldMask: ttnpb.FieldMask(
					"version_ids",
					"supports_join",
					"supports_class_b",
					"supports_class_c",
					"lorawan_version",
					"lorawan_phy_version",
					"mac_settings.supports_32_bit_f_cnt",
				),
			})
		})

		t.Run("ByProfile", func(t *testing.T) {
			template, err := s.GetTemplate(&ttnpb.GetTemplateRequest{
				VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
					BrandId: "foo-vendor",
				},
				EndDeviceProfileIds: &ttnpb.GetTemplateRequest_EndDeviceProfileIdentifiers{
					VendorId:        42,
					VendorProfileId: 0,
				},
			}, &store.EndDeviceProfile{
				VendorProfileID:           0,
				RegionalParametersVersion: "RP001-1.0.3-RevA",
				MACVersion:                ttnpb.MACVersion_MAC_V1_0_3,
				SupportsJoin:              true,
				Supports32BitFCnt:         true,
			})
			a.So(err, should.BeNil)
			a.So(template, should.Resemble, &ttnpb.EndDeviceTemplate{
				EndDevice: &ttnpb.EndDevice{
					VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
						BrandId: "foo-vendor",
					},
					LorawanPhyVersion: ttnpb.PHYVersion_PHY_V1_0_3_REV_A,
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					SupportsJoin:      true,
					MacSettings: &ttnpb.MACSettings{
						Supports_32BitFCnt: &ttnpb.BoolValue{
							Value: true,
						},
					},
				},
				FieldMask: ttnpb.FieldMask(
					"version_ids",
					"supports_join",
					"supports_class_b",
					"supports_class_c",
					"lorawan_version",
					"lorawan_phy_version",
					"mac_settings.supports_32_bit_f_cnt",
				),
			})
		})
	})

	t.Run("TestGetCodecs", func(t *testing.T) {
		t.Run("Missing", func(t *testing.T) {
			a := assertions.New(t)

			for _, ids := range []*ttnpb.EndDeviceVersionIdentifiers{
				{
					BrandId: "unknown-vendor",
				},
				{
					BrandId: "foo-vendor",
					ModelId: "unknown-model",
				},
				{
					BrandId:         "foo-vendor",
					ModelId:         "dev1",
					FirmwareVersion: "unknown-version",
				},
				{
					BrandId:         "foo-vendor",
					ModelId:         "dev1",
					FirmwareVersion: "1.0",
					BandId:          "unknown-band",
				},
			} {
				codec, err := s.GetDownlinkDecoder(&ttnpb.GetPayloadFormatterRequest{VersionIds: ids})
				a.So(errors.IsNotFound(err), should.BeTrue)
				a.So(codec, should.Equal, nil)
			}
		})
		for _, tc := range []struct {
			name  string
			f     func(store.GetCodecRequest) (interface{}, error)
			codec interface{}
		}{
			{
				name: "UplinkDecoder",
				f:    func(req store.GetCodecRequest) (interface{}, error) { return s.GetUplinkDecoder(req) },
				codec: &ttnpb.MessagePayloadDecoder{
					Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
					FormatterParameter: "// uplink decoder\n",
				},
			},
			{
				name: "DownlinkDecoder",
				f:    func(req store.GetCodecRequest) (interface{}, error) { return s.GetDownlinkDecoder(req) },
				codec: &ttnpb.MessagePayloadDecoder{
					Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
					FormatterParameter: "// downlink decoder\n",
				},
			},
			{
				name: "DownlinkEncoder",
				f:    func(req store.GetCodecRequest) (interface{}, error) { return s.GetDownlinkEncoder(req) },
				codec: &ttnpb.MessagePayloadEncoder{
					Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
					FormatterParameter: "// downlink encoder\n",
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				a := assertions.New(t)

				versionIDs := &ttnpb.EndDeviceVersionIdentifiers{
					BrandId:         "foo-vendor",
					ModelId:         "dev2",
					FirmwareVersion: "1.1",
					BandId:          "EU_433",
				}
				codec, err := tc.f(&ttnpb.GetPayloadFormatterRequest{VersionIds: versionIDs})
				a.So(err, should.BeNil)
				a.So(codec, should.Resemble, tc.codec)
			})
		}

		t.Run("Examples", func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				f     func(store.GetCodecRequest) (interface{}, error)
				codec interface{}
			}{
				{
					name: "UplinkDecoder",
					f:    func(req store.GetCodecRequest) (interface{}, error) { return s.GetUplinkDecoder(req) },
					codec: &ttnpb.MessagePayloadDecoder{
						Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
						FormatterParameter: "// uplink decoder\n",
						Examples: []*ttnpb.MessagePayloadDecoder_Example{{
							Description: "dummy example",
							Input: &ttnpb.EncodedMessagePayload{
								FPort:      10,
								FrmPayload: []byte{1, 1, 100},
							},
							Output: &ttnpb.DecodedMessagePayload{
								Data: mustStruct(map[string]interface{}{
									"type":  "BATTERY_STATUS",
									"value": 100,
									"nested": map[string]interface{}{
										"key":  "value",
										"list": []int{1, 2, 3},
									},
								}),
								Warnings: []string{"warn1"},
								Errors:   []string{"err1"},
							},
						}},
					},
				},
				{
					name: "DownlinkDecoder",
					f:    func(req store.GetCodecRequest) (interface{}, error) { return s.GetDownlinkDecoder(req) },
					codec: &ttnpb.MessagePayloadDecoder{
						Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
						FormatterParameter: "// downlink decoder\n",
						Examples: []*ttnpb.MessagePayloadDecoder_Example{{
							Description: "downlink decode example",
							Input: &ttnpb.EncodedMessagePayload{
								FPort:      20,
								FrmPayload: []byte{1, 5},
							},
							Output: &ttnpb.DecodedMessagePayload{
								Data: mustStruct(map[string]interface{}{
									"action": "DIM",
									"value":  5,
								}),
								Warnings: []string{"warn1"},
								Errors:   []string{"err1"},
							},
						}},
					},
				},
				{
					name: "DownlinkEncoder",
					f:    func(req store.GetCodecRequest) (interface{}, error) { return s.GetDownlinkEncoder(req) },
					codec: &ttnpb.MessagePayloadEncoder{
						Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
						FormatterParameter: "// downlink encoder\n",
						Examples: []*ttnpb.MessagePayloadEncoder_Example{{
							Description: "downlink encode example",
							Input: &ttnpb.DecodedMessagePayload{
								Data: mustStruct(map[string]interface{}{
									"action": "DIM",
									"value":  5,
								}),
							},
							Output: &ttnpb.EncodedMessagePayload{
								FPort:      20,
								FrmPayload: []byte{1, 5},
								Warnings:   []string{"warn1"},
								Errors:     []string{"err1"},
							},
						}},
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					a := assertions.New(t)

					versionIDs := &ttnpb.EndDeviceVersionIdentifiers{
						BrandId:         "foo-vendor",
						ModelId:         "dev2",
						FirmwareVersion: "1.1",
						BandId:          "EU_433",
					}
					codec, err := tc.f(&ttnpb.GetPayloadFormatterRequest{
						VersionIds: versionIDs,
						FieldMask:  ttnpb.FieldMask("examples"),
					})
					a.So(err, should.BeNil)
					a.So(codec, should.Resemble, tc.codec)
				})
			}
		})
	})
}

func mustStruct(d map[string]interface{}) *pbtypes.Struct {
	v, err := gogoproto.Struct(d)
	if err != nil {
		panic(err)
	}
	return v
}
