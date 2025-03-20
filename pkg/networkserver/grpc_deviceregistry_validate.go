// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal" // nolint: revive, stylecheck
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/proto"
)

// ifThenFuncFieldRight represents the RHS of a functional implication.
type ifThenFuncFieldRight struct {
	Func   func(m map[string]*ttnpb.EndDevice) (bool, string)
	Fields []string
}

var (
	ifZeroThenZeroFields = map[string][]string{
		"supports_join": {
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
			"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
			"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
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
			"pending_mac_state.recent_downlinks",
			"pending_mac_state.recent_mac_command_identifiers",
			"pending_mac_state.recent_uplinks",
			"pending_mac_state.rejected_adr_data_rate_indexes",
			"pending_mac_state.rejected_adr_tx_power_indexes",
			"pending_mac_state.rejected_data_rate_ranges",
			"pending_mac_state.rejected_frequencies",
			"pending_mac_state.rx_windows_available",
			"pending_session.dev_addr",
			"pending_session.keys.f_nwk_s_int_key.key",
			"pending_session.keys.nwk_s_enc_key.key",
			"pending_session.keys.s_nwk_s_int_key.key",
			"pending_session.keys.session_key_id",
			"session.keys.session_key_id",
		},
	}

	ifZeroThenNotZeroFields = map[string][]string{
		"supports_join": {
			"session.dev_addr",
			"session.keys.f_nwk_s_int_key.key",
			// NOTE: LoRaWAN-version specific fields are validated within Set directly.
		},
	}

	ifNotZeroThenZeroFields = map[string][]string{
		"multicast": {
			"mac_settings.desired_relay.mode.served.backoff",
			"mac_settings.desired_relay.mode.served.mode.always",
			"mac_settings.desired_relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_settings.desired_relay.mode.served.mode.end_device_controlled",
			"mac_settings.desired_relay.mode.served.second_channel.ack_offset",
			"mac_settings.desired_relay.mode.served.second_channel.data_rate_index",
			"mac_settings.desired_relay.mode.served.second_channel.frequency",
			"mac_settings.desired_relay.mode.served.serving_device_id",
			"mac_settings.desired_relay.mode.serving.cad_periodicity",
			"mac_settings.desired_relay.mode.serving.default_channel_index",
			"mac_settings.desired_relay.mode.serving.limits.join_requests.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.join_requests.reload_rate",
			"mac_settings.desired_relay.mode.serving.limits.notifications.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.notifications.reload_rate",
			"mac_settings.desired_relay.mode.serving.limits.overall.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.overall.reload_rate",
			"mac_settings.desired_relay.mode.serving.limits.reset_behavior",
			"mac_settings.desired_relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_settings.desired_relay.mode.serving.second_channel.ack_offset",
			"mac_settings.desired_relay.mode.serving.second_channel.data_rate_index",
			"mac_settings.desired_relay.mode.serving.second_channel.frequency",
			"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules",
			"mac_settings.relay.mode.served.backoff",
			"mac_settings.relay.mode.served.mode.always",
			"mac_settings.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_settings.relay.mode.served.mode.end_device_controlled",
			"mac_settings.relay.mode.served.second_channel.ack_offset",
			"mac_settings.relay.mode.served.second_channel.data_rate_index",
			"mac_settings.relay.mode.served.second_channel.frequency",
			"mac_settings.relay.mode.served.serving_device_id",
			"mac_settings.relay.mode.serving.cad_periodicity",
			"mac_settings.relay.mode.serving.default_channel_index",
			"mac_settings.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_settings.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_settings.relay.mode.serving.limits.notifications.bucket_size",
			"mac_settings.relay.mode.serving.limits.notifications.reload_rate",
			"mac_settings.relay.mode.serving.limits.overall.bucket_size",
			"mac_settings.relay.mode.serving.limits.overall.reload_rate",
			"mac_settings.relay.mode.serving.limits.reset_behavior",
			"mac_settings.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_settings.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_settings.relay.mode.serving.second_channel.ack_offset",
			"mac_settings.relay.mode.serving.second_channel.data_rate_index",
			"mac_settings.relay.mode.serving.second_channel.frequency",
			"mac_settings.relay.mode.serving.uplink_forwarding_rules",
			"mac_settings.schedule_downlinks.value",
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
			"mac_state.last_adr_change_f_cnt_up",
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_dev_status_f_cnt_up",
			"mac_state.pending_application_downlink",
			"mac_state.pending_requests",
			"mac_state.queued_responses",
			"mac_state.recent_mac_command_identifiers",
			"mac_state.recent_uplinks",
			"mac_state.rejected_adr_data_rate_indexes",
			"mac_state.rejected_adr_tx_power_indexes",
			"mac_state.rejected_data_rate_ranges",
			"mac_state.rejected_frequencies",
			"mac_state.rx_windows_available",
			"session.last_conf_f_cnt_down",
			"session.last_f_cnt_up",
			"supports_join",
		},
	}

	ifNotZeroThenNotZeroFields = map[string][]string{
		"supports_join": {
			"ids.dev_eui",
			"ids.join_eui",
		},
	}

	ifZeroThenFuncFields = map[string][]ifThenFuncFieldRight{
		"supports_join": {
			{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if dev, ok := m["ids.dev_eui"]; ok && !types.MustEUI64(dev.Ids.DevEui).OrZero().IsZero() {
						return true, ""
					}
					if m["lorawan_version"].GetLorawanVersion() == ttnpb.MACVersion_MAC_UNKNOWN {
						return false, "lorawan_version"
					}
					if macspec.RequireDevEUIForABP(m["lorawan_version"].LorawanVersion) && !m["multicast"].GetMulticast() {
						return false, "ids.dev_eui"
					}
					return true, ""
				},
				Fields: []string{
					"ids.dev_eui",
					"lorawan_version",
					"multicast",
				},
			},

			{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if !m["supports_class_b"].GetSupportsClassB() ||
						m["mac_settings.ping_slot_periodicity.value"].GetMacSettings().GetPingSlotPeriodicity() != nil {
						return true, ""
					}
					return false, "mac_settings.ping_slot_periodicity.value"
				},
				Fields: []string{
					"mac_settings.ping_slot_periodicity.value",
					"supports_class_b",
				},
			},
		},
	}

	ifNotZeroThenFuncFields = map[string][]ifThenFuncFieldRight{
		"multicast": append(func() (rs []ifThenFuncFieldRight) {
			for s, eq := range map[string]func(*ttnpb.MACParameters, *ttnpb.MACParameters) bool{
				"adr_ack_delay_exponent.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.AdrAckDelayExponent, b.AdrAckDelayExponent)
				},
				"adr_ack_limit_exponent.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.AdrAckLimitExponent, b.AdrAckLimitExponent)
				},
				"adr_data_rate_index": func(a, b *ttnpb.MACParameters) bool {
					return a.AdrDataRateIndex == b.AdrDataRateIndex
				},
				"adr_nb_trans": func(a, b *ttnpb.MACParameters) bool {
					return a.AdrNbTrans == b.AdrNbTrans
				},
				"adr_tx_power_index": func(a, b *ttnpb.MACParameters) bool {
					return a.AdrTxPowerIndex == b.AdrTxPowerIndex
				},
				"beacon_frequency": func(a, b *ttnpb.MACParameters) bool {
					return a.BeaconFrequency == b.BeaconFrequency
				},
				"channels": func(a, b *ttnpb.MACParameters) bool {
					if len(a.Channels) != len(b.Channels) {
						return false
					}
					for i, ch := range a.Channels {
						if !proto.Equal(ch, b.Channels[i]) {
							return false
						}
					}
					return true
				},
				"downlink_dwell_time.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.DownlinkDwellTime, b.DownlinkDwellTime)
				},
				"max_duty_cycle": func(a, b *ttnpb.MACParameters) bool {
					return a.MaxDutyCycle == b.MaxDutyCycle
				},
				"max_eirp": func(a, b *ttnpb.MACParameters) bool {
					return a.MaxEirp == b.MaxEirp
				},
				"ping_slot_data_rate_index_value.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.PingSlotDataRateIndexValue, b.PingSlotDataRateIndexValue)
				},
				"ping_slot_frequency": func(a, b *ttnpb.MACParameters) bool {
					return a.PingSlotFrequency == b.PingSlotFrequency
				},
				"rejoin_count_periodicity": func(a, b *ttnpb.MACParameters) bool {
					return a.RejoinCountPeriodicity == b.RejoinCountPeriodicity
				},
				"rejoin_time_periodicity": func(a, b *ttnpb.MACParameters) bool {
					return a.RejoinTimePeriodicity == b.RejoinTimePeriodicity
				},
				"rx1_data_rate_offset": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx1DataRateOffset == b.Rx1DataRateOffset
				},
				"rx1_delay": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx1Delay == b.Rx1Delay
				},
				"rx2_data_rate_index": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx2DataRateIndex == b.Rx2DataRateIndex
				},
				"rx2_frequency": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx2Frequency == b.Rx2Frequency
				},
				"uplink_dwell_time.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.UplinkDwellTime, b.UplinkDwellTime)
				},
			} {
				curPath := "mac_state.current_parameters." + s
				desPath := "mac_state.desired_parameters." + s
				rs = append(rs, ifThenFuncFieldRight{
					Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
						curDev := m[curPath]
						desDev := m[desPath]
						if curDev == nil || desDev == nil {
							if curDev != desDev {
								return false, desPath
							}
							return true, ""
						}
						if !eq(curDev.MacState.CurrentParameters, desDev.MacState.DesiredParameters) {
							return false, desPath
						}
						return true, ""
					},
					Fields: []string{
						curPath,
						desPath,
					},
				})
			}
			return rs
		}(),

			ifThenFuncFieldRight{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if !m["supports_class_b"].GetSupportsClassB() && !m["supports_class_c"].GetSupportsClassC() {
						return false, "supports_class_b"
					}
					return true, ""
				},
				Fields: []string{
					"supports_class_b",
					"supports_class_c",
				},
			},

			ifThenFuncFieldRight{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if !m["supports_class_b"].GetSupportsClassB() ||
						m["mac_settings.ping_slot_periodicity.value"].GetMacSettings().GetPingSlotPeriodicity() != nil {
						return true, ""
					}
					return false, "mac_settings.ping_slot_periodicity.value"
				},
				Fields: []string{
					"mac_settings.ping_slot_periodicity.value",
					"supports_class_b",
				},
			},
		),
	}

	// The downlinkInfluencingSetFields contains fields that can influence downlink scheduling,
	// e.g. trigger one or make a scheduled slot obsolete.
	downlinkInfluencingSetFields = [...]string{
		"last_dev_status_received_at",
		"mac_settings.schedule_downlinks.value",
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
	}

	legacyADRSettingsFields = []string{
		"mac_settings.adr_margin",
		"mac_settings.use_adr.value",
		"mac_settings.use_adr",
	}

	adrSettingsFields = []string{
		"mac_settings.adr.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow",
		"mac_settings.adr.mode.dynamic.channel_steering.mode",
		"mac_settings.adr.mode.dynamic.channel_steering",
		"mac_settings.adr.mode.dynamic.margin",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_data_rate_index",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_data_rate_index",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9",
		"mac_settings.adr.mode.dynamic.overrides",
		"mac_settings.adr.mode.dynamic",
		"mac_settings.adr.mode.static.data_rate_index",
		"mac_settings.adr.mode.static.nb_trans",
		"mac_settings.adr.mode.static.tx_power_index",
		"mac_settings.adr.mode.static",
		"mac_settings.adr.mode",
		"mac_settings.adr",
	}

	dynamicADRSettingsFields = []string{
		"mac_settings.adr.mode.dynamic.channel_steering.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow",
		"mac_settings.adr.mode.dynamic.channel_steering.mode",
		"mac_settings.adr.mode.dynamic.channel_steering",
		"mac_settings.adr.mode.dynamic.margin",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.min_nb_trans",
		"mac_settings.adr.mode.dynamic",
	}

	macSettingsFields = []string{
		"mac_settings",
		"mac_settings.adr",
		"mac_settings.adr.mode",
		"mac_settings.adr.mode.disabled",
		"mac_settings.adr.mode.dynamic",
		"mac_settings.adr.mode.dynamic.channel_steering",
		"mac_settings.adr.mode.dynamic.channel_steering.mode",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow",
		"mac_settings.adr.mode.dynamic.margin",
		"mac_settings.adr.mode.dynamic.max_data_rate_index",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic.overrides",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.min_nb_trans",
		"mac_settings.adr.mode.static",
		"mac_settings.adr.mode.static.data_rate_index",
		"mac_settings.adr.mode.static.nb_trans",
		"mac_settings.adr.mode.static.tx_power_index",
		"mac_settings.adr_margin",
		"mac_settings.beacon_frequency",
		"mac_settings.beacon_frequency.value",
		"mac_settings.class_b_c_downlink_interval",
		"mac_settings.class_b_timeout",
		"mac_settings.class_c_timeout",
		"mac_settings.desired_adr_ack_delay_exponent",
		"mac_settings.desired_adr_ack_delay_exponent.value",
		"mac_settings.desired_adr_ack_limit_exponent",
		"mac_settings.desired_adr_ack_limit_exponent.value",
		"mac_settings.desired_beacon_frequency",
		"mac_settings.desired_beacon_frequency.value",
		"mac_settings.desired_max_duty_cycle",
		"mac_settings.desired_max_duty_cycle.value",
		"mac_settings.desired_max_eirp",
		"mac_settings.desired_max_eirp.value",
		"mac_settings.desired_ping_slot_data_rate_index",
		"mac_settings.desired_ping_slot_data_rate_index.value",
		"mac_settings.desired_ping_slot_frequency",
		"mac_settings.desired_ping_slot_frequency.value",
		"mac_settings.desired_relay",
		"mac_settings.desired_relay.mode",
		"mac_settings.desired_relay.mode.served",
		"mac_settings.desired_relay.mode.served.backoff",
		"mac_settings.desired_relay.mode.served.mode",
		"mac_settings.desired_relay.mode.served.mode.always",
		"mac_settings.desired_relay.mode.served.mode.dynamic",
		"mac_settings.desired_relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_settings.desired_relay.mode.served.mode.end_device_controlled",
		"mac_settings.desired_relay.mode.served.second_channel",
		"mac_settings.desired_relay.mode.served.second_channel.ack_offset",
		"mac_settings.desired_relay.mode.served.second_channel.data_rate_index",
		"mac_settings.desired_relay.mode.served.second_channel.frequency",
		"mac_settings.desired_relay.mode.served.serving_device_id",
		"mac_settings.desired_relay.mode.serving",
		"mac_settings.desired_relay.mode.serving.cad_periodicity",
		"mac_settings.desired_relay.mode.serving.default_channel_index",
		"mac_settings.desired_relay.mode.serving.limits",
		"mac_settings.desired_relay.mode.serving.limits.join_requests",
		"mac_settings.desired_relay.mode.serving.limits.join_requests.bucket_size",
		"mac_settings.desired_relay.mode.serving.limits.join_requests.reload_rate",
		"mac_settings.desired_relay.mode.serving.limits.notifications",
		"mac_settings.desired_relay.mode.serving.limits.notifications.bucket_size",
		"mac_settings.desired_relay.mode.serving.limits.notifications.reload_rate",
		"mac_settings.desired_relay.mode.serving.limits.overall",
		"mac_settings.desired_relay.mode.serving.limits.overall.bucket_size",
		"mac_settings.desired_relay.mode.serving.limits.overall.reload_rate",
		"mac_settings.desired_relay.mode.serving.limits.reset_behavior",
		"mac_settings.desired_relay.mode.serving.limits.uplink_messages",
		"mac_settings.desired_relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_settings.desired_relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_settings.desired_relay.mode.serving.second_channel",
		"mac_settings.desired_relay.mode.serving.second_channel.ack_offset",
		"mac_settings.desired_relay.mode.serving.second_channel.data_rate_index",
		"mac_settings.desired_relay.mode.serving.second_channel.frequency",
		"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules",
		"mac_settings.desired_rx1_data_rate_offset",
		"mac_settings.desired_rx1_data_rate_offset.value",
		"mac_settings.desired_rx1_delay",
		"mac_settings.desired_rx1_delay.value",
		"mac_settings.desired_rx2_data_rate_index",
		"mac_settings.desired_rx2_data_rate_index.value",
		"mac_settings.desired_rx2_frequency",
		"mac_settings.desired_rx2_frequency.value",
		"mac_settings.downlink_dwell_time",
		"mac_settings.downlink_dwell_time.value",
		"mac_settings.factory_preset_frequencies",
		"mac_settings.max_duty_cycle",
		"mac_settings.max_duty_cycle.value",
		"mac_settings.ping_slot_data_rate_index",
		"mac_settings.ping_slot_data_rate_index.value",
		"mac_settings.ping_slot_frequency",
		"mac_settings.ping_slot_frequency.value",
		"mac_settings.ping_slot_periodicity",
		"mac_settings.ping_slot_periodicity.value",
		"mac_settings.relay",
		"mac_settings.relay.mode",
		"mac_settings.relay.mode.served",
		"mac_settings.relay.mode.served.backoff",
		"mac_settings.relay.mode.served.mode",
		"mac_settings.relay.mode.served.mode.always",
		"mac_settings.relay.mode.served.mode.dynamic",
		"mac_settings.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_settings.relay.mode.served.mode.end_device_controlled",
		"mac_settings.relay.mode.served.second_channel",
		"mac_settings.relay.mode.served.second_channel.ack_offset",
		"mac_settings.relay.mode.served.second_channel.data_rate_index",
		"mac_settings.relay.mode.served.second_channel.frequency",
		"mac_settings.relay.mode.served.serving_device_id",
		"mac_settings.relay.mode.serving",
		"mac_settings.relay.mode.serving.cad_periodicity",
		"mac_settings.relay.mode.serving.default_channel_index",
		"mac_settings.relay.mode.serving.limits",
		"mac_settings.relay.mode.serving.limits.join_requests",
		"mac_settings.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_settings.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_settings.relay.mode.serving.limits.notifications",
		"mac_settings.relay.mode.serving.limits.notifications.bucket_size",
		"mac_settings.relay.mode.serving.limits.notifications.reload_rate",
		"mac_settings.relay.mode.serving.limits.overall",
		"mac_settings.relay.mode.serving.limits.overall.bucket_size",
		"mac_settings.relay.mode.serving.limits.overall.reload_rate",
		"mac_settings.relay.mode.serving.limits.reset_behavior",
		"mac_settings.relay.mode.serving.limits.uplink_messages",
		"mac_settings.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_settings.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_settings.relay.mode.serving.second_channel",
		"mac_settings.relay.mode.serving.second_channel.ack_offset",
		"mac_settings.relay.mode.serving.second_channel.data_rate_index",
		"mac_settings.relay.mode.serving.second_channel.frequency",
		"mac_settings.relay.mode.serving.uplink_forwarding_rules",
		"mac_settings.resets_f_cnt",
		"mac_settings.resets_f_cnt.value",
		"mac_settings.rx1_data_rate_offset",
		"mac_settings.rx1_data_rate_offset.value",
		"mac_settings.rx1_delay",
		"mac_settings.rx1_delay.value",
		"mac_settings.rx2_data_rate_index",
		"mac_settings.rx2_data_rate_index.value",
		"mac_settings.rx2_frequency",
		"mac_settings.rx2_frequency.value",
		"mac_settings.schedule_downlinks",
		"mac_settings.schedule_downlinks.value",
		"mac_settings.status_count_periodicity",
		"mac_settings.status_time_periodicity",
		"mac_settings.supports_32_bit_f_cnt",
		"mac_settings.supports_32_bit_f_cnt.value",
		"mac_settings.uplink_dwell_time",
		"mac_settings.uplink_dwell_time.value",
		"mac_settings.use_adr",
		"mac_settings.use_adr.value",
	}
)

// Ensure ids.dev_addr and session.dev_addr are consistent.
func validateADR(st *setDeviceState) error {
	if st.HasSetField("ids.dev_addr") {
		if err := st.ValidateField(func(dev *ttnpb.EndDevice) bool {
			if st.Device.Ids.DevAddr == nil {
				return dev.GetSession() == nil
			}
			return dev.GetSession() != nil && bytes.Equal(dev.Session.DevAddr, st.Device.Ids.DevAddr)
		}, "session.dev_addr"); err != nil {
			return err
		}
	} else if st.HasSetField("session.dev_addr") {
		st.Device.Ids.DevAddr = nil
		if devAddr := types.MustDevAddr(st.Device.GetSession().GetDevAddr()); devAddr != nil {
			st.Device.Ids.DevAddr = devAddr.Bytes()
		}
		st.AddSetFields(
			"ids.dev_addr",
		)
	}
	return nil
}

func validateZeroFields(st *setDeviceState) error { // nolint: gocyclo
	// Ensure FieldIsZero(left) -> FieldIsZero(r), for each r in right.
	for left, right := range ifZeroThenZeroFields {
		if st.HasSetField(left) {
			if !st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreZero(right...); err != nil {
				return err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsNotZero(left); err != nil {
				return err
			}
		}
	}

	// Ensure FieldIsZero(left) -> !FieldIsZero(r), for each r in right.
	for left, right := range ifZeroThenNotZeroFields {
		if st.HasSetField(left) {
			if !st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreNotZero(right...); err != nil {
				return err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || !st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsNotZero(left); err != nil {
				return err
			}
		}
	}

	// Ensure FieldIsZero(left) -> r.Func(map rr -> *ttnpb.EndDevice), for each rr in r.Fields for each r in rs.
	for left, rs := range ifZeroThenFuncFields {
		for _, r := range rs {
			if st.HasSetField(left) {
				if !st.Device.FieldIsZero(left) {
					continue
				}
				if err := st.ValidateFields(r.Func, r.Fields...); err != nil {
					return err
				}
			}
			if !st.HasSetField(r.Fields...) {
				continue
			}

			if err := st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
				if !m[left].FieldIsZero(left) {
					return true, ""
				}
				return r.Func(m)
			}, append([]string{left}, r.Fields...)...); err != nil {
				return err
			}
		}
	}

	// Ensure !FieldIsZero(left) -> FieldIsZero(r), for each r in right.
	for left, right := range ifNotZeroThenZeroFields {
		if st.HasSetField(left) {
			if st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreZero(right...); err != nil {
				return err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsZero(left); err != nil {
				return err
			}
		}
	}

	// Ensure !FieldIsZero(left) -> !FieldIsZero(r), for each r in right.
	for left, right := range ifNotZeroThenNotZeroFields {
		if st.HasSetField(left) {
			if st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreNotZero(right...); err != nil {
				return err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || !st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsZero(left); err != nil {
				return err
			}
		}
	}

	// Ensure !FieldIsZero(left) -> r.Func(map rr -> *ttnpb.EndDevice), for each rr in r.Fields for each r in rs.
	for left, rs := range ifNotZeroThenFuncFields {
		for _, r := range rs {
			if st.HasSetField(left) {
				if st.Device.FieldIsZero(left) {
					continue
				}
				if err := st.ValidateFields(r.Func, r.Fields...); err != nil {
					return err
				}
			}
			if !st.HasSetField(r.Fields...) {
				continue
			}

			if err := st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
				if m[left].FieldIsZero(left) {
					return true, ""
				}
				return r.Func(m)
			}, append([]string{left}, r.Fields...)...); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateProfile( // nolint: gocyclo
	profile *ttnpb.MACSettings,
	dev *ttnpb.EndDevice,
	fps *frequencyplans.Store,
) error {
	fp, phy, err := DeviceFrequencyPlanAndBand(dev, fps)
	if err != nil {
		return err
	}
	if profile.GetRx2DataRateIndex() != nil {
		_, ok := phy.DataRates[profile.Rx2DataRateIndex.Value]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.rx2_data_rate_index.value")
		}
	}
	if profile.GetDesiredRx2DataRateIndex() != nil {
		_, ok := phy.DataRates[profile.DesiredRx2DataRateIndex.Value]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_rx2_data_rate_index.value")
		}
	}
	if profile.GetDesiredPingSlotDataRateIndex() != nil {
		_, ok := phy.DataRates[profile.DesiredPingSlotDataRateIndex.Value]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_ping_slot_data_rate_index.value")
		}
	}
	if profile.GetDesiredRelay().GetServed().GetSecondChannel() != nil {
		_, ok := phy.DataRates[profile.DesiredRelay.GetServed().SecondChannel.DataRateIndex]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_relay.served.second_channel.data_rate_index") // nolint: lll
		}
	}
	if chIdx := profile.GetDesiredRelay().GetServing().GetDefaultChannelIndex(); chIdx != nil {
		if chIdx.Value >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_relay.serving.default_channel_index")
		}
	}
	if profile.GetDesiredRelay().GetServing().GetSecondChannel() != nil {
		_, ok := phy.DataRates[profile.DesiredRelay.GetServing().SecondChannel.DataRateIndex]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_relay.serving.second_channel.data_rate_index") // nolint: lll
		}
	}
	if profile.GetAdr().GetDynamic().GetMaxDataRateIndex() != nil {
		drIdx := profile.Adr.GetDynamic().MaxDataRateIndex.Value
		_, ok := phy.DataRates[drIdx]
		if !ok || drIdx > phy.MaxADRDataRateIndex {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.dynamic.max_data_rate_index")
		}
	}
	if profile.GetAdr().GetDynamic().GetMinDataRateIndex() != nil {
		drIdx := profile.Adr.GetDynamic().MinDataRateIndex.Value
		_, ok := phy.DataRates[drIdx]
		if !ok || drIdx > phy.MaxADRDataRateIndex {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.dynamic.min_data_rate_index")
		}
	}
	if profile.GetAdr().GetDynamic().GetMaxTxPowerIndex() != nil {
		if profile.Adr.GetDynamic().MaxTxPowerIndex.Value > uint32(phy.MaxTxPowerIndex()) {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.dynamic.max_tx_power_index")
		}
	}
	if profile.GetAdr().GetDynamic().GetMinTxPowerIndex() != nil {
		if profile.Adr.GetDynamic().MinTxPowerIndex.Value > uint32(phy.MaxTxPowerIndex()) {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.dynamic.min_tx_power_index")
		}
	}
	if !phy.SupportsDynamicADR {
		if profile.GetAdr().GetDynamic() != nil {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.dynamic")
		}
	} else {
		if profile.GetAdr().GetDynamic() != nil {
			minDRI := profile.GetAdr().GetDynamic().GetMinDataRateIndex()
			maxDRI := profile.GetAdr().GetDynamic().GetMaxDataRateIndex()
			if minDRI != nil && maxDRI != nil && maxDRI.Value < minDRI.Value {
				return newInvalidFieldValueError("mac_settings.adr.mode.dynamic.max_data_rate_index.value")
			}

			minNbTrans := profile.GetAdr().GetDynamic().GetMinNbTrans()
			maxNbTrans := profile.GetAdr().GetDynamic().GetMaxNbTrans()
			if minNbTrans != nil && maxNbTrans != nil && maxNbTrans.Value < minNbTrans.Value {
				return newInvalidFieldValueError("mac_settings.adr.mode.dynamic.max_nb_trans")
			}

			for drIdx := ttnpb.DataRateIndex_DATA_RATE_0; drIdx <= ttnpb.DataRateIndex_DATA_RATE_15; drIdx++ {
				minOverrides := mac.DataRateIndexOverridesOf(profile.GetAdr().GetDynamic().GetOverrides(), drIdx).GetMinNbTrans() // nolint: lll
				maxOverrides := mac.DataRateIndexOverridesOf(profile.GetAdr().GetDynamic().GetOverrides(), drIdx).GetMaxNbTrans() // nolint: lll
				if minOverrides != nil && maxOverrides != nil && maxOverrides.Value < minOverrides.Value {
					return newInvalidFieldValueError(fmt.Sprintf("mac_settings.adr.mode.dynamic.overrides.data_rate_%d.max_nb_trans", drIdx)) // nolint: lll
				}
			}
		}
	}
	if profile.GetAdr().GetStatic() != nil {
		_, ok := phy.DataRates[profile.Adr.GetStatic().DataRateIndex]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.static.data_rate_index")
		}
		if profile.Adr.GetStatic().TxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.adr.mode.static.tx_power_index")
		}
	}
	if profile.GetUplinkDwellTime() != nil {
		if !phy.TxParamSetupReqSupport {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.uplink_dwell_time.value")
		}
	}
	if profile.GetDownlinkDwellTime() != nil {
		if !phy.TxParamSetupReqSupport {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.downlink_dwell_time.value")
		}
	}
	if profile.GetRelay().GetServed().GetSecondChannel() != nil {
		_, ok := phy.DataRates[profile.Relay.GetServed().SecondChannel.DataRateIndex]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.relay.served.second_channel.data_rate_index")
		}
	}
	if chIdx := profile.GetRelay().GetServing().GetDefaultChannelIndex(); chIdx != nil {
		if chIdx.Value >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.relay.serving.default_channel_index")
		}
	}
	if profile.GetRelay().GetServing().GetSecondChannel() != nil {
		_, ok := phy.DataRates[profile.Relay.GetServing().SecondChannel.DataRateIndex]
		if !ok {
			return newInvalidFieldValueError("mac_settings_profile.mac_settings.relay.serving.second_channel.data_rate_index")
		}
	}
	if len(profile.GetFactoryPresetFrequencies()) > 0 {
		switch phy.CFListType {
		case ttnpb.CFListType_FREQUENCIES:
			// Factory preset frequencies in bands which provide frequencies as part of the CFList
			// are interpreted as being used both for uplinks and downlinks.
			for _, frequency := range profile.FactoryPresetFrequencies {
				_, inSubBand := fp.FindSubBand(frequency)
				for _, sb := range phy.SubBands {
					if sb.MinFrequency <= frequency && frequency <= sb.MaxFrequency {
						inSubBand = true
						break
					}
				}
				if !inSubBand {
					return newInvalidFieldValueError("mac_settings_profile.mac_settings.factory_preset_frequencies")
				}
			}
		case ttnpb.CFListType_CHANNEL_MASKS:
			// Factory preset frequencies in bands which provide channel masks as part of the CFList
			// are interpreted as enabling explicit uplink channels.
			uplinkChannels := make(map[uint64]struct{}, len(phy.UplinkChannels))
			for _, ch := range phy.UplinkChannels {
				uplinkChannels[ch.Frequency] = struct{}{}
			}
			for _, frequency := range profile.FactoryPresetFrequencies {
				if _, ok := uplinkChannels[frequency]; !ok {
					return newInvalidFieldValueError("mac_settings_profile.mac_settings.factory_preset_frequencies")
				}
			}
		default:
			panic("unreachable")
		}
	}
	if dev.GetSupportsClassB() {
		if profile.GetPingSlotFrequency().GetValue() == 0 {
			if len(phy.PingSlotFrequencies) == 0 {
				return newInvalidFieldValueError("mac_settings_profile.mac_settings.ping_slot_frequency.value")
			}
		}
		if profile.GetDesiredPingSlotFrequency().GetValue() == 0 {
			if len(phy.PingSlotFrequencies) == 0 {
				return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_ping_slot_frequency.value")
			}
		}
		if profile.GetBeaconFrequency().GetValue() == 0 {
			if len(phy.Beacon.Frequencies) == 0 {
				return newInvalidFieldValueError("mac_settings_profile.mac_settings.beacon_frequency.value")
			}
		}
		if profile.GetDesiredBeaconFrequency().GetValue() == 0 {
			if len(phy.Beacon.Frequencies) == 0 {
				return newInvalidFieldValueError("mac_settings_profile.mac_settings.desired_beacon_frequency.value")
			}
		}
	}
	return nil
}

func validateBandSpecifications(st *setDeviceState, fps *frequencyplans.Store) error { // nolint: gocyclo
	var deferredPHYValidations []func(*band.Band, *frequencyplans.FrequencyPlan) error
	withPHY := func(f func(*band.Band, *frequencyplans.FrequencyPlan) error) error { // nolint: unparam
		deferredPHYValidations = append(deferredPHYValidations, f)
		return nil
	}
	if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
		fp, phy, err := DeviceFrequencyPlanAndBand(&ttnpb.EndDevice{
			FrequencyPlanId:   m["frequency_plan_id"].GetFrequencyPlanId(),
			LorawanPhyVersion: m["lorawan_phy_version"].GetLorawanPhyVersion(),
		}, fps)
		if err != nil {
			return err
		}
		withPHY = func(f func(*band.Band, *frequencyplans.FrequencyPlan) error) error {
			return f(phy, fp)
		}
		for _, f := range deferredPHYValidations {
			if err := f(phy, fp); err != nil {
				return err
			}
		}
		return nil
	},
		"frequency_plan_id",
		"lorawan_phy_version",
	); err != nil {
		return err
	}

	hasPHYUpdate := st.HasSetField(
		"frequency_plan_id",
		"lorawan_phy_version",
	)
	hasSetField := func(field string) (fieldToRetrieve string, validate bool) {
		return field, st.HasSetField(field) || hasPHYUpdate
	}

	setFields := func(fields ...string) []string {
		setFields := make([]string, 0, len(fields))
		for _, field := range fields {
			if st.HasSetField(field) {
				setFields = append(setFields, field)
			}
		}
		return setFields
	}

	if st.HasSetField(
		"frequency_plan_id",
		"version_ids.band_id",
	) {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(_ *band.Band, fp *frequencyplans.FrequencyPlan) error {
				if devBandID := dev.GetVersionIds().GetBandId(); devBandID != "" && devBandID != fp.BandID {
					return newInvalidFieldValueError("version_ids.band_id").WithCause(
						errDeviceAndFrequencyPlanBandMismatch.WithAttributes(
							"dev_band_id", devBandID,
							"fp_band_id", fp.BandID,
						),
					)
				}
				return nil
			})
		}, "version_ids.band_id"); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.rx2_data_rate_index.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetRx2DataRateIndex() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.Rx2DataRateIndex.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.desired_rx2_data_rate_index.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetDesiredRx2DataRateIndex() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.DesiredRx2DataRateIndex.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.ping_slot_data_rate_index.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetPingSlotDataRateIndex() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.PingSlotDataRateIndex.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.desired_ping_slot_data_rate_index.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetDesiredPingSlotDataRateIndex() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.DesiredPingSlotDataRateIndex.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.desired_relay.mode.served.second_channel.data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetDesiredRelay().GetServed().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.DesiredRelay.GetServed().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.desired_relay.mode.serving.default_channel_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				chIdx := dev.GetMacSettings().GetDesiredRelay().GetServing().GetDefaultChannelIndex()
				if chIdx == nil {
					return nil
				}
				if chIdx.Value >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.desired_relay.mode.serving.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetDesiredRelay().GetServing().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.DesiredRelay.GetServing().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.adr.mode.dynamic.max_data_rate_index.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetAdr().GetDynamic().GetMaxDataRateIndex() == nil {
					return nil
				}
				drIdx := dev.MacSettings.Adr.GetDynamic().MaxDataRateIndex.Value
				_, ok := phy.DataRates[drIdx]
				if !ok || drIdx > phy.MaxADRDataRateIndex {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.adr.mode.dynamic.min_data_rate_index.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetAdr().GetDynamic().GetMinDataRateIndex() == nil {
					return nil
				}
				drIdx := dev.MacSettings.Adr.GetDynamic().MinDataRateIndex.Value
				_, ok := phy.DataRates[drIdx]
				if !ok || drIdx > phy.MaxADRDataRateIndex {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.adr.mode.dynamic.max_tx_power_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetAdr().GetDynamic().GetMaxTxPowerIndex() == nil {
					return nil
				}
				if dev.MacSettings.Adr.GetDynamic().MaxTxPowerIndex.Value > uint32(phy.MaxTxPowerIndex()) {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.adr.mode.dynamic.min_tx_power_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetAdr().GetDynamic().GetMinTxPowerIndex() == nil {
					return nil
				}
				if dev.MacSettings.Adr.GetDynamic().MinTxPowerIndex.Value > uint32(phy.MaxTxPowerIndex()) {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if setFields := setFields(dynamicADRSettingsFields...); hasPHYUpdate || len(setFields) > 0 {
		fields := setFields
		if hasPHYUpdate {
			fields = append(fields, "mac_settings.adr.mode")
		}
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if phy.SupportsDynamicADR {
					return nil
				}
				for _, field := range fields {
					if m[field].GetMacSettings().GetAdr().GetDynamic() != nil {
						return newInvalidFieldValueError(field)
					}
				}
				return nil
			})
		},
			fields...,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.adr.mode.static.data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetAdr().GetStatic() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.Adr.GetStatic().DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.adr.mode.static.tx_power_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetAdr().GetStatic() == nil {
					return nil
				}
				if dev.MacSettings.Adr.GetStatic().TxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.uplink_dwell_time.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetUplinkDwellTime() == nil {
					return nil
				}
				if !phy.TxParamSetupReqSupport {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.downlink_dwell_time.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetDownlinkDwellTime() == nil {
					return nil
				}
				if !phy.TxParamSetupReqSupport {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.relay.mode.served.second_channel.data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetRelay().GetServed().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.Relay.GetServed().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.relay.mode.serving.default_channel_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				chIdx := dev.GetMacSettings().GetRelay().GetServing().GetDefaultChannelIndex()
				if chIdx == nil {
					return nil
				}
				if chIdx.Value >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_settings.relay.mode.serving.second_channel.data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacSettings().GetRelay().GetServing().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacSettings.Relay.GetServing().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.current_parameters.rx2_data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetMacState() == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.MacState.CurrentParameters.Rx2DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.desired_parameters.rx2_data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetMacState() == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.MacState.DesiredParameters.Rx2DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServed().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Relay.GetServed().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.current_parameters.relay.mode.serving.default_channel_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServing() == nil {
					return nil
				}
				chIdx := dev.PendingMacState.CurrentParameters.Relay.GetServing().DefaultChannelIndex
				if chIdx >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServing().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Relay.GetServing().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.current_parameters.rx2_data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetPendingMacState() == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Rx2DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServed().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.Relay.GetServed().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServing() == nil {
					return nil
				}
				chIdx := dev.PendingMacState.DesiredParameters.Relay.GetServing().DefaultChannelIndex
				if chIdx >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServing().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.Relay.GetServing().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.desired_parameters.rx2_data_rate_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetPendingMacState() == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.Rx2DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.current_parameters.ping_slot_data_rate_index_value.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetMacState() == nil || dev.MacState.CurrentParameters.PingSlotDataRateIndexValue == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.MacState.CurrentParameters.PingSlotDataRateIndexValue.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacState().GetCurrentParameters().GetRelay().GetServed().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacState.CurrentParameters.Relay.GetServed().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.current_parameters.relay.mode.serving.default_channel_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacState().GetCurrentParameters().GetRelay().GetServing() == nil {
					return nil
				}
				chIdx := dev.MacState.CurrentParameters.Relay.GetServing().DefaultChannelIndex
				if chIdx >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacState().GetCurrentParameters().GetRelay().GetServing().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacState.CurrentParameters.Relay.GetServing().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.desired_parameters.ping_slot_data_rate_index_value.value"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetMacState() == nil || dev.MacState.DesiredParameters.PingSlotDataRateIndexValue == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.MacState.DesiredParameters.PingSlotDataRateIndexValue.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacState().GetDesiredParameters().GetRelay().GetServed().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacState.DesiredParameters.Relay.GetServed().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.desired_parameters.relay.mode.serving.default_channel_index"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacState().GetDesiredParameters().GetRelay().GetServing() == nil {
					return nil
				}
				chIdx := dev.MacState.DesiredParameters.Relay.GetServing().DefaultChannelIndex
				if chIdx >= uint32(len(phy.Relay.WORChannels)) { // nolint: gosec
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if dev.GetMacState().GetDesiredParameters().GetRelay().GetServing().GetSecondChannel() == nil {
					return nil
				}
				_, ok := phy.DataRates[dev.MacState.DesiredParameters.Relay.GetServing().SecondChannel.DataRateIndex]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}

	if field, validate := hasSetField("pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetPendingMacState() == nil || dev.PendingMacState.CurrentParameters.PingSlotDataRateIndexValue == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.PingSlotDataRateIndexValue.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}
	if field, validate := hasSetField("pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value"); validate { // nolint: lll
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetPendingMacState() == nil || dev.PendingMacState.DesiredParameters.PingSlotDataRateIndexValue == nil {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.PingSlotDataRateIndexValue.Value]
				if !ok {
					return newInvalidFieldValueError(field)
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}

	if field, validate := hasSetField("mac_settings.factory_preset_frequencies"); validate {
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			if dev.GetMacSettings() == nil || len(dev.MacSettings.FactoryPresetFrequencies) == 0 {
				return nil
			}
			return withPHY(func(phy *band.Band, fp *frequencyplans.FrequencyPlan) error {
				switch phy.CFListType {
				case ttnpb.CFListType_FREQUENCIES:
					// Factory preset frequencies in bands which provide frequencies as part of the CFList
					// are interpreted as being used both for uplinks and downlinks.
					for _, frequency := range dev.MacSettings.FactoryPresetFrequencies {
						_, inSubBand := fp.FindSubBand(frequency)
						for _, sb := range phy.SubBands {
							if sb.MinFrequency <= frequency && frequency <= sb.MaxFrequency {
								inSubBand = true
								break
							}
						}
						if !inSubBand {
							return newInvalidFieldValueError(field)
						}
					}
				case ttnpb.CFListType_CHANNEL_MASKS:
					// Factory preset frequencies in bands which provide channel masks as part of the CFList
					// are interpreted as enabling explicit uplink channels.
					uplinkChannels := make(map[uint64]struct{}, len(phy.UplinkChannels))
					for _, ch := range phy.UplinkChannels {
						uplinkChannels[ch.Frequency] = struct{}{}
					}
					for _, frequency := range dev.MacSettings.FactoryPresetFrequencies {
						if _, ok := uplinkChannels[frequency]; !ok {
							return newInvalidFieldValueError(field)
						}
					}
				default:
					panic("unreachable")
				}
				return nil
			})
		},
			field,
		); err != nil {
			return err
		}
	}

	if hasPHYUpdate || st.HasSetField(
		"mac_settings.ping_slot_frequency.value",
		"supports_class_b",
	) {
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			if !m["supports_class_b"].GetSupportsClassB() ||
				m["mac_settings.ping_slot_frequency.value"].GetMacSettings().GetPingSlotFrequency().GetValue() > 0 {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if len(phy.PingSlotFrequencies) == 0 {
					return newInvalidFieldValueError("mac_settings.ping_slot_frequency.value")
				}
				return nil
			})
		},
			"mac_settings.ping_slot_frequency.value",
			"supports_class_b",
		); err != nil {
			return err
		}
	}

	if hasPHYUpdate || st.HasSetField(
		"mac_settings.desired_ping_slot_frequency.value",
		"supports_class_b",
	) {
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			if !m["supports_class_b"].GetSupportsClassB() ||
				m["mac_settings.desired_ping_slot_frequency.value"].GetMacSettings().GetDesiredPingSlotFrequency().GetValue() > 0 { // nolint: lll
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if len(phy.PingSlotFrequencies) == 0 {
					return newInvalidFieldValueError("mac_settings.desired_ping_slot_frequency.value")
				}
				return nil
			})
		},
			"mac_settings.desired_ping_slot_frequency.value",
			"supports_class_b",
		); err != nil {
			return err
		}
	}

	if hasPHYUpdate || st.HasSetField(
		"mac_settings.beacon_frequency.value",
		"supports_class_b",
	) {
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			if !m["supports_class_b"].GetSupportsClassB() ||
				m["mac_settings.beacon_frequency.value"].GetMacSettings().GetBeaconFrequency().GetValue() > 0 {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if len(phy.Beacon.Frequencies) == 0 {
					return newInvalidFieldValueError("mac_settings.beacon_frequency.value")
				}
				return nil
			})
		},
			"mac_settings.beacon_frequency.value",
			"supports_class_b",
		); err != nil {
			return err
		}
	}

	if hasPHYUpdate || st.HasSetField(
		"mac_settings.desired_beacon_frequency.value",
		"supports_class_b",
	) {
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			if !m["supports_class_b"].GetSupportsClassB() ||
				m["mac_settings.desired_beacon_frequency.value"].GetMacSettings().GetDesiredBeaconFrequency().GetValue() > 0 {
				return nil
			}
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if len(phy.Beacon.Frequencies) == 0 {
					return newInvalidFieldValueError("mac_settings.desired_beacon_frequency.value")
				}
				return nil
			})
		},
			"mac_settings.desired_beacon_frequency.value",
			"supports_class_b",
		); err != nil {
			return err
		}
	}

	for p, isValid := range map[string]func(*ttnpb.EndDevice, *band.Band) bool{
		"mac_settings.use_adr.value": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return !dev.GetMacSettings().GetUseAdr().GetValue() || phy.SupportsDynamicADR // nolint: staticcheck
		},
		"mac_state.current_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetMacState().GetCurrentParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
		},
		"mac_state.current_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetMacState().GetCurrentParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
		},
		"mac_state.current_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return len(dev.GetMacState().GetCurrentParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
		},
		"mac_state.desired_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetMacState().GetDesiredParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
		},
		"mac_state.desired_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetMacState().GetDesiredParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
		},
		"mac_state.desired_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return len(dev.GetMacState().GetDesiredParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
		},
		"pending_mac_state.current_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetPendingMacState().GetCurrentParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
		},
		"pending_mac_state.current_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetPendingMacState().GetCurrentParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
		},
		"pending_mac_state.current_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return len(dev.GetPendingMacState().GetCurrentParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
		},
		"pending_mac_state.desired_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetPendingMacState().GetDesiredParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
		},
		"pending_mac_state.desired_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return dev.GetPendingMacState().GetDesiredParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
		},
		"pending_mac_state.desired_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
			return len(dev.GetPendingMacState().GetDesiredParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
		},
	} {
		if !hasPHYUpdate && !st.HasSetField(p) {
			continue
		}
		if err := st.WithField(func(dev *ttnpb.EndDevice) error {
			return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
				if !isValid(dev, phy) {
					return newInvalidFieldValueError(p)
				}
				return nil
			})
		}, p); err != nil {
			return err
		}
	}
	return nil
}

// Ensure ADR dynamic parameters are monotonic.
// If one of the extrema is missing, the other extrema is considered to be valid.
func validateADRDynamicParameters(st *setDeviceState) error {
	return st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		{
			min := m["mac_settings.adr.mode.dynamic.min_data_rate_index.value"].GetMacSettings().GetAdr().GetDynamic().GetMinDataRateIndex() // nolint: revive,lll
			max := m["mac_settings.adr.mode.dynamic.max_data_rate_index.value"].GetMacSettings().GetAdr().GetDynamic().GetMaxDataRateIndex() // nolint: revive,lll

			if min != nil && max != nil && max.Value < min.Value {
				return false, "mac_settings.adr.mode.dynamic.max_data_rate_index.value"
			}
		}
		{
			min := m["mac_settings.adr.mode.dynamic.min_tx_power_index"].GetMacSettings().GetAdr().GetDynamic().GetMinTxPowerIndex() // nolint: revive,lll
			max := m["mac_settings.adr.mode.dynamic.max_tx_power_index"].GetMacSettings().GetAdr().GetDynamic().GetMaxTxPowerIndex() // nolint: revive,lll

			if min != nil && max != nil && max.Value < min.Value {
				return false, "mac_settings.adr.mode.dynamic.max_tx_power_index"
			}
		}
		{
			min := m["mac_settings.adr.mode.dynamic.min_nb_trans"].GetMacSettings().GetAdr().GetDynamic().GetMinNbTrans() // nolint: revive,lll
			max := m["mac_settings.adr.mode.dynamic.max_nb_trans"].GetMacSettings().GetAdr().GetDynamic().GetMaxNbTrans() // nolint: revive,lll

			if min != nil && max != nil && max.Value < min.Value {
				return false, "mac_settings.adr.mode.dynamic.max_nb_trans"
			}
		}
		for drIdx := ttnpb.DataRateIndex_DATA_RATE_0; drIdx <= ttnpb.DataRateIndex_DATA_RATE_15; drIdx++ {
			baseField := fmt.Sprintf("mac_settings.adr.mode.dynamic.overrides.data_rate_%d.", drIdx)
			min := mac.DataRateIndexOverridesOf(m[baseField+"min_nb_trans"].GetMacSettings().GetAdr().GetDynamic().GetOverrides(), drIdx).GetMinNbTrans() // nolint: revive,lll
			max := mac.DataRateIndexOverridesOf(m[baseField+"max_nb_trans"].GetMacSettings().GetAdr().GetDynamic().GetOverrides(), drIdx).GetMaxNbTrans() // nolint: revive,lll

			if min != nil && max != nil && max.Value < min.Value {
				return false, baseField + "max_nb_trans"
			}
		}
		return true, ""
	},
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_0.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_1.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_10.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_11.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_12.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_13.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_14.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_15.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_2.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_3.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_4.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_5.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_6.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_7.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_8.min_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.max_nb_trans",
		"mac_settings.adr.mode.dynamic.overrides.data_rate_9.min_nb_trans",
	)
}
