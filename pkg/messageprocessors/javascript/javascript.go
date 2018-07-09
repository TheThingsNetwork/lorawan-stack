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

// Package javascript contains the Javascript payload formatter message processors.
package javascript

import (
	"context"
	"fmt"
	"reflect"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/pkg/scripting"
	js "go.thethings.network/lorawan-stack/pkg/scripting/javascript"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type host struct {
	engine scripting.Engine
}

// New creates and returns a new Javascript payload encoder and decoder.
func New() messageprocessors.PayloadEncodeDecoder {
	return &host{
		engine: js.New(scripting.DefaultOptions),
	}
}

func (h *host) createEnvironment(model *ttnpb.EndDeviceVersion) map[string]interface{} {
	env := make(map[string]interface{})
	env["brand"] = model.BrandID
	env["model"] = model.ModelID
	env["hardware_version"] = model.HardwareVersion
	env["firmware_version"] = model.FirmwareVersion
	return env
}

var (
	errInvalidInput       = errors.DefineInvalidArgument("input", "invalid input")
	errInvalidOutput      = errors.Define("output", "invalid output")
	errInvalidOutputType  = errors.Define("output_type", "invalid output of type `{type}`")
	errInvalidOutputRange = errors.Define("output_range", "output value `{value}` does not fall between `{low}` and `{high}`")
	errMissingPayload     = errors.DefineInvalidArgument("missing_payload", "missing message payload")
)

// Encode encodes the message's MAC payload DecodedPayload to FRMPayload using script.
func (h *host) Encode(ctx context.Context, msg *ttnpb.DownlinkMessage, model *ttnpb.EndDeviceVersion, script string) (*ttnpb.DownlinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, errMissingPayload
	}

	decoded := payload.DecodedPayload
	if decoded == nil {
		return msg, nil
	}

	m, err := gogoproto.Map(decoded)
	if err != nil {
		return nil, errInvalidInput.WithCause(err)
	}

	env := h.createEnvironment(model)
	env["application_id"] = msg.ApplicationID
	env["device_id"] = msg.DeviceID
	env["dev_eui"] = msg.DevEUI
	env["join_eui"] = msg.JoinEUI
	env["payload"] = m
	env["f_port"] = payload.FPort

	script = fmt.Sprintf(`
		%s
		Encoder(env.payload, env.f_port)
	`, script)

	value, err := h.engine.Run(ctx, script, env)
	if err != nil {
		return nil, err
	}

	if value == nil || reflect.TypeOf(value).Kind() != reflect.Slice {
		return nil, errInvalidOutputType
	}

	slice := reflect.ValueOf(value)
	l := slice.Len()
	payload.FRMPayload = make([]byte, l)
	for i := 0; i < l; i++ {
		val := slice.Index(i).Interface()
		var b int64
		switch i := val.(type) {
		case int:
			b = int64(i)
		case int8:
			b = int64(i)
		case int16:
			b = int64(i)
		case int32:
			b = int64(i)
		case int64:
			b = i
		case uint8:
			b = int64(i)
		case uint16:
			b = int64(i)
		case uint32:
			b = int64(i)
		case uint64:
			b = int64(i)
		default:
			return nil, errInvalidOutputType.WithAttributes("type", fmt.Sprintf("%T", i))
		}
		if b < 0x00 || b > 0xFF {
			return nil, errInvalidOutputRange.WithAttributes(
				"value", b,
				"low", 0x00,
				"high", 0xFF,
			)
		}
		payload.FRMPayload[i] = byte(b)
	}

	return msg, nil
}

// Decode decodes the message's MAC payload FRMPayload to DecodedPayload using script.
func (h *host) Decode(ctx context.Context, msg *ttnpb.UplinkMessage, model *ttnpb.EndDeviceVersion, script string) (*ttnpb.UplinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, errMissingPayload
	}

	env := h.createEnvironment(model)
	env["application_id"] = msg.ApplicationID
	env["device_id"] = msg.DeviceID
	env["dev_eui"] = msg.DevEUI
	env["join_eui"] = msg.JoinEUI
	env["payload"] = payload.FRMPayload
	env["f_port"] = payload.FPort

	script = fmt.Sprintf(`
		%s
		Decoder(env.payload, env.f_port)
	`, script)

	value, err := h.engine.Run(ctx, script, env)
	if err != nil {
		return nil, err
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, errInvalidOutput
	}

	s, err := gogoproto.Struct(m)
	if err != nil {
		return nil, errInvalidOutput.WithCause(err)
	}

	payload.DecodedPayload = s
	return msg, nil
}
