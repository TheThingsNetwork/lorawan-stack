// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"runtime/trace"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/scripting"
	js "go.thethings.network/lorawan-stack/v3/pkg/scripting/javascript"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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

type encodeDownlinkInput struct {
	Data  map[string]interface{} `json:"data"`
	FPort *uint8                 `json:"fPort"`
}

type encodeDownlinkOutput struct {
	Bytes    []uint8  `json:"bytes"`
	FPort    *uint8   `json:"fPort"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

var (
	errInput        = errors.DefineInvalidArgument("input", "invalid input")
	errOutput       = errors.Define("output", "invalid output")
	errOutputErrors = errors.DefineAborted("output_errors", "{errors}")
)

// EncodeDownlink encodes the message's DecodedPayload to FRMPayload using the given script.
func (h *host) EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, script string) error {
	defer trace.StartRegion(ctx, "encode downlink message").End()

	decoded := msg.DecodedPayload
	if decoded == nil {
		return nil
	}
	data, err := gogoproto.Map(decoded)
	if err != nil {
		return errInput.WithCause(err)
	}
	fPort := uint8(msg.FPort)
	input := encodeDownlinkInput{
		Data:  data,
		FPort: &fPort,
	}

	// Fallback to legacy Encoder() function for backwards compatibility with The Things Network Stack V2 payload functions.
	script = fmt.Sprintf(`
		%s

		function main(input) {
			if (typeof encodeDownlink === 'function') {
				return encodeDownlink(input);
			}
			return {
				bytes: Encoder(input.data, input.fPort),
				fPort: input.fPort
			}
		}
	`, script)
	valueAs, err := h.engine.Run(ctx, script, "main", input)
	if err != nil {
		return err
	}

	var output encodeDownlinkOutput
	err = valueAs(&output)
	if err != nil {
		return errOutput.WithCause(err)
	}
	if len(output.Errors) > 0 {
		return errOutputErrors.WithAttributes("errors", strings.Join(output.Errors, ", "))
	}

	msg.FRMPayload = output.Bytes
	msg.DecodedPayloadWarnings = output.Warnings
	if output.FPort != nil {
		fPort := *output.FPort
		msg.FPort = uint32(fPort)
	} else if msg.FPort == 0 {
		msg.FPort = 1
	}
	return nil
}

type decodeUplinkInput struct {
	Bytes []uint8 `json:"bytes"`
	FPort uint8   `json:"fPort"`
}

type decodeUplinkOutput struct {
	Data     map[string]interface{} `json:"data"`
	Warnings []string               `json:"warnings"`
	Errors   []string               `json:"errors"`
}

// DecodeUplink decodes the message's FRMPayload to DecodedPayload using the given script.
func (h *host) DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, script string) error {
	defer trace.StartRegion(ctx, "decode uplink message").End()

	input := decodeUplinkInput{
		Bytes: msg.FRMPayload,
		FPort: uint8(msg.FPort),
	}

	// Fallback to legacy Decoder() function for backwards compatibility with The Things Network Stack V2 payload functions.
	script = fmt.Sprintf(`
		%s

		function main(input) {
			if (typeof decodeUplink === 'function') {
				return decodeUplink(input);
			}
			return {
				data: Decoder(input.bytes, input.fPort)
			}
		}
	`, script)
	valueAs, err := h.engine.Run(ctx, script, "main", input)
	if err != nil {
		return err
	}

	var output decodeUplinkOutput
	err = valueAs(&output)
	if err != nil {
		return errOutput.WithCause(err)
	}
	if len(output.Errors) > 0 {
		return errOutputErrors.WithAttributes("errors", strings.Join(output.Errors, ", "))
	}

	s, err := gogoproto.Struct(output.Data)
	if err != nil {
		return errOutput.WithCause(err)
	}
	msg.DecodedPayload = s
	msg.DecodedPayloadWarnings = output.Warnings
	return nil
}

type decodeDownlinkInput struct {
	Bytes []uint8 `json:"bytes"`
	FPort uint8   `json:"fPort"`
}

type decodeDownlinkOutput struct {
	Data     map[string]interface{} `json:"data"`
	Warnings []string               `json:"warnings"`
	Errors   []string               `json:"errors"`
}

// DecodeUplink decodes the message's FRMPayload to DecodedPayload using the given script.
func (h *host) DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, script string) error {
	defer trace.StartRegion(ctx, "decode downlink message").End()

	input := decodeDownlinkInput{
		Bytes: msg.FRMPayload,
		FPort: uint8(msg.FPort),
	}

	script = fmt.Sprintf(`
		%s

		function main(input) {
			return decodeDownlink(input);
		}
	`, script)
	valueAs, err := h.engine.Run(ctx, script, "main", input)
	if err != nil {
		return err
	}

	var output decodeDownlinkOutput
	err = valueAs(&output)
	if err != nil {
		return errOutput.WithCause(err)
	}
	if len(output.Errors) > 0 {
		return errOutputErrors.WithAttributes("errors", strings.Join(output.Errors, ", "))
	}

	s, err := gogoproto.Struct(output.Data)
	if err != nil {
		return errOutput.WithCause(err)
	}
	msg.DecodedPayload = s
	msg.DecodedPayloadWarnings = output.Warnings
	return nil
}
