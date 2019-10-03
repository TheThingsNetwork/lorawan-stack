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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/formatters"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestProtobufUpstream(t *testing.T) {
	formatter := formatters.Protobuf

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
						SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
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
						SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
						PendingSession: false,
					},
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			actual, err := formatter.FromUp(tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			expected, err := tc.Message.Marshal()
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(actual, should.Resemble, expected)
		})
	}
}

func TestProtobufDownstream(t *testing.T) {
	formatter := formatters.Protobuf

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
				input := tc.Input
				if input == nil {
					var err error
					if input, err = tc.Items.Marshal(); !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}
				res, err := formatter.ToDownlinks(input)
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
				Request: &ttnpb.DownlinkQueueRequest{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: "foo-app",
						},
						DeviceID: "foo-device",
					},
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
				input := tc.Input
				if input == nil {
					var err error
					if input, err = tc.Request.Marshal(); !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}
				res, err := formatter.ToDownlinkQueueRequest(input)
				if tc.ErrorAssertion != nil && !tc.ErrorAssertion(t, err) || tc.ErrorAssertion == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, tc.Request)
			})
		}
	})
}
