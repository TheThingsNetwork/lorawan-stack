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

package qrcode_test

import (
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/qrcode"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestParseEndDeviceAuthenticationCodes(t *testing.T) {
	for i, tc := range []struct {
		Data []byte
		ExpectedJoinEUI,
		ExpectedDevEUI types.EUI64
		ExpectedAuthenticationCode []byte
	}{
		{
			Data:                       []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
			ExpectedJoinEUI:            types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:             types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedAuthenticationCode: []byte{0x01, 0x02},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			data := test.Must(Parse(tc.Data)).(Data)
			intf, ok := data.(AuthenticatedEndDeviceIdentifiers)
			if !ok {
				t.Fatalf("Expected %T to implement AuthenticatedEndDeviceIdentifiers", data)
			}

			joinEUI, devEUI, authCode := intf.AuthenticatedEndDeviceIdentifiers()
			a.So(joinEUI, should.Resemble, tc.ExpectedJoinEUI)
			a.So(devEUI, should.Resemble, tc.ExpectedDevEUI)
			a.So(authCode, should.Resemble, tc.ExpectedAuthenticationCode)
		})
	}
}
