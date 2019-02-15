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
	return ns.devices.GetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, req.FieldMask.Paths)
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

	gets := append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...)
	if ttnpb.HasAnyField(req.FieldMask.Paths, "queued_application_downlinks") {
		gets = append(gets,
			"mac_state.device_class",
		)
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.device_class") {
		gets = append(gets,
			"mac_state.current_parameters",
			"mac_state.desired_parameters",
			"queued_application_downlinks",
		)
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.desired_parameters") {
		gets = append(gets,
			"mac_state.current_parameters",
			"mac_state.device_class",
		)
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.current_parameters") {
		gets = append(gets,
			"mac_state.desired_parameters",
			"mac_state.device_class",
		)
	}

	var addDownlinkTask bool
	dev, err := ns.devices.SetByID(ctx, req.Device.EndDeviceIdentifiers.ApplicationIdentifiers, req.Device.EndDeviceIdentifiers.DeviceID, gets, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if err := ttnpb.ProhibitFields(req.FieldMask.Paths,
			"mac_state",
			"pending_session",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}

		if dev != nil {
			if err := ttnpb.ProhibitFields(req.FieldMask.Paths,
				"session",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}

			addDownlinkTask = ttnpb.HasAnyField(req.FieldMask.Paths,
				"mac_state.current_parameters",
				"mac_state.desired_parameters",
				"mac_state.device_class",
				"queued_application_downlinks",
			) && req.Device.MACState.DeviceClass != ttnpb.CLASS_A &&
				(len(req.Device.QueuedApplicationDownlinks) > 0 || !req.Device.MACState.CurrentParameters.Equal(req.Device.MACState.DesiredParameters))
			return &req.Device, req.FieldMask.Paths, nil
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.dev_addr") && (req.Device.Session == nil || req.Device.Session.DevAddr.IsZero()) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.dev_addr")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.session_key_id") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetSessionKeyID()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.session_key_id")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.f_nwk_s_int_key.key") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetFNwkSIntKey().GetKey()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.f_nwk_s_int_key.key")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.f_nwk_s_int_key.kek_label") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetFNwkSIntKey().GetKEKLabel()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.f_nwk_s_int_key.kek_label")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.s_nwk_s_int_key.key") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetSNwkSIntKey().GetKey()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.s_nwk_s_int_key.key")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.s_nwk_s_int_key.kek_label") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetSNwkSIntKey().GetKEKLabel()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.s_nwk_s_int_key.kek_label")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.nwk_s_enc_key.key") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetNwkSEncKey().GetKey()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.nwk_s_enc_key.key")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.nwk_s_enc_key.kek_label") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetNwkSEncKey().GetKEKLabel()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.nwk_s_enc_key.kek_label")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.app_s_key.key") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetAppSKey().GetKey()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.app_s_key.key")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "session.keys.app_s_key.kek_label") && (req.Device.Session == nil || len(req.Device.Session.SessionKeys.GetAppSKey().GetKEKLabel()) == 0) {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.keys.app_s_key.kek_label")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.class_b_timeout") && req.Device.GetMACSettings().GetClassBTimeout() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.class_b_timeout")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.ping_slot_periodicity") && req.Device.GetMACSettings().GetPingSlotPeriodicity() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_periodicity")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.ping_slot_date_rate_index") && req.Device.GetMACSettings().GetPingSlotDataRateIndex() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_date_rate_index")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.ping_slot_frequency") && req.Device.GetMACSettings().GetPingSlotFrequency() == 0 {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_frequency")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.class_c_timeout") && req.Device.GetMACSettings().GetClassCTimeout() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.class_c_timeout")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx1_delay") && req.Device.GetMACSettings().GetRx1Delay() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx1_delay")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx1_data_rate_offset") && req.Device.GetMACSettings().GetRx1DataRateOffset() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx1_data_rate_offset")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx2_data_rate_index") && req.Device.GetMACSettings().GetRx2DataRateIndex() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx2_data_rate_index")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx2_frequency") && req.Device.GetMACSettings().GetRx2Frequency() == 0 {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx2_frequency")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.factory_preset_frequencies") && len(req.Device.GetMACSettings().GetFactoryPresetFrequencies()) == 0 {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.factory_preset_frequencies")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.max_duty_cycle") && req.Device.GetMACSettings().GetMaxDutyCycle() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.max_duty_cycle")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.supports_32_bit_f_cnt") && req.Device.GetMACSettings().GetSupports32BitFCnt() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.supports_32_bit_f_cnt")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.use_adr") && req.Device.GetMACSettings().GetUseADR() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.use_adr")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.adr_margin") && req.Device.GetMACSettings().GetADRMargin() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.adr_margin")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.resets_f_cnt") && req.Device.GetMACSettings().GetResetsFCnt() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.resets_f_cnt")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.status_time_periodicity") && req.Device.GetMACSettings().GetStatusTimePeriodicity() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.status_time_periodicity")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.status_count_periodicity") && req.Device.GetMACSettings().GetStatusCountPeriodicity() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.status_count_periodicity")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx1_delay") && req.Device.GetMACSettings().GetDesiredRx1Delay() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx1_delay")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx1_data_rate_offset") && req.Device.GetMACSettings().GetDesiredRx1DataRateOffset() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx1_data_rate_offset")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx2_data_rate_index") && req.Device.GetMACSettings().GetDesiredRx2DataRateIndex() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx2_data_rate_index")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx2_frequency") && req.Device.GetMACSettings().GetDesiredRx2Frequency() == 0 {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx2_frequency")
		}

		sets := append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...)
		if ttnpb.HasAnyField(sets, "version_ids") {
			// TODO: Apply version IDs (https://github.com/TheThingsIndustries/lorawan-stack/issues/1544)
		}

		if err := ttnpb.RequireFields(sets,
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
			"supports_join",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}
		if len(req.Device.FrequencyPlanID) == 0 {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "frequency_plan_id")
		}
		if err := req.Device.LoRaWANVersion.Validate(); err != nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "lorawan_version").WithCause(err)
		}
		if req.Device.LoRaWANPHYVersion == ttnpb.PHY_UNKNOWN {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "lorawan_phy_version")
		}

		if ttnpb.HasAnyField(sets, "supports_class_b") && req.Device.SupportsClassB {
			if err := ttnpb.RequireFields(sets,
				"mac_settings.ping_slot_date_rate_index",
				"mac_settings.ping_slot_frequency",
				"mac_settings.ping_slot_periodicity",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
		}

		if req.Device.SupportsJoin && !ttnpb.HasAnyField(sets, "session") {
			return &req.Device, sets, nil
		}

		if err := ttnpb.RequireFields(sets,
			"session.dev_addr",
			"session.keys.f_nwk_s_int_key.key",
		); err != nil {
			return nil, nil, errInvalidFieldMask.WithCause(err)
		}
		req.Device.EndDeviceIdentifiers.DevAddr = &req.Device.Session.DevAddr

		if !validABPSessionKey(req.Device.Session.FNwkSIntKey) {
			return nil, nil, errInvalidFNwkSIntKey
		}

		if req.Device.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			if err := ttnpb.RequireFields(sets,
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
			if err := ttnpb.ProhibitFields(sets,
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
			); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			// TODO: Encrypt (https://github.com/TheThingsIndustries/lorawan-stack/issues/1562)
			req.Device.Session.SNwkSIntKey = req.Device.Session.FNwkSIntKey
			req.Device.Session.NwkSEncKey = req.Device.Session.FNwkSIntKey
			sets = append(sets, "session.keys.s_nwk_s_int_key", "session.keys.nwk_s_enc_key")
		}

		if ttnpb.HasAnyField(sets, "session.started_at") && req.Device.GetSession().GetStartedAt().IsZero() {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "session.started_at")
		} else if !ttnpb.HasAnyField(sets, "session.started_at") {
			req.Device.Session.StartedAt = time.Now().UTC()
			sets = append(sets, "session.started_at")
		}

		if err := resetMACState(&req.Device, ns.FrequencyPlans, ns.defaultMACSettings); err != nil {
			return nil, nil, err
		}

		addDownlinkTask = req.Device.MACState.DeviceClass != ttnpb.CLASS_A &&
			(len(req.Device.QueuedApplicationDownlinks) > 0 || !req.Device.MACState.CurrentParameters.Equal(req.Device.MACState.DesiredParameters))

		return &req.Device, sets, nil
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
