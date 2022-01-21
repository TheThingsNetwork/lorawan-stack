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

package qrcodegenerator_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/enddevice"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestQRCodeParser(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := componenttest.NewComponent(t, &component.Config{})
	qrg, err := New(c, &Config{})
	test.Must(qrg, err)
	laTr005 := new(enddevice.LoRaAllianceTR005Format)
	qrg.RegisterEndDeviceFormat(laTr005.ID(), laTr005)
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_QR_CODE_GENERATOR)

	genClient := ttnpb.NewEndDeviceQRCodeGeneratorClient(c.LoopbackConn())
	parserClient := ttnpb.NewQRCodeParserClient(c.LoopbackConn())
	format, err := genClient.GetFormat(ctx, &ttnpb.GetQRCodeFormatRequest{
		FormatId: laTr005.ID(),
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(format.Name, should.Equal, "LoRa Alliance TR005")

	for _, tc := range []struct {
		Name      string
		FormatID  string
		GetQRData func() []byte
		Assertion func(*assertions.Assertion, *ttnpb.ParseQRCodeResponse, error) bool
	}{
		{
			Name: "EmptyData",
			GetQRData: func() []byte {
				return []byte{}
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseQRCodeResponse, err error) bool {
				if !a.So(resp, should.BeNil) {
					return false
				}
				if !a.So(errors.IsInvalidArgument(err), should.BeTrue) {
					return false
				}
				return true
			},
		},
		{
			Name:     "UnknownFormat",
			FormatID: "unknown",
			GetQRData: func() []byte {
				return []byte(`LW:D0:1111111111111111:2222222222222222:00000000:O123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseQRCodeResponse, err error) bool {
				if !a.So(resp, should.BeNil) {
					return false
				}
				if !a.So(errors.IsInvalidArgument(err), should.BeTrue) {
					return false
				}
				return true
			},
		},
		{
			Name:     "InvalidValue",
			FormatID: "tr005",
			GetQRData: func() []byte {
				return []byte(`LW:D0:1111111111111111:222222222222222:00000000:O123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseQRCodeResponse, err error) bool {
				if !a.So(resp, should.BeNil) {
					return false
				}
				if !a.So(errors.IsInvalidArgument(err), should.BeTrue) {
					return false
				}
				return true
			},
		},
		{
			Name: "ValidBytes",
			GetQRData: func() []byte {
				return []byte(`LW:D0:1111111111111111:2222222222222222:00000001:O123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseQRCodeResponse, err error) bool {
				if !a.So(resp, should.NotBeNil) {
					return false
				}
				if !a.So(err, should.BeNil) {
					return false
				}
				a.So(resp.EntityOnboardingData.FormatId, should.Equal, laTr005.ID())
				a.So(*resp.EntityOnboardingData.GetEndDeviceOnboardingData().JoinEui, should.Equal, types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11})
				a.So(*resp.EntityOnboardingData.GetEndDeviceOnboardingData().DevEui, should.Equal, types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22})
				a.So(resp.EntityOnboardingData.GetEndDeviceOnboardingData().ModelId, should.Resemble, []byte{0x00, 0x01})
				return true
			},
		},
		{
			Name: "ValidGenerated",
			GetQRData: func() []byte {
				resp, err := genClient.Generate(ctx, &ttnpb.GenerateEndDeviceQRCodeRequest{
					FormatId: laTr005.ID(),
					EndDevice: &ttnpb.EndDevice{
						Ids: &ttnpb.EndDeviceIdentifiers{
							DeviceId: "test-dev",
							ApplicationIds: &ttnpb.ApplicationIdentifiers{
								ApplicationId: "test-app",
							},
							DevEui:  &types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
							JoinEui: &types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						},
						ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
							Value: "123456",
						},
					},
				})
				if !a.So(err, should.BeNil) {
					panic("could not generate QR Code")
				}
				return []byte(resp.Text)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseQRCodeResponse, err error) bool {
				if !a.So(resp, should.NotBeNil) {
					return false
				}
				if !a.So(err, should.BeNil) {
					return false
				}
				a.So(resp.EntityOnboardingData.FormatId, should.Equal, laTr005.ID())
				a.So(*resp.EntityOnboardingData.GetEndDeviceOnboardingData().JoinEui, should.Equal, types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11})
				a.So(*resp.EntityOnboardingData.GetEndDeviceOnboardingData().DevEui, should.Equal, types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22})
				return true
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := parserClient.Parse(ctx, &ttnpb.ParseQRCodeRequest{
				FormatId: tc.FormatID,
				QrCode:   tc.GetQRData(),
			})
			if !a.So(tc.Assertion(a, resp, err), should.BeTrue) {
				t.FailNow()
			}
		})
	}

}
