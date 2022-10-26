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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"gopkg.in/yaml.v2"
)

func durationPtr(t time.Duration) *time.Duration { return &t }

func TestProfile(t *testing.T) {
	for _, tc := range []struct {
		profile  string
		codec    string
		template *ttnpb.EndDeviceTemplate
	}{
		{
			profile: "example-1",
			template: &ttnpb.EndDeviceTemplate{
				EndDevice: &ttnpb.EndDevice{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_2,
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
					SupportsJoin:      true,
					MacSettings: &ttnpb.MACSettings{
						Supports_32BitFCnt: &ttnpb.BoolValue{Value: true},
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
			},
		},
		{
			profile: "example-2",
			codec:   "codec",
			template: &ttnpb.EndDeviceTemplate{
				EndDevice: &ttnpb.EndDevice{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
					},
					MacSettings: &ttnpb.MACSettings{
						Rx1Delay:          &ttnpb.RxDelayValue{Value: ttnpb.RxDelay_RX_DELAY_1},
						Rx1DataRateOffset: &ttnpb.DataRateOffsetValue{Value: ttnpb.DataRateOffset_DATA_RATE_OFFSET_0},
						Rx2DataRateIndex:  &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex_DATA_RATE_3},
						Rx2Frequency:      &ttnpb.FrequencyValue{Value: 869525000},
						FactoryPresetFrequencies: []uint64{
							868100000,
							868300000,
							868500000,
							867100000,
							867300000,
							867500000,
							867700000,
							867900000,
						},
						Supports_32BitFCnt: &ttnpb.BoolValue{Value: true},
					},
					MacState: &ttnpb.MACState{
						DesiredParameters: &ttnpb.MACParameters{
							MaxEirp: 14,
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
					"formatters",
					"mac_settings.rx1_delay",
					"mac_settings.rx1_data_rate_offset",
					"mac_settings.rx2_data_rate_index",
					"mac_settings.rx2_frequency",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.factory_preset_frequencies",
					"mac_state.desired_parameters.max_eirp",
				),
			},
		},
		{
			profile: "class-b-profile",
			template: &ttnpb.EndDeviceTemplate{
				EndDevice: &ttnpb.EndDevice{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_2,
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
					SupportsClassB:    true,
					SupportsJoin:      true,
					MacSettings: &ttnpb.MACSettings{
						ClassBTimeout:         ttnpb.ProtoDurationPtr(8 * time.Second),
						PingSlotPeriodicity:   &ttnpb.PingSlotPeriodValue{Value: ttnpb.PingSlotPeriod_PING_EVERY_16S},
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex_DATA_RATE_3},
						PingSlotFrequency:     &ttnpb.ZeroableFrequencyValue{Value: 868300000},
					},
				},
				FieldMask: ttnpb.FieldMask(
					"version_ids",
					"supports_join",
					"supports_class_b",
					"supports_class_c",
					"lorawan_version",
					"lorawan_phy_version",
					"mac_settings.class_b_timeout",
					"mac_settings.ping_slot_data_rate_index",
					"mac_settings.ping_slot_frequency",
					"mac_settings.ping_slot_periodicity",
				),
			},
		},
		{
			profile: "class-c-profile",
			template: &ttnpb.EndDeviceTemplate{
				EndDevice: &ttnpb.EndDevice{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_2,
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
					SupportsClassC:    true,
					SupportsJoin:      true,
					MacSettings: &ttnpb.MACSettings{
						ClassCTimeout: ttnpb.ProtoDurationPtr(64 * time.Second),
					},
				},
				FieldMask: ttnpb.FieldMask(
					"version_ids",
					"supports_join",
					"supports_class_b",
					"supports_class_c",
					"lorawan_version",
					"lorawan_phy_version",
					"mac_settings.class_c_timeout",
				),
			},
		},
	} {
		t.Run(tc.profile, func(t *testing.T) {
			a := assertions.New(t)
			b, err := os.ReadFile(filepath.Join("testdata", "vendor", "full-vendor", tc.profile+".yaml"))
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			profile := &store.EndDeviceProfile{}
			err = yaml.UnmarshalStrict(b, profile)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			ids := &ttnpb.EndDeviceVersionIdentifiers{}
			fwProfile := &ttnpb.EndDeviceModel_FirmwareVersion_Profile{
				CodecId: tc.codec,
			}

			tc.template.EndDevice.VersionIds = ids
			template, err := profile.ToTemplatePB(ids, fwProfile)
			a.So(err, should.BeNil)
			a.So(template, should.Resemble, tc.template)
		})
	}
}
