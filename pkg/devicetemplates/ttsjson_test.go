// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/devicetemplates"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

const (
	otaaDevice = `{
		"ids": {
			"device_id": "otaa-device",
			"application_ids": {
				"application_id": "test-app"
			},
			"dev_eui": "0102030405060708",
			"join_eui": "0807060504030201"
		},
		"frequency_plan_id": "EU_863_870",
		"lorawan_version": "1.0.2",
		"lorawan_phy_version": "1.0.2-b",
		"root_keys": {
			"app_key": {
				"key": "01020304010203040102030401020304"
			}
		},
		"supports_join": true
	}`

	abpDevice = `{
		"ids": {
			"device_id": "abp-device",
			"application_ids": {
				"application_id": "test-app"
			},
			"dev_eui": "0102030405060708",
			"join_eui": "0807060504030201"
		},
		"frequency_plan_id": "US_902_928_FSB_2",
		"lorawan_version": "1.0.2",
		"lorawan_phy_version": "1.0.2-b",
		"mac_settings": {
			"rx1_delay": null
		},
		"supports_join": false,
		"session": {
			"dev_addr": "01010101"
		}
	}`

	abpDeviceWithoutSession = `{
		"ids": {
			"device_id": "abp-device-error",
			"application_ids": {
				"application_id": "test-app"
			},
			"dev_eui": "0102030405060708",
			"join_eui": "0807060504030201"
		},
		"frequency_plan_id": "US_902_928_FSB_2",
		"lorawan_version": "1.0.2",
		"lorawan_phy_version": "1.0.2-b",
		"mac_settings": {
			"rx1_delay": null
		},
		"supports_join": false
	}`

	otaaWithSession = `{
		"ids": {
			"device_id": "industrial-tracker",
			"application_ids": {
				"application_id": "ttn-tabs"
			},
			"dev_eui": "E8E1E100010146B1",
			"join_eui": "E8E1E1000101363E"
		},
		"name": "industrial-tracker",
		"lorawan_version": "1.0.2",
		"lorawan_phy_version": "1.0.2-b",
		"frequency_plan_id": "EU_863_870",
		"supports_join": true,
		"root_keys": {
			"app_key": {
				"key": "00112233445566778899AABBCCDDEEFF"
			}
		},
		"mac_settings": {
			"rx1_delay": 1
		},
		"session": {
			"dev_addr": "260125FD",
			"keys": {
				"app_s_key": {
					"key": "00112233445566778899AABBCCDDEEFF"
				},
				"f_nwk_s_int_key": {
					"key": "00112233445566778899AABBCCDDEEFF"
				}
			},
			"last_f_cnt_up": 0,
			"last_n_f_cnt_down": 0
		}
	}`

	devWithDevAddr = `{
		"ids": {
			"device_id": "otaa-device",
			"application_ids": {
				"application_id": "test-app"
			},
			"dev_eui": "0102030405060708",
			"join_eui": "0807060504030201"
		},
		"dev_addr": "01010101",
		"frequency_plan_id": "EU_863_870",
		"lorawan_version": "1.0.2",
		"lorawan_phy_version": "1.0.2-b",
		"root_keys": {
			"app_key": {
				"key": "01020304010203040102030401020304"
			}
		},
		"supports_join": true
	}`
)

func TestTTSJSONConverter(t *testing.T) {
	tts := GetConverter("the-things-stack")
	a := assertions.New(t)
	if !a.So(tts, should.NotBeNil) {
		t.FailNow()
	}

	for _, tc := range []struct {
		name              string
		body              []byte
		validateError     func(t *testing.T, err error)
		validateResult    func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int)
		nExpect           int
		expectedTemplates []*ttnpb.EndDeviceTemplate
	}{
		{
			name: "InvalidJSON",
			body: []byte("invalid json"),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			name: "OneDevice",
			body: []byte(otaaDevice),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			nExpect:        1,
			validateResult: validateTemplates,
		},
		{
			name: "OneABPOneOTAA",
			body: []byte(abpDevice + "\n\n" + otaaDevice),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: validateTemplates,
			nExpect:        2,
		},
		{
			name: "OneOKOneError",
			body: []byte(abpDevice + "\n\n" + "invalid json"),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			validateResult: validateTemplates,
			nExpect:        1,
		},
		{
			name: "OneWithSession",
			body: []byte(otaaWithSession),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: validateTemplates,
			nExpect:        1,
		},
		{
			name: "OneWithSessionOneWithout",
			body: []byte(otaaWithSession + "\n\n" + abpDeviceWithoutSession),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: validateTemplates,
			nExpect:        2,
		},
		{
			name: "RemovesDevAddrFromRoot",
			body: []byte(devWithDevAddr),
			validateError: func(t *testing.T, err error) {
				assertions.New(t).So(err, should.BeNil)
			},
			validateResult: func(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int) {
				a := assertions.New(t)
				if !a.So(len(templates), should.Equal, count) {
					t.FailNow()
				}
				tmpl := templates[0]
				if !a.So(tmpl, should.NotBeNil) {
					t.FailNow()
				}

				a.So(tmpl.EndDevice.Ids.DevAddr, should.BeNil)
				a.So(tmpl.FieldMask.GetPaths(), should.NotContain, "dev_addr")
			},
			nExpect: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := test.Context()

			templates := []*ttnpb.EndDeviceTemplate{}
			err := tts.Convert(ctx, bytes.NewReader(tc.body), func(tmpl *ttnpb.EndDeviceTemplate) error {
				templates = append(templates, tmpl)
				return nil
			})

			tc.validateError(t, err)
			if tc.validateResult != nil {
				tc.validateResult(t, templates, tc.nExpect)
			}
		})
	}
}
