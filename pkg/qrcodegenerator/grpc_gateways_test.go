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

package qrcodegenerator_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestGatewayQRCodeParsing(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := componenttest.NewComponent(t, &component.Config{})
	ttigpro1 := new(gateways.TTIGPRO1Format)
	qrg, err := New(c, &Config{}, WithGatewayFormat(ttigpro1.ID(), ttigpro1))
	test.Must(qrg, err)

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_QR_CODE_GENERATOR)

	client := ttnpb.NewGatewayQRCodeGeneratorClient(c.LoopbackConn())

	for _, tc := range []struct {
		Name      string
		FormatID  string
		GetQRData func() []byte
		Assertion func(*assertions.Assertion, *ttnpb.ParseGatewayQRCodeResponse, error) bool
	}{
		{
			Name: "EmptyData",
			GetQRData: func() []byte {
				return []byte{}
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseGatewayQRCodeResponse, err error) bool {
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
				return []byte(`https://ttig.pro/c/ec656efffe000128/abcdef123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseGatewayQRCodeResponse, err error) bool {
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
			Name:     "InvalidFormat",
			FormatID: "tr005",
			GetQRData: func() []byte {
				return []byte(`https://ttig.pro/c/ec656efffe000128/abcdef123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseGatewayQRCodeResponse, err error) bool {
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
			Name:     "ValidTTIGPRO1",
			FormatID: ttigpro1.ID(),
			GetQRData: func() []byte {
				return []byte(`https://ttig.pro/c/ec656efffe000128/abcdef123456`)
			},
			Assertion: func(a *assertions.Assertion, resp *ttnpb.ParseGatewayQRCodeResponse, err error) bool {
				if !a.So(resp, should.NotBeNil) {
					return false
				}
				if !a.So(err, should.BeNil) {
					return false
				}
				a.So(resp.FormatId, should.Equal, ttigpro1.ID())

				return true
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := client.Parse(ctx, &ttnpb.ParseGatewayQRCodeRequest{
				FormatId: tc.FormatID,
				QrCode:   tc.GetQRData(),
			}, c.WithClusterAuth())
			if !a.So(tc.Assertion(a, resp, err), should.BeTrue) {
				t.FailNow()
			}
		})
	}
}
