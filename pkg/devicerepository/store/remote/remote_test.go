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
					BrandID: "foo-vendor",
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
					BrandID: "full-vendor",
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
					BrandID:              "foo-vendor",
					Name:                 "Foo Vendor",
					LoRaAllianceVendorID: 42,
				},
				{
					BrandID:                       "full-vendor",
					Name:                          "Full Vendor",
					LoRaAllianceVendorID:          44,
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
					BrandID: "foo-vendor",
					ModelID: "dev1",
					Name:    "Device 1",
				},
				{
					BrandID: "foo-vendor",
					ModelID: "dev2",
					Name:    "Device 2",
				},
				{
					BrandID: "full-vendor",
					ModelID: "full-device",
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
					BrandID: "foo-vendor",
					ModelID: "dev1",
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
					BrandID: "foo-vendor",
					ModelID: "dev2",
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
					BrandID:     "foo-vendor",
					ModelID:     "dev1",
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
									ProfileID:        "profile1",
									LoRaWANCertified: true,
								},
								"US_902_928": {
									CodecID:          "foo-codec",
									ProfileID:        "profile2",
									LoRaWANCertified: true,
								},
							},
						},
					},
				},
				{
					BrandID:     "foo-vendor",
					ModelID:     "dev2",
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
									CodecID:          "foo-codec",
									ProfileID:        "profile2",
									LoRaWANCertified: true,
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
				BrandID:     "full-vendor",
				ModelID:     "full-device",
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
								CodecID:   "",
								ProfileID: "full-profile2",
							},
							"US_902_928": {
								CodecID:   "codec",
								ProfileID: "full-profile",
							},
						},
					},
				},
				Sensors: []string{"temperature", "gas"},
				Dimensions: &ttnpb.EndDeviceModel_Dimensions{
					Width:    &pbtypes.FloatValue{Value: 1},
					Height:   &pbtypes.FloatValue{Value: 2},
					Diameter: &pbtypes.FloatValue{Value: 3},
					Length:   &pbtypes.FloatValue{Value: 4},
				},
				Weight: &pbtypes.FloatValue{Value: 5},
				Battery: &ttnpb.EndDeviceModel_Battery{
					Replaceable: &pbtypes.BoolValue{Value: true},
					Type:        "AAA",
				},
				OperatingConditions: &ttnpb.EndDeviceModel_OperatingConditions{
					Temperature: &ttnpb.EndDeviceModel_OperatingConditions_Limits{
						Min: &pbtypes.FloatValue{Value: 1},
						Max: &pbtypes.FloatValue{Value: 2},
					},
					RelativeHumidity: &ttnpb.EndDeviceModel_OperatingConditions_Limits{
						Min: &pbtypes.FloatValue{Value: 3},
						Max: &pbtypes.FloatValue{Value: 4},
					},
				},
				IPCode:          "IP67",
				KeyProvisioning: []ttnpb.KeyProvisioning{ttnpb.KEY_PROVISIONING_CUSTOM},
				KeySecurity:     ttnpb.KEY_SECURITY_READ_PROTECTED,
				Photos: &ttnpb.EndDeviceModel_Photos{
					Main:  "a.jpg",
					Other: []string{"b.jpg", "c.jpg"},
				},
				Videos: &ttnpb.EndDeviceModel_Videos{
					Main:  "a.mp4",
					Other: []string{"b.mp4", "https://youtube.com/watch?v=c.mp4"},
				},
				ProductURL:   "https://product.vendor.io",
				DatasheetURL: "https://production.vendor.io/datasheet.pdf",
				Resellers: []*ttnpb.EndDeviceModel_Reseller{
					{
						Name:   "Reseller 1",
						Region: []string{"European Union"},
						URL:    "https://example.com/eu",
					},
					{
						Name:   "Reseller 2",
						Region: []string{"United States", "Canada"},
						URL:    "https://example.com/na",
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

	t.Run("TestGetCodecs", func(t *testing.T) {
		t.Run("Missing", func(t *testing.T) {
			a := assertions.New(t)

			for _, ids := range []ttnpb.EndDeviceVersionIdentifiers{
				{
					BrandID: "unknown-vendor",
				},
				{
					BrandID: "foo-vendor",
					ModelID: "unknown-model",
				},
				{
					BrandID:         "foo-vendor",
					ModelID:         "dev1",
					FirmwareVersion: "unknown-version",
				},
				{
					BrandID:         "foo-vendor",
					ModelID:         "dev1",
					FirmwareVersion: "1.0",
					BandID:          "unknown-band",
				},
			} {
				codec, err := s.GetDownlinkDecoder(store.GetCodecRequest{VersionIDs: &ids})
				a.So(errors.IsNotFound(err), should.BeTrue)
				a.So(codec, should.Equal, nil)
			}
		})
		for _, tc := range []struct {
			name  string
			f     func(store.GetCodecRequest) (*ttnpb.MessagePayloadFormatter, error)
			codec string
		}{
			{
				name:  "UplinkDecoder",
				f:     s.GetUplinkDecoder,
				codec: "// uplink decoder\n",
			},
			{
				name:  "DownlinkDecoder",
				f:     s.GetDownlinkDecoder,
				codec: "// downlink decoder\n",
			},
			{
				name:  "DownlinkEncoder",
				f:     s.GetDownlinkEncoder,
				codec: "// downlink encoder\n",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				a := assertions.New(t)

				versionIDs := &ttnpb.EndDeviceVersionIdentifiers{
					BrandID:         "foo-vendor",
					ModelID:         "dev2",
					FirmwareVersion: "1.1",
					BandID:          "EU_433",
				}
				codec, err := tc.f(store.GetCodecRequest{VersionIDs: versionIDs})
				a.So(err, should.BeNil)
				a.So(codec, should.Resemble, &ttnpb.MessagePayloadFormatter{
					Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
					FormatterParameter: tc.codec,
				})
			})
		}

		t.Run("Examples", func(t *testing.T) {
			for _, tc := range []struct {
				name     string
				codec    string
				f        func(store.GetCodecRequest) (*ttnpb.MessagePayloadFormatter, error)
				examples []*ttnpb.MessagePayloadFormatter_Example
			}{
				{
					name:  "UplinkDecoder",
					codec: "// uplink decoder\n",
					f:     s.GetUplinkDecoder,
					examples: []*ttnpb.MessagePayloadFormatter_Example{{
						Description: "dummy example",
						Input: mustStruct(map[string]interface{}{
							"fPort": 10,
							"bytes": []int{1, 1, 100},
						}),
						Output: mustStruct(map[string]interface{}{
							"type":  "BATTERY_STATUS",
							"value": 100,
						}),
					}},
				},
				{
					name:  "DownlinkDecoder",
					codec: "// downlink decoder\n",
					f:     s.GetDownlinkDecoder,
					examples: []*ttnpb.MessagePayloadFormatter_Example{{
						Description: "downlink decode example",
						Input: mustStruct(map[string]interface{}{
							"action": "DIM",
							"value":  5,
						}),
						Output: mustStruct(map[string]interface{}{
							"fPort": 20,
							"bytes": []int{1, 5},
						}),
					}},
				},
				{
					name:  "DownlinkEncoder",
					codec: "// downlink encoder\n",
					f:     s.GetDownlinkEncoder,
					examples: []*ttnpb.MessagePayloadFormatter_Example{{
						Description: "downlink encode example",
						Input: mustStruct(map[string]interface{}{
							"fPort": 20,
							"bytes": []int{1, 5},
						}),
						Output: mustStruct(map[string]interface{}{
							"action": "DIM",
							"value":  5,
						}),
					}},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					a := assertions.New(t)

					versionIDs := &ttnpb.EndDeviceVersionIdentifiers{
						BrandID:         "foo-vendor",
						ModelID:         "dev2",
						FirmwareVersion: "1.1",
						BandID:          "EU_433",
					}
					codec, err := tc.f(store.GetCodecRequest{
						VersionIDs: versionIDs,
						Paths:      []string{"examples"},
					})
					a.So(err, should.BeNil)
					a.So(codec, should.Resemble, &ttnpb.MessagePayloadFormatter{
						Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
						FormatterParameter: tc.codec,
						Examples:           tc.examples,
					})
				})
			}
		})
	})

	t.Run("GetTemplate", func(t *testing.T) {
		t.Run("Missing", func(t *testing.T) {
			a := assertions.New(t)

			for _, ids := range []ttnpb.EndDeviceVersionIdentifiers{
				{
					BrandID: "unknown-vendor",
				},
				{
					BrandID: "foo-vendor",
					ModelID: "unknown-model",
				},
				{
					BrandID:         "foo-vendor",
					ModelID:         "dev1",
					FirmwareVersion: "unknown-version",
				},
				{
					BrandID:         "foo-vendor",
					ModelID:         "dev1",
					FirmwareVersion: "1.0",
					BandID:          "unknown-band",
				},
			} {
				tmpl, err := s.GetTemplate(&ids)
				a.So(errors.IsNotFound(err), should.BeTrue)
				a.So(tmpl, should.BeNil)
			}
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			tmpl, err := s.GetTemplate(&ttnpb.EndDeviceVersionIdentifiers{
				BrandID:         "foo-vendor",
				ModelID:         "dev2",
				FirmwareVersion: "1.1",
				HardwareVersion: "2.0",
				BandID:          "EU_433",
			})
			a.So(err, should.BeNil)
			a.So(tmpl, should.NotBeNil)
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
