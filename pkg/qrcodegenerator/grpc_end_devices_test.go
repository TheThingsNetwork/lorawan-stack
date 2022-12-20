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

package qrcodegenerator_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/enddevices"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestGenerateEndDeviceQRCode(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := componenttest.NewComponent(t, &component.Config{})
	testFormat := new(mockFormat)
	qrg, err := New(c, &Config{}, WithEndDeviceFormat("test", testFormat))
	test.Must(qrg, err)
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_QR_CODE_GENERATOR)

	client := ttnpb.NewEndDeviceQRCodeGeneratorClient(c.LoopbackConn())

	format, err := client.GetFormat(ctx, &ttnpb.GetQRCodeFormatRequest{
		FormatId: "test",
	}, c.WithClusterAuth())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(format.Name, should.Equal, "Test")

	formats, err := client.ListFormats(ctx, ttnpb.Empty, c.WithClusterAuth())
	a.So(err, should.BeNil)
	a.So(formats.Formats["test"], should.Resemble, &ttnpb.QRCodeFormat{
		Name:        "Test",
		Description: "Test",
		FieldMask:   ttnpb.FieldMask("ids"),
	})

	dev := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: "test",
			},
			DeviceId: "test",
			JoinEui:  types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}.Bytes(),
			DevEui:   types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
		},
	}

	res, err := client.Generate(ctx, &ttnpb.GenerateEndDeviceQRCodeRequest{
		FormatId:  "test",
		EndDevice: dev,
		Image: &ttnpb.GenerateEndDeviceQRCodeRequest_Image{
			ImageSize: 100,
		},
	}, c.WithClusterAuth())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(res.Text, should.Equal, "70B3D57ED0000000:0102030405060708")
	if !a.So(res.Image.GetEmbedded(), should.NotBeNil) {
		t.FailNow()
	}
	a.So(res.Image.Embedded.MimeType, should.Equal, "image/png")
	img, err := png.Decode(bytes.NewReader(res.Image.Embedded.Data))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(img.Bounds(), should.Resemble, image.Rectangle{Max: image.Point{100, 100}})
}

func TestGenerateEndDeviceQRCodeParsing(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := componenttest.NewComponent(t, &component.Config{})
	laTr005 := new(enddevices.LoRaAllianceTR005Format)
	qrg, err := New(c, &Config{}, WithEndDeviceFormat(laTr005.ID(), laTr005))
	test.Must(qrg, err)
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_QR_CODE_GENERATOR)

	genClient := ttnpb.NewEndDeviceQRCodeGeneratorClient(c.LoopbackConn())
	format, err := genClient.GetFormat(ctx, &ttnpb.GetQRCodeFormatRequest{
		FormatId: laTr005.ID(),
	}, c.WithClusterAuth())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(format.Name, should.Equal, "LoRa Alliance TR005")

	for _, tc := range []struct {
		Name      string
		FormatID  string
		GetQRData func() []byte
		Assertion func(*assertions.Assertion, *ttnpb.ParseEndDeviceQRCodeResponse, error) bool
	}{
		{
			Name: "EmptyData",
			GetQRData: func() []byte {
				return []byte{}
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseEndDeviceQRCodeResponse, err error) bool {
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
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseEndDeviceQRCodeResponse, err error) bool {
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
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseEndDeviceQRCodeResponse, err error) bool {
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
			Name:     "ValidBytes",
			FormatID: laTr005.ID(),
			GetQRData: func() []byte {
				return []byte(`LW:D0:1111111111111111:2222222222222222:AABB1122:O123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseEndDeviceQRCodeResponse, err error) bool {
				if !a.So(resp, should.NotBeNil) {
					return false
				}
				if !a.So(err, should.BeNil) {
					return false
				}
				a.So(resp.FormatId, should.Equal, laTr005.ID())
				endDeviceTemplate := resp.GetEndDeviceTemplate()
				a.So(endDeviceTemplate, should.NotBeNil)
				endDevice := endDeviceTemplate.EndDevice
				a.So(endDevice, should.NotBeNil)

				endDeviceIDs := endDevice.Ids.GetEntityIdentifiers().GetDeviceIds()
				a.So(endDeviceIDs, should.NotBeNil)
				a.So(endDeviceIDs.JoinEui, should.Resemble, types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes())
				a.So(endDeviceIDs.DevEui, should.Resemble, types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}.Bytes())

				a.So(endDevice.ClaimAuthenticationCode, should.NotBeNil)
				a.So(endDevice.ClaimAuthenticationCode.Value, should.Equal, "123456")

				tr005Ids := endDevice.GetLoraAllianceProfileIds()
				a.So(tr005Ids, should.NotBeNil)
				a.So(tr005Ids.VendorId, should.Equal, 0xAABB)
				a.So(tr005Ids.VendorProfileId, should.Equal, 0x1122)
				return true
			},
		},
		{
			Name:     "ValidBytesWithSerialNumber",
			FormatID: laTr005.ID(),
			GetQRData: func() []byte {
				return []byte(`LW:D0:1111111111111111:2222222222222222:ABCDABCD:O123456:S12345678`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseEndDeviceQRCodeResponse, err error) bool {
				if !a.So(resp, should.NotBeNil) {
					return false
				}
				if !a.So(err, should.BeNil) {
					return false
				}
				a.So(resp.FormatId, should.Equal, laTr005.ID())
				endDeviceTemplate := resp.GetEndDeviceTemplate()
				a.So(endDeviceTemplate, should.NotBeNil)
				endDevice := endDeviceTemplate.GetEndDevice()
				a.So(endDevice, should.NotBeNil)
				a.So(endDevice.SerialNumber, should.Equal, "12345678")

				endDeviceIDs := endDevice.GetIds()
				a.So(endDeviceIDs, should.NotBeNil)
				a.So(endDeviceIDs.JoinEui, should.Resemble, types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes())
				a.So(endDeviceIDs.DevEui, should.Resemble, types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}.Bytes())

				a.So(endDevice.ClaimAuthenticationCode, should.NotBeNil)
				a.So(endDevice.ClaimAuthenticationCode.Value, should.Equal, "123456")

				tr005Ids := endDevice.GetLoraAllianceProfileIds()
				a.So(tr005Ids, should.NotBeNil)
				a.So(tr005Ids.VendorId, should.Equal, 0xABCD)
				a.So(tr005Ids.VendorProfileId, should.Equal, 0xABCD)
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
							DevEui:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}.Bytes(),
							JoinEui: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
						},
						ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
							Value: "123456",
						},
					},
				}, c.WithClusterAuth())
				if !a.So(err, should.BeNil) {
					panic("could not generate QR Code")
				}
				return []byte(resp.Text)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseEndDeviceQRCodeResponse, err error) bool {
				if !a.So(resp, should.NotBeNil) {
					return false
				}
				if !a.So(err, should.BeNil) {
					return false
				}
				a.So(resp.FormatId, should.Equal, laTr005.ID())
				endDeviceTemplate := resp.GetEndDeviceTemplate()
				a.So(endDeviceTemplate, should.NotBeNil)
				endDevice := endDeviceTemplate.GetEndDevice()
				a.So(endDevice, should.NotBeNil)

				endDeviceIDs := endDevice.GetIds()
				a.So(endDeviceIDs, should.NotBeNil)
				a.So(endDeviceIDs.JoinEui, should.Resemble, types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes())
				a.So(endDeviceIDs.DevEui, should.Resemble, types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}.Bytes())

				a.So(endDevice.ClaimAuthenticationCode, should.NotBeNil)
				a.So(endDevice.ClaimAuthenticationCode.Value, should.Equal, "123456")
				return true
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := genClient.Parse(ctx, &ttnpb.ParseEndDeviceQRCodeRequest{
				FormatId: tc.FormatID,
				QrCode:   tc.GetQRData(),
			}, c.WithClusterAuth())
			if !a.So(tc.Assertion(a, resp, err), should.BeTrue) {
				t.FailNow()
			}
		})
	}
}
