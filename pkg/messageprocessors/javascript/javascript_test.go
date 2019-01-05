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

package javascript

import (
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEncode(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	host := New()

	eui := types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-app",
		},
		DeviceID: "foo-device",
		DevEUI:   &eui,
	}
	version := &ttnpb.EndDeviceVersionIdentifiers{
		BrandID:         "The Things Products",
		ModelID:         "The Things Uno",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0.0",
	}

	message := &ttnpb.ApplicationDownlink{
		DecodedPayload: &pbtypes.Struct{
			Fields: map[string]*pbtypes.Value{
				"temperature": {
					Kind: &pbtypes.Value_NumberValue{
						NumberValue: -21.3,
					},
				},
			},
		},
	}

	// Return constant byte array.
	{
		script := `
		function Encoder(payload, f_port) {
			return [1, 2, 3]
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.BeNil)
		a.So(message.FRMPayload, should.Resemble, []byte{1, 2, 3})
	}

	// Encode temperature.
	{
		script := `
		function Encoder(payload, f_port) {
			var val = payload.temperature * 100
			return [
				(val >> 8) & 0xff,
				val & 0xff
			]
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.BeNil)
		a.So(message.FRMPayload, should.Resemble, []byte{247, 174})
	}

	// Encode temperature based on a specific model.
	{
		script := `
		function Encoder(payload, f_port) {
			switch (env.model) {
			case "The Things Uno":
				var val = payload.temperature * 100
				return [
					(val >> 8) & 0xff,
					val & 0xff
				]
			default:
				throw Error('unknown version')
			}
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.BeNil)
		a.So(message.FRMPayload, should.Resemble, []byte{247, 174})

		version.ModelID = "L-Tek FF1705"
		err = host.Encode(ctx, ids, version, message, script)
		a.So(err, should.NotBeNil)
	}

	// Return out of range values.
	{
		script := `
		function Encoder(payload, f_port) {
			return [300, 0, 1]
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errOutputRange)
	}

	// Return invalid type.
	{
		script := `
		function Encoder(payload, f_port) {
			return ['test']
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errOutputType)
	}

	// Return nothing.
	{
		script := `
		function Encoder(payload, f_port) {
			return null
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errOutputType)
	}

	// Return an object.
	{
		script := `
		function Encoder(payload, f_port) {
			return {
				value: 42
			}
		}
		`
		err := host.Encode(ctx, ids, version, message, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errOutputType)
	}
}

func TestDecode(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	host := New()

	eui := types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-app",
		},
		DeviceID: "foo-device",
		DevEUI:   &eui,
	}
	version := &ttnpb.EndDeviceVersionIdentifiers{
		BrandID:         "The Things Products",
		ModelID:         "The Things Uno",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0.0",
	}

	message := &ttnpb.ApplicationUplink{
		FRMPayload: []byte{247, 174},
	}

	// Return constant object.
	{
		script := `
		function Decoder(payload, f_port) {
			return {
				temperature: -21.3
			}
		}
		`
		err := host.Decode(ctx, ids, version, message, script)
		a.So(err, should.BeNil)
		m, err := gogoproto.Map(message.DecodedPayload)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string]interface{}{
			"temperature": -21.3,
		})
	}

	// Parse and take DevEUI, brand and version into account.
	{
		script := `
		function Decoder(payload, f_port) {
			return {
				temperature: ((payload[0] & 0x80 ? 0xffff : 0x0000) << 16 | payload[0] << 8 | payload[1]) / 100,
				dev_eui: env.dev_eui,
				brand: env.brand,
				model: env.model,
			}
		}
		`
		err := host.Decode(ctx, ids, version, message, script)
		a.So(err, should.BeNil)
		m, err := gogoproto.Map(message.DecodedPayload)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string]interface{}{
			"dev_eui":     "0102030405060708",
			"brand":       "The Things Products",
			"model":       "The Things Uno",
			"temperature": -21.3,
		})
	}

	// Return invalid type.
	{
		script := `
		function Decoder(payload, f_port) {
			return 42
		}
		`
		err := host.Decode(ctx, ids, version, message, script)
		a.So(err, should.NotBeNil)
	}

	// Catch error.
	{
		script := `
		function Decoder(payload, f_port) {
			throw Error('unknown error')
		}
		`
		err := host.Decode(ctx, ids, version, message, script)
		a.So(err, should.NotBeNil)
	}
}
