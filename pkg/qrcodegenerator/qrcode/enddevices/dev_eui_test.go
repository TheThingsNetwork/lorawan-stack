// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package enddevices

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDevEUI(t *testing.T) {
	t.Run("Encode", func(t *testing.T) {
		for _, tc := range []struct {
			Name     string
			Device   *ttnpb.EndDevice
			Expected devEUIData
		}{
			{
				Name: "Simple",
				Device: &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						DevEui: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}.Bytes(),
					},
				},
				Expected: devEUIData{
					DevEUI: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)
				var res devEUIData
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
			Expected       devEUIData
			ErrorAssertion func(t *testing.T, err error) bool
		}{
			{
				Name: "Simple",
				Data: []byte("4242FFFFFFFFFFFF"),
				Expected: devEUIData{
					DevEUI: types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			},
			{
				Name: "Invalid",
				Data: []byte("424FFFFFFFFF"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				var data devEUIData
				err := data.UnmarshalText(tc.Data)
				if tc.ErrorAssertion != nil {
					a.So(tc.ErrorAssertion(t, err), should.BeTrue)
					return
				}
				if !a.So(err, should.BeNil) || !a.So(data, should.Resemble, tc.Expected) {
					t.FailNow()
				}

				text := test.Must(data.MarshalText())
				a.So(string(text), should.Equal, string(tc.Data))
			})
		}
	})
}
