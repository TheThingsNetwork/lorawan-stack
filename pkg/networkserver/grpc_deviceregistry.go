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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateEndDevice = events.Define(
		"ns.end_device.create", "create end device",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateEndDevice = events.Define(
		"ns.end_device.update", "update end device",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteEndDevice = events.Define(
		"ns.end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

const maxRequiredDeviceReadRightCount = 3

func appendRequiredDeviceReadRights(rights []ttnpb.Right, gets ...string) []ttnpb.Right {
	if len(gets) == 0 {
		return rights
	}
	rights = append(rights,
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	if ttnpb.HasAnyField(gets,
		"pending_session.queued_application_downlinks",
		"queued_application_downlinks",
		"session.queued_application_downlinks",
	) {
		rights = append(rights, ttnpb.RIGHT_APPLICATION_LINK)
	}
	if ttnpb.HasAnyField(gets,
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		rights = append(rights, ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS)
	}
	return rights
}

func addDeviceGetPaths(paths ...string) []string {
	gets := paths
	if ttnpb.HasAnyField(paths,
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		if ttnpb.HasAnyField(paths,
			"pending_session.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.f_nwk_s_int_key.encrypted_key",
				"pending_session.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_session.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.nwk_s_enc_key.encrypted_key",
				"pending_session.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_session.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.s_nwk_s_int_key.encrypted_key",
				"pending_session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(paths,
			"session.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.f_nwk_s_int_key.encrypted_key",
				"session.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"session.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"session.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(paths,
			"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(paths,
			"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
		}
	}
	return gets
}

func unwrapSelectedSessionKeys(ctx context.Context, kv crypto.KeyVault, dev *ttnpb.EndDevice, paths ...string) error {
	if dev.PendingSession != nil && ttnpb.HasAnyField(paths,
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingSession.SessionKeys, "pending_session.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingSession.SessionKeys = sk
	}
	if dev.Session != nil && ttnpb.HasAnyField(paths,
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.Session.SessionKeys, "session.keys", paths...)
		if err != nil {
			return err
		}
		dev.Session.SessionKeys = sk
	}

	if dev.PendingMACState.GetQueuedJoinAccept() != nil && ttnpb.HasAnyField(paths,
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingMACState.QueuedJoinAccept.Keys, "pending_mac_state.queued_join_accept.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingMACState.QueuedJoinAccept.Keys = sk
	}
	if dev.MACState.GetQueuedJoinAccept() != nil && ttnpb.HasAnyField(paths,
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.MACState.QueuedJoinAccept.Keys, "mac_state.queued_join_accept.keys", paths...)
		if err != nil {
			return err
		}
		dev.MACState.QueuedJoinAccept.Keys = sk
	}
	return nil
}

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, appendRequiredDeviceReadRights(
		make([]ttnpb.Right, 0, maxRequiredDeviceReadRightCount),
		req.FieldMask.Paths...,
	)...); err != nil {
		return nil, err
	}

	dev, ctx, err := ns.devices.GetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, addDeviceGetPaths(req.FieldMask.Paths...))
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get device from registry")
		return nil, err
	}
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyVault, dev, req.FieldMask.Paths...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_id") && req.EndDevice.FrequencyPlanID == "" {
		return nil, errInvalidFieldValue.WithAttributes("field", "frequency_plan_id")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "lorawan_phy_version") {
		if err := req.EndDevice.LoRaWANPHYVersion.Validate(); err != nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "lorawan_phy_version").WithCause(err)
		}
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "lorawan_version") {
		if err := req.EndDevice.LoRaWANVersion.Validate(); err != nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "lorawan_version").WithCause(err)
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.dev_addr") && (req.EndDevice.Session == nil || req.EndDevice.Session.DevAddr.IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.dev_addr")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.f_nwk_s_int_key.key") && (req.EndDevice.Session == nil || req.EndDevice.Session.FNwkSIntKey.GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.f_nwk_s_int_key.key")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.nwk_s_enc_key.key") && (req.EndDevice.Session == nil || req.EndDevice.Session.NwkSEncKey.GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.nwk_s_enc_key.key")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.s_nwk_s_int_key.key") && (req.EndDevice.Session == nil || req.EndDevice.Session.SNwkSIntKey.GetKey().IsZero()) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.s_nwk_s_int_key.key")
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.session_key_id") && (req.EndDevice.Session == nil || len(req.EndDevice.Session.SessionKeyID) == 0) {
		return nil, errInvalidFieldValue.WithAttributes("field", "session.keys.session_key_id")
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "multicast") && ttnpb.HasAnyField(req.FieldMask.Paths, "supports_join") && req.EndDevice.Multicast && req.EndDevice.SupportsJoin {
		return nil, errInvalidFieldValue.WithAttributes("field", "supports_join")
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "session") && ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state") {
		if req.EndDevice.Session == nil && req.EndDevice.MACState != nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "session")
		}
		if req.EndDevice.Session != nil && req.EndDevice.MACState == nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "mac_state")
		}
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "pending_session") && ttnpb.HasAnyField(req.FieldMask.Paths, "pending_mac_state") {
		if req.EndDevice.PendingSession == nil && req.EndDevice.PendingMACState != nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "pending_session")
		}
		if req.EndDevice.PendingSession != nil && req.EndDevice.PendingMACState == nil {
			return nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state")
		}
	}

	requiredRights := []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths,
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.session_key_id",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
		"session.keys.session_key_id",
	) {
		requiredRights = append(requiredRights, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS)
	}
	if err = rights.RequireApplication(ctx, req.EndDevice.ApplicationIdentifiers, requiredRights...); err != nil {
		return nil, err
	}

	sets := append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...)
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.f_nwk_s_int_key.key") {
		fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *req.EndDevice.Session.FNwkSIntKey.Key, ns.deviceKEKLabel, ns.KeyVault)
		if err != nil {
			return nil, err
		}
		defer func(ke ttnpb.KeyEnvelope) {
			if dev != nil {
				dev.Session.FNwkSIntKey = &ke
			}
		}(*req.EndDevice.Session.FNwkSIntKey)
		req.EndDevice.Session.FNwkSIntKey = fNwkSIntKey
		sets = ttnpb.AddFields(sets,
			"session.keys.f_nwk_s_int_key.encrypted_key",
			"session.keys.f_nwk_s_int_key.kek_label",
		)
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.nwk_s_enc_key.key") {
		nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, *req.EndDevice.Session.NwkSEncKey.Key, ns.deviceKEKLabel, ns.KeyVault)
		if err != nil {
			return nil, err
		}
		defer func(ke ttnpb.KeyEnvelope) {
			if dev != nil {
				dev.Session.NwkSEncKey = &ke
			}
		}(*req.EndDevice.Session.NwkSEncKey)
		req.EndDevice.Session.NwkSEncKey = nwkSEncKey
		sets = ttnpb.AddFields(sets,
			"session.keys.nwk_s_enc_key.encrypted_key",
			"session.keys.nwk_s_enc_key.kek_label",
		)
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.s_nwk_s_int_key.key") {
		sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *req.EndDevice.Session.SNwkSIntKey.Key, ns.deviceKEKLabel, ns.KeyVault)
		if err != nil {
			return nil, err
		}
		defer func(ke ttnpb.KeyEnvelope) {
			if dev != nil {
				dev.Session.SNwkSIntKey = &ke
			}
		}(*req.EndDevice.Session.SNwkSIntKey)
		req.EndDevice.Session.SNwkSIntKey = sNwkSIntKey
		sets = ttnpb.AddFields(sets,
			"session.keys.s_nwk_s_int_key.encrypted_key",
			"session.keys.s_nwk_s_int_key.kek_label",
		)
	}

	gets := append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...)
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
			"session.dev_addr",
			"session.last_conf_f_cnt_down",
			"session.last_f_cnt_up",
			"session.last_n_f_cnt_down",
			"session.queued_application_downlinks",
		)
		needsDownlinkCheck = true
	}

	var evt events.Event
	dev, ctx, err = ns.devices.SetByID(ctx, req.EndDevice.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDevice.EndDeviceIdentifiers.DeviceID, gets, func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if stored != nil {
			evt = evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, req.EndDevice.EndDeviceIdentifiers, req.FieldMask.Paths)
			if err := ttnpb.ProhibitFields(sets,
				"ids.dev_addr",
				"multicast",
				"supports_join",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			if ttnpb.HasAnyField(sets, "session.dev_addr") {
				req.EndDevice.DevAddr = &req.EndDevice.Session.DevAddr
				sets = ttnpb.AddFields(sets,
					"ids.dev_addr",
				)
			}
			if ttnpb.HasAnyField(sets,
				"frequency_plan_id",
				"lorawan_phy_version",
				"mac_settings.use_adr.value",
				"mac_state.current_parameters.adr_data_rate_index",
				"mac_state.current_parameters.adr_tx_power_index",
				"mac_state.desired_parameters.adr_data_rate_index",
				"mac_state.desired_parameters.adr_tx_power_index",
				"pending_mac_state.current_parameters.adr_data_rate_index",
				"pending_mac_state.current_parameters.adr_tx_power_index",
				"pending_mac_state.desired_parameters.adr_data_rate_index",
				"pending_mac_state.desired_parameters.adr_tx_power_index",
			) {
				if !ttnpb.HasAnyField(sets, "frequency_plan_id") {
					req.EndDevice.FrequencyPlanID = stored.FrequencyPlanID
				}
				if !ttnpb.HasAnyField(sets, "lorawan_phy_version") {
					req.EndDevice.LoRaWANPHYVersion = stored.LoRaWANPHYVersion
				}
				phy, err := DeviceBand(&req.EndDevice, ns.FrequencyPlans)
				if err != nil {
					return nil, nil, err
				}

				if ttnpb.HasAnyField(sets, "mac_settings.use_adr.value") && req.EndDevice.GetMACSettings().GetUseADR().GetValue() && !phy.EnableADR {
					return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.use_adr.value")
				}
				if req.EndDevice.MACState != nil {
					if ttnpb.HasAnyField(sets, "mac_state.current_parameters.adr_data_rate_index") && req.EndDevice.MACState.CurrentParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.current_parameters.adr_data_rate_index")
					}
					if ttnpb.HasAnyField(sets, "mac_state.desired_parameters.adr_data_rate_index") && req.EndDevice.MACState.DesiredParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.desired_parameters.adr_data_rate_index")
					}
					if ttnpb.HasAnyField(sets, "mac_state.current_parameters.adr_tx_power_index") && req.EndDevice.MACState.CurrentParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.current_parameters.adr_tx_power_index")
					}
					if ttnpb.HasAnyField(sets, "mac_state.desired_parameters.adr_tx_power_index") && req.EndDevice.MACState.DesiredParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.desired_parameters.adr_tx_power_index")
					}
				}
				if req.EndDevice.PendingMACState != nil {
					if ttnpb.HasAnyField(sets, "pending_mac_state.current_parameters.adr_data_rate_index") && req.EndDevice.PendingMACState.CurrentParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.current_parameters.adr_data_rate_index")
					}
					if ttnpb.HasAnyField(sets, "pending_mac_state.desired_parameters.adr_data_rate_index") && req.EndDevice.PendingMACState.DesiredParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.desired_parameters.adr_data_rate_index")
					}
					if ttnpb.HasAnyField(sets, "pending_mac_state.current_parameters.adr_tx_power_index") && req.EndDevice.PendingMACState.CurrentParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.current_parameters.adr_tx_power_index")
					}
					if ttnpb.HasAnyField(sets, "pending_mac_state.desired_parameters.adr_tx_power_index") && req.EndDevice.PendingMACState.DesiredParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
						return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.desired_parameters.adr_tx_power_index")
					}
				}
			}
			return &req.EndDevice, sets, nil
		}

		evt = evtCreateEndDevice.NewWithIdentifiersAndData(ctx, req.EndDevice.EndDeviceIdentifiers, nil)
		if err := ttnpb.RequireFields(sets,
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
			"supports_join",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}

		phy, err := DeviceBand(&req.EndDevice, ns.FrequencyPlans)
		if err != nil {
			return nil, nil, err
		}

		if ttnpb.HasAnyField(sets, "mac_settings.use_adr.value") && req.EndDevice.GetMACSettings().GetUseADR().GetValue() && !phy.EnableADR {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.use_adr.value")
		}

		if ttnpb.HasAnyField(sets, "supports_class_b") && req.EndDevice.SupportsClassB {
			if ns.defaultMACSettings.PingSlotFrequency == nil && phy.PingSlotFrequency == nil {
				if err := ttnpb.RequireFields(sets,
					"mac_settings.ping_slot_frequency.value",
				); err != nil {
					return nil, nil, errInvalidFieldMask.WithCause(err)
				}
				if req.EndDevice.GetMACSettings().GetPingSlotFrequency() == nil {
					return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_frequency")
				}
			}
			if ns.defaultMACSettings.PingSlotPeriodicity == nil && ttnpb.HasAnyField(req.FieldMask.Paths, "multicast") && req.EndDevice.Multicast {
				if err := ttnpb.RequireFields(sets,
					"mac_settings.ping_slot_periodicity.value",
				); err != nil {
					return nil, nil, errInvalidFieldMask.WithCause(err)
				}
				if req.EndDevice.GetMACSettings().GetPingSlotPeriodicity() == nil {
					return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_periodicity")
				}
			}
		}

		if req.EndDevice.DevAddr != nil {
			if !ttnpb.HasAnyField(sets, "session.dev_addr") || !req.EndDevice.DevAddr.Equal(req.EndDevice.Session.DevAddr) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "ids.dev_addr")
			}
		}

		sets = ttnpb.AddFields(sets,
			"ids.application_ids",
			"ids.device_id",
		)
		if req.EndDevice.JoinEUI != nil {
			sets = ttnpb.AddFields(sets,
				"ids.join_eui",
			)
		}
		if req.EndDevice.DevEUI != nil && !req.EndDevice.DevEUI.IsZero() {
			sets = ttnpb.AddFields(sets,
				"ids.dev_eui",
			)
		}

		if req.EndDevice.MACState != nil {
			if ttnpb.HasAnyField(sets, "mac_state.current_parameters.adr_data_rate_index") && req.EndDevice.MACState.CurrentParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.current_parameters.adr_data_rate_index")
			}
			if ttnpb.HasAnyField(sets, "mac_state.desired_parameters.adr_data_rate_index") && req.EndDevice.MACState.DesiredParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.desired_parameters.adr_data_rate_index")
			}
			if ttnpb.HasAnyField(sets, "mac_state.current_parameters.adr_tx_power_index") && req.EndDevice.MACState.CurrentParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.current_parameters.adr_tx_power_index")
			}
			if ttnpb.HasAnyField(sets, "mac_state.desired_parameters.adr_tx_power_index") && req.EndDevice.MACState.DesiredParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_state.desired_parameters.adr_tx_power_index")
			}
		}
		if req.EndDevice.PendingMACState != nil {
			if ttnpb.HasAnyField(sets, "pending_mac_state.current_parameters.adr_data_rate_index") && req.EndDevice.PendingMACState.CurrentParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.current_parameters.adr_data_rate_index")
			}
			if ttnpb.HasAnyField(sets, "pending_mac_state.desired_parameters.adr_data_rate_index") && req.EndDevice.PendingMACState.DesiredParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.desired_parameters.adr_data_rate_index")
			}
			if ttnpb.HasAnyField(sets, "pending_mac_state.current_parameters.adr_tx_power_index") && req.EndDevice.PendingMACState.CurrentParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.current_parameters.adr_tx_power_index")
			}
			if ttnpb.HasAnyField(sets, "pending_mac_state.desired_parameters.adr_tx_power_index") && req.EndDevice.PendingMACState.DesiredParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "pending_mac_state.desired_parameters.adr_tx_power_index")
			}
		}

		if req.EndDevice.SupportsJoin {
			if req.EndDevice.JoinEUI == nil {
				return nil, nil, errNoJoinEUI.New()
			}
			if req.EndDevice.DevEUI == nil {
				return nil, nil, errNoDevEUI.New()
			}
			if !ttnpb.HasAnyField([]string{"session"}, sets...) || req.EndDevice.Session == nil {
				return &req.EndDevice, sets, nil
			}
		} else if req.EndDevice.LoRaWANVersion.RequireDevEUIForABP() && req.EndDevice.DevEUI == nil {
			return nil, nil, errNoDevEUI.New()
		}

		if err := ttnpb.RequireFields(sets,
			"session.dev_addr",
			"session.keys.f_nwk_s_int_key.key",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}
		req.EndDevice.DevAddr = &req.EndDevice.Session.DevAddr
		sets = ttnpb.AddFields(sets,
			"ids.dev_addr",
		)

		if req.EndDevice.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			if err := ttnpb.RequireFields(sets,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
		} else {
			if err := ttnpb.ProhibitFields(sets,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			req.EndDevice.Session.NwkSEncKey = req.EndDevice.Session.FNwkSIntKey
			req.EndDevice.Session.SNwkSIntKey = req.EndDevice.Session.FNwkSIntKey
			sets = ttnpb.AddFields(sets,
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(sets, "session.started_at") && req.EndDevice.GetSession().GetStartedAt().IsZero() {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.started_at")
		} else if !ttnpb.HasAnyField(sets, "session.started_at") {
			req.EndDevice.Session.StartedAt = time.Now().UTC()
			sets = ttnpb.AddFields(sets,
				"session.started_at",
			)
		}

		macState, err := mac.NewState(&req.EndDevice, ns.FrequencyPlans, ns.defaultMACSettings)
		if err != nil {
			return nil, nil, err
		}
		req.EndDevice.MACState = macState
		sets = ttnpb.AddFields(sets, "mac_state")

		return &req.EndDevice, sets, nil
	})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to set device in registry")
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}

	if !needsDownlinkCheck {
		return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
	}

	if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after device set")
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

// ResetFactoryDefaults implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) ResetFactoryDefaults(ctx context.Context, req *ttnpb.ResetAndGetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, appendRequiredDeviceReadRights(
		append(make([]ttnpb.Right, 0, 1+maxRequiredDeviceReadRightCount), ttnpb.RIGHT_APPLICATION_DEVICES_WRITE),
		req.FieldMask.Paths...,
	)...); err != nil {
		return nil, err
	}

	dev, _, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, addDeviceGetPaths(ttnpb.AddFields(append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...),
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"session.dev_addr",
		"session.queued_application_downlinks",
		"session.keys",
		"supports_join",
	)...), func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if stored == nil {
			return nil, nil, errDeviceNotFound.New()
		}

		stored.BatteryPercentage = nil
		stored.DownlinkMargin = 0
		stored.LastDevStatusReceivedAt = nil
		stored.MACState = nil
		stored.PendingMACState = nil
		stored.PendingSession = nil
		stored.PowerState = ttnpb.PowerState_POWER_UNKNOWN
		if stored.SupportsJoin {
			stored.Session = nil
		} else {
			if stored.Session == nil {
				return nil, nil, errCorruptedMACState.New()
			}

			macState, err := mac.NewState(stored, ns.FrequencyPlans, ns.defaultMACSettings)
			if err != nil {
				return nil, nil, err
			}
			stored.MACState = macState
			stored.Session = &ttnpb.Session{
				DevAddr:                    stored.Session.DevAddr,
				SessionKeys:                stored.Session.SessionKeys,
				StartedAt:                  time.Now().UTC(),
				QueuedApplicationDownlinks: stored.Session.QueuedApplicationDownlinks,
			}
		}
		return stored, []string{
			"battery_percentage",
			"downlink_margin",
			"last_dev_status_received_at",
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"session",
		}, nil
	})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to reset device state in registry")
		return nil, err
	}
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyVault, dev, req.FieldMask.Paths...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, _, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, nil, func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev == nil {
			return nil, nil, errDeviceNotFound.New()
		}
		evt = evtDeleteEndDevice.NewWithIdentifiersAndData(ctx, req, nil)
		return nil, nil, nil
	})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete device from registry")
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	return ttnpb.Empty, nil
}
