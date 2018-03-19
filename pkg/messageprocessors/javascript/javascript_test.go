// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package javascript_test

import (
	"context"
	"testing"

	errshould "github.com/TheThingsNetwork/ttn/pkg/errors/should"
	"github.com/TheThingsNetwork/ttn/pkg/gogoproto"
	"github.com/TheThingsNetwork/ttn/pkg/messageprocessors"
	"github.com/TheThingsNetwork/ttn/pkg/messageprocessors/javascript"
	"github.com/TheThingsNetwork/ttn/pkg/scripting"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestEncode(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	host := javascript.New()

	model := &ttnpb.EndDeviceVersion{
		EndDeviceModel: ttnpb.EndDeviceModel{
			BrandID: "The Things Products",
			ModelID: "The Things Uno",
		},
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
		output, err := host.Encode(ctx, message, model, script)
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
		output, err := host.Encode(ctx, message, model, script)
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
				throw Error('unknown model')
			}
		}
		`
		output, err := host.Encode(ctx, message, model, script)
		a.So(err, should.BeNil)
		a.So(output.Payload.GetMACPayload().FRMPayload, should.Resemble, []byte{247, 174})

		model.EndDeviceModel.ModelID = "L-Tek FF1705"
		_, err = host.Encode(ctx, message, model, script)
		a.So(err, errshould.Describe, scripting.ErrRuntime)
	}

	// Return out of range values.
	{
		script := `
		function Encoder(payload, f_port) {
			return [300, 0, 1]
		}
		`
		_, err := host.Encode(ctx, message, model, script)
		a.So(err, errshould.Describe, messageprocessors.ErrInvalidOutputRange)
	}

	// Return invalid type.
	{
		script := `
		function Encoder(payload, f_port) {
			return ['test']
		}
		`
		_, err := host.Encode(ctx, message, model, script)
		a.So(err, errshould.Describe, messageprocessors.ErrInvalidOutput)
	}

	// Return nothing.
	{
		script := `
		function Encoder(payload, f_port) {
			return null
		}
		`
		_, err := host.Encode(ctx, message, model, script)
		a.So(err, errshould.Describe, messageprocessors.ErrInvalidOutputType)
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
		_, err := host.Encode(ctx, message, model, script)
		a.So(err, errshould.Describe, messageprocessors.ErrInvalidOutputType)
	}
}

func TestDecode(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	host := javascript.New()

	model := &ttnpb.EndDeviceVersion{
		EndDeviceModel: ttnpb.EndDeviceModel{
			BrandID: "The Things Products",
			ModelID: "The Things Uno",
		},
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
		output, err := host.Decode(ctx, message, model, script)
		a.So(err, should.BeNil)
		m, err := gogoproto.Map(output.Payload.GetMACPayload().DecodedPayload)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string]interface{}{
			"temperature": -21.3,
		})
	}

	// Parse and take brand and model into account.
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
		output, err := host.Decode(ctx, message, model, script)
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
		_, err := host.Decode(ctx, message, model, script)
		a.So(err, should.NotBeNil)
	}

	// Catch error.
	{
		script := `
		function Decoder(payload, f_port) {
			throw Error('unknown error')
		}
		`
		_, err := host.Decode(ctx, message, model, script)
		a.So(err, should.NotBeNil)
	}
}
