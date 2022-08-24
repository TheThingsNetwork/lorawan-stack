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
	"reflect"
	"runtime/trace"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/normalizedpayload"
	"go.thethings.network/lorawan-stack/v3/pkg/scripting"
	js "go.thethings.network/lorawan-stack/v3/pkg/scripting/javascript"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type host struct {
	engine scripting.AheadOfTimeEngine
}

// New creates and returns a new Javascript payload encoder and decoder.
func New() messageprocessors.CompilablePayloadEncoderDecoder {
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

func wrapDownlinkEncoderScript(script string) string {
	// Fallback to Encoder() for backwards compatibility with The Things Network Stack V2 payload functions.
	return fmt.Sprintf(`
		%s

		function main(input) {
			const { data, fPort } = input;
			if (typeof encodeDownlink === 'function') {
				return encodeDownlink({ data, fPort });
			}
			return {
				bytes: Encoder(data, fPort),
				fPort: input.fPort
			}
		}
	`, script)
}

// CompileDownlinkEncoder generates a downlink encoder from the provided script.
func (h *host) CompileDownlinkEncoder(
	ctx context.Context, script string,
) (
	func(
		context.Context,
		*ttnpb.EndDeviceIdentifiers,
		*ttnpb.EndDeviceVersionIdentifiers,
		*ttnpb.ApplicationDownlink,
	) error,
	error,
) {
	defer trace.StartRegion(ctx, "compile downlink encoder").End()

	run, err := h.engine.Compile(ctx, wrapDownlinkEncoderScript(script))
	if err != nil {
		return nil, err
	}

	return func(
		ctx context.Context,
		ids *ttnpb.EndDeviceIdentifiers,
		version *ttnpb.EndDeviceVersionIdentifiers,
		msg *ttnpb.ApplicationDownlink,
	) error {
		return h.encodeDownlink(ctx, msg, run)
	}, nil
}

// EncodeDownlink encodes the message's DecodedPayload to FRMPayload using the given script.
func (h *host) EncodeDownlink(
	ctx context.Context,
	_ *ttnpb.EndDeviceIdentifiers,
	_ *ttnpb.EndDeviceVersionIdentifiers,
	msg *ttnpb.ApplicationDownlink,
	script string,
) error {
	run := func(ctx context.Context, fn string, params ...interface{}) (func(interface{}) error, error) {
		return h.engine.Run(ctx, wrapDownlinkEncoderScript(script), fn, params...)
	}
	return h.encodeDownlink(ctx, msg, run)
}

func (*host) encodeDownlink(
	ctx context.Context,
	msg *ttnpb.ApplicationDownlink,
	run func(context.Context, string, ...interface{}) (func(interface{}) error, error),
) error {
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

	valueAs, err := run(ctx, "main", input)
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

	msg.FrmPayload = output.Bytes
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

type normalizeUplinkOutput struct {
	Data     interface{} `json:"data"`
	Warnings []string    `json:"warnings"`
	Errors   []string    `json:"errors"`
}

type uplinkDecoderOutput struct {
	Decoded    decodeUplinkOutput     `json:"decoded"`
	Normalized *normalizeUplinkOutput `json:"normalized"`
}

func wrapUplinkDecoderScript(script string) string {
	// This wrapper executes decodeUplink() if it is defined. Then, it executes normalizeUplink() if it is defined too,
	// and if the output of decodeUplink() didn't return errors.
	// Fallback to Decoder() for backwards compatibility with The Things Network Stack V2 payload functions.
	return fmt.Sprintf(`
		%s

		function main(input) {
			const bytes = input.bytes.slice();
			const { fPort } = input;
			if (typeof decodeUplink === 'function') {
				const decoded = decodeUplink({ bytes, fPort });
				let normalized;
				const { data, errors } = decoded;
				if ((!errors || !errors.length) && data && typeof normalizeUplink === 'function') {
					normalized = normalizeUplink({ data });
				}
				return { decoded, normalized };
			}
			return {
				decoded: {
					data: Decoder(bytes, fPort)
				}
			}
		}
	`, script)
}

// CompileUplinkDecoder generates an uplink decoder from the provided script.
func (h *host) CompileUplinkDecoder(
	ctx context.Context, script string,
) (
	func(
		context.Context,
		*ttnpb.EndDeviceIdentifiers,
		*ttnpb.EndDeviceVersionIdentifiers,
		*ttnpb.ApplicationUplink,
	) error,
	error,
) {
	defer trace.StartRegion(ctx, "compile uplink decoder").End()

	run, err := h.engine.Compile(ctx, wrapUplinkDecoderScript(script))
	if err != nil {
		return nil, err
	}

	return func(
		ctx context.Context,
		ids *ttnpb.EndDeviceIdentifiers,
		version *ttnpb.EndDeviceVersionIdentifiers,
		msg *ttnpb.ApplicationUplink,
	) error {
		return h.decodeUplink(ctx, msg, run)
	}, nil
}

// DecodeUplink decodes the message's FRMPayload to DecodedPayload using the given script.
func (h *host) DecodeUplink(
	ctx context.Context,
	_ *ttnpb.EndDeviceIdentifiers,
	_ *ttnpb.EndDeviceVersionIdentifiers,
	msg *ttnpb.ApplicationUplink,
	script string,
) error {
	run := func(ctx context.Context, fn string, params ...interface{}) (func(interface{}) error, error) {
		return h.engine.Run(ctx, wrapUplinkDecoderScript(script), fn, params...)
	}
	return h.decodeUplink(ctx, msg, run)
}

func (*host) decodeUplink(
	ctx context.Context,
	msg *ttnpb.ApplicationUplink,
	run func(context.Context, string, ...interface{}) (func(interface{}) error, error),
) error {
	defer trace.StartRegion(ctx, "decode uplink message").End()

	input := decodeUplinkInput{
		Bytes: msg.FrmPayload,
		FPort: uint8(msg.FPort),
	}

	valueAs, err := run(ctx, "main", input)
	if err != nil {
		return err
	}

	var output uplinkDecoderOutput
	err = valueAs(&output)
	if err != nil {
		return errOutput.WithCause(err)
	}

	if errs := output.Decoded.Errors; len(errs) > 0 {
		return errOutputErrors.WithAttributes("errors", strings.Join(errs, ", "))
	}
	decodedPayload, err := gogoproto.Struct(output.Decoded.Data)
	if err != nil {
		return errOutput.WithCause(err)
	}
	msg.DecodedPayload, msg.DecodedPayloadWarnings = decodedPayload, output.Decoded.Warnings

	if normalized := output.Normalized; normalized != nil {
		if errs := normalized.Errors; len(errs) > 0 {
			return errOutputErrors.WithAttributes("errors", strings.Join(errs, ", "))
		}
		if normalized.Data == nil {
			return errOutput.New()
		}
		// The returned data can be an array of measurements or a single measurement object.
		var measurements []map[string]interface{}
		if val := reflect.ValueOf(normalized.Data); val.Kind() == reflect.Slice {
			measurements = make([]map[string]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				measurement, ok := val.Index(i).Interface().(map[string]interface{})
				if !ok {
					return errOutput.New()
				}
				measurements[i] = measurement
			}
		} else {
			measurement, ok := normalized.Data.(map[string]interface{})
			if !ok {
				return errOutput.New()
			}
			measurements = []map[string]interface{}{measurement}
		}
		normalizedPayload := make([]*pbtypes.Struct, len(measurements))
		for i := range measurements {
			pb, err := gogoproto.Struct(measurements[i])
			if err != nil {
				return errOutput.WithCause(err)
			}
			normalizedPayload[i] = pb
		}
		// Validate the normalized payload.
		_, err := normalizedpayload.Parse(normalizedPayload)
		if err != nil {
			return errOutput.WithCause(err)
		}
		msg.NormalizedPayload, msg.NormalizedPayloadWarnings = normalizedPayload, normalized.Warnings
	} else {
		// If the normalizer is not set, the decoder may return already normalized payload.
		// This is a best effort attempt to parse the decoded payload as normalized payload.
		// If that does not return an error, the decoded payload is assumed to be normalized.
		normalizedPayload := []*pbtypes.Struct{
			decodedPayload,
		}
		_, err := normalizedpayload.Parse(normalizedPayload)
		if err == nil {
			msg.NormalizedPayload, msg.NormalizedPayloadWarnings = normalizedPayload, nil
		}
	}
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

func wrapDownlinkDecoderScript(script string) string {
	return fmt.Sprintf(`
		%s

		function main(input) {
			const bytes = input.bytes.slice();
			const { fPort } = input;
			return decodeDownlink({ bytes, fPort });
		}
	`, script)
}

// CompileDownlinkDecoder generates a downlink decoder from the provided script.
func (h *host) CompileDownlinkDecoder(
	ctx context.Context, script string,
) (
	func(
		context.Context,
		*ttnpb.EndDeviceIdentifiers,
		*ttnpb.EndDeviceVersionIdentifiers,
		*ttnpb.ApplicationDownlink,
	) error,
	error,
) {
	defer trace.StartRegion(ctx, "compile downlink decoder").End()

	run, err := h.engine.Compile(ctx, wrapDownlinkDecoderScript(script))
	if err != nil {
		return nil, err
	}

	return func(
		ctx context.Context,
		ids *ttnpb.EndDeviceIdentifiers,
		version *ttnpb.EndDeviceVersionIdentifiers,
		msg *ttnpb.ApplicationDownlink,
	) error {
		return h.decodeDownlink(ctx, msg, run)
	}, nil
}

// DecodeUplink decodes the message's FRMPayload to DecodedPayload using the given script.
func (h *host) DecodeDownlink(
	ctx context.Context,
	_ *ttnpb.EndDeviceIdentifiers,
	_ *ttnpb.EndDeviceVersionIdentifiers,
	msg *ttnpb.ApplicationDownlink,
	script string,
) error {
	run := func(ctx context.Context, fn string, params ...interface{}) (func(interface{}) error, error) {
		return h.engine.Run(ctx, wrapDownlinkDecoderScript(script), fn, params...)
	}
	return h.decodeDownlink(ctx, msg, run)
}

func (*host) decodeDownlink(
	ctx context.Context,
	msg *ttnpb.ApplicationDownlink,
	run func(context.Context, string, ...interface{}) (func(interface{}) error, error),
) error {
	defer trace.StartRegion(ctx, "decode downlink message").End()

	input := decodeDownlinkInput{
		Bytes: msg.FrmPayload,
		FPort: uint8(msg.FPort),
	}

	valueAs, err := run(ctx, "main", input)
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
