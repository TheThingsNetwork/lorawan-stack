// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package fmt_test

import (
	"strconv"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/fmt"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestJSONEncode(t *testing.T) {
	formatter := fmt.JSON

	for i, tc := range []struct {
		Message *ttnpb.ApplicationUp
		Result  string
	}{
		{
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "foo-app",
					},
					DeviceID: "foo-device",
				},
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{
						SessionKeyID: "test",
						FPort:        42,
						FCnt:         42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
						DecodedPayload: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"test_key": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 42,
									},
								},
							},
						},
					},
				},
			},
			Result: `{"end_device_ids":{"device_id":"foo-device","application_ids":{"application_id":"foo-app"}},"uplink_message":{"session_key_id":"test","f_port":42,"f_cnt":42,"frm_payload":"AQID","decoded_payload":{"test_key":42},"settings":{}}}`,
		},
		{
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "foo-app",
					},
					DeviceID: "foo-device",
				},
				Up: &ttnpb.ApplicationUp_JoinAccept{
					JoinAccept: &ttnpb.ApplicationJoinAccept{
						SessionKeyID:     "test",
						PendingSession:   false,
						SessionStartedAt: time.Date(2018, 11, 27, 15, 12, 0, 0, time.UTC),
					},
				},
			},
			Result: `{"end_device_ids":{"device_id":"foo-device","application_ids":{"application_id":"foo-app"}},"join_accept":{"session_key_id":"test","session_started_at":"2018-11-27T15:12:00Z"}}`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			buf, err := formatter.Encode(tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(string(buf), should.Equal, tc.Result)
		})
	}
}

func TestJSONDecode(t *testing.T) {
	formatter := fmt.JSON

	for i, tc := range []struct {
		Input []byte
		Items *ttnpb.ApplicationDownlinks
	}{
		{
			Input: []byte(`{"downlinks":[{"f_port":42,"frm_payload":"AQEB","confirmed":true},{"f_port":42,"frm_payload":"AgIC","confirmed":true}]}`),
			Items: &ttnpb.ApplicationDownlinks{
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x1, 0x1, 0x1},
						Confirmed:  true,
					},
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2},
						Confirmed:  true,
					},
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			res, err := formatter.Decode(tc.Input)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(res, should.Resemble, tc.Items)
		})
	}
}
