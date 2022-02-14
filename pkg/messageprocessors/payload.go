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

package messageprocessors

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// PayloadEncoderDecoder provides an interface to encoding and decoding messages.
type PayloadEncoderDecoder interface {
	EncodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error
	DecodeUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error
	DecodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error
}

// CompilablePayloadEncoderDecoder extends PayloadEncoderDecoder with the ability
// to compile the parameters ahead of time.
type CompilablePayloadEncoderDecoder interface {
	PayloadEncoderDecoder

	CompileDownlinkEncoder(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error)
	CompileUplinkDecoder(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationUplink) error, error)
	CompileDownlinkDecoder(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error)
}

// PayloadProcessor provides an interface to processing payloads of multiple formats.
type PayloadProcessor interface {
	EncodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error
	DecodeUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error
	DecodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error
}

// MapPayloadProcessor implements PayloadProcessor using a mapping between ttnpb.PayloadFormatter and PayloadEncoderDecoder.
type MapPayloadProcessor map[ttnpb.PayloadFormatter]PayloadEncoderDecoder

var errFormatterNotConfigured = errors.DefineFailedPrecondition("formatter_not_configured", "formatter `{formatter}` is not configured")

// EncodeDownlink implements PayloadProcessor.
func (p MapPayloadProcessor) EncodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.EncodeDownlink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}

// DecodeUplink implements PayloadProcessor.
func (p MapPayloadProcessor) DecodeUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.DecodeUplink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}

// DecodeDownlink implements PayloadProcessor.
func (p MapPayloadProcessor) DecodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.DecodeDownlink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}

// GetPayloadEncoderDecoder returns the underlying PayloadEncoderDecoder for the provided format.
func (p MapPayloadProcessor) GetPayloadEncoderDecoder(ctx context.Context, formatter ttnpb.PayloadFormatter) (PayloadEncoderDecoder, error) {
	mp, ok := p[formatter]
	if !ok {
		return nil, errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	return mp, nil
}
