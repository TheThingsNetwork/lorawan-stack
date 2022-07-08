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
)

func TestTTSCSVConverter(t *testing.T) {
	tts := GetConverter("the-things-stack-csv")
	a := assertions.New(t)
	if !a.So(tts, should.NotBeNil) {
		t.FailNow()
	}

	for _, tc := range []struct {
		name           string
		reader         io.Reader
		validateError  func(t *testing.T, err error)
		validateResult func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int)
		nExpect        int
	}{
		{
			name:   "AllColumns",
			reader: bytes.NewBufferString(csvAllColumns),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int) {
				a := assertions.New(t)
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
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int) {
				a := assertions.New(t)
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
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.NotBeNil)
			},
			nExpect: 0,
		},
		{
			name:   "InvalidDevEUI",
			reader: bytes.NewBufferString(csvInvalidDevEUI),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.NotBeNil)
			},
			nExpect: 0,
		},
		{
			name:   "GenerateDeviceID",
			reader: bytes.NewBufferString(csvGenerateDeviceID),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int) {
				a := assertions.New(t)
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(dev.EndDevice.Ids.DeviceId, should.Equal, "111111111111111a")
			},
			nExpect: 1,
		},
		{
			name:   "AppEUI",
			reader: bytes.NewBufferString(csvAppEUI),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int) {
				a := assertions.New(t)
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				dev := templates[0]
				a.So(dev.EndDevice.Ids.JoinEui, should.Resemble, &types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11})
				a.So(ttnpb.RequireFields(dev.FieldMask.Paths,
					"ids.dev_eui",
					"ids.join_eui",
				), should.BeNil)
			},
			nExpect: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
				for i := 0; i < tc.nExpect; i++ {
					templates = append(templates, <-ch)
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

			tc.validateError(t, err)
			if tc.validateResult != nil {
				tc.validateResult(t, templates, tc.nExpect)
			}
		})
	}
}
