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

package gateways

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestTTIGPRO1(t *testing.T) {
	t.Run("Decode", func(t *testing.T) {
		for _, tc := range []struct {
			Name           string
			Data           []byte
			Expected       ttigpro1
			ErrorAssertion func(t *testing.T, err error) bool
		}{
			{
				Name: "CorrectQRCode",
				Data: []byte("https://ttig.pro/c/ec656efffe000128/abcdef123456"),
				Expected: ttigpro1{
					gatewayEUI: types.EUI64{0xec, 0x65, 0x6e, 0xff, 0xfe, 0x00, 0x01, 0x28},
					ownerToken: "abcdef123456",
				},
			},
			{
				Name: "InvalidURLPrefix",
				Data: []byte("https://example.com/c/ec656efffe000128/abcdef12"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/EUINotLowercase",
				Data: []byte("https://ttig.pro/c/EC656effFe000128/abcdef12"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/EUILength",
				Data: []byte("https://ttig.pro/c/ec656efffe00012/abcdef12"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/EUINotBase16",
				Data: []byte("https://ttig.pro/c/ec656efffe00012g/abcdef12"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/OwnerTokenLength",
				Data: []byte("https://ttig.pro/c/ec656efffe000128/abcdef123"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
			{
				Name: "Invalid/OwnerTokenNotBase62",
				Data: []byte("https://ttig.pro/c/ec656efffe000128/abcdef12!"),
				ErrorAssertion: func(t *testing.T, err error) bool {
					return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				var data ttigpro1
				err := data.UnmarshalText(tc.Data)
				if tc.ErrorAssertion != nil {
					a.So(tc.ErrorAssertion(t, err), should.BeTrue)
					return
				}
				if !a.So(err, should.BeNil) || !a.So(data, should.Resemble, tc.Expected) {
					t.FailNow()
				}
			})
		}
	})
}
