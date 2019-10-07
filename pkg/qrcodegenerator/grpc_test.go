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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/qrcode"
	. "go.thethings.network/lorawan-stack/pkg/qrcodegenerator"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestGenerateEndDeviceQRCode(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	qrcode.RegisterEndDeviceFormat("test", new(mockFormat))

	c := componenttest.NewComponent(t, &component.Config{})
	test.Must(New(c, &Config{}))
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_QR_CODE_GENERATOR)

	client := ttnpb.NewEndDeviceQRCodeGeneratorClient(c.LoopbackConn())

	format, err := client.GetFormat(ctx, &ttnpb.GetQRCodeFormatRequest{
		FormatID: "test",
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(format.Name, should.Equal, "Test")

	formats, err := client.ListFormats(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	a.So(formats.Formats["test"], should.Resemble, &ttnpb.QRCodeFormat{
		Name:        "Test",
		Description: "Test",
		FieldMask: pbtypes.FieldMask{
			Paths: []string{"ids"},
		},
	})

	dev := ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test",
			},
			DeviceID: "test",
			JoinEUI:  eui64Ptr(types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}),
			DevEUI:   eui64Ptr(types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}),
		},
	}

	res, err := client.Generate(ctx, &ttnpb.GenerateEndDeviceQRCodeRequest{
		FormatID:  "test",
		EndDevice: dev,
		Image: &ttnpb.GenerateEndDeviceQRCodeRequest_Image{
			ImageSize: 100,
		},
	})
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
