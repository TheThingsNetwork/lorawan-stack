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
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errNoPayload = errors.Define("no_payload", "no payload")

func (as *ApplicationServer) encodeAndEncrypt(ctx context.Context, dev *ttnpb.EndDevice, session *ttnpb.Session, downlink *ttnpb.ApplicationDownlink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	if session == nil || session.AppSKey == nil {
		return errNoAppSKey
	}
	if downlink.FRMPayload == nil && downlink.DecodedPayload == nil {
		return errNoPayload
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
			if err := as.formatter.Encode(ctx, dev.EndDeviceIdentifiers, dev.VersionIDs, downlink, formatter, parameter); err != nil {
				return err
			}
		}
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, *session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.EncryptDownlink(appSKey, session.DevAddr, downlink.FCnt, downlink.FRMPayload)
	if err != nil {
		return err
	}
	downlink.FRMPayload = frmPayload
	return nil
}

func (as *ApplicationServer) decryptAndDecode(ctx context.Context, dev *ttnpb.EndDevice, uplink *ttnpb.ApplicationUplink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	if dev.Session == nil || dev.Session.AppSKey == nil {
		return errNoAppSKey
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, *dev.Session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.DecryptUplink(appSKey, dev.Session.DevAddr, uplink.FCnt, uplink.FRMPayload)
	if err != nil {
		return err
	}
	uplink.FRMPayload = frmPayload
	var formatter ttnpb.PayloadFormatter
	var parameter string
	if dev.Formatters != nil {
		formatter, parameter = dev.Formatters.UpFormatter, dev.Formatters.UpFormatterParameter
	} else if defaultFormatters != nil {
		formatter, parameter = defaultFormatters.UpFormatter, defaultFormatters.UpFormatterParameter
	}
	if formatter != ttnpb.PayloadFormatter_FORMATTER_NONE {
		if err := as.formatter.Decode(ctx, dev.EndDeviceIdentifiers, dev.VersionIDs, uplink, formatter, parameter); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Payload decoding failed")
			events.Publish(evtDecodeFailDataUp(ctx, dev.EndDeviceIdentifiers, err))
		}
	}
	return nil
}

type payloadFormatter struct {
	repository     *devicerepository.Client
	upFormatters   map[ttnpb.PayloadFormatter]messageprocessors.PayloadDecoder
	downFormatters map[ttnpb.PayloadFormatter]messageprocessors.PayloadEncoder
}

var (
	errNoVersion          = errors.DefineFailedPrecondition("no_version", "no end device version")
	errVersionUnavailable = errors.DefineUnavailable("version_unavailable", "end device version is unavailable in the repository")
)

func (p payloadFormatter) getRepositoryFormatters(version *ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.MessagePayloadFormatters, error) {
	if version == nil || p.repository == nil {
		return nil, errNoVersion
	}
	versions, err := p.repository.DeviceVersions(version.BrandID, version.ModelID)
	if err != nil {
		return nil, errVersionUnavailable.WithCause(err)
	}
	for _, v := range versions {
		if v.FirmwareVersion == version.FirmwareVersion && v.HardwareVersion == version.HardwareVersion {
			return &v.DefaultFormatters, nil
		}
	}
	return nil, errVersionUnavailable
}

var errFormatterNotConfigured = errors.DefineFailedPrecondition("formatter_not_configured", "formatter `{formatter}` is not configured")

func (p payloadFormatter) Encode(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	if formatter == ttnpb.PayloadFormatter_FORMATTER_REPOSITORY {
		formatters, err := p.getRepositoryFormatters(version)
		if err != nil {
			return err
		}
		formatter, parameter = formatters.DownFormatter, formatters.DownFormatterParameter
	}
	mp, ok := p.downFormatters[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.Encode(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}

func (p payloadFormatter) Decode(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
	if formatter == ttnpb.PayloadFormatter_FORMATTER_REPOSITORY {
		formatters, err := p.getRepositoryFormatters(version)
		if err != nil {
			return err
		}
		formatter, parameter = formatters.UpFormatter, formatters.UpFormatterParameter
	}
	mp, ok := p.upFormatters[formatter]
	if !ok {
		return errFormatterNotConfigured.WithAttributes("formatter", formatter)
	}
	if err := mp.Decode(ctx, ids, version, msg, parameter); err != nil {
		return err
	}
	return nil
}
