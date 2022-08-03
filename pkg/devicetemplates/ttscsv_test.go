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
	_ "embed"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/devicetemplates"
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
)

func TestTTSCSVConverter(t *testing.T) {
	t.Parallel()

	tts := GetConverter("the-things-stack-csv")
	a := assertions.New(t)
	if !a.So(tts, should.NotBeNil) {
		t.FailNow()
	}

	for _, tc := range []struct {
		name           string
		reader         io.Reader
		validateError  func(a *assertions.Assertion, err error)
		validateResult func(a *assertions.Assertion, templates []*ttnpb.EndDeviceTemplate, count int)
		nExpect        int
	}{
		{
			name:   "AllColumns",
			reader: bytes.NewBufferString(csvAllColumns),
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
			name:   "ExtraColumns",
			reader: bytes.NewBufferString(csvExtraColumns),
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
			name:   "EmptyString",
			reader: bytes.NewBufferString(""),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.NotBeNil)
			},
			nExpect: 0,
		},
		{
			name:   "InvalidDevEUI",
			reader: bytes.NewBufferString(csvInvalidDevEUI),
			validateError: func(a *assertions.Assertion, err error) {
				a.So(err, should.NotBeNil)
			},
			nExpect: 0,
		},
		{
			name:   "GenerateDeviceID",
			reader: bytes.NewBufferString(csvGenerateDeviceID),
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
			name:   "AppEUI",
			reader: bytes.NewBufferString(csvAppEUI),
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
			name:   "CommaSeparated",
			reader: bytes.NewBufferString(csvCommaSeparated),
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
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.Context()
			ch := make(chan *ttnpb.EndDeviceTemplate)

			wg := sync.WaitGroup{}
			wg.Add(2)
			var err error
			templates := []*ttnpb.EndDeviceTemplate{}
			go func() {
				err = tts.Convert(ctx, tc.reader, ch)
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

			a := assertions.New(t)
			tc.validateError(a, err)
			if tc.validateResult != nil {
				tc.validateResult(a, templates, tc.nExpect)
			}
		})
	}
}
