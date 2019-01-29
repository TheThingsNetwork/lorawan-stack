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

package networkserver

import (
	"bytes"
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	return ns.devices.GetByEUI(ctx, *req.EndDeviceIdentifiers.JoinEUI, *req.EndDeviceIdentifiers.DevEUI, req.FieldMask.Paths)
}

func validABPSessionKey(key *ttnpb.KeyEnvelope) bool {
	return key != nil &&
		key.KEKLabel == "" &&
		len(key.Key) == 16 &&
		!bytes.Equal(key.Key, bytes.Repeat([]byte{0}, 16))
}

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.Device.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var addDownlinkTask bool
	dev, err := ns.devices.SetByID(ctx, req.Device.EndDeviceIdentifiers.ApplicationIdentifiers, req.Device.EndDeviceIdentifiers.DeviceID, req.FieldMask.Paths, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		paths := req.FieldMask.Paths
		if dev != nil {
			addDownlinkTask = ttnpb.HasAnyField(paths, "mac_state.device_class") && req.Device.MACState.DeviceClass != ttnpb.CLASS_A ||
				ttnpb.HasAnyField(paths, "queued_application_downlinks") && len(req.Device.QueuedApplicationDownlinks) > 0
			return &req.Device, paths, nil
		}

		if ttnpb.HasAnyField(paths, "version_ids") {
			// TODO: Apply version IDs (https://github.com/TheThingsIndustries/lorawan-stack/issues/1544)
		}

		if req.Device.MACSettings == nil {
			return nil, nil, errNoMACSettings
		}

		if !ttnpb.HasAnyField(paths, "mac_settings.adr_margin") {
			// TODO: Apply NS-wide default (https://github.com/TheThingsIndustries/lorawan-stack/issues/1544)
			req.Device.MACSettings.ADRMargin = 15
			paths = append(paths, "mac_settings.adr_margin")
		} else if req.Device.MACSettings.ADRMargin == 0 {
			return nil, nil, errInvalidADRMargin
		}

		if err := ttnpb.RequireFields(paths,
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
			"mac_settings.use_adr",
			"resets_f_cnt",
			"resets_join_nonces",
			"supports_class_b",
			"supports_class_c",
			"supports_join",
			"uses_32_bit_f_cnt",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}

		if ttnpb.HasAnyField(paths, "supports_class_b") {
			if !ttnpb.HasAnyField(paths, "mac_settings.class_b_timeout") {
				// TODO: Apply NS-wide default if not set (https://github.com/TheThingsIndustries/lorawan-stack/issues/1544)
				req.Device.MACSettings.ClassBTimeout = time.Minute
				paths = append(paths, "mac_settings.class_b_timeout")
			} else if req.Device.MACSettings.ClassBTimeout == 0 {
				return nil, nil, errInvalidClassBTimeout
			}
		}

		if ttnpb.HasAnyField(paths, "supports_class_c") {
			if !ttnpb.HasAnyField(paths, "mac_settings.class_c_timeout") {
				// TODO: Apply NS-wide default if not set (https://github.com/TheThingsIndustries/lorawan-stack/issues/1544)
				req.Device.MACSettings.ClassCTimeout = 10 * time.Second
				paths = append(paths, "mac_settings.class_c_timeout")
			} else if req.Device.MACSettings.ClassCTimeout == 0 {
				return nil, nil, errInvalidClassCTimeout
			}
		}

		if req.Device.SupportsJoin {
			return &req.Device, paths, nil
		}

		if err := ttnpb.RequireFields(paths,
			"session.dev_addr",
			"session.keys.f_nwk_s_int_key.key",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}
		if req.Device.Session == nil {
			return nil, nil, errEmptySession
		}

		if !validABPSessionKey(req.Device.Session.FNwkSIntKey) {
			return nil, nil, errInvalidFNwkSIntKey
		}

		if req.Device.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			if err := ttnpb.RequireFields(paths,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}

			if !validABPSessionKey(req.Device.Session.SNwkSIntKey) {
				return nil, nil, errInvalidSNwkSIntKey
			}

			if !validABPSessionKey(req.Device.Session.NwkSEncKey) {
				return nil, nil, errInvalidNwkSEncKey
			}
		} else {
			if err := ttnpb.ProhibitFields(paths,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			// TODO: Encrypt (https://github.com/TheThingsIndustries/lorawan-stack/issues/1562)
			req.Device.Session.SNwkSIntKey = req.Device.Session.FNwkSIntKey
			req.Device.Session.NwkSEncKey = req.Device.Session.FNwkSIntKey
			paths = append(paths, "session.keys.s_nwk_s_int_key", "session.keys.nwk_s_enc_key")
		}
		req.Device.Session.StartedAt = time.Now().UTC()
		paths = append(paths, "session.started_at")

		if err := resetMACState(&req.Device, ns.FrequencyPlans); err != nil {
			return nil, nil, err
		}
		if req.Device.MACState.DeviceClass != ttnpb.CLASS_A {
			addDownlinkTask = len(req.Device.QueuedApplicationDownlinks) > 0 ||
				!req.Device.MACState.CurrentParameters.Equal(req.Device.MACState.DesiredParameters)
		}
		paths = append(paths, "mac_state")

		return &req.Device, paths, nil
	})
	if err != nil {
		return nil, err
	}
	if addDownlinkTask {
		if err = ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, time.Now()); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to add downlink task for device after set")
		}
	}
	return dev, nil
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	_, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, err
}
