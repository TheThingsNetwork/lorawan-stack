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
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	evtCreateEndDevice = events.Define(
		"ns.end_device.create", "create end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtUpdateEndDevice = events.Define(
		"ns.end_device.update", "update end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtDeleteEndDevice = events.Define(
		"ns.end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtBatchDeleteEndDevices = events.Define(
		"ns.end_device.batch.delete", "batch delete end devices",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithDataType(&ttnpb.EndDeviceIdentifiersList{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
)

const maxRequiredDeviceReadRightCount = 3

func appendRequiredDeviceReadRights(rights []ttnpb.Right, gets ...string) []ttnpb.Right {
	if len(gets) == 0 {
		return rights
	}
	rights = append(rights,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	)
	if ttnpb.HasAnyField(gets,
		"pending_session.queued_application_downlinks",
		"queued_application_downlinks",
		"session.queued_application_downlinks",
	) {
		rights = append(rights, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ)
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
		rights = append(rights, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS)
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

func unwrapSelectedSessionKeys(ctx context.Context, kv crypto.KeyService, dev *ttnpb.EndDevice, paths ...string) error {
	if dev.PendingSession != nil && ttnpb.HasAnyField(paths,
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingSession.Keys, "pending_session.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingSession.Keys = sk
	}
	if dev.Session != nil && ttnpb.HasAnyField(paths,
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.Session.Keys, "session.keys", paths...)
		if err != nil {
			return err
		}
		dev.Session.Keys = sk
	}
	if dev.PendingMacState.GetQueuedJoinAccept() != nil && ttnpb.HasAnyField(paths,
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingMacState.QueuedJoinAccept.Keys, "pending_mac_state.queued_join_accept.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingMacState.QueuedJoinAccept.Keys = sk
	}
	return nil
}

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIds, appendRequiredDeviceReadRights(
		make([]ttnpb.Right, 0, maxRequiredDeviceReadRightCount),
		req.FieldMask.GetPaths()...,
	)...); err != nil {
		return nil, err
	}

	dev, ctx, err := ns.devices.GetByID(ctx, req.EndDeviceIds.ApplicationIds, req.EndDeviceIds.DeviceId, addDeviceGetPaths(req.FieldMask.GetPaths()...))
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get device from registry")
		return nil, err
	}
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyService(), dev, req.FieldMask.GetPaths()...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

func newInvalidFieldValueError(field string) *errors.Error {
	return errInvalidFieldValue.WithAttributes("field", field)
}

func setKeyIsZero(m map[string]*ttnpb.EndDevice, get func(*ttnpb.EndDevice) *ttnpb.KeyEnvelope, path string) bool {
	if dev, ok := m[path+".key"]; ok {
		if ke := get(dev); !types.MustAES128Key(ke.GetKey()).OrZero().IsZero() {
			return false
		}
	}
	if dev, ok := m[path+".encrypted_key"]; ok {
		if ke := get(dev); len(ke.GetEncryptedKey()) != 0 {
			return false
		}
	}
	return true
}

func setKeyEqual(m map[string]*ttnpb.EndDevice, getA, getB func(*ttnpb.EndDevice) *ttnpb.KeyEnvelope, pathA, pathB string) bool {
	if a, b := getA(m[pathA+".key"]).GetKey(), getB(m[pathB+".key"]).GetKey(); a == nil && b != nil ||
		a != nil && b == nil ||
		a != nil && b != nil && !types.MustAES128Key(a).Equal(*types.MustAES128Key(b)) {
		return false
	}
	if a, b := getA(m[pathA+".encrypted_key"]).GetEncryptedKey(), getB(m[pathB+".encrypted_key"]).GetEncryptedKey(); !bytes.Equal(a, b) {
		return false
	}
	if a, b := getA(m[pathA+".kek_label"]).GetKekLabel(), getB(m[pathB+".kek_label"]).GetKekLabel(); a != b {
		return false
	}
	return true
}

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) { // nolint: gocyclo,lll
	st := newSetDeviceState(req.EndDevice, req.FieldMask.GetPaths()...)

	requiredRights := append(make([]ttnpb.Right, 0, 2),
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	)
	if st.HasSetField(
		"pending_mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.app_s_key.kek_label",
		"pending_mac_state.queued_join_accept.keys.app_s_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.session_key_id",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"pending_session.keys.session_key_id",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
		"session.keys.session_key_id",
	) {
		requiredRights = append(requiredRights, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS)
	}
	if err := rights.RequireApplication(ctx, st.Device.Ids.ApplicationIds, requiredRights...); err != nil {
		return nil, err
	}

	// Account for CLI not sending ids.* paths.
	st.AddSetFields(
		"ids.application_ids",
		"ids.device_id",
	)
	if st.Device.Ids.JoinEui != nil {
		st.AddSetFields(
			"ids.join_eui",
		)
	}
	if st.Device.Ids.DevEui != nil {
		st.AddSetFields(
			"ids.dev_eui",
		)
	}
	if st.Device.Ids.DevAddr != nil {
		st.AddSetFields(
			"ids.dev_addr",
		)
	}

	if err := st.ValidateSetField(
		func() bool { return st.Device.FrequencyPlanId != "" },
		"frequency_plan_id",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		st.Device.LorawanPhyVersion.Validate,
		"lorawan_phy_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		st.Device.LorawanVersion.Validate,
		"lorawan_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		func() error {
			if st.Device.MacState == nil {
				return nil
			}
			return st.Device.MacState.LorawanVersion.Validate()
		},
		"mac_state.lorawan_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		func() error {
			if st.Device.PendingMacState == nil {
				return nil
			}
			return st.Device.PendingMacState.LorawanVersion.Validate()
		},
		"pending_mac_state.lorawan_version",
	); err != nil {
		return nil, err
	}

	fps, err := ns.FrequencyPlansStore(ctx)
	if err != nil {
		return nil, err
	}

	var (
		profile                   *ttnpb.MACSettingsProfile
		macSettingsProfileChanged bool
	)

	if st.HasSetField(
		"mac_settings_profile_ids",
		"mac_settings_profile_ids.application_ids",
		"mac_settings_profile_ids.application_ids.application_id",
		"mac_settings_profile_ids.profile_id",
	) {
		if st.Device.MacSettingsProfileIds != nil {
			// If mac_settings_profile_ids is set, mac_settings must not be set.
			if st.HasSetField(macSettingsFields...) {
				return nil, newInvalidFieldValueError("mac_settings")
			}
			profile, err = ns.macSettingsProfiles.Get(
				ctx,
				st.Device.MacSettingsProfileIds,
				[]string{"ids", "mac_settings"},
			)
			if err != nil {
				return nil, err
			}

			if err = validateProfile(profile.GetMacSettings(), st.Device, fps); err != nil {
				return nil, err
			}

			// If mac_settings_profile_ids is set, mac_settings must not be set.
			st.Device.MacSettings = nil
			st.AddSetFields(macSettingsFields...)
		}
		macSettingsProfileChanged = true
	}

	if err := validateADR(st); err != nil {
		return nil, err
	}

	if err := validateZeroFields(st); err != nil {
		return nil, err
	}

	// Ensure parameters are consistent with band specifications.
	if st.HasSetField(
		"frequency_plan_id",
		"lorawan_phy_version",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.static.data_rate_index",
		"mac_settings.adr.mode.static.tx_power_index",
		"mac_settings.desired_ping_slot_data_rate_index.value",
		"mac_settings.desired_relay.mode.served.second_channel.data_rate_index",
		"mac_settings.desired_relay.mode.serving.default_channel_index",
		"mac_settings.desired_relay.mode.serving.second_channel.data_rate_index",
		"mac_settings.desired_rx2_data_rate_index.value",
		"mac_settings.downlink_dwell_time.value",
		"mac_settings.factory_preset_frequencies",
		"mac_settings.ping_slot_data_rate_index.value",
		"mac_settings.ping_slot_frequency.value",
		"mac_settings.relay.mode.served.second_channel.data_rate_index",
		"mac_settings.relay.mode.serving.default_channel_index",
		"mac_settings.relay.mode.serving.second_channel.data_rate_index",
		"mac_settings.rx2_data_rate_index.value",
		"mac_settings.uplink_dwell_time.value",
		"mac_settings.use_adr.value",
		"mac_state.current_parameters.adr_data_rate_index",
		"mac_state.current_parameters.adr_tx_power_index",
		"mac_state.current_parameters.channels",
		"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.current_parameters.rx2_data_rate_index",
		"mac_state.desired_parameters.adr_data_rate_index",
		"mac_state.desired_parameters.adr_tx_power_index",
		"mac_state.desired_parameters.channels",
		"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.desired_parameters.rx2_data_rate_index",
		"pending_mac_state.current_parameters.adr_data_rate_index",
		"pending_mac_state.current_parameters.adr_tx_power_index",
		"pending_mac_state.current_parameters.channels",
		"pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.rx2_data_rate_index",
		"pending_mac_state.desired_parameters.adr_data_rate_index",
		"pending_mac_state.desired_parameters.adr_tx_power_index",
		"pending_mac_state.desired_parameters.channels",
		"pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.rx2_data_rate_index",
		"supports_class_b",
	) {
		if err := validateBandSpecifications(st, fps); err != nil {
			return nil, err
		}
	}

	if err := validateADRDynamicParameters(st); err != nil {
		return nil, err
	}

	var getTransforms []func(*ttnpb.EndDevice)
	if st.Device.Session != nil {
		for p, isZero := range map[string]func() bool{
			"session.dev_addr":                 types.MustDevAddr(st.Device.Session.DevAddr).OrZero().IsZero,
			"session.keys.f_nwk_s_int_key.key": st.Device.Session.Keys.GetFNwkSIntKey().IsZero,
			"session.keys.nwk_s_enc_key.key": func() bool {
				return st.Device.Session.Keys.GetNwkSEncKey() != nil && st.Device.Session.Keys.NwkSEncKey.IsZero()
			},
			"session.keys.s_nwk_s_int_key.key": func() bool {
				return st.Device.Session.Keys.GetSNwkSIntKey() != nil && st.Device.Session.Keys.SNwkSIntKey.IsZero()
			},
		} {
			p, isZero := p, isZero
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("session.keys.f_nwk_s_int_key.key") {
			k := st.Device.Session.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.Session.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"session.keys.f_nwk_s_int_key.encrypted_key",
				"session.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if k := st.Device.Session.Keys.GetNwkSEncKey().GetKey(); k != nil && st.HasSetField("session.keys.nwk_s_enc_key.key") {
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.Session.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if k := st.Device.Session.Keys.GetSNwkSIntKey().GetKey(); k != nil && st.HasSetField("session.keys.s_nwk_s_int_key.key") {
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.Session.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}
	if st.Device.PendingSession != nil {
		for p, isZero := range map[string]func() bool{
			"pending_session.dev_addr":                 types.MustDevAddr(st.Device.PendingSession.DevAddr).OrZero().IsZero,
			"pending_session.keys.f_nwk_s_int_key.key": st.Device.PendingSession.Keys.GetFNwkSIntKey().IsZero,
			"pending_session.keys.nwk_s_enc_key.key":   st.Device.PendingSession.Keys.GetNwkSEncKey().IsZero,
			"pending_session.keys.s_nwk_s_int_key.key": st.Device.PendingSession.Keys.GetSNwkSIntKey().IsZero,
			"pending_session.keys.session_key_id": func() bool {
				return len(st.Device.PendingSession.Keys.GetSessionKeyId()) == 0
			},
		} {
			p, isZero := p, isZero
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("pending_session.keys.f_nwk_s_int_key.key") {
			k := st.Device.PendingSession.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"pending_session.keys.f_nwk_s_int_key.encrypted_key",
				"pending_session.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_session.keys.nwk_s_enc_key.key") {
			k := st.Device.PendingSession.Keys.NwkSEncKey.Key
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"pending_session.keys.nwk_s_enc_key.encrypted_key",
				"pending_session.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_session.keys.s_nwk_s_int_key.key") {
			k := st.Device.PendingSession.Keys.SNwkSIntKey.Key
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"pending_session.keys.s_nwk_s_int_key.encrypted_key",
				"pending_session.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}
	if st.Device.PendingMacState.GetQueuedJoinAccept() != nil {
		for p, isZero := range map[string]func() bool{
			"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key": st.Device.PendingMacState.QueuedJoinAccept.Keys.GetFNwkSIntKey().IsZero,
			"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key":   st.Device.PendingMacState.QueuedJoinAccept.Keys.GetNwkSEncKey().IsZero,
			"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key": st.Device.PendingMacState.QueuedJoinAccept.Keys.GetSNwkSIntKey().IsZero,
			"pending_mac_state.queued_join_accept.keys.session_key_id":      func() bool { return len(st.Device.PendingMacState.QueuedJoinAccept.Keys.GetSessionKeyId()) == 0 },
			"pending_mac_state.queued_join_accept.payload":                  func() bool { return len(st.Device.PendingMacState.QueuedJoinAccept.Payload) == 0 },
			"pending_mac_state.queued_join_accept.dev_addr": types.MustDevAddr(
				st.Device.PendingMacState.QueuedJoinAccept.DevAddr,
			).OrZero().IsZero,
		} {
			p, isZero := p, isZero
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key") {
			k := st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key") {
			k := st.Device.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey.Key
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key") {
			k := st.Device.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey.Key
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}

	var (
		// hasSession indicates whether the effective device model contains a non-zero session.
		hasSession bool

		// hasMACState indicates whether the effective device model contains a non-zero MAC state.
		hasMACState bool
	)
	if err := st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		for k, v := range m {
			switch {
			case strings.HasPrefix(k, "mac_state."):
				if v.MacState != nil {
					hasMACState = true
				}
			case strings.HasPrefix(k, "session."):
				if v.Session != nil {
					hasSession = true
				}
			}
			if hasMACState && hasSession {
				break
			}
		}

		isMulticast := m["multicast"].GetMulticast()
		switch {
		case !hasMACState && !hasSession && !isMulticast:
			return true, ""

		case !hasSession:
			return false, "session"

		case !hasMACState && st.HasSetField("mac_state"):
			return false, "mac_state"
		}

		var macVersion ttnpb.MACVersion
		if hasMACState {
			// NOTE: If not set, this will be derived from top-level device model.
			if isMulticast {
				if dev, ok := m["mac_state.device_class"]; ok && dev.MacState.GetDeviceClass() == ttnpb.Class_CLASS_A {
					return false, "mac_state.device_class"
				}
			}
			// NOTE: If not set, this will be derived from top-level device model.
			if dev, ok := m["mac_state.lorawan_version"]; ok && dev.MacState == nil {
				return false, "mac_state.lorawan_version"
			} else if !ok {
				macVersion = m["lorawan_version"].LorawanVersion
			} else {
				macVersion = dev.MacState.LorawanVersion
			}
		} else {
			macVersion = m["lorawan_version"].LorawanVersion
		}

		if dev, ok := m["session.dev_addr"]; !ok || dev.Session == nil {
			return false, "session.dev_addr"
		}

		getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetKeys().GetFNwkSIntKey()
		}
		if setKeyIsZero(m, getFNwkSIntKey, "session.keys.f_nwk_s_int_key") {
			return false, "session.keys.f_nwk_s_int_key.key"
		}

		getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetKeys().GetNwkSEncKey()
		}
		getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetKeys().GetSNwkSIntKey()
		}
		isZero := struct {
			NwkSEncKey  bool
			SNwkSIntKey bool
		}{
			NwkSEncKey:  setKeyIsZero(m, getNwkSEncKey, "session.keys.nwk_s_enc_key"),
			SNwkSIntKey: setKeyIsZero(m, getSNwkSIntKey, "session.keys.s_nwk_s_int_key"),
		}
		if macspec.UseNwkKey(macVersion) {
			if isZero.NwkSEncKey {
				return false, "session.keys.nwk_s_enc_key.key"
			}
			if isZero.SNwkSIntKey {
				return false, "session.keys.s_nwk_s_int_key.key"
			}
		} else {
			if st.HasSetField("session.keys.nwk_s_enc_key.key") &&
				!setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "session.keys.f_nwk_s_int_key", "session.keys.nwk_s_enc_key") {
				return false, "session.keys.nwk_s_enc_key.key"
			}
			if st.HasSetField("session.keys.s_nwk_s_int_key.key") &&
				!setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "session.keys.f_nwk_s_int_key", "session.keys.s_nwk_s_int_key") {
				return false, "session.keys.s_nwk_s_int_key.key"
			}
		}
		if m["supports_join"].GetSupportsJoin() {
			if dev, ok := m["session.keys.session_key_id"]; !ok || dev.Session == nil {
				return false, "session.keys.session_key_id"
			}
		}
		return true, ""
	},
		"lorawan_version",
		"mac_state.current_parameters.adr_ack_delay_exponent.value",
		"mac_state.current_parameters.adr_ack_limit_exponent.value",
		"mac_state.current_parameters.adr_data_rate_index",
		"mac_state.current_parameters.adr_nb_trans",
		"mac_state.current_parameters.adr_tx_power_index",
		"mac_state.current_parameters.beacon_frequency",
		"mac_state.current_parameters.channels",
		"mac_state.current_parameters.downlink_dwell_time.value",
		"mac_state.current_parameters.max_duty_cycle",
		"mac_state.current_parameters.max_eirp",
		"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.current_parameters.ping_slot_frequency",
		"mac_state.current_parameters.rejoin_count_periodicity",
		"mac_state.current_parameters.rejoin_time_periodicity",
		"mac_state.current_parameters.relay.mode.served.backoff",
		"mac_state.current_parameters.relay.mode.served.mode.always",
		"mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
		"mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
		"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.served.second_channel.frequency",
		"mac_state.current_parameters.relay.mode.served.serving_device_id",
		"mac_state.current_parameters.relay.mode.serving.cad_periodicity",
		"mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
		"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
		"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
		"mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
		"mac_state.current_parameters.rx1_data_rate_offset",
		"mac_state.current_parameters.rx1_delay",
		"mac_state.current_parameters.rx2_data_rate_index",
		"mac_state.current_parameters.rx2_frequency",
		"mac_state.current_parameters.uplink_dwell_time.value",
		"mac_state.desired_parameters.adr_ack_delay_exponent.value",
		"mac_state.desired_parameters.adr_ack_limit_exponent.value",
		"mac_state.desired_parameters.adr_data_rate_index",
		"mac_state.desired_parameters.adr_nb_trans",
		"mac_state.desired_parameters.adr_tx_power_index",
		"mac_state.desired_parameters.beacon_frequency",
		"mac_state.desired_parameters.channels",
		"mac_state.desired_parameters.downlink_dwell_time.value",
		"mac_state.desired_parameters.max_duty_cycle",
		"mac_state.desired_parameters.max_eirp",
		"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.desired_parameters.ping_slot_frequency",
		"mac_state.desired_parameters.rejoin_count_periodicity",
		"mac_state.desired_parameters.rejoin_time_periodicity",
		"mac_state.desired_parameters.relay.mode.served.backoff",
		"mac_state.desired_parameters.relay.mode.served.mode.always",
		"mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
		"mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
		"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
		"mac_state.desired_parameters.relay.mode.served.serving_device_id",
		"mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
		"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
		"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
		"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
		"mac_state.desired_parameters.rx1_data_rate_offset",
		"mac_state.desired_parameters.rx1_delay",
		"mac_state.desired_parameters.rx2_data_rate_index",
		"mac_state.desired_parameters.rx2_frequency",
		"mac_state.desired_parameters.uplink_dwell_time.value",
		"mac_state.device_class",
		"mac_state.last_adr_change_f_cnt_up",
		"mac_state.last_confirmed_downlink_at",
		"mac_state.last_dev_status_f_cnt_up",
		"mac_state.last_downlink_at",
		"mac_state.last_network_initiated_downlink_at",
		"mac_state.lorawan_version",
		"mac_state.pending_application_downlink.class_b_c.absolute_time",
		"mac_state.pending_application_downlink.class_b_c.gateways",
		"mac_state.pending_application_downlink.confirmed",
		"mac_state.pending_application_downlink.correlation_ids",
		"mac_state.pending_application_downlink.f_cnt",
		"mac_state.pending_application_downlink.f_port",
		"mac_state.pending_application_downlink.frm_payload",
		"mac_state.pending_application_downlink.network_ids",
		"mac_state.pending_application_downlink.network_ids.cluster_address",
		"mac_state.pending_application_downlink.network_ids.cluster_id",
		"mac_state.pending_application_downlink.network_ids.net_id",
		"mac_state.pending_application_downlink.network_ids.ns_id",
		"mac_state.pending_application_downlink.network_ids.tenant_address",
		"mac_state.pending_application_downlink.network_ids.tenant_id",
		"mac_state.pending_application_downlink.priority",
		"mac_state.pending_application_downlink.session_key_id",
		"mac_state.pending_relay_downlink.raw_payload",
		"mac_state.pending_relay_downlink",
		"mac_state.pending_requests",
		"mac_state.ping_slot_periodicity.value",
		"mac_state.queued_responses",
		"mac_state.recent_downlinks",
		"mac_state.recent_mac_command_identifiers",
		"mac_state.recent_uplinks",
		"mac_state.rejected_adr_data_rate_indexes",
		"mac_state.rejected_adr_tx_power_indexes",
		"mac_state.rejected_data_rate_ranges",
		"mac_state.rejected_frequencies",
		"mac_state.rx_windows_available",
		"multicast",
		"session.dev_addr",
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
		"session.last_conf_f_cnt_down",
		"session.last_f_cnt_up",
		"session.last_n_f_cnt_down",
		"session.started_at",
		"supports_join",
	); err != nil {
		return nil, err
	}

	var (
		// hasPendingSession indicates whether the effective device model contains a non-zero pending session.
		hasPendingSession bool

		// hasQueuedJoinAccept indicates whether the effective device model contains a non-zero queued join-accept.
		hasQueuedJoinAccept bool
	)
	if err := st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		var hasPendingMACState bool
		for k, v := range m {
			switch {
			case strings.HasPrefix(k, "pending_mac_state."):
				if v.PendingMacState != nil {
					hasPendingMACState = true
				}
			case strings.HasPrefix(k, "pending_session."):
				if v.PendingSession != nil {
					hasPendingSession = true
				}
			}
			if hasPendingMACState && hasPendingSession {
				break
			}
		}
		switch {
		case !hasPendingMACState && !hasPendingSession:
			return true, ""
		case !hasPendingMACState:
			return false, "pending_mac_state"
		}
		for k, v := range m {
			if strings.HasPrefix(k, "pending_mac_state.queued_join_accept.") && v.PendingMacState.GetQueuedJoinAccept() != nil {
				hasQueuedJoinAccept = true
				break
			}
		}

		var macVersion ttnpb.MACVersion
		if dev, ok := m["pending_mac_state.lorawan_version"]; !ok || dev.PendingMacState == nil {
			return false, "pending_mac_state.lorawan_version"
		} else {
			macVersion = dev.PendingMacState.LorawanVersion
		}
		useNwkKey := macspec.UseNwkKey(macVersion)

		if hasPendingSession {
			// NOTE: PendingMACState may be set before PendingSession is set by downlink routine.
			if dev, ok := m["pending_session.dev_addr"]; !ok || dev.PendingSession == nil {
				return false, "pending_session.dev_addr"
			}

			getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				return dev.GetPendingSession().GetKeys().GetFNwkSIntKey()
			}
			if setKeyIsZero(m, getFNwkSIntKey, "pending_session.keys.f_nwk_s_int_key") {
				return false, "pending_session.keys.f_nwk_s_int_key.key"
			}
			getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				return dev.GetPendingSession().GetKeys().GetNwkSEncKey()
			}
			if setKeyIsZero(m, getNwkSEncKey, "pending_session.keys.nwk_s_enc_key") {
				return false, "pending_session.keys.nwk_s_enc_key.key"
			}
			getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				return dev.GetPendingSession().GetKeys().GetSNwkSIntKey()
			}
			if setKeyIsZero(m, getSNwkSIntKey, "pending_session.keys.s_nwk_s_int_key") {
				return false, "pending_session.keys.s_nwk_s_int_key.key"
			}
			if !useNwkKey {
				if !setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "pending_session.keys.f_nwk_s_int_key", "pending_session.keys.nwk_s_enc_key") {
					return false, "pending_session.keys.nwk_s_enc_key.key"
				}
				if !setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "pending_session.keys.f_nwk_s_int_key", "pending_session.keys.s_nwk_s_int_key") {
					return false, "pending_session.keys.s_nwk_s_int_key.key"
				}
			}
			if dev, ok := m["pending_session.keys.session_key_id"]; !ok || dev.PendingSession == nil {
				return false, "pending_session.keys.session_key_id"
			}
		} else if !hasQueuedJoinAccept {
			return false, "pending_mac_state.queued_join_accept"
		}

		if hasQueuedJoinAccept {
			getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMacState().GetQueuedJoinAccept().GetKeys()
				return keys.GetFNwkSIntKey()
			}
			if setKeyIsZero(m, getFNwkSIntKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key") {
				return false, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key"
			}
			getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMacState().GetQueuedJoinAccept().GetKeys()
				return keys.GetNwkSEncKey()
			}
			if setKeyIsZero(m, getNwkSEncKey, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key") {
				return false, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key"
			}
			getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMacState().GetQueuedJoinAccept().GetKeys()
				return keys.GetSNwkSIntKey()
			}
			if setKeyIsZero(m, getSNwkSIntKey, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key") {
				return false, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key"
			}

			if !useNwkKey {
				if !setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key", "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key") {
					return false, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key"
				}
				if !setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key", "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key") {
					return false, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key"
				}
			}

			if dev, ok := m["pending_mac_state.queued_join_accept.keys.session_key_id"]; !ok || dev.PendingMacState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.keys.session_key_id"
			}
			if dev, ok := m["pending_mac_state.queued_join_accept.payload"]; !ok || dev.PendingMacState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.payload"
			}
			if dev, ok := m["pending_mac_state.queued_join_accept.request.dev_addr"]; !ok || dev.PendingMacState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.request.dev_addr"
			}
		}
		return true, ""
	},
		"pending_mac_state.current_parameters.adr_ack_delay_exponent.value",
		"pending_mac_state.current_parameters.adr_ack_limit_exponent.value",
		"pending_mac_state.current_parameters.adr_data_rate_index",
		"pending_mac_state.current_parameters.adr_nb_trans",
		"pending_mac_state.current_parameters.adr_tx_power_index",
		"pending_mac_state.current_parameters.beacon_frequency",
		"pending_mac_state.current_parameters.channels",
		"pending_mac_state.current_parameters.downlink_dwell_time.value",
		"pending_mac_state.current_parameters.max_duty_cycle",
		"pending_mac_state.current_parameters.max_eirp",
		"pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.current_parameters.ping_slot_frequency",
		"pending_mac_state.current_parameters.rejoin_count_periodicity",
		"pending_mac_state.current_parameters.rejoin_time_periodicity",
		"pending_mac_state.current_parameters.relay.mode.served.backoff",
		"pending_mac_state.current_parameters.relay.mode.served.mode.always",
		"pending_mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"pending_mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.frequency",
		"pending_mac_state.current_parameters.relay.mode.served.serving_device_id",
		"pending_mac_state.current_parameters.relay.mode.serving.cad_periodicity",
		"pending_mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
		"pending_mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
		"pending_mac_state.current_parameters.rx1_data_rate_offset",
		"pending_mac_state.current_parameters.rx1_delay",
		"pending_mac_state.current_parameters.rx2_data_rate_index",
		"pending_mac_state.current_parameters.rx2_frequency",
		"pending_mac_state.current_parameters.uplink_dwell_time.value",
		"pending_mac_state.desired_parameters.adr_ack_delay_exponent.value",
		"pending_mac_state.desired_parameters.adr_ack_limit_exponent.value",
		"pending_mac_state.desired_parameters.adr_data_rate_index",
		"pending_mac_state.desired_parameters.adr_nb_trans",
		"pending_mac_state.desired_parameters.adr_tx_power_index",
		"pending_mac_state.desired_parameters.beacon_frequency",
		"pending_mac_state.desired_parameters.channels",
		"pending_mac_state.desired_parameters.downlink_dwell_time.value",
		"pending_mac_state.desired_parameters.max_duty_cycle",
		"pending_mac_state.desired_parameters.max_eirp",
		"pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.desired_parameters.ping_slot_frequency",
		"pending_mac_state.desired_parameters.rejoin_count_periodicity",
		"pending_mac_state.desired_parameters.rejoin_time_periodicity",
		"pending_mac_state.desired_parameters.relay.mode.served.backoff",
		"pending_mac_state.desired_parameters.relay.mode.served.mode.always",
		"pending_mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"pending_mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
		"pending_mac_state.desired_parameters.relay.mode.served.serving_device_id",
		"pending_mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
		"pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
		"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
		"pending_mac_state.desired_parameters.rx1_data_rate_offset",
		"pending_mac_state.desired_parameters.rx1_delay",
		"pending_mac_state.desired_parameters.rx2_data_rate_index",
		"pending_mac_state.desired_parameters.rx2_frequency",
		"pending_mac_state.desired_parameters.uplink_dwell_time.value",
		"pending_mac_state.device_class",
		"pending_mac_state.last_adr_change_f_cnt_up",
		"pending_mac_state.last_confirmed_downlink_at",
		"pending_mac_state.last_dev_status_f_cnt_up",
		"pending_mac_state.last_downlink_at",
		"pending_mac_state.last_network_initiated_downlink_at",
		"pending_mac_state.lorawan_version",
		"pending_mac_state.pending_join_request.cf_list.ch_masks",
		"pending_mac_state.pending_join_request.cf_list.freq",
		"pending_mac_state.pending_join_request.cf_list.type",
		"pending_mac_state.pending_join_request.downlink_settings.opt_neg",
		"pending_mac_state.pending_join_request.downlink_settings.rx1_dr_offset",
		"pending_mac_state.pending_join_request.downlink_settings.rx2_dr",
		"pending_mac_state.pending_join_request.rx_delay",
		"pending_mac_state.ping_slot_periodicity.value",
		"pending_mac_state.queued_join_accept.correlation_ids",
		"pending_mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.app_s_key.kek_label",
		"pending_mac_state.queued_join_accept.keys.app_s_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.session_key_id",
		"pending_mac_state.queued_join_accept.payload",
		"pending_mac_state.queued_join_accept.request.cf_list.ch_masks",
		"pending_mac_state.queued_join_accept.request.cf_list.freq",
		"pending_mac_state.queued_join_accept.request.cf_list.type",
		"pending_mac_state.queued_join_accept.request.dev_addr",
		"pending_mac_state.queued_join_accept.request.downlink_settings.opt_neg",
		"pending_mac_state.queued_join_accept.request.downlink_settings.rx1_dr_offset",
		"pending_mac_state.queued_join_accept.request.downlink_settings.rx2_dr",
		"pending_mac_state.queued_join_accept.request.net_id",
		"pending_mac_state.queued_join_accept.request.rx_delay",
		"pending_mac_state.queued_responses",
		"pending_mac_state.recent_downlinks",
		"pending_mac_state.recent_mac_command_identifiers",
		"pending_mac_state.recent_uplinks",
		"pending_mac_state.rejected_adr_data_rate_indexes",
		"pending_mac_state.rejected_adr_tx_power_indexes",
		"pending_mac_state.rejected_data_rate_ranges",
		"pending_mac_state.rejected_frequencies",
		"pending_mac_state.rx_windows_available",
		"pending_session.dev_addr",
		"pending_session.keys.f_nwk_s_int_key.encrypted_key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.encrypted_key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.encrypted_key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"pending_session.keys.session_key_id",
	); err != nil {
		return nil, err
	}

	needsDownlinkCheck := st.HasSetField(downlinkInfluencingSetFields[:]...)
	if needsDownlinkCheck {
		st.AddGetFields(
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
			"mac_state.current_parameters.adr_ack_delay_exponent.value",
			"mac_state.current_parameters.adr_ack_limit_exponent.value",
			"mac_state.current_parameters.adr_data_rate_index",
			"mac_state.current_parameters.adr_nb_trans",
			"mac_state.current_parameters.adr_tx_power_index",
			"mac_state.current_parameters.beacon_frequency",
			"mac_state.current_parameters.channels",
			"mac_state.current_parameters.downlink_dwell_time.value",
			"mac_state.current_parameters.max_duty_cycle",
			"mac_state.current_parameters.max_eirp",
			"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
			"mac_state.current_parameters.ping_slot_frequency",
			"mac_state.current_parameters.rejoin_count_periodicity",
			"mac_state.current_parameters.rejoin_time_periodicity",
			"mac_state.current_parameters.relay.mode.served.backoff",
			"mac_state.current_parameters.relay.mode.served.mode.always",
			"mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
			"mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
			"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
			"mac_state.current_parameters.relay.mode.served.second_channel.frequency",
			"mac_state.current_parameters.relay.mode.served.serving_device_id",
			"mac_state.current_parameters.relay.mode.serving.cad_periodicity",
			"mac_state.current_parameters.relay.mode.serving.default_channel_index",
			"mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
			"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
			"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
			"mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
			"mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
			"mac_state.current_parameters.rx1_data_rate_offset",
			"mac_state.current_parameters.rx1_delay",
			"mac_state.current_parameters.rx2_data_rate_index",
			"mac_state.current_parameters.rx2_frequency",
			"mac_state.current_parameters.uplink_dwell_time.value",
			"mac_state.desired_parameters.adr_ack_delay_exponent.value",
			"mac_state.desired_parameters.adr_ack_limit_exponent.value",
			"mac_state.desired_parameters.adr_data_rate_index",
			"mac_state.desired_parameters.adr_nb_trans",
			"mac_state.desired_parameters.adr_tx_power_index",
			"mac_state.desired_parameters.beacon_frequency",
			"mac_state.desired_parameters.channels",
			"mac_state.desired_parameters.downlink_dwell_time.value",
			"mac_state.desired_parameters.max_duty_cycle",
			"mac_state.desired_parameters.max_eirp",
			"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
			"mac_state.desired_parameters.ping_slot_frequency",
			"mac_state.desired_parameters.rejoin_count_periodicity",
			"mac_state.desired_parameters.rejoin_time_periodicity",
			"mac_state.desired_parameters.relay.mode.served.backoff",
			"mac_state.desired_parameters.relay.mode.served.mode.always",
			"mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
			"mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
			"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
			"mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
			"mac_state.desired_parameters.relay.mode.served.serving_device_id",
			"mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
			"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
			"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
			"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
			"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
			"mac_state.desired_parameters.rx1_data_rate_offset",
			"mac_state.desired_parameters.rx1_delay",
			"mac_state.desired_parameters.rx2_data_rate_index",
			"mac_state.desired_parameters.rx2_frequency",
			"mac_state.desired_parameters.uplink_dwell_time.value",
			"mac_state.device_class",
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_dev_status_f_cnt_up",
			"mac_state.last_downlink_at",
			"mac_state.last_network_initiated_downlink_at",
			"mac_state.lorawan_version",
			"mac_state.ping_slot_periodicity.value",
			"mac_state.queued_responses",
			"mac_state.recent_mac_command_identifiers",
			"mac_state.recent_uplinks",
			"mac_state.rejected_adr_data_rate_indexes",
			"mac_state.rejected_adr_tx_power_indexes",
			"mac_state.rejected_data_rate_ranges",
			"mac_state.rejected_frequencies",
			"mac_state.rx_windows_available",
			"multicast",
			"session.dev_addr",
			"session.last_conf_f_cnt_down",
			"session.last_f_cnt_up",
			"session.last_n_f_cnt_down",
			"session.queued_application_downlinks",
			"supports_join",
		)
	}

	var evt events.Event
	dev, ctx, err := ns.devices.SetByID(ctx, st.Device.Ids.ApplicationIds, st.Device.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel, st.SetFunc(func(ctx context.Context, stored *ttnpb.EndDevice) error {
		if nonZeroFields := ttnpb.NonZeroFields(stored, st.GetFields()...); len(nonZeroFields) > 0 {
			newStored := &ttnpb.EndDevice{}
			if err := newStored.SetFields(stored, nonZeroFields...); err != nil {
				return err
			}
			stored = newStored
		}
		if hasSession {
			macVersion := stored.GetMacState().GetLorawanVersion()
			if stored.GetMacState() == nil && !st.HasSetField("mac_state") {
				fps, err := ns.FrequencyPlansStore(ctx)
				if err != nil {
					return err
				}
				macState, err := mac.NewState(st.Device, fps, ns.defaultMACSettings, profile.GetMacSettings())
				if err != nil {
					return err
				}
				if macSets := ttnpb.FieldsWithoutPrefix("mac_state", st.SetFields()...); len(macSets) != 0 {
					if err := macState.SetFields(st.Device.MacState, macSets...); err != nil {
						return err
					}
				}
				st.Device.MacState = macState
				st.AddSetFields(
					"mac_state",
				)
				macVersion = macState.LorawanVersion
			} else if st.HasSetField("mac_state.lorawan_version") {
				macVersion = st.Device.MacState.LorawanVersion
			}

			if st.HasSetField("session.keys.f_nwk_s_int_key.key") && !macspec.UseNwkKey(macVersion) {
				st.Device.Session.Keys.NwkSEncKey = st.Device.Session.Keys.FNwkSIntKey
				st.Device.Session.Keys.SNwkSIntKey = st.Device.Session.Keys.FNwkSIntKey
				st.AddSetFields(
					"session.keys.nwk_s_enc_key.encrypted_key",
					"session.keys.nwk_s_enc_key.kek_label",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.encrypted_key",
					"session.keys.s_nwk_s_int_key.kek_label",
					"session.keys.s_nwk_s_int_key.key",
				)
			}
			if st.HasSetField("session.started_at") && st.Device.GetSession().GetStartedAt() == nil ||
				st.HasSetField("session.session_key_id") && !bytes.Equal(st.Device.GetSession().GetKeys().GetSessionKeyId(), stored.GetSession().GetKeys().GetSessionKeyId()) ||
				stored.GetSession().GetStartedAt() == nil {
				st.Device.Session.StartedAt = timestamppb.New(time.Now()) // NOTE: This is not equivalent to timestamppb.Now().
				st.AddSetFields(
					"session.started_at",
				)
			}
		}
		if hasPendingSession {
			var macVersion ttnpb.MACVersion
			if st.HasSetField("pending_mac_state.lorawan_version") {
				macVersion = st.Device.GetPendingMacState().GetLorawanVersion()
			} else {
				macVersion = stored.GetPendingMacState().GetLorawanVersion()
			}

			useNwkKey := macspec.UseNwkKey(macVersion)
			if st.HasSetField("pending_session.keys.f_nwk_s_int_key.key") && !useNwkKey {
				st.Device.PendingSession.Keys.NwkSEncKey = st.Device.PendingSession.Keys.FNwkSIntKey
				st.Device.PendingSession.Keys.SNwkSIntKey = st.Device.PendingSession.Keys.FNwkSIntKey
				st.AddSetFields(
					"pending_session.keys.nwk_s_enc_key.encrypted_key",
					"pending_session.keys.nwk_s_enc_key.kek_label",
					"pending_session.keys.nwk_s_enc_key.key",
					"pending_session.keys.s_nwk_s_int_key.encrypted_key",
					"pending_session.keys.s_nwk_s_int_key.kek_label",
					"pending_session.keys.s_nwk_s_int_key.key",
				)
			}
			if st.HasSetField("pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key") && hasQueuedJoinAccept && !useNwkKey {
				st.Device.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey = st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey
				st.Device.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey = st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey
				st.AddSetFields(
					"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
					"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
					"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
					"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
					"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
					"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
				)
			}
		}
		if macSettingsProfileChanged &&
			!(stored.GetMacSettingsProfileIds() == nil &&
				st.Device.MacSettingsProfileIds == nil) {
			id := st.Device.MacSettingsProfileIds
			if id == nil {
				id = stored.GetMacSettingsProfileIds()
			}

			profile, err = ns.macSettingsProfiles.Get(
				ctx,
				id,
				[]string{"ids", "mac_settings", "end_devices_count"},
			)
			if err != nil {
				return err
			}

			var changed string
			// Mac Settings profile is added
			if stored.GetMacSettingsProfileIds() == nil && st.Device.MacSettingsProfileIds != nil {
				changed = "inc"
			}

			// Mac Settings profile is deleted
			if stored.GetMacSettingsProfileIds() != nil && st.Device.MacSettingsProfileIds == nil {
				changed = "dec"

				st.Device.MacSettings = profile.MacSettings
				st.AddSetFields(macSettingsFields...)
			}

			_, err := ns.macSettingsProfiles.Set(
				ctx,
				id,
				[]string{"ids", "mac_settings", "end_devices_count"},
				func(_ context.Context, existing *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error) {
					switch changed {
					case "inc":
						existing.EndDevicesCount++
					case "dec":
						if existing.EndDevicesCount > 0 {
							existing.EndDevicesCount--
						}
					}
					return existing, []string{"ids", "mac_settings", "end_devices_count"}, nil
				})
			if err != nil {
				return err
			}
		}

		if stored == nil {
			evt = evtCreateEndDevice.NewWithIdentifiersAndData(ctx, st.Device.Ids, nil)
			return nil
		}

		evt = evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, st.Device.Ids, req.FieldMask.GetPaths())
		if st.HasSetField("multicast") && st.Device.Multicast != stored.Multicast {
			return newInvalidFieldValueError("multicast")
		}
		if st.HasSetField("supports_join") && st.Device.SupportsJoin != stored.SupportsJoin {
			return newInvalidFieldValueError("supports_join")
		}
		return nil
	}))
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to set device in registry")
		return nil, err
	}
	for _, f := range getTransforms {
		f(dev)
	}

	if evt != nil {
		events.Publish(evt)
	}

	if !needsDownlinkCheck {
		return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
	}

	if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after device set")
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

// ResetFactoryDefaults implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) ResetFactoryDefaults(ctx context.Context, req *ttnpb.ResetAndGetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIds, appendRequiredDeviceReadRights(
		append(make([]ttnpb.Right, 0, 1+maxRequiredDeviceReadRightCount), ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE),
		req.FieldMask.GetPaths()...,
	)...); err != nil {
		return nil, err
	}

	dev, _, err := ns.devices.SetByID(ctx, req.EndDeviceIds.ApplicationIds, req.EndDeviceIds.DeviceId, addDeviceGetPaths(ttnpb.AddFields(append(req.FieldMask.GetPaths()[:0:0], req.FieldMask.GetPaths()...),
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"multicast",
		"session.dev_addr",
		"session.keys",
		"session.queued_application_downlinks",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
	)...), func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if stored == nil {
			return nil, nil, errDeviceNotFound.New()
		}

		stored.BatteryPercentage = nil
		stored.DownlinkMargin = 0
		stored.LastDevStatusReceivedAt = nil
		stored.MacState = nil
		stored.PendingMacState = nil
		stored.PendingSession = nil
		stored.PowerState = ttnpb.PowerState_POWER_UNKNOWN
		if stored.SupportsJoin {
			stored.Session = nil
		} else {
			if stored.Session == nil {
				return nil, nil, ErrCorruptedMACState.
					WithCause(ErrSession)
			}

			fps, err := ns.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, nil, err
			}
			var profile *ttnpb.MACSettingsProfile
			if stored.MacSettingsProfileIds != nil {
				profile, err = ns.macSettingsProfiles.Get(ctx, stored.MacSettingsProfileIds, []string{"mac_settings"})
				if err != nil {
					return nil, nil, err
				}
			}
			macState, err := mac.NewState(stored, fps, ns.defaultMACSettings, profile.GetMacSettings())
			if err != nil {
				return nil, nil, err
			}
			stored.MacState = macState
			stored.Session = &ttnpb.Session{
				DevAddr:                    stored.Session.DevAddr,
				Keys:                       stored.Session.Keys,
				StartedAt:                  timestamppb.New(time.Now()), // NOTE: This is not equivalent to timestamppb.Now().
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
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyService(), dev, req.FieldMask.GetPaths()...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*emptypb.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, _, err := ns.devices.SetByID(ctx, req.ApplicationIds, req.DeviceId, nil, func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
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

type nsEndDeviceBatchRegistry struct {
	ttnpb.UnimplementedNsEndDeviceBatchRegistryServer

	devices             DeviceRegistry
	macSettingsProfiles MACSettingsProfileRegistry
	frequencyPlans      func(context.Context) (*frequencyplans.Store, error)
}

// Delete implements ttipb.NsEndDeviceBatchRegistryServer.
func (srv *nsEndDeviceBatchRegistry) Delete(
	ctx context.Context,
	req *ttnpb.BatchDeleteEndDevicesRequest,
) (*emptypb.Empty, error) {
	// Check if the user has rights on the application.
	if err := rights.RequireApplication(
		ctx,
		req.ApplicationIds,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	deleted, err := srv.devices.BatchDelete(ctx, req.ApplicationIds, req.DeviceIds)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete device from registry")
		return nil, err
	}

	if len(deleted) != 0 {
		events.Publish(
			evtBatchDeleteEndDevices.NewWithIdentifiersAndData(
				ctx, req.ApplicationIds, &ttnpb.EndDeviceIdentifiersList{
					EndDeviceIds: deleted,
				},
			),
		)
	}

	return ttnpb.Empty, nil
}

func (srv *nsEndDeviceBatchRegistry) SetMACSettingsProfile( // nolint: gocyclo
	ctx context.Context,
	req *ttnpb.BatchSetMACSettingsProfileRequest,
) (*ttnpb.EndDevices, error) {
	// Check if the user has rights on the application.
	if err := rights.RequireApplication(
		ctx,
		req.ApplicationIds,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}

	if req.MacSettingsProfileIds != nil {
		devices, err := srv.devices.BatchGetByID(ctx, req.ApplicationIds, req.DeviceIds, ttnpb.EndDeviceFieldPathsTopLevel)
		if err != nil {
			logRegistryRPCError(ctx, err, "Failed to get devices from registry")
			return nil, err
		}

		fps, err := srv.frequencyPlans(ctx)
		if err != nil {
			return nil, err
		}

		profile, err := srv.macSettingsProfiles.Get(
			ctx,
			req.MacSettingsProfileIds,
			[]string{"ids", "mac_settings", "end_devices_count"},
		)
		if err != nil {
			return nil, err
		}

		for _, dev := range devices {
			err = validateProfile(profile.GetMacSettings(), dev, fps)
			if err != nil {
				logRegistryRPCError(ctx, err, "Failed to validate mac settings profile")
				return nil, err
			}
		}
	}

	type change struct {
		changed string
		count   uint32
	}
	changes := make(map[*ttnpb.MACSettingsProfileIdentifiers]change, len(req.DeviceIds))

	storedDevices, err := srv.devices.BatchSetByID(
		ctx,
		req.ApplicationIds,
		req.DeviceIds,
		func(stored *ttnpb.EndDevice) error {
			if req.MacSettingsProfileIds != nil && stored.MacSettingsProfileIds == nil {
				if c, ok := changes[req.MacSettingsProfileIds]; !ok {
					changes[req.MacSettingsProfileIds] = change{"inc", 1}
				} else {
					c.count++
					changes[req.MacSettingsProfileIds] = c
				}
				stored.MacSettingsProfileIds = req.MacSettingsProfileIds
				stored.MacSettings = nil
			}
			if req.MacSettingsProfileIds != nil && stored.MacSettingsProfileIds != nil &&
				req.MacSettingsProfileIds != stored.MacSettingsProfileIds {
				if c, ok := changes[req.MacSettingsProfileIds]; !ok {
					changes[req.MacSettingsProfileIds] = change{"inc", 1}
				} else {
					c.count++
					changes[req.MacSettingsProfileIds] = c
				}
				if c, ok := changes[stored.MacSettingsProfileIds]; !ok {
					changes[stored.MacSettingsProfileIds] = change{"dec", 1}
				} else {
					c.count++
					changes[stored.MacSettingsProfileIds] = c
				}
				stored.MacSettingsProfileIds = req.MacSettingsProfileIds
				stored.MacSettings = nil
			}
			if req.MacSettingsProfileIds == nil && stored.MacSettingsProfileIds != nil {
				profile, err := srv.macSettingsProfiles.Get(
					ctx,
					stored.MacSettingsProfileIds,
					[]string{"ids", "mac_settings", "end_devices_count"},
				)
				if err != nil {
					return err
				}
				if c, ok := changes[stored.MacSettingsProfileIds]; !ok {
					changes[stored.MacSettingsProfileIds] = change{"dec", 1}
				} else {
					c.count++
					changes[stored.MacSettingsProfileIds] = c
				}
				stored.MacSettings = profile.MacSettings
				stored.MacSettingsProfileIds = nil
			}
			return nil
		})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to set devices in registry")
		return nil, err
	}

	for id, change := range changes {
		_, err = srv.macSettingsProfiles.Set(
			ctx,
			id,
			[]string{"ids", "mac_settings", "end_devices_count"},
			func(_ context.Context, existing *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error) {
				switch change.changed {
				case "inc":
					existing.EndDevicesCount += change.count
				case "dec":
					if existing.EndDevicesCount > 0 {
						existing.EndDevicesCount -= change.count
					}
				}
				return existing, []string{"ids", "mac_settings", "end_devices_count"}, nil
			})
		if err != nil {
			logRegistryRPCError(ctx, err, "Failed to set mac settings profile in registry")
			return nil, err
		}
	}

	return &ttnpb.EndDevices{
		EndDevices: storedDevices,
	}, nil
}

func init() {
	// The legacy and modern ADR fields should be mutually exclusive.
	// As such, specifying one of the fields means that every other field of the opposite
	// type should be zero.
	for _, field := range adrSettingsFields {
		ifNotZeroThenZeroFields[field] = append(ifNotZeroThenZeroFields[field], legacyADRSettingsFields...)
	}
	for _, field := range legacyADRSettingsFields {
		ifNotZeroThenZeroFields[field] = append(ifNotZeroThenNotZeroFields[field], "mac_settings.adr")
	}
}
