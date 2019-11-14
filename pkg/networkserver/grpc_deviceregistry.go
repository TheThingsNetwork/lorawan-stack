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
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtCreateEndDevice = events.Define(
		"ns.end_device.create", "create end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtUpdateEndDevice = events.Define(
		"ns.end_device.update", "update end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtDeleteEndDevice = events.Define(
		"ns.end_device.delete", "delete end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
)

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "queued_application_downlinks") {
		if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
			return nil, err
		}
	}

	gets := req.FieldMask.Paths
	if ttnpb.HasAnyField(req.FieldMask.Paths,
		"mac_state.queued_join_accept.keys.app_s_key.key",
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS); err != nil {
			return nil, err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"pending_session.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.f_nwk_s_int_key.encrypted_key",
				"pending_session.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"pending_session.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.nwk_s_enc_key.encrypted_key",
				"pending_session.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"pending_session.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.s_nwk_s_int_key.encrypted_key",
				"pending_session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"session.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.f_nwk_s_int_key.encrypted_key",
				"session.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"session.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"session.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"mac_state.queued_join_accept.keys.app_s_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
				"mac_state.queued_join_accept.keys.app_s_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths,
			"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.current_parameters.adr_ack_delay") && !ttnpb.HasAnyField(gets, "mac_state.current_parameters.adr_ack_delay_exponent") {
		gets = ttnpb.AddFields(gets, "mac_state.current_parameters.adr_ack_delay_exponent")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.current_parameters.adr_ack_limit") && !ttnpb.HasAnyField(gets, "mac_state.current_parameters.adr_ack_limit_exponent") {
		gets = ttnpb.AddFields(gets, "mac_state.current_parameters.adr_ack_limit_exponent")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.desired_parameters.adr_ack_delay") && !ttnpb.HasAnyField(gets, "mac_state.desired_parameters.adr_ack_delay_exponent") {
		gets = ttnpb.AddFields(gets, "mac_state.desired_parameters.adr_ack_delay_exponent")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.desired_parameters.adr_ack_limit") && !ttnpb.HasAnyField(gets, "mac_state.desired_parameters.adr_ack_limit_exponent") {
		gets = ttnpb.AddFields(gets, "mac_state.desired_parameters.adr_ack_limit_exponent")
	}

	dev, err := ns.devices.GetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, gets)
	if err != nil {
		return nil, err
	}

	if dev.GetMACState().GetQueuedJoinAccept() != nil && ttnpb.HasAnyField(req.FieldMask.Paths,
		"mac_state.queued_join_accept.keys.app_s_key.key",
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
	) {
		appSKey := dev.MACState.QueuedJoinAccept.Keys.AppSKey
		dev.MACState.QueuedJoinAccept.Keys.AppSKey = nil
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, ns.KeyVault, dev.MACState.QueuedJoinAccept.Keys, "mac_state.queued_join_accept.keys", req.FieldMask.Paths...)
		if err != nil {
			return nil, err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.queued_join_accept.keys.app_s_key.key") && appSKey != nil {
			key, err := cryptoutil.UnwrapAES128Key(ctx, *appSKey, ns.KeyVault)
			if err != nil && !errors.IsNotFound(err) {
				return nil, err
			} else if err == nil {
				sk.AppSKey = &ttnpb.KeyEnvelope{Key: &key}
			}
		}
		dev.MACState.QueuedJoinAccept.Keys = sk
	}
	if dev.GetPendingSession() != nil && ttnpb.HasAnyField(req.FieldMask.Paths,
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, ns.KeyVault, dev.PendingSession.SessionKeys, "pending_session.keys", req.FieldMask.Paths...)
		if err != nil {
			return nil, err
		}
		dev.PendingSession.SessionKeys = sk
	}
	if dev.GetSession() != nil && ttnpb.HasAnyField(req.FieldMask.Paths,
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, ns.KeyVault, dev.Session.SessionKeys, "session.keys", req.FieldMask.Paths...)
		if err != nil {
			return nil, err
		}
		dev.Session.SessionKeys = sk
	}

	if dev.MACState != nil {
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.current_parameters.adr_ack_delay") && dev.MACState.CurrentParameters.ADRAckDelayExponent != nil {
			dev.MACState.CurrentParameters.ADRAckDelay = lorawan.ADRAckDelayExponentToUint32(dev.MACState.CurrentParameters.ADRAckDelayExponent.Value)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.current_parameters.adr_ack_limit") && dev.MACState.CurrentParameters.ADRAckLimitExponent != nil {
			dev.MACState.CurrentParameters.ADRAckLimit = lorawan.ADRAckLimitExponentToUint32(dev.MACState.CurrentParameters.ADRAckLimitExponent.Value)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.desired_parameters.adr_ack_delay") && dev.MACState.DesiredParameters.ADRAckDelayExponent != nil {
			dev.MACState.DesiredParameters.ADRAckDelay = lorawan.ADRAckDelayExponentToUint32(dev.MACState.DesiredParameters.ADRAckDelayExponent.Value)
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.desired_parameters.adr_ack_limit") && dev.MACState.DesiredParameters.ADRAckLimitExponent != nil {
			dev.MACState.DesiredParameters.ADRAckLimit = lorawan.ADRAckLimitExponentToUint32(dev.MACState.DesiredParameters.ADRAckLimitExponent.Value)
		}
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

func validABPSessionKey(key *ttnpb.KeyEnvelope) bool {
	return key != nil && key.KEKLabel == "" && !key.Key.IsZero()
}

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.dev_addr") && (req.EndDevice.Session == nil || req.EndDevice.Session.DevAddr.IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.dev_addr")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.session_key_id") && (req.EndDevice.Session == nil || len(req.EndDevice.Session.SessionKeys.GetSessionKeyID()) == 0) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.session_key_id")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.f_nwk_s_int_key.key") && (req.EndDevice.Session == nil || req.EndDevice.Session.SessionKeys.GetFNwkSIntKey().GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.f_nwk_s_int_key.key")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.s_nwk_s_int_key.key") && (req.EndDevice.Session == nil || req.EndDevice.Session.SessionKeys.GetSNwkSIntKey().GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.s_nwk_s_int_key.key")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.nwk_s_enc_key.key") && (req.EndDevice.Session == nil || req.EndDevice.Session.SessionKeys.GetNwkSEncKey().GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.nwk_s_enc_key.key")
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_id") && req.EndDevice.FrequencyPlanID == "" {
		return nil, errInvalidFieldValue.WithAttributes("field", "frequency_plan_id")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "lorawan_version") {
		if err := req.EndDevice.LoRaWANVersion.Validate(); err != nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "lorawan_version").WithCause(err)
		}
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "lorawan_phy_version") {
		if err := req.EndDevice.LoRaWANPHYVersion.Validate(); err != nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "lorawan_phy_version").WithCause(err)
		}
	}

	if err := rights.RequireApplication(ctx, req.EndDevice.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths,
		"mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
		"mac_state.queued_join_accept.keys.app_s_key.kek_label",
		"mac_state.queued_join_accept.keys.app_s_key.key",
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.session_key_id",
		"pending_session.keys.f_nwk_s_int_key.encrypted_key",
		"pending_session.keys.f_nwk_s_int_key.kek_label",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.encrypted_key",
		"pending_session.keys.nwk_s_enc_key.kek_label",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.encrypted_key",
		"pending_session.keys.s_nwk_s_int_key.kek_label",
		"pending_session.keys.s_nwk_s_int_key.key",
		"pending_session.keys.session_key_id",
		"session.keys.f_nwk_s_int_key.encrypted_key",
		"session.keys.f_nwk_s_int_key.kek_label",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.encrypted_key",
		"session.keys.nwk_s_enc_key.kek_label",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.encrypted_key",
		"session.keys.s_nwk_s_int_key.kek_label",
		"session.keys.s_nwk_s_int_key.key",
		"session.keys.session_key_id",
	) {
		if err := rights.RequireApplication(ctx, req.EndDevice.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
			return nil, err
		}
	}

	gets := req.FieldMask.Paths
	var needsDownlinkCheck bool
	if ttnpb.HasAnyField([]string{
		"frequency_plan_id",
		"lorawan_phy_version",
		"mac_settings",
		"mac_state",
		"session",
	}, req.FieldMask.Paths...) {
		gets = ttnpb.AddFields(gets,
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
			"mac_state",
			"multicast",
			"queued_application_downlinks",
			"recent_uplinks",
			"session.dev_addr",
			"session.last_conf_f_cnt_down",
			"session.last_f_cnt_up",
			"session.last_n_f_cnt_down",
		)
		needsDownlinkCheck = true
	}

	var evt events.Event
	dev, err := ns.devices.SetByID(ctx, req.EndDevice.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDevice.EndDeviceIdentifiers.DeviceID, gets, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		sets := req.FieldMask.Paths
		if ttnpb.HasAnyField(sets, "version_ids") {
			// TODO: Apply version IDs (https://github.com/TheThingsIndustries/lorawan-stack/issues/1544)
		}

		if dev == nil {
			evt = evtCreateEndDevice(ctx, req.EndDevice.EndDeviceIdentifiers, nil)
		} else {
			evt = evtUpdateEndDevice(ctx, req.EndDevice.EndDeviceIdentifiers, req.FieldMask.Paths)
			if err := ttnpb.ProhibitFields(req.FieldMask.Paths,
				"ids.dev_addr",
				"multicast",
				"supports_join",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			if ttnpb.HasAnyField(req.FieldMask.Paths, "session.dev_addr") {
				req.EndDevice.DevAddr = &req.EndDevice.Session.DevAddr
				sets = append(sets, "ids.dev_addr")
			}
			return &req.EndDevice, sets, nil
		}

		if err := ttnpb.RequireFields(sets,
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
			"supports_join",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}

		if ttnpb.HasAnyField(sets, "supports_class_b") && req.EndDevice.SupportsClassB {
			if err := ttnpb.RequireFields(sets,
				"mac_settings.ping_slot_date_rate_index",
				"mac_settings.ping_slot_frequency",
				"mac_settings.ping_slot_periodicity",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
		}

		sets = ttnpb.AddFields(sets,
			"ids.application_ids",
			"ids.device_id",
		)
		if req.EndDevice.JoinEUI != nil && !req.EndDevice.JoinEUI.IsZero() {
			sets = ttnpb.AddFields(sets,
				"ids.join_eui",
			)
		}
		if req.EndDevice.DevEUI != nil && !req.EndDevice.DevEUI.IsZero() {
			sets = ttnpb.AddFields(sets,
				"ids.dev_eui",
			)
		}
		if req.EndDevice.DevAddr != nil {
			if !ttnpb.HasAnyField(req.FieldMask.Paths, "session.dev_addr") || !req.EndDevice.DevAddr.Equal(req.EndDevice.Session.DevAddr) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "ids.dev_addr")
			}
		}

		if req.EndDevice.SupportsJoin {
			if ttnpb.HasAnyField(sets, "multicast") && req.EndDevice.Multicast {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "supports_join")
			}
			if req.EndDevice.JoinEUI == nil {
				return nil, nil, errNoJoinEUI
			}
			if req.EndDevice.DevEUI == nil {
				return nil, nil, errNoDevEUI
			}
			if !ttnpb.HasAnyField([]string{"session"}, sets...) {
				return &req.EndDevice, sets, nil
			}
		}

		if err := ttnpb.RequireFields(sets,
			"session.dev_addr",
			"session.keys.f_nwk_s_int_key.key",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}
		req.EndDevice.EndDeviceIdentifiers.DevAddr = &req.EndDevice.Session.DevAddr
		sets = ttnpb.AddFields(sets,
			"ids.dev_addr",
		)

		if !validABPSessionKey(req.EndDevice.Session.FNwkSIntKey) {
			return nil, nil, errInvalidFNwkSIntKey
		}

		if req.EndDevice.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			if err := ttnpb.RequireFields(sets,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}

			if !validABPSessionKey(req.EndDevice.Session.SNwkSIntKey) {
				return nil, nil, errInvalidSNwkSIntKey
			}

			if !validABPSessionKey(req.EndDevice.Session.NwkSEncKey) {
				return nil, nil, errInvalidNwkSEncKey
			}
		} else {
			if err := ttnpb.ProhibitFields(sets,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			// TODO: Encrypt (https://github.com/TheThingsIndustries/lorawan-stack/issues/1562)
			req.EndDevice.Session.SNwkSIntKey = req.EndDevice.Session.FNwkSIntKey
			req.EndDevice.Session.NwkSEncKey = req.EndDevice.Session.FNwkSIntKey
			sets = append(sets,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			)
		}

		if ttnpb.HasAnyField(sets, "session.started_at") && req.EndDevice.GetSession().GetStartedAt().IsZero() {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.tarted_at")
		} else if !ttnpb.HasAnyField(sets, "session.started_at") {
			req.EndDevice.Session.StartedAt = timeNow().UTC()
			sets = append(sets, "session.started_at")
		}

		macState, err := newMACState(&req.EndDevice, ns.FrequencyPlans, ns.defaultMACSettings)
		if err != nil {
			return nil, nil, err
		}
		req.EndDevice.MACState = macState
		sets = ttnpb.AddFields(sets, "mac_state")

		return &req.EndDevice, sets, nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}

	if !needsDownlinkCheck {
		return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
	}

	var downAt time.Time
	_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to determine device band")
		downAt = timeNow().UTC()
	} else {
		var ok bool
		downAt, ok = nextDataDownlinkAt(ctx, dev, phy, ns.defaultMACSettings)
		if !ok {
			return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
		}
	}
	downAt = downAt.Add(-nsScheduleWindow)
	log.FromContext(ctx).WithField("start_at", downAt).Debug("Add downlink task after device set")
	if err := ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, downAt, true); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to add downlink task after device set")
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, nil, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			evt = evtDeleteEndDevice(ctx, req, nil)
		}
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	return ttnpb.Empty, err
}
