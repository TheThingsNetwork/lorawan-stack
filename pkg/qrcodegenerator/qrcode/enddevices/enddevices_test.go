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

package enddevices_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/enddevices"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestParseEndDeviceAuthenticationCodes(t *testing.T) {
	for i, tc := range []struct {
		FormatID string
		Data     []byte
		ExpectedJoinEUI,
		ExpectedDevEUI types.EUI64
		ExpectedAuthenticationCode string
	}{
		{
			FormatID:                   "tr005draft3",
			Data:                       []byte("URN:DEV:LW:42FFFFFFFFFFFFFF_4242FFFFFFFFFFFF_42FFFF42_V0102"),
			ExpectedJoinEUI:            types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:             types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedAuthenticationCode: "0102",
		},
		{
			FormatID:                   "tr005draft2",
			Data:                       []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
			ExpectedJoinEUI:            types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:             types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedAuthenticationCode: "0102",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			qrCode := New(context.Background())

			d, err := qrCode.Parse("", tc.Data)
			data := test.Must(d, err)

			edt := data.EndDeviceTemplate()
			a.So(edt, should.NotBeNil)
			a.So(data.FormatID(), should.Equal, tc.FormatID)
			endDevice := edt.GetEndDevice()

			a.So(endDevice, should.NotBeEmpty)
			ids := endDevice.GetIds()
			a.So(ids, should.NotBeEmpty)
			a.So(ids.JoinEui, should.Resemble, tc.ExpectedJoinEUI.Bytes())
			a.So(ids.DevEui, should.Resemble, tc.ExpectedDevEUI.Bytes())
			a.So(endDevice.ClaimAuthenticationCode, should.NotBeEmpty)
			a.So(endDevice.ClaimAuthenticationCode.Value, should.Resemble, tc.ExpectedAuthenticationCode)
		})
	}
}

type mock struct{}

func (mock) Validate() error                              { return nil }
func (*mock) Encode(*ttnpb.EndDevice) error               { return nil }
func (mock) MarshalText() ([]byte, error)                 { return nil, nil }
func (*mock) UnmarshalText([]byte) error                  { return nil }
func (*mock) EndDeviceTemplate() *ttnpb.EndDeviceTemplate { return nil }
func (*mock) FormatID() string                            { return "mock" }

type mockFormat struct{}

func (mockFormat) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:      "test",
		FieldMask: ttnpb.FieldMask("ids"),
	}
}

func (mockFormat) New() Data {
	return new(mock)
}

func (mockFormat) ID() string {
	return "mock"
}

func TestQRCodeFormats(t *testing.T) {
	a := assertions.New(t)
	qrCode := New(context.Background())

	a.So(qrCode.GetEndDeviceFormat("mock"), should.BeNil)

	mockFormat := new(mockFormat)
	qrCode.RegisterEndDeviceFormat(mockFormat.ID(), mockFormat)
	f := qrCode.GetEndDeviceFormat("mock")
	if !a.So(f, should.NotBeNil) {
		t.FailNow()
	}
	a.So(f.Format().Name, should.Equal, "test")

	fs := qrCode.GetEndDeviceFormats()
	a.So(fs["mock"], should.Equal, f)
}
