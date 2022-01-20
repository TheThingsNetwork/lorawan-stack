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
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	evtCreateEndDevice = events.Define(
		"as.end_device.create", "create end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateEndDevice = events.Define(
		"as.end_device.update", "update end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteEndDevice = events.Define(
		"as.end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

type asEndDeviceRegistryServer struct {
	AS       *ApplicationServer
	kekLabel string
}

func (r asEndDeviceRegistryServer) retrieveSessionKeys(ctx context.Context, dev *ttnpb.EndDevice, paths []string) error {
	unwrapKeys := func(ctx context.Context, session *ttnpb.Session, prefix string, paths ...string) error {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, r.AS.KeyVault, session.Keys, prefix, paths...)
		if err != nil {
			return err
		}
		session.Keys = sk
		return nil
	}

	needsPendingAppSKey := dev.GetPendingSession() != nil && ttnpb.HasAnyField(paths,
		"pending_session.keys.app_s_key.key",
	)
	needsAppSKey := dev.GetSession() != nil && ttnpb.HasAnyField(paths,
		"session.keys.app_s_key.key",
	)

	if !needsAppSKey && !needsPendingAppSKey {
		return nil
	}

	link, err := r.AS.getLink(ctx, dev.Ids.ApplicationIds, []string{
		"skip_payload_crypto",
	})
	if err != nil {
		return err
	}

	for _, k := range []struct {
		Name    string
		Needed  bool
		Session *ttnpb.Session
		Prefix  string
	}{
		{
			Name:    "current session",
			Needed:  needsAppSKey,
			Session: dev.Session,
			Prefix:  "session.keys",
		},
		{
			Name:    "pending session",
			Needed:  needsPendingAppSKey,
			Session: dev.PendingSession,
			Prefix:  "pending_session.keys",
		},
	} {
		if !k.Needed {
			continue
		}
		if r.AS.skipPayloadCrypto(ctx, link, dev, k.Session) {
			continue
		}
		if err := unwrapKeys(ctx, k.Session, k.Prefix, paths...); err != nil {
			warning.Add(ctx, fmt.Sprintf("Failed to unwrap the session keys for the %s: %v", k.Name, err))
		}
	}

	return nil
}

// Get implements ttnpb.AsEndDeviceRegistryServer.
func (r asEndDeviceRegistryServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}

	gets := req.FieldMask.GetPaths()
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(),
		"pending_session.keys.app_s_key.key",
		"session.keys.app_s_key.key",
	) {
		if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS); err != nil {
			return nil, err
		}
		gets = ttnpb.AddFields(gets, "skip_payload_crypto_override")
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(),
			"pending_session.keys.app_s_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.app_s_key.encrypted_key",
				"pending_session.keys.app_s_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(),
			"session.keys.app_s_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.app_s_key.encrypted_key",
				"session.keys.app_s_key.kek_label",
			)
		}
	}

	dev, err := r.AS.deviceRegistry.Get(ctx, req.EndDeviceIds, gets)
	if err != nil {
		return nil, err
	}

	if err := r.retrieveSessionKeys(ctx, dev, req.FieldMask.GetPaths()); err != nil {
		return nil, err
	}

	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

var (
	errInvalidFieldMask        = errors.DefineInvalidArgument("field_mask", "invalid field mask")
	errFormatterScriptTooLarge = errors.DefineInvalidArgument("formatter_script_too_large", "formatter script size exceeds maximum allowed size", "size", "max_size")
)

// Set implements ttnpb.AsEndDeviceRegistryServer.
func (r asEndDeviceRegistryServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "session.dev_addr") && (req.EndDevice.Session == nil || req.EndDevice.Session.DevAddr.IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.dev_addr")
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "session.keys.app_s_key.key") && (req.EndDevice.GetSession().GetKeys().GetAppSKey().GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.app_s_key.key")
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "formatters.up_formatter_parameter") {
		if size := len(req.EndDevice.GetFormatters().GetUpFormatterParameter()); size > r.AS.config.Formatters.MaxParameterLength {
			return nil, errInvalidFieldValue.WithAttributes("field", "formatters.up_formatter_parameter").WithCause(
				errFormatterScriptTooLarge.WithAttributes("size", size, "max_size", r.AS.config.Formatters.MaxParameterLength),
			)
		}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "formatters.down_formatter_parameter") {
		if size := len(req.EndDevice.GetFormatters().GetDownFormatterParameter()); size > r.AS.config.Formatters.MaxParameterLength {
			return nil, errInvalidFieldValue.WithAttributes("field", "formatters.down_formatter_parameter").WithCause(
				errFormatterScriptTooLarge.WithAttributes("size", size, "max_size", r.AS.config.Formatters.MaxParameterLength),
			)
		}
	}
	if err := rights.RequireApplication(ctx, *req.EndDevice.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(),
		"session.keys.app_s_key.key",
		"session.keys.session_key_id",
	) {
		if err := rights.RequireApplication(ctx, *req.EndDevice.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
			return nil, err
		}
	}

	sets := append(req.FieldMask.GetPaths()[:0:0], req.FieldMask.GetPaths()...)
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "session.keys.app_s_key.key") {
		appSKey, err := cryptoutil.WrapAES128Key(ctx, *req.EndDevice.Session.Keys.AppSKey.Key, r.kekLabel, r.AS.KeyVault)
		if err != nil {
			return nil, err
		}
		defer func(ke ttnpb.KeyEnvelope) {
			if dev != nil {
				dev.Session.Keys.AppSKey = &ke
			}
		}(*req.EndDevice.Session.Keys.AppSKey)
		req.EndDevice.Session.Keys.AppSKey = appSKey
		sets = ttnpb.AddFields(sets,
			"session.keys.app_s_key.encrypted_key",
			"session.keys.app_s_key.kek_label",
		)
	}

	var evt events.Event
	dev, err = r.AS.deviceRegistry.Set(ctx, req.EndDevice.Ids, req.FieldMask.GetPaths(), func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			evt = evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, req.EndDevice.Ids, req.FieldMask.GetPaths())
			if err := ttnpb.ProhibitFields(sets,
				"ids.dev_addr",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			if ttnpb.HasAnyField(sets, "session.dev_addr") {
				req.EndDevice.Ids.DevAddr = &req.EndDevice.Session.DevAddr
				sets = ttnpb.AddFields(sets,
					"ids.dev_addr",
				)
			}
			return &req.EndDevice, sets, nil
		}

		evt = evtCreateEndDevice.NewWithIdentifiersAndData(ctx, req.EndDevice.Ids, nil)

		if req.EndDevice.Ids.DevAddr != nil {
			if !ttnpb.HasAnyField(sets, "session.dev_addr") || !req.EndDevice.Ids.DevAddr.Equal(req.EndDevice.Session.DevAddr) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "ids.dev_addr")
			}
		}

		sets = ttnpb.AddFields(sets,
			"ids.application_ids",
			"ids.device_id",
		)
		if req.EndDevice.Ids.JoinEui != nil {
			sets = ttnpb.AddFields(sets,
				"ids.join_eui",
			)
		}
		if req.EndDevice.Ids.DevEui != nil && !req.EndDevice.Ids.DevEui.IsZero() {
			sets = ttnpb.AddFields(sets,
				"ids.dev_eui",
			)
		}
		return &req.EndDevice, sets, nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

// Delete implements ttnpb.AsEndDeviceRegistryServer.
func (r asEndDeviceRegistryServer) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, err := r.AS.deviceRegistry.Set(ctx, ids, nil, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev == nil {
			return nil, nil, errDeviceNotFound.New().WithAttributes("device_uid", unique.ID(ctx, ids))
		}
		evt = evtDeleteEndDevice.NewWithIdentifiersAndData(ctx, ids, nil)
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	if err := r.AS.appUpsRegistry.Clear(ctx, ids); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
