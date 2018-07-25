// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEncode(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	host := New()

	version := &ttnpb.EndDeviceVersion{
		BrandID:         "The Things Products",
		ModelID:         "The Things Uno",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0.0",
	}

	message := &ttnpb.DownlinkMessage{
		Payload: ttnpb.Message{
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					DecodedPayload: &types.Struct{
						Fields: map[string]*types.Value{
							"temperature": {
								Kind: &types.Value_NumberValue{
									NumberValue: -21.3,
								},
							},
						},
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
		output, err := host.Encode(ctx, message, version, script)
		a.So(err, should.BeNil)
		a.So(output.Payload.GetMACPayload().FRMPayload, should.Resemble, []byte{1, 2, 3})
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
		output, err := host.Encode(ctx, message, version, script)
		a.So(err, should.BeNil)
		a.So(output.Payload.GetMACPayload().FRMPayload, should.Resemble, []byte{247, 174})
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
		output, err := host.Encode(ctx, message, version, script)
		a.So(err, should.BeNil)
		a.So(output.Payload.GetMACPayload().FRMPayload, should.Resemble, []byte{247, 174})

		version.ModelID = "L-Tek FF1705"
		_, err = host.Encode(ctx, message, version, script)
		a.So(err, should.NotBeNil)
	}

	// Return out of range values.
	{
		script := `
		function Encoder(payload, f_port) {
			return [300, 0, 1]
		}
		`
		_, err := host.Encode(ctx, message, version, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidOutputRange)
	}

	// Return invalid type.
	{
		script := `
		function Encoder(payload, f_port) {
			return ['test']
		}
		`
		_, err := host.Encode(ctx, message, version, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidOutputType)
	}

	// Return nothing.
	{
		script := `
		function Encoder(payload, f_port) {
			return null
		}
		`
		_, err := host.Encode(ctx, message, version, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidOutputType)
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
		_, err := host.Encode(ctx, message, version, script)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidOutputType)
	}
}

func TestDecode(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	host := New()

	version := &ttnpb.EndDeviceVersion{
		BrandID:         "The Things Products",
		ModelID:         "The Things Uno",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0.0",
	}

	message := &ttnpb.UplinkMessage{
		Payload: ttnpb.Message{
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FRMPayload: []byte{247, 174},
				},
			},
		},
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
		output, err := host.Decode(ctx, message, version, script)
		a.So(err, should.BeNil)
		m, err := gogoproto.Map(output.Payload.GetMACPayload().DecodedPayload)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string]interface{}{
			"temperature": -21.3,
		})
	}

	// Parse and take brand and version into account.
	{
		script := `
		function Decoder(payload, f_port) {
			return {
				temperature: ((payload[0] & 0x80 ? 0xffff : 0x0000) << 16 | payload[0] << 8 | payload[1]) / 100,
				brand: env.brand,
				model: env.model,
			}
		}
		`
		output, err := host.Decode(ctx, message, version, script)
		a.So(err, should.BeNil)
		m, err := gogoproto.Map(output.Payload.GetMACPayload().DecodedPayload)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string]interface{}{
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
		_, err := host.Decode(ctx, message, version, script)
		a.So(err, should.NotBeNil)
	}

	// Catch error.
	{
		script := `
		function Decoder(payload, f_port) {
			throw Error('unknown error')
		}
		`
		_, err := host.Decode(ctx, message, version, script)
		a.So(err, should.NotBeNil)
	}
}
