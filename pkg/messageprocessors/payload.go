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

// PayloadEncodeDecoder provides an interface to encoding and decoding messages.
type PayloadEncodeDecoder interface {
	EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error
	DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error
	DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error
}

// PayloadProcessor provides an interface to processing payloads of multiple formats.
type PayloadProcessor interface {
	EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error
	DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error
	DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error
}

// MapPayloadProcessor implements PayloadProcessor using a mapping between ttnpb.PayloadFormatter and PayloadEncodeDecoder.
type MapPayloadProcessor map[ttnpb.PayloadFormatter]PayloadEncodeDecoder

var errFormatterNotConfigured = errors.DefineFailedPrecondition("formatter_not_configured", "formatter `{formatter}` is not configured")

// EncodeDownlink implements PayloadProcessor.
func (p MapPayloadProcessor) EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
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
func (p MapPayloadProcessor) DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
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
func (p MapPayloadProcessor) DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.DecodeDownlink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}
