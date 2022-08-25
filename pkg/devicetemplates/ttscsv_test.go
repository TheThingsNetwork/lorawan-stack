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

package devicetemplates_test

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	mockdr "go.thethings.network/lorawan-stack/v3/pkg/devicerepository/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/devicetemplateconverter/profilefetcher"
	. "go.thethings.network/lorawan-stack/v3/pkg/devicetemplates"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	//go:embed testdata/all_columns.csv
	csvAllColumns string
	//go:embed testdata/extra_columns.csv
	csvExtraColumns string
	//go:embed testdata/invalid_deveui.csv
	csvInvalidDevEUI string
	//go:embed testdata/generate_device_id.csv
	csvGenerateDeviceID string
	//go:embed testdata/appeui.csv
	csvAppEUI string
	//go:embed testdata/comma_separated.csv
	csvCommaSeparated string
	//go:embed testdata/with_version_id.csv
	csvVendorID string
	//go:embed testdata/without_version_id.csv
	csvNoVendorID string
)

var mockErr = errors.DefineInternal("mock_internal_error", "An error is expected")

func TestTTSCSVConverter(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		reader         io.Reader
		converter      Converter
		fillContext    func(context.Context) context.Context
		validateError  func(a *assertions.Assertion, err error)
		validateResult func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int)
		nExpect        int
	}{
		{
			name:      "AllColumns",
			reader:    bytes.NewBufferString(csvAllColumns),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[2]
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"description",
					"frequency_plan_id",
					"ids.dev_eui",
					"ids.device_id",
					"ids.join_eui",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.rx1_delay",
					"mac_settings.supports_32_bit_f_cnt",
					"name",
					"root_keys.app_key.key",
					"root_keys.app_key.key",
					"root_keys.nwk_key.key",
					"session.dev_addr",
					"session.last_a_f_cnt_down",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"supports_class_c",
					"supports_join",
					"version_ids.band_id",
					"version_ids.brand_id",
					"version_ids.firmware_version",
					"version_ids.hardware_version",
					"version_ids.model_id",
				), should.BeNil)
			},
			nExpect: 3,
		},
		{
			name:      "ExtraColumns",
			reader:    bytes.NewBufferString(csvExtraColumns),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"ids.dev_eui",
				), should.BeNil)
			},
			nExpect: 1,
		},
		{
			name:      "EmptyString",
			reader:    bytes.NewBufferString(""),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.NotBeNil)
			},
			nExpect: 0,
		},
		{
			name:      "InvalidDevEUI",
			reader:    bytes.NewBufferString(csvInvalidDevEUI),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.NotBeNil)
			},
			nExpect: 0,
		},
		{
			name:      "GenerateDeviceID",
			reader:    bytes.NewBufferString(csvGenerateDeviceID),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(dev.EndDevice.Ids.DeviceId, should.Equal, "eui-111111111111111a")
			},
			nExpect: 1,
		},
		{
			name:      "AppEUI",
			reader:    bytes.NewBufferString(csvAppEUI),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(
					dev.EndDevice.Ids.JoinEui,
					should.Resemble,
					types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
				)
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"ids.dev_eui",
					"ids.join_eui",
				), should.BeNil)
			},
			nExpect: 1,
		},
		{
			name:      "CommaSeparated",
			reader:    bytes.NewBufferString(csvCommaSeparated),
			converter: GetConverter("the-things-stack-csv"),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"ids.device_id",
					"ids.dev_eui",
					"ids.join_eui",
					"root_keys.app_key.key",
					"root_keys.nwk_key.key",
				), should.BeNil)
			},
			nExpect: 3,
		},
		{
			name:      "VendorID/No valid fields",
			reader:    bytes.NewBufferString(csvNoVendorID),
			converter: GetConverter("the-things-stack-csv"),
			fillContext: func(ctx context.Context) context.Context {
				return profilefetcher.NewContextWithFetcher(ctx, &mockTemplateFetcher{})
			},
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(
					dev.EndDevice.Ids.JoinEui,
					should.Resemble,
					types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
				)
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"frequency_plan_id",
					"ids.dev_eui",
					"ids.device_id",
					"ids.join_eui",
				), should.BeNil)
			},
			nExpect: 1,
		},
		{
			name:      "VendorID/Error fetching",
			reader:    bytes.NewBufferString(csvVendorID),
			converter: GetConverter("the-things-stack-csv"),
			fillContext: func(ctx context.Context) context.Context {
				return profilefetcher.NewContextWithFetcher(ctx, &mockTemplateFetcher{
					Err: mockErr.New(),
				})
			},
			validateError: func(a *assertions.Assertion, err error) {
				a.So(errors.IsInternal(err), should.BeTrue)
			},
			nExpect: 1,
		},
		{
			name:      "VendorID/Valid",
			reader:    bytes.NewBufferString(csvVendorID),
			converter: GetConverter("the-things-stack-csv"),
			fillContext: func(ctx context.Context) context.Context {
				return profilefetcher.NewContextWithFetcher(ctx, &mockTemplateFetcher{
					MockDR: func() *mockdr.MockDR {
						dr := mockdr.New()
						dr.EndDeviceTemplate = mockProfileTemplate
						return dr
					}(),
				})
			},
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(
					dev.EndDevice.Ids.JoinEui,
					should.Resemble,
					types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
				)
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"frequency_plan_id",
					"ids.dev_eui",
					"ids.device_id",
					"ids.join_eui",
					"lorawan_version",
					"lorawan_phy_version",
					"supports_class_c",
				), should.BeNil)

				// validating if end device profile information was applied
				a.So(dev.EndDevice.LorawanPhyVersion, should.Equal, ttnpb.PHYVersion_PHY_V1_0_2_REV_A)
				a.So(dev.EndDevice.SupportsClassC, should.BeTrue)
				// LoRaWAN Version was fetched from profile as MAC_V1_0_2 but provided by csv as MAC_V1_0_4.
				a.So(dev.EndDevice.LorawanVersion, should.Equal, ttnpb.MACVersion_MAC_V1_0_4)
			},
			nExpect: 1,
		},
		{
			name:      "VendorID/Valid fallback value",
			reader:    bytes.NewBufferString(csvNoVendorID),
			converter: GetConverter("the-things-stack-csv"),
			fillContext: func(ctx context.Context) context.Context {
				ctx = profilefetcher.NewContextWithFetcher(ctx, &mockTemplateFetcher{
					MockDR: func() *mockdr.MockDR {
						dr := mockdr.New()
						dr.EndDeviceTemplate = mockProfileTemplate
						return dr
					}(),
				})
				return NewContextWithProfileIDs(ctx, &ttnpb.EndDeviceVersionIdentifiers{
					BandId:          "EU_863_870",
					BrandId:         "the-things-industries",
					FirmwareVersion: "1.0",
					HardwareVersion: "1.1",
					ModelId:         "generic-node-sensor-edition",
				})
			},
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.BeNil)
			},
			validateResult: func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int) {
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(
					dev.EndDevice.Ids.JoinEui,
					should.Resemble,
					types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
				)
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"ids.dev_eui",
					"ids.join_eui",
					"lorawan_version",
					"lorawan_phy_version",
					"supports_class_c",
				), should.BeNil)

				// validating if end device profile information was applied
				a.So(dev.EndDevice.LorawanPhyVersion, should.Equal, ttnpb.PHYVersion_PHY_V1_0_2_REV_A)
				a.So(dev.EndDevice.SupportsClassC, should.BeTrue)
				// LoRaWAN Version was fetched from profile as MAC_V1_0_2 but provided by csv as MAC_V1_0_4.
				a.So(dev.EndDevice.LorawanVersion, should.Equal, ttnpb.MACVersion_MAC_V1_0_4)
			},
			nExpect: 1,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a, ctx := test.New(t)
			if tc.fillContext != nil {
				ctx = tc.fillContext(ctx)
			}
			ch := make(chan *ttnpb.EndDeviceTemplate)

			wg := sync.WaitGroup{}
			wg.Add(2)
			var err error
			templates := []*ttnpb.EndDeviceTemplate{}
			go func() {
				err = tc.converter.Convert(ctx, tc.reader, ch)
				wg.Done()
			}()
			go func() {
				for t := range ch {
					templates = append(templates, t)
				}
				wg.Done()
			}()

			complete := make(chan struct{})
			go func() {
				defer func() {
					complete <- struct{}{}
				}()
				wg.Wait()
			}()

			select {
			case <-complete:
			case <-time.After(time.Second):
				t.Error("Timed out waiting for converter")
				t.FailNow()
			}

			tc.validateError(a, err)
			if tc.validateResult != nil {
				tc.validateResult(a, templates, tc.nExpect)
			}
		})
	}
}

// mockProfileTemplate is used in the ValidID tests.
var mockProfileTemplate = &ttnpb.EndDeviceTemplate{
	EndDevice: &ttnpb.EndDevice{
		LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_2,
		LorawanPhyVersion: ttnpb.PHYVersion_PHY_V1_0_2_REV_A,
		SupportsClassC:    true,
	},
	FieldMask: ttnpb.FieldMask("lorawan_version", "lorawan_phy_version", "supports_class_c"),
}

type mockTemplateFetcher struct {
	MockDR *mockdr.MockDR
	Err    error
}

// GetTemplate makes a request to the Device Repository server with its predefined call options.
func (tf *mockTemplateFetcher) GetTemplate(
	ctx context.Context,
	in *ttnpb.GetTemplateRequest,
) (*ttnpb.EndDeviceTemplate, error) {
	if tf.Err != nil {
		return nil, tf.Err
	}
	if tf.MockDR == nil {
		return nil, nil
	}
	return tf.MockDR.GetTemplate(ctx, in)
}
