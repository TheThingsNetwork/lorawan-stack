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

func (h *host) createEnvironment(version *ttnpb.EndDeviceVersionIdentifiers) map[string]interface{} {
	env := make(map[string]interface{})
	env["brand"] = version.BrandID
	env["model"] = version.ModelID
	env["hardware_version"] = version.HardwareVersion
	env["firmware_version"] = version.FirmwareVersion
	return env
}

var (
	errInput       = errors.DefineInvalidArgument("input", "invalid input")
	errOutput      = errors.Define("output", "invalid output")
	errOutputType  = errors.Define("output_type", "invalid output of type `{type}`")
	errOutputRange = errors.Define("output_range", "output value `{value}` does not fall between `{low}` and `{high}`")
)

// Encode encodes the message's DecodedPayload to FRMPayload using the given script.
func (h *host) Encode(ctx context.Context, msg *ttnpb.ApplicationDownlink, version *ttnpb.EndDeviceVersionIdentifiers, script string) (*ttnpb.ApplicationDownlink, error) {
	decoded := msg.DecodedPayload
	if decoded == nil {
		return msg, nil
	}
	m, err := gogoproto.Map(decoded)
	if err != nil {
		return nil, errInput.WithCause(err)
	}
	env := h.createEnvironment(version)
	env["payload"] = m
	env["f_port"] = msg.FPort
	script = fmt.Sprintf(`
		%s
		Encoder(env.payload, env.f_port)
	`, script)
	value, err := h.engine.Run(ctx, script, env)
	if err != nil {
		return nil, err
	}
	if value == nil || reflect.TypeOf(value).Kind() != reflect.Slice {
		return nil, errOutputType
	}
	slice := reflect.ValueOf(value)
	l := slice.Len()
	msg.FRMPayload = make([]byte, l)
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
			return nil, errOutputType.WithAttributes("type", fmt.Sprintf("%T", i))
		}
		if b < 0x00 || b > 0xFF {
			return nil, errOutputRange.WithAttributes(
				"value", b,
				"low", 0x00,
				"high", 0xFF,
			)
		}
		msg.FRMPayload[i] = byte(b)
	}
	return msg, nil
}

// Decode decodes the message's FRMPayload to DecodedPayload using the given script.
func (h *host) Decode(ctx context.Context, msg *ttnpb.ApplicationUplink, version *ttnpb.EndDeviceVersionIdentifiers, script string) (*ttnpb.ApplicationUplink, error) {
	env := h.createEnvironment(version)
	env["payload"] = msg.FRMPayload
	env["f_port"] = msg.FPort
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
		return nil, errOutput
	}
	s, err := gogoproto.Struct(m)
	if err != nil {
		return nil, errOutput.WithCause(err)
	}
	msg.DecodedPayload = s
	return msg, nil
}
