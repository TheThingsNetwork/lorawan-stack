// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package gateways_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestParseGatewaysAuthenticationCodes(t *testing.T) {
	for i, tc := range []struct {
		FormatID string
		Data     []byte
		ExpectedEUI,
		ExpectedOwnerToken string
	}{
		{
			FormatID:           "ttigpro1",
			Data:               []byte("https://ttig.pro/c/ec656efffe000128/abcdef123456"),
			ExpectedEUI:        "ec656efffe000128",
			ExpectedOwnerToken: "abcdef123456",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			qrCode := New(context.Background())

			d, err := qrCode.Parse(tc.FormatID, tc.Data)
			data := test.Must(d, err)

			a.So(data, should.NotBeNil)
		})
	}
}
