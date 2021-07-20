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

	apppayload "go.thethings.network/lorawan-application-payload"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var errNoPayload = errors.Define("no_payload", "no payload")

func (as *ApplicationServer) encodeAndEncryptDownlinks(ctx context.Context, dev *ttnpb.EndDevice, link *ttnpb.ApplicationLink, items []*ttnpb.ApplicationDownlink, sessions []*ttnpb.Session) ([]*ttnpb.ApplicationDownlink, error) {
	var encryptedItems []*ttnpb.ApplicationDownlink
	for _, session := range sessions {
		skipPayloadCrypto := as.skipPayloadCrypto(ctx, link, dev, session)
		for _, item := range items {
			fCnt := session.LastAFCntDown + 1
			sessionKeyID := session.SessionKeyID
			if skipPayloadCrypto {
				fCnt = item.FCnt

				if len(item.SessionKeyID) > 0 {
					sessionKeyID = item.SessionKeyID
				}
			}
			encryptedItem := &ttnpb.ApplicationDownlink{
				SessionKeyID:   sessionKeyID,
				FPort:          item.FPort,
				FCnt:           fCnt,
				FRMPayload:     item.FRMPayload,
				DecodedPayload: item.DecodedPayload,
				Confirmed:      item.Confirmed,
				ClassBC:        item.ClassBC,
				Priority:       item.Priority,
				CorrelationIDs: item.CorrelationIDs,
			}
			if !skipPayloadCrypto {
				if err := as.encodeAndEncryptDownlink(ctx, dev, session, encryptedItem, link.DefaultFormatters); err != nil {
					log.FromContext(ctx).WithError(err).Warn("Encoding and encryption of downlink message failed; drop item")
					return nil, err
				}
			}
			encryptedItem.DecodedPayload = nil
			session.LastAFCntDown = encryptedItem.FCnt
			encryptedItems = append(encryptedItems, encryptedItem)
		}
	}
	return encryptedItems, nil
}

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
				events.Publish(evtEncodeFailDataDown.NewWithIdentifiersAndData(ctx, &dev.EndDeviceIdentifiers, err))
				return err
			}
			if len(downlink.DecodedPayloadWarnings) > 0 {
				events.Publish(evtEncodeWarningDataDown.NewWithIdentifiersAndData(ctx, &dev.EndDeviceIdentifiers, downlink))
			}
		}
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, session.AppSKey, as.KeyVault)
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
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.AppSKey, as.KeyVault)
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
			events.Publish(evtDecodeFailDataUp.NewWithIdentifiersAndData(ctx, &dev.EndDeviceIdentifiers, err))
		} else if len(uplink.DecodedPayloadWarnings) > 0 {
			events.Publish(evtDecodeWarningDataUp.NewWithIdentifiersAndData(ctx, &dev.EndDeviceIdentifiers, uplink))
		}
	}
	return nil
}

func (as *ApplicationServer) decryptAndDecodeDownlink(ctx context.Context, dev *ttnpb.EndDevice, downlink *ttnpb.ApplicationDownlink, defaultFormatters *ttnpb.MessagePayloadFormatters) error {
	if err := as.decryptDownlink(ctx, dev, downlink, nil); err != nil {
		return err
	}
	return as.decodeDownlink(ctx, dev, downlink, defaultFormatters)
}

func (as *ApplicationServer) decryptDownlink(ctx context.Context, dev *ttnpb.EndDevice, downlink *ttnpb.ApplicationDownlink, ternarySession *ttnpb.Session) error {
	var session *ttnpb.Session
	switch {
	case dev.Session != nil && bytes.Equal(dev.Session.SessionKeyID, downlink.SessionKeyID):
		session = dev.Session
	case dev.PendingSession != nil && bytes.Equal(dev.PendingSession.SessionKeyID, downlink.SessionKeyID):
		session = dev.PendingSession
	case ternarySession != nil && bytes.Equal(ternarySession.SessionKeyID, downlink.SessionKeyID):
		session = ternarySession
	default:
		return errUnknownSession.New()
	}
	if session.GetAppSKey() == nil {
		return errNoAppSKey.New()
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(ctx, session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.DecryptDownlink(appSKey, session.DevAddr, downlink.FCnt, downlink.FRMPayload, false)
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
			events.Publish(evtDecodeFailDataDown.NewWithIdentifiersAndData(ctx, &dev.EndDeviceIdentifiers, err))
		} else if len(downlink.DecodedPayloadWarnings) > 0 {
			events.Publish(evtDecodeWarningDataDown.NewWithIdentifiersAndData(ctx, &dev.EndDeviceIdentifiers, downlink))
		}
	}
	return nil
}

func (as *ApplicationServer) locationFromDecodedPayload(uplink *ttnpb.ApplicationUplink) (res *ttnpb.Location) {
	m, err := gogoproto.Map(uplink.DecodedPayload)
	if err != nil {
		return nil
	}
	loc, ok := apppayload.InferLocation(m)
	if !ok {
		return nil
	}
	return &ttnpb.Location{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Altitude:  int32(loc.Altitude),
		Accuracy:  int32(loc.Accuracy),
		Source:    ttnpb.SOURCE_GPS,
	}
}
