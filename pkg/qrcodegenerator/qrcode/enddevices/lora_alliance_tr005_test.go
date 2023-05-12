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

package enddevices_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/enddevices"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestLoRaAllianceTR005(t *testing.T) {
	t.Run("Encode", func(t *testing.T) {
		for _, tc := range []struct {
			Name     string
			Device   *ttnpb.EndDevice
			Expected LoRaAllianceTR005
		}{
			{
				Name: "Simple",
				Device: &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						JoinEui: types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}.Bytes(),
						DevEui:  types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}.Bytes(),
					},
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "ABCD",
					},
				},
				Expected: LoRaAllianceTR005{
					JoinEUI:    types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00},
					DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					OwnerToken: "ABCD",
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)
				var res LoRaAllianceTR005
				err := res.Encode(tc.Device)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, tc.Expected)
			})
		}
	})

	t.Run("Decode", func(t *testing.T) {
		for _, tc := range []struct {
			Name           string
			Data           []byte
			CanonicalData  []byte
			Expected       LoRaAllianceTR005
			ErrorAssertion func(t *testing.T, err error) bool
		}{
			{
				Name: "Simple",
				Data: []byte("LW:D0:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42"),
				Expected: LoRaAllianceTR005{
					JoinEUI:  types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEUI:   types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					VendorID: [2]byte{0x42, 0xff},
					ModelID:  [2]byte{0xff, 0x42},
				},
			},
			{
				Name: "Extensions",
				Data: []byte("LW:D0:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:CCHECKSUM:O0102:SSERIAL:PPROPRIETARY"),
				Expected: LoRaAllianceTR005{
					JoinEUI:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					VendorID:     [2]byte{0x42, 0xff},
					ModelID:      [2]byte{0xff, 0x42},
					Checksum:     "CHECKSUM",
					OwnerToken:   "0102",
					SerialNumber: "SERIAL",
					Proprietary:  "PROPRIETARY",
				},
			},
			{
				Name:          "EmptyExtensions",
				Data:          []byte("LW:D0:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:O:S:P"),
				CanonicalData: []byte("LW:D0:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42"),
				Expected: LoRaAllianceTR005{
					JoinEUI:    types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEUI:     types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					VendorID:   [2]byte{0x42, 0xff},
					ModelID:    [2]byte{0xff, 0x42},
					OwnerToken: "",
				},
			},
			{
				Name: "Invalid/Type",
				Data: []byte{0x42, 0xff, 0x42, 0x42},
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/Parts",
				Data: []byte("LW:D0:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/EUI",
				Data: []byte("LW:D0:42FFFFFFFF:4242FFFFFFFFFFFF:42FFFF42"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/ProdID",
				Data: []byte("LW:D0:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42AABB"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				var data LoRaAllianceTR005
				err := data.UnmarshalText(tc.Data)
				if tc.ErrorAssertion != nil {
					a.So(tc.ErrorAssertion(t, err), should.BeTrue)
					return
				}
				if !a.So(err, should.BeNil) || !a.So(data, should.Resemble, tc.Expected) {
					t.FailNow()
				}

				canonical := tc.CanonicalData
				if canonical == nil {
					canonical = tc.Data
				}

				text := test.Must(data.MarshalText())
				a.So(string(text), should.Equal, string(canonical))
			})
		}
	})
}
