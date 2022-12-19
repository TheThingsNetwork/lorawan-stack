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

package formatters_test

import (
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/formatters"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestJSONUpstream(t *testing.T) {
	formatter := formatters.JSON

	for i, tc := range []struct {
		Message *ttnpb.ApplicationUp
		Result  string
	}{
		{
			Message: &ttnpb.ApplicationUp{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "foo-app",
					},
					DeviceId: "foo-device",
				},
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{
						SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
						FPort:        42,
						FCnt:         42,
						FrmPayload:   []byte{0x1, 0x2, 0x3},
						DecodedPayload: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"test_key": {
									Kind: &structpb.Value_NumberValue{
										NumberValue: 42,
									},
								},
							},
						},
					},
				},
			},
			Result: `{"end_device_ids":{"device_id":"foo-device","application_ids":{"application_id":"foo-app"}},"uplink_message":{"session_key_id":"ESIzRA==","f_port":42,"f_cnt":42,"frm_payload":"AQID","decoded_payload":{"test_key":42}}}`,
		},
		{
			Message: &ttnpb.ApplicationUp{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "foo-app",
					},
					DeviceId: "foo-device",
				},
				Up: &ttnpb.ApplicationUp_JoinAccept{
					JoinAccept: &ttnpb.ApplicationJoinAccept{
						SessionKeyId:   []byte{0x11, 0x22, 0x33, 0x44},
						PendingSession: false,
					},
				},
			},
			Result: `{"end_device_ids":{"device_id":"foo-device","application_ids":{"application_id":"foo-app"}},"join_accept":{"session_key_id":"ESIzRA=="}}`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			buf, err := formatter.FromUp(tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(string(buf), should.Equal, tc.Result)
		})
	}
}

func TestJSONDownstream(t *testing.T) {
	formatter := formatters.JSON

	t.Run("Downlinks", func(t *testing.T) {
		for i, tc := range []struct {
			Input          []byte
			Items          *ttnpb.ApplicationDownlinks
			ErrorAssertion func(*testing.T, error) bool
		}{
			{
				Input: []byte(`garbage`),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(err, should.NotBeNil)
				},
			},
			{
				Input: []byte(`{"downlinks":[{"f_port":42,"frm_payload":"AQEB","confirmed":true},{"f_port":42,"frm_payload":"AgIC","confirmed":true}]}`),
				Items: &ttnpb.ApplicationDownlinks{
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FrmPayload: []byte{0x1, 0x1, 0x1},
							Confirmed:  true,
						},
						{
							FPort:      42,
							FrmPayload: []byte{0x2, 0x2, 0x2},
							Confirmed:  true,
						},
					},
				},
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)
				res, err := formatter.ToDownlinks(tc.Input)
				if tc.ErrorAssertion != nil && !tc.ErrorAssertion(t, err) || tc.ErrorAssertion == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, tc.Items)
			})
		}
	})

	t.Run("DownlinkQueueRequest", func(t *testing.T) {
		for i, tc := range []struct {
			Input          []byte
			Request        *ttnpb.DownlinkQueueRequest
			ErrorAssertion func(*testing.T, error) bool
		}{
			{
				Input: []byte(`garbage`),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(err, should.NotBeNil)
				},
			},
			{
				Input: []byte(`{"end_device_ids":{"application_ids":{"application_id":"foo-app"},"device_id":"foo-device"},"downlinks":[{"f_port":42,"frm_payload":"AQEB","confirmed":true},{"f_port":42,"frm_payload":"AgIC","confirmed":true}]}}`),
				Request: &ttnpb.DownlinkQueueRequest{
					EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: "foo-app",
						},
						DeviceId: "foo-device",
					},
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FrmPayload: []byte{0x1, 0x1, 0x1},
							Confirmed:  true,
						},
						{
							FPort:      42,
							FrmPayload: []byte{0x2, 0x2, 0x2},
							Confirmed:  true,
						},
					},
				},
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)
				res, err := formatter.ToDownlinkQueueRequest(tc.Input)
				if tc.ErrorAssertion != nil && !tc.ErrorAssertion(t, err) || tc.ErrorAssertion == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, tc.Request)
			})
		}
	})
}
