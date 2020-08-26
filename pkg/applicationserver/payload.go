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

package applicationserver

import (
	"bytes"
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var errNoPayload = errors.Define("no_payload", "no payload")

func (as *ApplicationServer) encodeAndEncryptDownlink(ctx context.Context, dev *ttnpb.EndDevice, session *ttnpb.Session, downlink *ttnpb.ApplicationDownlink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	if session == nil || session.AppSKey == nil {
		return errNoAppSKey.New()
	}
	if downlink.FRMPayload == nil && downlink.DecodedPayload == nil {
		return errNoPayload.New()
	}
	if downlink.FRMPayload == nil && downlink.DecodedPayload != nil {
		var formatter ttnpb.PayloadFormatter
		var parameter string
		if dev.Formatters != nil {
			formatter, parameter = dev.Formatters.DownFormatter, dev.Formatters.DownFormatterParameter
		} else if defaultFormatters != nil {
			formatter, parameter = defaultFormatters.DownFormatter, defaultFormatters.DownFormatterParameter
		}
		if formatter != ttnpb.PayloadFormatter_FORMATTER_NONE {
			if err := as.formatters.EncodeDownlink(ctx, dev.EndDeviceIdentifiers, dev.VersionIDs, downlink, formatter, parameter); err != nil {
				events.Publish(evtEncodeFailDataDown.NewWithIdentifiersAndData(ctx, dev.EndDeviceIdentifiers, err))
				return err
			}
			if len(downlink.DecodedPayloadWarnings) > 0 {
				events.Publish(evtEncodeWarningDataDown.NewWithIdentifiersAndData(ctx, dev.EndDeviceIdentifiers, downlink))
			}
		}
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, *session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.EncryptDownlink(appSKey, session.DevAddr, downlink.FCnt, downlink.FRMPayload, false)
	if err != nil {
		return err
	}
	downlink.FRMPayload = frmPayload
	return nil
}

func (as *ApplicationServer) decryptAndDecodeUplink(ctx context.Context, dev *ttnpb.EndDevice, uplink *ttnpb.ApplicationUplink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	if err := as.decryptUplink(ctx, dev, uplink); err != nil {
		return err
	}
	return as.decodeUplink(ctx, dev, uplink, defaultFormatters)
}

func (as *ApplicationServer) decryptUplink(ctx context.Context, dev *ttnpb.EndDevice, uplink *ttnpb.ApplicationUplink) error {
	if dev.Session == nil || dev.Session.AppSKey == nil {
		return errNoAppSKey.New()
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, *dev.Session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.DecryptUplink(appSKey, dev.Session.DevAddr, uplink.FCnt, uplink.FRMPayload, false)
	if err != nil {
		return err
	}
	uplink.FRMPayload = frmPayload
	return nil
}

func (as *ApplicationServer) decodeUplink(ctx context.Context, dev *ttnpb.EndDevice, uplink *ttnpb.ApplicationUplink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	var formatter ttnpb.PayloadFormatter
	var parameter string
	if dev.Formatters != nil {
		formatter, parameter = dev.Formatters.UpFormatter, dev.Formatters.UpFormatterParameter
	} else if defaultFormatters != nil {
		formatter, parameter = defaultFormatters.UpFormatter, defaultFormatters.UpFormatterParameter
	}
	if formatter != ttnpb.PayloadFormatter_FORMATTER_NONE {
		if err := as.formatters.DecodeUplink(ctx, dev.EndDeviceIdentifiers, dev.VersionIDs, uplink, formatter, parameter); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to decode uplink")
			events.Publish(evtDecodeFailDataUp.NewWithIdentifiersAndData(ctx, dev.EndDeviceIdentifiers, err))
		} else if len(uplink.DecodedPayloadWarnings) > 0 {
			events.Publish(evtDecodeWarningDataUp.NewWithIdentifiersAndData(ctx, dev.EndDeviceIdentifiers, uplink))
		}
	}
	return nil
}

func (as *ApplicationServer) decryptAndDecodeDownlink(ctx context.Context, dev *ttnpb.EndDevice, downlink *ttnpb.ApplicationDownlink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	if err := as.decryptDownlink(ctx, dev, downlink); err != nil {
		return err
	}
	return as.decodeDownlink(ctx, dev, downlink, defaultFormatters)
}

func (as *ApplicationServer) decryptDownlink(ctx context.Context, dev *ttnpb.EndDevice, downlink *ttnpb.ApplicationDownlink) error {
	var session *ttnpb.Session
	switch {
	case dev.Session != nil && bytes.Equal(dev.Session.SessionKeyID, downlink.SessionKeyID):
		session = dev.Session
	case dev.PendingSession != nil && bytes.Equal(dev.PendingSession.SessionKeyID, downlink.SessionKeyID):
		session = dev.PendingSession
	default:
		return errUnknownSession.New()
	}
	if session.GetAppSKey() == nil {
		return errNoAppSKey.New()
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, *dev.Session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.DecryptDownlink(appSKey, dev.Session.DevAddr, downlink.FCnt, downlink.FRMPayload, false)
	if err != nil {
		return err
	}
	downlink.FRMPayload = frmPayload
	return nil
}

func (as *ApplicationServer) decodeDownlink(ctx context.Context, dev *ttnpb.EndDevice, downlink *ttnpb.ApplicationDownlink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	var formatter ttnpb.PayloadFormatter
	var parameter string
	if dev.Formatters != nil {
		formatter, parameter = dev.Formatters.DownFormatter, dev.Formatters.DownFormatterParameter
	} else if defaultFormatters != nil {
		formatter, parameter = defaultFormatters.DownFormatter, defaultFormatters.DownFormatterParameter
	}
	if formatter != ttnpb.PayloadFormatter_FORMATTER_NONE {
		if err := as.formatters.DecodeDownlink(ctx, dev.EndDeviceIdentifiers, dev.VersionIDs, downlink, formatter, parameter); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to decode downlink")
			events.Publish(evtDecodeFailDataDown.NewWithIdentifiersAndData(ctx, dev.EndDeviceIdentifiers, err))
		} else if len(downlink.DecodedPayloadWarnings) > 0 {
			events.Publish(evtDecodeWarningDataDown.NewWithIdentifiersAndData(ctx, dev.EndDeviceIdentifiers, downlink))
		}
	}
	return nil
}

type payloadFormatters map[ttnpb.PayloadFormatter]messageprocessors.PayloadEncodeDecoder

var errFormatterNotConfigured = errors.DefineFailedPrecondition("formatter_not_configured", "formatter `{formatter}` is not configured")

func (p payloadFormatters) EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.EncodeDownlink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}

func (p payloadFormatters) DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.DecodeUplink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}

func (p payloadFormatters) DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	mp, ok := p[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.DecodeDownlink(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}
