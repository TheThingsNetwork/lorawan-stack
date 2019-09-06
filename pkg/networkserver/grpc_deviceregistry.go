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
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
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
	dev, err := ns.devices.GetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, req.FieldMask.Paths)
	if err != nil {
		return nil, err
	}
	for _, s := range []struct {
		val  *ttnpb.Session
		path string
	}{
		{
			val:  dev.Session,
			path: "session",
		},
		{
			val:  dev.PendingSession,
			path: "pending_session",
		},
	} {
		if s.val == nil {
			continue
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, s.path+".keys.f_nwk_s_int_key") && s.val.FNwkSIntKey != nil {
			key, err := cryptoutil.UnwrapAES128Key(*s.val.FNwkSIntKey, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			s.val.FNwkSIntKey = &ttnpb.KeyEnvelope{Key: &key}
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, s.path+".keys.s_nwk_s_int_key") && s.val.SNwkSIntKey != nil {
			key, err := cryptoutil.UnwrapAES128Key(*s.val.SNwkSIntKey, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			s.val.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: &key}
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, s.path+".keys.nwk_s_enc_key") && s.val.NwkSEncKey != nil {
			key, err := cryptoutil.UnwrapAES128Key(*s.val.NwkSEncKey, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			s.val.NwkSEncKey = &ttnpb.KeyEnvelope{Key: &key}
		}
	}
	return dev, nil
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

	gets := req.FieldMask.Paths
	var setsMACState bool
	for _, p := range req.FieldMask.Paths {
		if p == "mac_state" {
			setsMACState = true
			break
		}
		if strings.HasPrefix(p, "mac_state.") {
			setsMACState = true
			gets = append(gets,
				"mac_state",
				"queued_application_downlinks",
			)
			break
		}
	}

	var evt events.Event
	var addDownlinkTask bool
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
			addDownlinkTask = setsMACState &&
				(ttnpb.HasAnyField(req.FieldMask.Paths, "mac_state.device_class") && req.EndDevice.MACState.DeviceClass != ttnpb.CLASS_A ||
					dev.GetMACState().GetDeviceClass() != ttnpb.CLASS_A) &&
				(len(dev.QueuedApplicationDownlinks) > 0 ||
					!dev.MACState.CurrentParameters.Equal(dev.MACState.DesiredParameters))

			return &req.EndDevice, sets, nil
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.class_b_timeout") && req.EndDevice.GetMACSettings().GetClassBTimeout() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.class_b_timeout")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.ping_slot_periodicity.value") && req.EndDevice.GetMACSettings().GetPingSlotPeriodicity() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_periodicity")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.ping_slot_date_rate_index.value") && req.EndDevice.GetMACSettings().GetPingSlotDataRateIndex() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_date_rate_index")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.ping_slot_frequency.value") && req.EndDevice.GetMACSettings().GetPingSlotFrequency() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.ping_slot_frequency")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.class_c_timeout") && req.EndDevice.GetMACSettings().GetClassCTimeout() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.class_c_timeout")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx1_delay.value") && req.EndDevice.GetMACSettings().GetRx1Delay() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx1_delay")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx1_data_rate_offset.value") && req.EndDevice.GetMACSettings().GetRx1DataRateOffset() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx1_data_rate_offset")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx2_data_rate_index.value") && req.EndDevice.GetMACSettings().GetRx2DataRateIndex() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx2_data_rate_index")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.rx2_frequency.value") && req.EndDevice.GetMACSettings().GetRx2Frequency() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.rx2_frequency")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.factory_preset_frequencies") && len(req.EndDevice.GetMACSettings().GetFactoryPresetFrequencies()) == 0 {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.factory_preset_frequencies")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.max_duty_cycle.value") && req.EndDevice.GetMACSettings().GetMaxDutyCycle() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.max_duty_cycle")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.supports_32_bit_f_cnt.value") && req.EndDevice.GetMACSettings().GetSupports32BitFCnt() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.supports_32_bit_f_cnt")
		}

		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.use_adr.value") && req.EndDevice.GetMACSettings().GetUseADR() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.use_adr")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.adr_margin.value") && req.EndDevice.GetMACSettings().GetADRMargin() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.adr_margin")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.resets_f_cnt.value") && req.EndDevice.GetMACSettings().GetResetsFCnt() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.resets_f_cnt")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.status_time_periodicity") && req.EndDevice.GetMACSettings().GetStatusTimePeriodicity() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.status_time_periodicity")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.status_count_periodicity.value") && req.EndDevice.GetMACSettings().GetStatusCountPeriodicity() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.status_count_periodicity")
		}

		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx1_data_rate_offset.value") && req.EndDevice.GetMACSettings().GetDesiredRx1DataRateOffset() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx1_data_rate_offset")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx1_delay.value") && req.EndDevice.GetMACSettings().GetDesiredRx1Delay() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx1_delay")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx2_data_rate_index.value") && req.EndDevice.GetMACSettings().GetDesiredRx2DataRateIndex() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx2_data_rate_index")
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "mac_settings.desired_rx2_frequency.value") && req.EndDevice.GetMACSettings().GetDesiredRx2Frequency() == nil {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "mac_settings.desired_rx2_frequency")
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

		if req.EndDevice.Multicast && req.EndDevice.SupportsJoin {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "supports_join")
		}

		sets = append(sets,
			"ids.application_ids",
			"ids.device_id",
		)
		if req.EndDevice.JoinEUI != nil && !req.EndDevice.JoinEUI.IsZero() {
			sets = append(sets,
				"ids.join_eui",
			)
		}
		if req.EndDevice.DevEUI != nil && !req.EndDevice.DevEUI.IsZero() {
			sets = append(sets,
				"ids.dev_eui",
			)
		}
		if req.EndDevice.DevAddr != nil {
			if !ttnpb.HasAnyField(req.FieldMask.Paths, "session.dev_addr") || !req.EndDevice.DevAddr.Equal(req.EndDevice.Session.DevAddr) {
				return nil, nil, errInvalidFieldValue.WithAttributes("field", "ids.dev_addr")
			}
		}

		if req.EndDevice.SupportsJoin {
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
		sets = append(sets,
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
			req.EndDevice.Session.StartedAt = time.Now().UTC()
			sets = append(sets, "session.started_at")
		}

		macState, err := newMACState(&req.EndDevice, ns.FrequencyPlans, ns.defaultMACSettings)
		if err != nil {
			return nil, nil, err
		}
		req.EndDevice.MACState = macState
		sets = append(sets, "mac_state")

		return &req.EndDevice, sets, nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	if addDownlinkTask {
		startAt := time.Now().UTC()
		log.FromContext(ctx).WithField("start_at", startAt).Debug("Add downlink task")
		if err = ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, startAt, true); err != nil {
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
