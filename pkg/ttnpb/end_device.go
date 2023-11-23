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

package ttnpb

import (
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// XXX_WellKnownType ensures BoolValue is encoded as upstream BoolValue.
func (v *BoolValue) XXX_WellKnownType() string {
	return "BoolValue"
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *BoolValue) MarshalText() ([]byte, error) {
	if !v.GetValue() {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *BoolValue) UnmarshalText(b []byte) error {
	switch s := string(b); s {
	case "true":
		*v = BoolValue{Value: true}
	case "false":
		*v = BoolValue{}
	default:
		return errCouldNotParse("BoolValue")(s)
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *BoolValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return !v.Value
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ServedRelayParameters) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "backoff":
		return v.Backoff == 0
	case "mode":
		return v.Mode == nil
	case "mode.always":
		return v.GetAlways() == nil
	case "mode.dynamic":
		return v.GetDynamic() == nil
	case "mode.dynamic.smart_enable_level":
		return v.GetDynamic().FieldIsZero("smart_enable_level")
	case "mode.end_device_controlled":
		return v.GetEndDeviceControlled() == nil
	case "second_channel":
		return v.SecondChannel == nil
	case "second_channel.ack_offset":
		return v.SecondChannel.FieldIsZero("ack_offset")
	case "second_channel.data_rate_index":
		return v.SecondChannel.FieldIsZero("data_rate_index")
	case "second_channel.frequency":
		return v.SecondChannel.FieldIsZero("frequency")
	case "serving_device_id":
		return v.ServingDeviceId == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *RelayForwardLimits) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "bucket_size":
		return v.BucketSize == 0
	case "reload_rate":
		return v.ReloadRate == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ServingRelayForwardingLimits) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "reset_behavior":
		return v.ResetBehavior == 0
	case "join_requests":
		return v.JoinRequests == nil
	case "join_requests.bucket_size":
		return v.JoinRequests.FieldIsZero("bucket_size")
	case "join_requests.reload_rate":
		return v.JoinRequests.FieldIsZero("reload_rate")
	case "notifications":
		return v.Notifications == nil
	case "notifications.bucket_size":
		return v.Notifications.FieldIsZero("bucket_size")
	case "notifications.reload_rate":
		return v.Notifications.FieldIsZero("reload_rate")
	case "uplink_messages":
		return v.UplinkMessages == nil
	case "uplink_messages.bucket_size":
		return v.UplinkMessages.FieldIsZero("bucket_size")
	case "uplink_messages.reload_rate":
		return v.UplinkMessages.FieldIsZero("reload_rate")
	case "overall":
		return v.Overall == nil
	case "overall.bucket_size":
		return v.Overall.FieldIsZero("bucket_size")
	case "overall.reload_rate":
		return v.Overall.FieldIsZero("reload_rate")
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ServingRelayParameters) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "second_channel":
		return v.SecondChannel == nil
	case "second_channel.ack_offset":
		return v.SecondChannel.FieldIsZero("ack_offset")
	case "second_channel.data_rate_index":
		return v.SecondChannel.FieldIsZero("data_rate_index")
	case "second_channel.frequency":
		return v.SecondChannel.FieldIsZero("frequency")
	case "default_channel_index":
		return v.DefaultChannelIndex == 0
	case "cad_periodicity":
		return v.CadPeriodicity == 0
	case "uplink_forwarding_rules":
		return v.UplinkForwardingRules == nil
	case "limits":
		return v.Limits == nil
	case "limits.reset_behavior":
		return v.Limits.FieldIsZero("reset_behavior")
	case "limits.join_requests":
		return v.Limits.FieldIsZero("join_requests")
	case "limits.join_requests.bucket_size":
		return v.Limits.FieldIsZero("join_requests.bucket_size")
	case "limits.join_requests.reload_rate":
		return v.Limits.FieldIsZero("join_requests.reload_rate")
	case "limits.notifications":
		return v.Limits.FieldIsZero("notifications")
	case "limits.notifications.bucket_size":
		return v.Limits.FieldIsZero("notifications.bucket_size")
	case "limits.notifications.reload_rate":
		return v.Limits.FieldIsZero("notifications.reload_rate")
	case "limits.uplink_messages":
		return v.Limits.FieldIsZero("uplink_messages")
	case "limits.uplink_messages.bucket_size":
		return v.Limits.FieldIsZero("uplink_messages.bucket_size")
	case "limits.uplink_messages.reload_rate":
		return v.Limits.FieldIsZero("uplink_messages.reload_rate")
	case "limits.overall":
		return v.Limits.FieldIsZero("overall")
	case "limits.overall.bucket_size":
		return v.Limits.FieldIsZero("overall.bucket_size")
	case "limits.overall.reload_rate":
		return v.Limits.FieldIsZero("overall.reload_rate")
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *RelayParameters) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "mode":
		return v.Mode == nil
	case "mode.served":
		return v.GetServed() == nil
	case "mode.served.backoff":
		return v.GetServed().FieldIsZero("backoff")
	case "mode.served.mode":
		return v.GetServed().FieldIsZero("mode")
	case "mode.served.mode.always":
		return v.GetServed().FieldIsZero("mode.always")
	case "mode.served.mode.dynamic":
		return v.GetServed().FieldIsZero("mode.dynamic")
	case "mode.served.mode.dynamic.smart_enable_level":
		return v.GetServed().FieldIsZero("mode.dynamic.smart_enable_level")
	case "mode.served.mode.end_device_controlled":
		return v.GetServed().FieldIsZero("mode.end_device_controlled")
	case "mode.served.second_channel":
		return v.GetServed().FieldIsZero("second_channel")
	case "mode.served.second_channel.ack_offset":
		return v.GetServed().FieldIsZero("second_channel.ack_offset")
	case "mode.served.second_channel.data_rate_index":
		return v.GetServed().FieldIsZero("second_channel.data_rate_index")
	case "mode.served.second_channel.frequency":
		return v.GetServed().FieldIsZero("second_channel.frequency")
	case "mode.served.default_channel_index":
		return v.GetServed().FieldIsZero("default_channel_index")
	case "mode.served.serving_device_id":
		return v.GetServed().FieldIsZero("serving_device_id")
	case "mode.serving":
		return v.GetServing() == nil
	case "mode.serving.second_channel":
		return v.GetServing().FieldIsZero("second_channel")
	case "mode.serving.second_channel.ack_offset":
		return v.GetServing().FieldIsZero("second_channel.ack_offset")
	case "mode.serving.second_channel.data_rate_index":
		return v.GetServing().FieldIsZero("second_channel.data_rate_index")
	case "mode.serving.second_channel.frequency":
		return v.GetServing().FieldIsZero("second_channel.frequency")
	case "mode.serving.default_channel_index":
		return v.GetServing().FieldIsZero("default_channel_index")
	case "mode.serving.cad_periodicity":
		return v.GetServing().FieldIsZero("cad_periodicity")
	case "mode.serving.uplink_forwarding_rules":
		return v.GetServing().FieldIsZero("uplink_forwarding_rules")
	case "mode.serving.limits":
		return v.GetServing().FieldIsZero("limits")
	case "mode.serving.limits.reset_behavior":
		return v.GetServing().FieldIsZero("limits.reset_behavior")
	case "mode.serving.limits.join_requests":
		return v.GetServing().FieldIsZero("limits.join_requests")
	case "mode.serving.limits.join_requests.bucket_size":
		return v.GetServing().FieldIsZero("limits.join_requests.bucket_size")
	case "mode.serving.limits.join_requests.reload_rate":
		return v.GetServing().FieldIsZero("limits.join_requests.reload_rate")
	case "mode.serving.limits.notifications":
		return v.GetServing().FieldIsZero("limits.notifications")
	case "mode.serving.limits.notifications.bucket_size":
		return v.GetServing().FieldIsZero("limits.notifications.bucket_size")
	case "mode.serving.limits.notifications.reload_rate":
		return v.GetServing().FieldIsZero("limits.notifications.reload_rate")
	case "mode.serving.limits.uplink_messages":
		return v.GetServing().FieldIsZero("limits.uplink_messages")
	case "mode.serving.limits.uplink_messages.bucket_size":
		return v.GetServing().FieldIsZero("limits.uplink_messages.bucket_size")
	case "mode.serving.limits.uplink_messages.reload_rate":
		return v.GetServing().FieldIsZero("limits.uplink_messages.reload_rate")
	case "mode.serving.limits.overall":
		return v.GetServing().FieldIsZero("limits.overall")
	case "mode.serving.limits.overall.bucket_size":
		return v.GetServing().FieldIsZero("limits.overall.bucket_size")
	case "mode.serving.limits.overall.reload_rate":
		return v.GetServing().FieldIsZero("limits.overall.reload_rate")
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *EndDeviceAuthenticationCode) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "valid_from":
		return v.ValidFrom == nil
	case "valid_to":
		return v.ValidTo == nil
	case "value":
		return v.Value == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ADRSettings_StaticMode) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "data_rate_index":
		return v.DataRateIndex == 0
	case "tx_power_index":
		return v.TxPowerIndex == 0
	case "nb_trans":
		return v.NbTrans == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ADRSettings_DynamicMode_ChannelSteeringSettings) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "mode":
		return v.Mode == nil
	case "mode.disabled":
		return v.GetDisabled() == nil
	case "mode.lora_narrow":
		return v.GetLoraNarrow() == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ADRSettings_DynamicMode) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "margin":
		return v.Margin == nil
	case "channel_steering":
		return v.ChannelSteering == nil
	case "channel_steering.mode":
		return v.ChannelSteering.FieldIsZero("mode")
	case "channel_steering.mode.disabled":
		return v.ChannelSteering.FieldIsZero("mode.disabled")
	case "channel_steering.mode.lora_narrow":
		return v.ChannelSteering.FieldIsZero("mode.lora_narrow")
	case "min_data_rate_index":
		return v.MinDataRateIndex == nil
	case "min_data_rate_index.value":
		return v.MinDataRateIndex.FieldIsZero("value")
	case "max_data_rate_index":
		return v.MaxDataRateIndex == nil
	case "max_data_rate_index.value":
		return v.MaxDataRateIndex.FieldIsZero("value")
	case "min_tx_power_index":
		return v.MinTxPowerIndex == nil
	case "max_tx_power_index":
		return v.MaxTxPowerIndex == nil
	case "min_nb_trans":
		return v.MinNbTrans == nil
	case "max_nb_trans":
		return v.MaxNbTrans == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ADRSettings) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "mode":
		return v.Mode == nil
	case "mode.static":
		return v.GetStatic() == nil
	case "mode.static.data_rate_index":
		return v.GetStatic().FieldIsZero("data_rate_index")
	case "mode.static.tx_power_index":
		return v.GetStatic().FieldIsZero("tx_power_index")
	case "mode.static.nb_trans":
		return v.GetStatic().FieldIsZero("nb_trans")
	case "mode.dynamic":
		return v.GetDynamic() == nil
	case "mode.dynamic.channel_steering":
		return v.GetDynamic().FieldIsZero("channel_steering")
	case "mode.dynamic.channel_steering.mode":
		return v.GetDynamic().FieldIsZero("channel_steering.mode")
	case "mode.dynamic.channel_steering.mode.disabled":
		return v.GetDynamic().FieldIsZero("channel_steering.mode.disabled")
	case "mode.dynamic.channel_steering.mode.lora_narrow":
		return v.GetDynamic().FieldIsZero("channel_steering.mode.lora_narrow")
	case "mode.dynamic.margin":
		return v.GetDynamic().FieldIsZero("margin")
	case "mode.dynamic.min_data_rate_index":
		return v.GetDynamic().FieldIsZero("min_data_rate_index")
	case "mode.dynamic.min_data_rate_index.value":
		return v.GetDynamic().FieldIsZero("min_data_rate_index.value")
	case "mode.dynamic.max_data_rate_index":
		return v.GetDynamic().FieldIsZero("max_data_rate_index")
	case "mode.dynamic.max_data_rate_index.value":
		return v.GetDynamic().FieldIsZero("max_data_rate_index.value")
	case "mode.dynamic.min_tx_power_index":
		return v.GetDynamic().FieldIsZero("min_tx_power_index")
	case "mode.dynamic.max_tx_power_index":
		return v.GetDynamic().FieldIsZero("max_tx_power_index")
	case "mode.dynamic.min_nb_trans":
		return v.GetDynamic().FieldIsZero("min_nb_trans")
	case "mode.dynamic.max_nb_trans":
		return v.GetDynamic().FieldIsZero("max_nb_trans")
	case "mode.disabled":
		return v.GetDisabled() == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MACSettings) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "adr":
		return v.Adr == nil
	case "adr.mode":
		return v.Adr.FieldIsZero("mode")
	case "adr.mode.static":
		return v.Adr.FieldIsZero("mode.static")
	case "adr.mode.static.data_rate_index":
		return v.Adr.FieldIsZero("mode.static.data_rate_index")
	case "adr.mode.static.tx_power_index":
		return v.Adr.FieldIsZero("mode.static.tx_power_index")
	case "adr.mode.static.nb_trans":
		return v.Adr.FieldIsZero("mode.static.nb_trans")
	case "adr.mode.dynamic":
		return v.Adr.FieldIsZero("mode.dynamic")
	case "adr.mode.dynamic.channel_steering":
		return v.Adr.FieldIsZero("mode.dynamic.channel_steering")
	case "adr.mode.dynamic.channel_steering.mode":
		return v.Adr.FieldIsZero("mode.dynamic.channel_steering.mode")
	case "adr.mode.dynamic.channel_steering.mode.disabled":
		return v.Adr.FieldIsZero("mode.dynamic.channel_steering.mode.disabled")
	case "adr.mode.dynamic.channel_steering.mode.lora_narrow":
		return v.Adr.FieldIsZero("mode.dynamic.channel_steering.mode.lora_narrow")
	case "adr.mode.dynamic.margin":
		return v.Adr.FieldIsZero("mode.dynamic.margin")
	case "adr.mode.dynamic.min_data_rate_index":
		return v.Adr.FieldIsZero("mode.dynamic.min_data_rate_index")
	case "adr.mode.dynamic.min_data_rate_index.value":
		return v.Adr.FieldIsZero("mode.dynamic.min_data_rate_index.value")
	case "adr.mode.dynamic.max_data_rate_index":
		return v.Adr.FieldIsZero("mode.dynamic.max_data_rate_index")
	case "adr.mode.dynamic.max_data_rate_index.value":
		return v.Adr.FieldIsZero("mode.dynamic.max_data_rate_index.value")
	case "adr.mode.dynamic.min_tx_power_index":
		return v.Adr.FieldIsZero("mode.dynamic.min_tx_power_index")
	case "adr.mode.dynamic.max_tx_power_index":
		return v.Adr.FieldIsZero("mode.dynamic.max_tx_power_index")
	case "adr.mode.dynamic.min_nb_trans":
		return v.Adr.FieldIsZero("mode.dynamic.min_nb_trans")
	case "adr.mode.dynamic.max_nb_trans":
		return v.Adr.FieldIsZero("mode.dynamic.max_nb_trans")
	case "adr.mode.disabled":
		return v.Adr.FieldIsZero("mode.disabled")
	case "adr_margin":
		return v.AdrMargin == nil
	case "beacon_frequency":
		return v.BeaconFrequency == nil
	case "beacon_frequency.value":
		return v.BeaconFrequency.FieldIsZero("value")
	case "class_b_timeout":
		return v.ClassBTimeout == nil
	case "class_b_c_downlink_interval":
		return v.ClassBCDownlinkInterval == nil
	case "class_c_timeout":
		return v.ClassCTimeout == nil
	case "desired_adr_ack_delay_exponent":
		return v.DesiredAdrAckDelayExponent == nil
	case "desired_adr_ack_delay_exponent.value":
		return v.DesiredAdrAckDelayExponent.FieldIsZero("value")
	case "desired_adr_ack_limit_exponent":
		return v.DesiredAdrAckLimitExponent == nil
	case "desired_adr_ack_limit_exponent.value":
		return v.DesiredAdrAckLimitExponent.FieldIsZero("value")
	case "desired_beacon_frequency":
		return v.DesiredBeaconFrequency == nil
	case "desired_beacon_frequency.value":
		return v.DesiredBeaconFrequency.FieldIsZero("value")
	case "desired_max_duty_cycle":
		return v.DesiredMaxDutyCycle == nil
	case "desired_max_duty_cycle.value":
		return v.DesiredMaxDutyCycle.FieldIsZero("value")
	case "desired_max_eirp":
		return v.DesiredMaxEirp == nil
	case "desired_max_eirp.value":
		return v.DesiredMaxEirp.FieldIsZero("value")
	case "desired_ping_slot_data_rate_index":
		return v.DesiredPingSlotDataRateIndex == nil
	case "desired_ping_slot_data_rate_index.value":
		return v.DesiredPingSlotDataRateIndex.FieldIsZero("value")
	case "desired_ping_slot_frequency":
		return v.DesiredPingSlotFrequency == nil
	case "desired_ping_slot_frequency.value":
		return v.DesiredPingSlotFrequency.FieldIsZero("value")
	case "desired_rx1_data_rate_offset":
		return v.DesiredRx1DataRateOffset == nil
	case "desired_rx1_data_rate_offset.value":
		return v.DesiredRx1DataRateOffset.FieldIsZero("value")
	case "desired_rx1_delay":
		return v.DesiredRx1Delay == nil
	case "desired_rx1_delay.value":
		return v.DesiredRx1Delay.FieldIsZero("value")
	case "desired_rx2_data_rate_index":
		return v.DesiredRx2DataRateIndex == nil
	case "desired_rx2_data_rate_index.value":
		return v.DesiredRx2DataRateIndex.FieldIsZero("value")
	case "desired_rx2_frequency":
		return v.DesiredRx2Frequency == nil
	case "desired_rx2_frequency.value":
		return v.DesiredRx2Frequency.FieldIsZero("value")
	case "factory_preset_frequencies":
		return v.FactoryPresetFrequencies == nil
	case "max_duty_cycle":
		return v.MaxDutyCycle == nil
	case "max_duty_cycle.value":
		return v.MaxDutyCycle.FieldIsZero("value")
	case "ping_slot_data_rate_index":
		return v.PingSlotDataRateIndex == nil
	case "ping_slot_data_rate_index.value":
		return v.PingSlotDataRateIndex.FieldIsZero("value")
	case "ping_slot_frequency":
		return v.PingSlotFrequency == nil
	case "ping_slot_frequency.value":
		return v.PingSlotFrequency.FieldIsZero("value")
	case "ping_slot_periodicity":
		return v.PingSlotPeriodicity == nil
	case "ping_slot_periodicity.value":
		return v.PingSlotPeriodicity.FieldIsZero("value")
	case "relay":
		return v.Relay == nil
	case "relay.mode":
		return v.Relay.FieldIsZero("mode")
	case "relay.mode.served":
		return v.Relay.FieldIsZero("mode.served")
	case "relay.mode.served.backoff":
		return v.Relay.FieldIsZero("mode.served.backoff")
	case "relay.mode.served.mode":
		return v.Relay.FieldIsZero("mode.served.mode")
	case "relay.mode.served.mode.always":
		return v.Relay.FieldIsZero("mode.served.mode.always")
	case "relay.mode.served.mode.dynamic":
		return v.Relay.FieldIsZero("mode.served.mode.dynamic")
	case "relay.mode.served.mode.dynamic.smart_enable_level":
		return v.Relay.FieldIsZero("mode.served.mode.dynamic.smart_enable_level")
	case "relay.mode.served.mode.end_device_controlled":
		return v.Relay.FieldIsZero("mode.served.mode.end_device_controlled")
	case "relay.mode.served.second_channel":
		return v.Relay.FieldIsZero("mode.served.second_channel")
	case "relay.mode.served.second_channel.ack_offset":
		return v.Relay.FieldIsZero("mode.served.second_channel.ack_offset")
	case "relay.mode.served.second_channel.data_rate_index":
		return v.Relay.FieldIsZero("mode.served.second_channel.data_rate_index")
	case "relay.mode.served.second_channel.frequency":
		return v.Relay.FieldIsZero("mode.served.second_channel.frequency")
	case "relay.mode.served.default_channel_index":
		return v.Relay.FieldIsZero("mode.served.default_channel_index")
	case "relay.mode.served.serving_device_id":
		return v.Relay.FieldIsZero("mode.served.serving_device_id")
	case "relay.mode.serving":
		return v.Relay.FieldIsZero("mode.serving")
	case "relay.mode.serving.second_channel":
		return v.Relay.FieldIsZero("mode.serving.second_channel")
	case "relay.mode.serving.second_channel.ack_offset":
		return v.Relay.FieldIsZero("mode.serving.second_channel.ack_offset")
	case "relay.mode.serving.second_channel.data_rate_index":
		return v.Relay.FieldIsZero("mode.serving.second_channel.data_rate_index")
	case "relay.mode.serving.second_channel.frequency":
		return v.Relay.FieldIsZero("mode.serving.second_channel.frequency")
	case "relay.mode.serving.default_channel_index":
		return v.Relay.FieldIsZero("mode.serving.default_channel_index")
	case "relay.mode.serving.cad_periodicity":
		return v.Relay.FieldIsZero("mode.serving.cad_periodicity")
	case "relay.mode.serving.uplink_forwarding_rules":
		return v.Relay.FieldIsZero("mode.serving.uplink_forwarding_rules")
	case "relay.mode.serving.limits":
		return v.Relay.FieldIsZero("mode.serving.limits")
	case "relay.mode.serving.limits.reset_behavior":
		return v.Relay.FieldIsZero("mode.serving.limits.reset_behavior")
	case "relay.mode.serving.limits.join_requests":
		return v.Relay.FieldIsZero("mode.serving.limits.join_requests")
	case "relay.mode.serving.limits.join_requests.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.join_requests.bucket_size")
	case "relay.mode.serving.limits.join_requests.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.join_requests.reload_rate")
	case "relay.mode.serving.limits.notifications":
		return v.Relay.FieldIsZero("mode.serving.limits.notifications")
	case "relay.mode.serving.limits.notifications.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.notifications.bucket_size")
	case "relay.mode.serving.limits.notifications.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.notifications.reload_rate")
	case "relay.mode.serving.limits.uplink_messages":
		return v.Relay.FieldIsZero("mode.serving.limits.uplink_messages")
	case "relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.uplink_messages.bucket_size")
	case "relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.uplink_messages.reload_rate")
	case "relay.mode.serving.limits.overall":
		return v.Relay.FieldIsZero("mode.serving.limits.overall")
	case "relay.mode.serving.limits.overall.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.overall.bucket_size")
	case "relay.mode.serving.limits.overall.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.overall.reload_rate")
	case "desired_relay":
		return v.DesiredRelay == nil
	case "desired_relay.mode":
		return v.DesiredRelay.FieldIsZero("mode")
	case "desired_relay.mode.served":
		return v.DesiredRelay.FieldIsZero("mode.served")
	case "desired_relay.mode.served.backoff":
		return v.DesiredRelay.FieldIsZero("mode.served.backoff")
	case "desired_relay.mode.served.mode":
		return v.DesiredRelay.FieldIsZero("mode.served.mode")
	case "desired_relay.mode.served.mode.always":
		return v.DesiredRelay.FieldIsZero("mode.served.mode.always")
	case "desired_relay.mode.served.mode.dynamic":
		return v.DesiredRelay.FieldIsZero("mode.served.mode.dynamic")
	case "desired_relay.mode.served.mode.dynamic.smart_enable_level":
		return v.DesiredRelay.FieldIsZero("mode.served.mode.dynamic.smart_enable_level")
	case "desired_relay.mode.served.mode.end_device_controlled":
		return v.DesiredRelay.FieldIsZero("mode.served.mode.end_device_controlled")
	case "desired_relay.mode.served.second_channel":
		return v.DesiredRelay.FieldIsZero("mode.served.second_channel")
	case "desired_relay.mode.served.second_channel.ack_offset":
		return v.DesiredRelay.FieldIsZero("mode.served.second_channel.ack_offset")
	case "desired_relay.mode.served.second_channel.data_rate_index":
		return v.DesiredRelay.FieldIsZero("mode.served.second_channel.data_rate_index")
	case "desired_relay.mode.served.second_channel.frequency":
		return v.DesiredRelay.FieldIsZero("mode.served.second_channel.frequency")
	case "desired_relay.mode.served.default_channel_index":
		return v.DesiredRelay.FieldIsZero("mode.served.default_channel_index")
	case "desired_relay.mode.served.serving_device_id":
		return v.DesiredRelay.FieldIsZero("mode.served.serving_device_id")
	case "desired_relay.mode.serving":
		return v.DesiredRelay.FieldIsZero("mode.serving")
	case "desired_relay.mode.serving.second_channel":
		return v.DesiredRelay.FieldIsZero("mode.serving.second_channel")
	case "desired_relay.mode.serving.second_channel.ack_offset":
		return v.DesiredRelay.FieldIsZero("mode.serving.second_channel.ack_offset")
	case "desired_relay.mode.serving.second_channel.data_rate_index":
		return v.DesiredRelay.FieldIsZero("mode.serving.second_channel.data_rate_index")
	case "desired_relay.mode.serving.second_channel.frequency":
		return v.DesiredRelay.FieldIsZero("mode.serving.second_channel.frequency")
	case "desired_relay.mode.serving.default_channel_index":
		return v.DesiredRelay.FieldIsZero("mode.serving.default_channel_index")
	case "desired_relay.mode.serving.cad_periodicity":
		return v.DesiredRelay.FieldIsZero("mode.serving.cad_periodicity")
	case "desired_relay.mode.serving.uplink_forwarding_rules":
		return v.DesiredRelay.FieldIsZero("mode.serving.uplink_forwarding_rules")
	case "desired_relay.mode.serving.limits":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits")
	case "desired_relay.mode.serving.limits.reset_behavior":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.reset_behavior")
	case "desired_relay.mode.serving.limits.join_requests":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.join_requests")
	case "desired_relay.mode.serving.limits.join_requests.bucket_size":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.join_requests.bucket_size")
	case "desired_relay.mode.serving.limits.join_requests.reload_rate":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.join_requests.reload_rate")
	case "desired_relay.mode.serving.limits.notifications":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.notifications")
	case "desired_relay.mode.serving.limits.notifications.bucket_size":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.notifications.bucket_size")
	case "desired_relay.mode.serving.limits.notifications.reload_rate":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.notifications.reload_rate")
	case "desired_relay.mode.serving.limits.uplink_messages":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.uplink_messages")
	case "desired_relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.uplink_messages.bucket_size")
	case "desired_relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.uplink_messages.reload_rate")
	case "desired_relay.mode.serving.limits.overall":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.overall")
	case "desired_relay.mode.serving.limits.overall.bucket_size":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.overall.bucket_size")
	case "desired_relay.mode.serving.limits.overall.reload_rate":
		return v.DesiredRelay.FieldIsZero("mode.serving.limits.overall.reload_rate")
	case "resets_f_cnt":
		return v.ResetsFCnt == nil
	case "resets_f_cnt.value":
		return v.ResetsFCnt.FieldIsZero("value")
	case "rx1_data_rate_offset":
		return v.Rx1DataRateOffset == nil
	case "rx1_data_rate_offset.value":
		return v.Rx1DataRateOffset.FieldIsZero("value")
	case "rx1_delay":
		return v.Rx1Delay == nil
	case "rx1_delay.value":
		return v.Rx1Delay.FieldIsZero("value")
	case "rx2_data_rate_index":
		return v.Rx2DataRateIndex == nil
	case "rx2_data_rate_index.value":
		return v.Rx2DataRateIndex.FieldIsZero("value")
	case "rx2_frequency":
		return v.Rx2Frequency == nil
	case "rx2_frequency.value":
		return v.Rx2Frequency.FieldIsZero("value")
	case "schedule_downlinks":
		return v.ScheduleDownlinks == nil
	case "schedule_downlinks.value":
		return v.ScheduleDownlinks.FieldIsZero("value")
	case "status_count_periodicity":
		return v.StatusCountPeriodicity == nil
	case "status_time_periodicity":
		return v.StatusTimePeriodicity == nil
	case "supports_32_bit_f_cnt":
		return v.Supports_32BitFCnt == nil
	case "supports_32_bit_f_cnt.value":
		return v.Supports_32BitFCnt.FieldIsZero("value")
	case "use_adr":
		return v.UseAdr == nil
	case "use_adr.value":
		return v.UseAdr.FieldIsZero("value")
	case "uplink_dwell_time":
		return v.UplinkDwellTime == nil
	case "uplink_dwell_time.value":
		return v.UplinkDwellTime.FieldIsZero("value")
	case "downlink_dwell_time":
		return v.DownlinkDwellTime == nil
	case "downlink_dwell_time.value":
		return v.DownlinkDwellTime.FieldIsZero("value")
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MACParameters) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "adr_ack_delay":
		return v.AdrAckDelay == 0
	case "adr_ack_delay_exponent":
		return v.AdrAckDelayExponent == nil
	case "adr_ack_delay_exponent.value":
		return v.AdrAckDelayExponent.FieldIsZero("value")
	case "adr_ack_limit":
		return v.AdrAckLimit == 0
	case "adr_ack_limit_exponent":
		return v.AdrAckLimitExponent == nil
	case "adr_ack_limit_exponent.value":
		return v.AdrAckLimitExponent.FieldIsZero("value")
	case "adr_data_rate_index":
		return v.AdrDataRateIndex == 0
	case "adr_nb_trans":
		return v.AdrNbTrans == 0
	case "adr_tx_power_index":
		return v.AdrTxPowerIndex == 0
	case "beacon_frequency":
		return v.BeaconFrequency == 0
	case "channels":
		return v.Channels == nil
	case "downlink_dwell_time":
		return v.DownlinkDwellTime == nil
	case "downlink_dwell_time.value":
		return v.DownlinkDwellTime.FieldIsZero("value")
	case "max_duty_cycle":
		return v.MaxDutyCycle == 0
	case "max_eirp":
		return v.MaxEirp == 0
	case "ping_slot_data_rate_index":
		return v.PingSlotDataRateIndex == 0
	case "ping_slot_data_rate_index_value":
		return v.PingSlotDataRateIndexValue == nil
	case "ping_slot_data_rate_index_value.value":
		return v.PingSlotDataRateIndexValue.FieldIsZero("value")
	case "ping_slot_frequency":
		return v.PingSlotFrequency == 0
	case "rejoin_count_periodicity":
		return v.RejoinCountPeriodicity == 0
	case "rejoin_time_periodicity":
		return v.RejoinTimePeriodicity == 0
	case "relay":
		return v.Relay == nil
	case "relay.mode":
		return v.Relay.FieldIsZero("mode")
	case "relay.mode.served":
		return v.Relay.FieldIsZero("mode.served")
	case "relay.mode.served.backoff":
		return v.Relay.FieldIsZero("mode.served.backoff")
	case "relay.mode.served.mode":
		return v.Relay.FieldIsZero("mode.served.mode")
	case "relay.mode.served.mode.always":
		return v.Relay.FieldIsZero("mode.served.mode.always")
	case "relay.mode.served.mode.dynamic":
		return v.Relay.FieldIsZero("mode.served.mode.dynamic")
	case "relay.mode.served.mode.dynamic.smart_enable_level":
		return v.Relay.FieldIsZero("mode.served.mode.dynamic.smart_enable_level")
	case "relay.mode.served.mode.end_device_controlled":
		return v.Relay.FieldIsZero("mode.served.mode.end_device_controlled")
	case "relay.mode.served.second_channel":
		return v.Relay.FieldIsZero("mode.served.second_channel")
	case "relay.mode.served.second_channel.ack_offset":
		return v.Relay.FieldIsZero("mode.served.second_channel.ack_offset")
	case "relay.mode.served.second_channel.data_rate_index":
		return v.Relay.FieldIsZero("mode.served.second_channel.data_rate_index")
	case "relay.mode.served.second_channel.frequency":
		return v.Relay.FieldIsZero("mode.served.second_channel.frequency")
	case "relay.mode.served.default_channel_index":
		return v.Relay.FieldIsZero("mode.served.default_channel_index")
	case "relay.mode.served.serving_device_id":
		return v.Relay.FieldIsZero("mode.served.serving_device_id")
	case "relay.mode.serving":
		return v.Relay.FieldIsZero("mode.serving")
	case "relay.mode.serving.second_channel":
		return v.Relay.FieldIsZero("mode.serving.second_channel")
	case "relay.mode.serving.second_channel.ack_offset":
		return v.Relay.FieldIsZero("mode.serving.second_channel.ack_offset")
	case "relay.mode.serving.second_channel.data_rate_index":
		return v.Relay.FieldIsZero("mode.serving.second_channel.data_rate_index")
	case "relay.mode.serving.second_channel.frequency":
		return v.Relay.FieldIsZero("mode.serving.second_channel.frequency")
	case "relay.mode.serving.default_channel_index":
		return v.Relay.FieldIsZero("mode.serving.default_channel_index")
	case "relay.mode.serving.cad_periodicity":
		return v.Relay.FieldIsZero("mode.serving.cad_periodicity")
	case "relay.mode.serving.uplink_forwarding_rules":
		return v.Relay.FieldIsZero("mode.serving.uplink_forwarding_rules")
	case "relay.mode.serving.limits":
		return v.Relay.FieldIsZero("mode.serving.limits")
	case "relay.mode.serving.limits.reset_behavior":
		return v.Relay.FieldIsZero("mode.serving.limits.reset_behavior")
	case "relay.mode.serving.limits.join_requests":
		return v.Relay.FieldIsZero("mode.serving.limits.join_requests")
	case "relay.mode.serving.limits.join_requests.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.join_requests.bucket_size")
	case "relay.mode.serving.limits.join_requests.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.join_requests.reload_rate")
	case "relay.mode.serving.limits.notifications":
		return v.Relay.FieldIsZero("mode.serving.limits.notifications")
	case "relay.mode.serving.limits.notifications.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.notifications.bucket_size")
	case "relay.mode.serving.limits.notifications.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.notifications.reload_rate")
	case "relay.mode.serving.limits.uplink_messages":
		return v.Relay.FieldIsZero("mode.serving.limits.uplink_messages")
	case "relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.uplink_messages.bucket_size")
	case "relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.uplink_messages.reload_rate")
	case "relay.mode.serving.limits.overall":
		return v.Relay.FieldIsZero("mode.serving.limits.overall")
	case "relay.mode.serving.limits.overall.bucket_size":
		return v.Relay.FieldIsZero("mode.serving.limits.overall.bucket_size")
	case "relay.mode.serving.limits.overall.reload_rate":
		return v.Relay.FieldIsZero("mode.serving.limits.overall.reload_rate")
	case "rx1_data_rate_offset":
		return v.Rx1DataRateOffset == 0
	case "rx1_delay":
		return v.Rx1Delay == 0
	case "rx2_data_rate_index":
		return v.Rx2DataRateIndex == 0
	case "rx2_frequency":
		return v.Rx2Frequency == 0
	case "uplink_dwell_time":
		return v.UplinkDwellTime == nil
	case "uplink_dwell_time.value":
		return v.UplinkDwellTime.FieldIsZero("value")
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MACState_JoinRequest) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "cf_list":
		return v.CfList == nil
	case "cf_list.ch_masks":
		return v.CfList.FieldIsZero("ch_masks")
	case "cf_list.freq":
		return v.CfList.FieldIsZero("freq")
	case "cf_list.type":
		return v.CfList.FieldIsZero("type")
	case "downlink_settings":
		return v.DownlinkSettings == nil
	case "downlink_settings.opt_neg":
		return v.DownlinkSettings.FieldIsZero("opt_neg")
	case "downlink_settings.rx1_dr_offset":
		return v.DownlinkSettings.FieldIsZero("rx1_dr_offset")
	case "downlink_settings.rx2_dr":
		return v.DownlinkSettings.FieldIsZero("rx2_dr")
	case "rx_delay":
		return v.RxDelay == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MACState_JoinAccept) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "correlation_ids":
		return v.CorrelationIds == nil
	case "dev_addr":
		return types.MustDevAddr(v.DevAddr).OrZero().IsZero()
	case "keys":
		return v.Keys == nil
	case "keys.app_s_key":
		return v.Keys.FieldIsZero("app_s_key")
	case "keys.app_s_key.encrypted_key":
		return v.Keys.FieldIsZero("app_s_key.encrypted_key")
	case "keys.app_s_key.kek_label":
		return v.Keys.FieldIsZero("app_s_key.kek_label")
	case "keys.app_s_key.key":
		return v.Keys.FieldIsZero("app_s_key.key")
	case "keys.f_nwk_s_int_key":
		return v.Keys.FieldIsZero("f_nwk_s_int_key")
	case "keys.f_nwk_s_int_key.encrypted_key":
		return v.Keys.FieldIsZero("f_nwk_s_int_key.encrypted_key")
	case "keys.f_nwk_s_int_key.kek_label":
		return v.Keys.FieldIsZero("f_nwk_s_int_key.kek_label")
	case "keys.f_nwk_s_int_key.key":
		return v.Keys.FieldIsZero("f_nwk_s_int_key.key")
	case "keys.nwk_s_enc_key":
		return v.Keys.FieldIsZero("nwk_s_enc_key")
	case "keys.nwk_s_enc_key.encrypted_key":
		return v.Keys.FieldIsZero("nwk_s_enc_key.encrypted_key")
	case "keys.nwk_s_enc_key.kek_label":
		return v.Keys.FieldIsZero("nwk_s_enc_key.kek_label")
	case "keys.nwk_s_enc_key.key":
		return v.Keys.FieldIsZero("nwk_s_enc_key.key")
	case "keys.s_nwk_s_int_key":
		return v.Keys.FieldIsZero("s_nwk_s_int_key")
	case "keys.s_nwk_s_int_key.encrypted_key":
		return v.Keys.FieldIsZero("s_nwk_s_int_key.encrypted_key")
	case "keys.s_nwk_s_int_key.kek_label":
		return v.Keys.FieldIsZero("s_nwk_s_int_key.kek_label")
	case "keys.s_nwk_s_int_key.key":
		return v.Keys.FieldIsZero("s_nwk_s_int_key.key")
	case "keys.session_key_id":
		return v.Keys.FieldIsZero("session_key_id")
	case "net_id":
		return types.MustNetID(v.NetId).OrZero().IsZero()
	case "payload":
		return v.Payload == nil
	case "request":
		return v.Request == nil
	case "request.cf_list":
		return v.Request.FieldIsZero("cf_list")
	case "request.cf_list.ch_masks":
		return v.Request.FieldIsZero("cf_list.ch_masks")
	case "request.cf_list.freq":
		return v.Request.FieldIsZero("cf_list.freq")
	case "request.cf_list.type":
		return v.Request.FieldIsZero("cf_list.type")
	case "request.downlink_settings":
		return v.Request.FieldIsZero("downlink_settings")
	case "request.downlink_settings.opt_neg":
		return v.Request.FieldIsZero("downlink_settings.opt_neg")
	case "request.downlink_settings.rx1_dr_offset":
		return v.Request.FieldIsZero("downlink_settings.rx1_dr_offset")
	case "request.downlink_settings.rx2_dr":
		return v.Request.FieldIsZero("downlink_settings.rx2_dr")
	case "request.rx_delay":
		return v.Request.FieldIsZero("rx_delay")
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MACState) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "current_parameters":
		return v.CurrentParameters == nil
	case "current_parameters.adr_ack_delay":
		return v.CurrentParameters.FieldIsZero("adr_ack_delay")
	case "current_parameters.adr_ack_delay_exponent":
		return v.CurrentParameters.FieldIsZero("adr_ack_delay_exponent")
	case "current_parameters.adr_ack_delay_exponent.value":
		return v.CurrentParameters.FieldIsZero("adr_ack_delay_exponent.value")
	case "current_parameters.adr_ack_limit":
		return v.CurrentParameters.FieldIsZero("adr_ack_limit")
	case "current_parameters.adr_ack_limit_exponent":
		return v.CurrentParameters.FieldIsZero("adr_ack_limit_exponent")
	case "current_parameters.adr_ack_limit_exponent.value":
		return v.CurrentParameters.FieldIsZero("adr_ack_limit_exponent.value")
	case "current_parameters.adr_data_rate_index":
		return v.CurrentParameters.FieldIsZero("adr_data_rate_index")
	case "current_parameters.adr_nb_trans":
		return v.CurrentParameters.FieldIsZero("adr_nb_trans")
	case "current_parameters.adr_tx_power_index":
		return v.CurrentParameters.FieldIsZero("adr_tx_power_index")
	case "current_parameters.beacon_frequency":
		return v.CurrentParameters.FieldIsZero("beacon_frequency")
	case "current_parameters.channels":
		return v.CurrentParameters.FieldIsZero("channels")
	case "current_parameters.downlink_dwell_time":
		return v.CurrentParameters.FieldIsZero("downlink_dwell_time")
	case "current_parameters.downlink_dwell_time.value":
		return v.CurrentParameters.FieldIsZero("downlink_dwell_time.value")
	case "current_parameters.max_duty_cycle":
		return v.CurrentParameters.FieldIsZero("max_duty_cycle")
	case "current_parameters.max_eirp":
		return v.CurrentParameters.FieldIsZero("max_eirp")
	case "current_parameters.ping_slot_data_rate_index":
		return v.CurrentParameters.FieldIsZero("ping_slot_data_rate_index")
	case "current_parameters.ping_slot_data_rate_index_value":
		return v.CurrentParameters.FieldIsZero("ping_slot_data_rate_index_value")
	case "current_parameters.ping_slot_data_rate_index_value.value":
		return v.CurrentParameters.FieldIsZero("ping_slot_data_rate_index_value.value")
	case "current_parameters.ping_slot_frequency":
		return v.CurrentParameters.FieldIsZero("ping_slot_frequency")
	case "current_parameters.rejoin_count_periodicity":
		return v.CurrentParameters.FieldIsZero("rejoin_count_periodicity")
	case "current_parameters.rejoin_time_periodicity":
		return v.CurrentParameters.FieldIsZero("rejoin_time_periodicity")
	case "current_parameters.relay":
		return v.CurrentParameters.FieldIsZero("relay")
	case "current_parameters.relay.mode":
		return v.CurrentParameters.FieldIsZero("relay.mode")
	case "current_parameters.relay.mode.served":
		return v.CurrentParameters.FieldIsZero("relay.mode.served")
	case "current_parameters.relay.mode.served.backoff":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.backoff")
	case "current_parameters.relay.mode.served.mode":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.mode")
	case "current_parameters.relay.mode.served.mode.always":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.mode.always")
	case "current_parameters.relay.mode.served.mode.dynamic":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.mode.dynamic")
	case "current_parameters.relay.mode.served.mode.dynamic.smart_enable_level":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.mode.dynamic.smart_enable_level")
	case "current_parameters.relay.mode.served.mode.end_device_controlled":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.mode.end_device_controlled")
	case "current_parameters.relay.mode.served.second_channel":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.second_channel")
	case "current_parameters.relay.mode.served.second_channel.ack_offset":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.second_channel.ack_offset")
	case "current_parameters.relay.mode.served.second_channel.data_rate_index":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.second_channel.data_rate_index")
	case "current_parameters.relay.mode.served.second_channel.frequency":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.second_channel.frequency")
	case "current_parameters.relay.mode.served.default_channel_index":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.default_channel_index")
	case "current_parameters.relay.mode.served.serving_device_id":
		return v.CurrentParameters.FieldIsZero("relay.mode.served.serving_device_id")
	case "current_parameters.relay.mode.serving":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving")
	case "current_parameters.relay.mode.serving.second_channel":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.second_channel")
	case "current_parameters.relay.mode.serving.second_channel.ack_offset":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.second_channel.ack_offset")
	case "current_parameters.relay.mode.serving.second_channel.data_rate_index":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.second_channel.data_rate_index")
	case "current_parameters.relay.mode.serving.second_channel.frequency":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.second_channel.frequency")
	case "current_parameters.relay.mode.serving.default_channel_index":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.default_channel_index")
	case "current_parameters.relay.mode.serving.cad_periodicity":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.cad_periodicity")
	case "current_parameters.relay.mode.serving.uplink_forwarding_rules":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.uplink_forwarding_rules")
	case "current_parameters.relay.mode.serving.limits":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits")
	case "current_parameters.relay.mode.serving.limits.reset_behavior":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.reset_behavior")
	case "current_parameters.relay.mode.serving.limits.join_requests":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.join_requests")
	case "current_parameters.relay.mode.serving.limits.join_requests.bucket_size":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.join_requests.bucket_size")
	case "current_parameters.relay.mode.serving.limits.join_requests.reload_rate":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.join_requests.reload_rate")
	case "current_parameters.relay.mode.serving.limits.notifications":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.notifications")
	case "current_parameters.relay.mode.serving.limits.notifications.bucket_size":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.notifications.bucket_size")
	case "current_parameters.relay.mode.serving.limits.notifications.reload_rate":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.notifications.reload_rate")
	case "current_parameters.relay.mode.serving.limits.uplink_messages":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.uplink_messages")
	case "current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.uplink_messages.bucket_size")
	case "current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.uplink_messages.reload_rate")
	case "current_parameters.relay.mode.serving.limits.overall":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.overall")
	case "current_parameters.relay.mode.serving.limits.overall.bucket_size":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.overall.bucket_size")
	case "current_parameters.relay.mode.serving.limits.overall.reload_rate":
		return v.CurrentParameters.FieldIsZero("relay.mode.serving.limits.overall.reload_rate")
	case "current_parameters.rx1_data_rate_offset":
		return v.CurrentParameters.FieldIsZero("rx1_data_rate_offset")
	case "current_parameters.rx1_delay":
		return v.CurrentParameters.FieldIsZero("rx1_delay")
	case "current_parameters.rx2_data_rate_index":
		return v.CurrentParameters.FieldIsZero("rx2_data_rate_index")
	case "current_parameters.rx2_frequency":
		return v.CurrentParameters.FieldIsZero("rx2_frequency")
	case "current_parameters.uplink_dwell_time":
		return v.CurrentParameters.FieldIsZero("uplink_dwell_time")
	case "current_parameters.uplink_dwell_time.value":
		return v.CurrentParameters.FieldIsZero("uplink_dwell_time.value")
	case "desired_parameters":
		return v.DesiredParameters == nil
	case "desired_parameters.adr_ack_delay":
		return v.DesiredParameters.FieldIsZero("adr_ack_delay")
	case "desired_parameters.adr_ack_delay_exponent":
		return v.DesiredParameters.FieldIsZero("adr_ack_delay_exponent")
	case "desired_parameters.adr_ack_delay_exponent.value":
		return v.DesiredParameters.FieldIsZero("adr_ack_delay_exponent.value")
	case "desired_parameters.adr_ack_limit":
		return v.DesiredParameters.FieldIsZero("adr_ack_limit")
	case "desired_parameters.adr_ack_limit_exponent":
		return v.DesiredParameters.FieldIsZero("adr_ack_limit_exponent")
	case "desired_parameters.adr_ack_limit_exponent.value":
		return v.DesiredParameters.FieldIsZero("adr_ack_limit_exponent.value")
	case "desired_parameters.adr_data_rate_index":
		return v.DesiredParameters.FieldIsZero("adr_data_rate_index")
	case "desired_parameters.adr_nb_trans":
		return v.DesiredParameters.FieldIsZero("adr_nb_trans")
	case "desired_parameters.adr_tx_power_index":
		return v.DesiredParameters.FieldIsZero("adr_tx_power_index")
	case "desired_parameters.beacon_frequency":
		return v.DesiredParameters.FieldIsZero("beacon_frequency")
	case "desired_parameters.channels":
		return v.DesiredParameters.FieldIsZero("channels")
	case "desired_parameters.downlink_dwell_time":
		return v.DesiredParameters.FieldIsZero("downlink_dwell_time")
	case "desired_parameters.downlink_dwell_time.value":
		return v.DesiredParameters.FieldIsZero("downlink_dwell_time.value")
	case "desired_parameters.max_duty_cycle":
		return v.DesiredParameters.FieldIsZero("max_duty_cycle")
	case "desired_parameters.max_eirp":
		return v.DesiredParameters.FieldIsZero("max_eirp")
	case "desired_parameters.ping_slot_data_rate_index":
		return v.DesiredParameters.FieldIsZero("ping_slot_data_rate_index")
	case "desired_parameters.ping_slot_data_rate_index_value":
		return v.DesiredParameters.FieldIsZero("ping_slot_data_rate_index_value")
	case "desired_parameters.ping_slot_data_rate_index_value.value":
		return v.DesiredParameters.FieldIsZero("ping_slot_data_rate_index_value.value")
	case "desired_parameters.ping_slot_frequency":
		return v.DesiredParameters.FieldIsZero("ping_slot_frequency")
	case "desired_parameters.rejoin_count_periodicity":
		return v.DesiredParameters.FieldIsZero("rejoin_count_periodicity")
	case "desired_parameters.rejoin_time_periodicity":
		return v.DesiredParameters.FieldIsZero("rejoin_time_periodicity")
	case "desired_parameters.relay":
		return v.DesiredParameters.FieldIsZero("relay")
	case "desired_parameters.relay.mode":
		return v.DesiredParameters.FieldIsZero("relay.mode")
	case "desired_parameters.relay.mode.served":
		return v.DesiredParameters.FieldIsZero("relay.mode.served")
	case "desired_parameters.relay.mode.served.backoff":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.backoff")
	case "desired_parameters.relay.mode.served.mode":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.mode")
	case "desired_parameters.relay.mode.served.mode.always":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.mode.always")
	case "desired_parameters.relay.mode.served.mode.dynamic":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.mode.dynamic")
	case "desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.mode.dynamic.smart_enable_level")
	case "desired_parameters.relay.mode.served.mode.end_device_controlled":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.mode.end_device_controlled")
	case "desired_parameters.relay.mode.served.second_channel":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.second_channel")
	case "desired_parameters.relay.mode.served.second_channel.ack_offset":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.second_channel.ack_offset")
	case "desired_parameters.relay.mode.served.second_channel.data_rate_index":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.second_channel.data_rate_index")
	case "desired_parameters.relay.mode.served.second_channel.frequency":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.second_channel.frequency")
	case "desired_parameters.relay.mode.served.default_channel_index":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.default_channel_index")
	case "desired_parameters.relay.mode.served.serving_device_id":
		return v.DesiredParameters.FieldIsZero("relay.mode.served.serving_device_id")
	case "desired_parameters.relay.mode.serving":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving")
	case "desired_parameters.relay.mode.serving.second_channel":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.second_channel")
	case "desired_parameters.relay.mode.serving.second_channel.ack_offset":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.second_channel.ack_offset")
	case "desired_parameters.relay.mode.serving.second_channel.data_rate_index":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.second_channel.data_rate_index")
	case "desired_parameters.relay.mode.serving.second_channel.frequency":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.second_channel.frequency")
	case "desired_parameters.relay.mode.serving.default_channel_index":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.default_channel_index")
	case "desired_parameters.relay.mode.serving.cad_periodicity":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.cad_periodicity")
	case "desired_parameters.relay.mode.serving.uplink_forwarding_rules":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.uplink_forwarding_rules")
	case "desired_parameters.relay.mode.serving.limits":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits")
	case "desired_parameters.relay.mode.serving.limits.reset_behavior":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.reset_behavior")
	case "desired_parameters.relay.mode.serving.limits.join_requests":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.join_requests")
	case "desired_parameters.relay.mode.serving.limits.join_requests.bucket_size":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.join_requests.bucket_size")
	case "desired_parameters.relay.mode.serving.limits.join_requests.reload_rate":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.join_requests.reload_rate")
	case "desired_parameters.relay.mode.serving.limits.notifications":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.notifications")
	case "desired_parameters.relay.mode.serving.limits.notifications.bucket_size":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.notifications.bucket_size")
	case "desired_parameters.relay.mode.serving.limits.notifications.reload_rate":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.notifications.reload_rate")
	case "desired_parameters.relay.mode.serving.limits.uplink_messages":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.uplink_messages")
	case "desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.uplink_messages.bucket_size")
	case "desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.uplink_messages.reload_rate")
	case "desired_parameters.relay.mode.serving.limits.overall":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.overall")
	case "desired_parameters.relay.mode.serving.limits.overall.bucket_size":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.overall.bucket_size")
	case "desired_parameters.relay.mode.serving.limits.overall.reload_rate":
		return v.DesiredParameters.FieldIsZero("relay.mode.serving.limits.overall.reload_rate")
	case "desired_parameters.rx1_data_rate_offset":
		return v.DesiredParameters.FieldIsZero("rx1_data_rate_offset")
	case "desired_parameters.rx1_delay":
		return v.DesiredParameters.FieldIsZero("rx1_delay")
	case "desired_parameters.rx2_data_rate_index":
		return v.DesiredParameters.FieldIsZero("rx2_data_rate_index")
	case "desired_parameters.rx2_frequency":
		return v.DesiredParameters.FieldIsZero("rx2_frequency")
	case "desired_parameters.uplink_dwell_time":
		return v.DesiredParameters.FieldIsZero("uplink_dwell_time")
	case "desired_parameters.uplink_dwell_time.value":
		return v.DesiredParameters.FieldIsZero("uplink_dwell_time.value")
	case "device_class":
		return v.DeviceClass == 0
	case "last_adr_change_f_cnt_up":
		return v.LastAdrChangeFCntUp == 0
	case "last_confirmed_downlink_at":
		return v.LastConfirmedDownlinkAt == nil
	case "last_dev_status_f_cnt_up":
		return v.LastDevStatusFCntUp == 0
	case "last_downlink_at":
		return v.LastDownlinkAt == nil
	case "last_network_initiated_downlink_at":
		return v.LastNetworkInitiatedDownlinkAt == nil
	case "lorawan_version":
		return v.LorawanVersion == 0
	case "pending_application_downlink":
		return v.PendingApplicationDownlink == nil
	case "pending_application_downlink.class_b_c":
		return v.PendingApplicationDownlink.FieldIsZero("class_b_c")
	case "pending_application_downlink.class_b_c.absolute_time":
		return v.PendingApplicationDownlink.FieldIsZero("class_b_c.absolute_time")
	case "pending_application_downlink.class_b_c.gateways":
		return v.PendingApplicationDownlink.FieldIsZero("class_b_c.gateways")
	case "pending_application_downlink.confirmed":
		return v.PendingApplicationDownlink.FieldIsZero("confirmed")
	case "pending_application_downlink.correlation_ids":
		return v.PendingApplicationDownlink.FieldIsZero("correlation_ids")
	case "pending_application_downlink.confirmed_retry":
		return v.PendingApplicationDownlink.FieldIsZero("confirmed_retry")
	case "pending_application_downlink.confirmed_retry.attempt":
		return v.PendingApplicationDownlink.FieldIsZero("confirmed_retry.attempt")
	case "pending_application_downlink.confirmed_retry.max_attempts":
		return v.PendingApplicationDownlink.FieldIsZero("confirmed_retry.max_attempts")
	case "pending_application_downlink.decoded_payload":
		return v.PendingApplicationDownlink.FieldIsZero("decoded_payload")
	case "pending_application_downlink.decoded_payload_warnings":
		return v.PendingApplicationDownlink.FieldIsZero("decoded_payload_warnings")
	case "pending_application_downlink.f_cnt":
		return v.PendingApplicationDownlink.FieldIsZero("f_cnt")
	case "pending_application_downlink.f_port":
		return v.PendingApplicationDownlink.FieldIsZero("f_port")
	case "pending_application_downlink.frm_payload":
		return v.PendingApplicationDownlink.FieldIsZero("frm_payload")
	case "pending_application_downlink.priority":
		return v.PendingApplicationDownlink.FieldIsZero("priority")
	case "pending_application_downlink.session_key_id":
		return v.PendingApplicationDownlink.FieldIsZero("session_key_id")
	case "pending_join_request":
		return v.PendingJoinRequest == nil
	case "pending_join_request.cf_list":
		return v.PendingJoinRequest.FieldIsZero("cf_list")
	case "pending_join_request.cf_list.ch_masks":
		return v.PendingJoinRequest.FieldIsZero("cf_list.ch_masks")
	case "pending_join_request.cf_list.freq":
		return v.PendingJoinRequest.FieldIsZero("cf_list.freq")
	case "pending_join_request.cf_list.type":
		return v.PendingJoinRequest.FieldIsZero("cf_list.type")
	case "pending_join_request.consumed_airtime":
		return v.PendingJoinRequest.FieldIsZero("consumed_airtime")
	case "pending_join_request.correlation_ids":
		return v.PendingJoinRequest.FieldIsZero("correlation_ids")
	case "pending_join_request.dev_addr":
		return v.PendingJoinRequest.FieldIsZero("dev_addr")
	case "pending_join_request.downlink_settings":
		return v.PendingJoinRequest.FieldIsZero("downlink_settings")
	case "pending_join_request.downlink_settings.opt_neg":
		return v.PendingJoinRequest.FieldIsZero("downlink_settings.opt_neg")
	case "pending_join_request.downlink_settings.rx1_dr_offset":
		return v.PendingJoinRequest.FieldIsZero("downlink_settings.rx1_dr_offset")
	case "pending_join_request.downlink_settings.rx2_dr":
		return v.PendingJoinRequest.FieldIsZero("downlink_settings.rx2_dr")
	case "pending_join_request.net_id":
		return v.PendingJoinRequest.FieldIsZero("net_id")
	case "pending_join_request.payload":
		return v.PendingJoinRequest.FieldIsZero("payload")
	case "pending_join_request.payload.Payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload")
	case "pending_join_request.payload.Payload.join_accept_payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload")
	case "pending_join_request.payload.Payload.join_accept_payload.cf_list":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.cf_list")
	case "pending_join_request.payload.Payload.join_accept_payload.cf_list.ch_masks":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.cf_list.ch_masks")
	case "pending_join_request.payload.Payload.join_accept_payload.cf_list.freq":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.cf_list.freq")
	case "pending_join_request.payload.Payload.join_accept_payload.cf_list.type":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.cf_list.type")
	case "pending_join_request.payload.Payload.join_accept_payload.dev_addr":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.dev_addr")
	case "pending_join_request.payload.Payload.join_accept_payload.dl_settings":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.dl_settings")
	case "pending_join_request.payload.Payload.join_accept_payload.dl_settings.opt_neg":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.dl_settings.opt_neg")
	case "pending_join_request.payload.Payload.join_accept_payload.dl_settings.rx1_dr_offset":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.dl_settings.rx1_dr_offset")
	case "pending_join_request.payload.Payload.join_accept_payload.dl_settings.rx2_dr":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.dl_settings.rx2_dr")
	case "pending_join_request.payload.Payload.join_accept_payload.encrypted":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.encrypted")
	case "pending_join_request.payload.Payload.join_accept_payload.join_nonce":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.join_nonce")
	case "pending_join_request.payload.Payload.join_accept_payload.net_id":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.net_id")
	case "pending_join_request.payload.Payload.join_accept_payload.rx_delay":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_accept_payload.rx_delay")
	case "pending_join_request.payload.Payload.join_request_payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_request_payload")
	case "pending_join_request.payload.Payload.join_request_payload.dev_eui":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_request_payload.dev_eui")
	case "pending_join_request.payload.Payload.join_request_payload.dev_nonce":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_request_payload.dev_nonce")
	case "pending_join_request.payload.Payload.join_request_payload.join_eui":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.join_request_payload.join_eui")
	case "pending_join_request.payload.Payload.mac_payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload")
	case "pending_join_request.payload.Payload.mac_payload.decoded_payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.decoded_payload")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.dev_addr":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.dev_addr")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_cnt":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_cnt")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_ctrl":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_ctrl")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_ctrl.ack":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_ctrl.ack")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_ctrl.adr":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_ctrl.adr")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_ctrl.adr_ack_req":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_ctrl.adr_ack_req")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_ctrl.class_b":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_ctrl.class_b")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_ctrl.f_pending":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_ctrl.f_pending")
	case "pending_join_request.payload.Payload.mac_payload.f_hdr.f_opts":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_hdr.f_opts")
	case "pending_join_request.payload.Payload.mac_payload.f_port":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.f_port")
	case "pending_join_request.payload.Payload.mac_payload.frm_payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.frm_payload")
	case "pending_join_request.payload.Payload.mac_payload.full_f_cnt":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.mac_payload.full_f_cnt")
	case "pending_join_request.payload.Payload.rejoin_request_payload":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.rejoin_request_payload")
	case "pending_join_request.payload.Payload.rejoin_request_payload.dev_eui":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.rejoin_request_payload.dev_eui")
	case "pending_join_request.payload.Payload.rejoin_request_payload.join_eui":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.rejoin_request_payload.join_eui")
	case "pending_join_request.payload.Payload.rejoin_request_payload.net_id":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.rejoin_request_payload.net_id")
	case "pending_join_request.payload.Payload.rejoin_request_payload.rejoin_cnt":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.rejoin_request_payload.rejoin_cnt")
	case "pending_join_request.payload.Payload.rejoin_request_payload.rejoin_type":
		return v.PendingJoinRequest.FieldIsZero("payload.Payload.rejoin_request_payload.rejoin_type")
	case "pending_join_request.payload.m_hdr":
		return v.PendingJoinRequest.FieldIsZero("payload.m_hdr")
	case "pending_join_request.payload.m_hdr.m_type":
		return v.PendingJoinRequest.FieldIsZero("payload.m_hdr.m_type")
	case "pending_join_request.payload.m_hdr.major":
		return v.PendingJoinRequest.FieldIsZero("payload.m_hdr.major")
	case "pending_join_request.payload.mic":
		return v.PendingJoinRequest.FieldIsZero("payload.mic")
	case "pending_join_request.raw_payload":
		return v.PendingJoinRequest.FieldIsZero("raw_payload")
	case "pending_join_request.rx_delay":
		return v.PendingJoinRequest.FieldIsZero("rx_delay")
	case "pending_join_request.selected_mac_version":
		return v.PendingJoinRequest.FieldIsZero("selected_mac_version")
	case "pending_relay_downlink":
		return v.PendingRelayDownlink == nil
	case "pending_relay_downlink.raw_payload":
		return v.PendingRelayDownlink.FieldIsZero("raw_payload")
	case "pending_requests":
		return v.PendingRequests == nil
	case "ping_slot_periodicity":
		return v.PingSlotPeriodicity == nil
	case "ping_slot_periodicity.value":
		return v.PingSlotPeriodicity.FieldIsZero("value")
	case "queued_join_accept":
		return v.QueuedJoinAccept == nil
	case "queued_join_accept.correlation_ids":
		return v.QueuedJoinAccept.FieldIsZero("correlation_ids")
	case "queued_join_accept.dev_addr":
		return v.QueuedJoinAccept.FieldIsZero("dev_addr")
	case "queued_join_accept.keys":
		return v.QueuedJoinAccept.FieldIsZero("keys")
	case "queued_join_accept.keys.app_s_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.app_s_key")
	case "queued_join_accept.keys.app_s_key.encrypted_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.app_s_key.encrypted_key")
	case "queued_join_accept.keys.app_s_key.kek_label":
		return v.QueuedJoinAccept.FieldIsZero("keys.app_s_key.kek_label")
	case "queued_join_accept.keys.app_s_key.key":
		return v.QueuedJoinAccept.FieldIsZero("keys.app_s_key.key")
	case "queued_join_accept.keys.f_nwk_s_int_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.f_nwk_s_int_key")
	case "queued_join_accept.keys.f_nwk_s_int_key.encrypted_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.f_nwk_s_int_key.encrypted_key")
	case "queued_join_accept.keys.f_nwk_s_int_key.kek_label":
		return v.QueuedJoinAccept.FieldIsZero("keys.f_nwk_s_int_key.kek_label")
	case "queued_join_accept.keys.f_nwk_s_int_key.key":
		return v.QueuedJoinAccept.FieldIsZero("keys.f_nwk_s_int_key.key")
	case "queued_join_accept.keys.nwk_s_enc_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.nwk_s_enc_key")
	case "queued_join_accept.keys.nwk_s_enc_key.encrypted_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.nwk_s_enc_key.encrypted_key")
	case "queued_join_accept.keys.nwk_s_enc_key.kek_label":
		return v.QueuedJoinAccept.FieldIsZero("keys.nwk_s_enc_key.kek_label")
	case "queued_join_accept.keys.nwk_s_enc_key.key":
		return v.QueuedJoinAccept.FieldIsZero("keys.nwk_s_enc_key.key")
	case "queued_join_accept.keys.s_nwk_s_int_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.s_nwk_s_int_key")
	case "queued_join_accept.keys.s_nwk_s_int_key.encrypted_key":
		return v.QueuedJoinAccept.FieldIsZero("keys.s_nwk_s_int_key.encrypted_key")
	case "queued_join_accept.keys.s_nwk_s_int_key.kek_label":
		return v.QueuedJoinAccept.FieldIsZero("keys.s_nwk_s_int_key.kek_label")
	case "queued_join_accept.keys.s_nwk_s_int_key.key":
		return v.QueuedJoinAccept.FieldIsZero("keys.s_nwk_s_int_key.key")
	case "queued_join_accept.keys.session_key_id":
		return v.QueuedJoinAccept.FieldIsZero("keys.session_key_id")
	case "queued_join_accept.net_id":
		return v.QueuedJoinAccept.FieldIsZero("net_id")
	case "queued_join_accept.payload":
		return v.QueuedJoinAccept.FieldIsZero("payload")
	case "queued_join_accept.request":
		return v.QueuedJoinAccept.FieldIsZero("request")
	case "queued_join_accept.request.cf_list":
		return v.QueuedJoinAccept.FieldIsZero("request.cf_list")
	case "queued_join_accept.request.cf_list.ch_masks":
		return v.QueuedJoinAccept.FieldIsZero("request.cf_list.ch_masks")
	case "queued_join_accept.request.cf_list.freq":
		return v.QueuedJoinAccept.FieldIsZero("request.cf_list.freq")
	case "queued_join_accept.request.cf_list.type":
		return v.QueuedJoinAccept.FieldIsZero("request.cf_list.type")
	case "queued_join_accept.request.downlink_settings":
		return v.QueuedJoinAccept.FieldIsZero("request.downlink_settings")
	case "queued_join_accept.request.downlink_settings.opt_neg":
		return v.QueuedJoinAccept.FieldIsZero("request.downlink_settings.opt_neg")
	case "queued_join_accept.request.downlink_settings.rx1_dr_offset":
		return v.QueuedJoinAccept.FieldIsZero("request.downlink_settings.rx1_dr_offset")
	case "queued_join_accept.request.downlink_settings.rx2_dr":
		return v.QueuedJoinAccept.FieldIsZero("request.downlink_settings.rx2_dr")
	case "queued_join_accept.request.rx_delay":
		return v.QueuedJoinAccept.FieldIsZero("request.rx_delay")
	case "queued_responses":
		return v.QueuedResponses == nil
	case "recent_downlinks":
		return v.RecentDownlinks == nil
	case "recent_mac_command_identifiers":
		return v.RecentMacCommandIdentifiers == nil
	case "recent_uplinks":
		return v.RecentUplinks == nil
	case "rejected_adr_data_rate_indexes":
		return v.RejectedAdrDataRateIndexes == nil
	case "rejected_adr_tx_power_indexes":
		return v.RejectedAdrTxPowerIndexes == nil
	case "rejected_data_rate_ranges":
		return v.RejectedDataRateRanges == nil
	case "rejected_frequencies":
		return v.RejectedFrequencies == nil
	case "rx_windows_available":
		return !v.RxWindowsAvailable
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *Session) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "dev_addr":
		return types.MustDevAddr(v.DevAddr).OrZero().IsZero()
	case "keys":
		return fieldsAreZero(v.Keys, SessionKeysFieldPathsTopLevel...)
	case "keys.app_s_key":
		return v.GetKeys().FieldIsZero("app_s_key")
	case "keys.app_s_key.encrypted_key":
		return v.GetKeys().FieldIsZero("app_s_key.encrypted_key")
	case "keys.app_s_key.kek_label":
		return v.GetKeys().FieldIsZero("app_s_key.kek_label")
	case "keys.app_s_key.key":
		return v.GetKeys().FieldIsZero("app_s_key.key")
	case "keys.f_nwk_s_int_key":
		return v.GetKeys().FieldIsZero("f_nwk_s_int_key")
	case "keys.f_nwk_s_int_key.encrypted_key":
		return v.GetKeys().FieldIsZero("f_nwk_s_int_key.encrypted_key")
	case "keys.f_nwk_s_int_key.kek_label":
		return v.GetKeys().FieldIsZero("f_nwk_s_int_key.kek_label")
	case "keys.f_nwk_s_int_key.key":
		return v.GetKeys().FieldIsZero("f_nwk_s_int_key.key")
	case "keys.nwk_s_enc_key":
		return v.GetKeys().FieldIsZero("nwk_s_enc_key")
	case "keys.nwk_s_enc_key.encrypted_key":
		return v.GetKeys().FieldIsZero("nwk_s_enc_key.encrypted_key")
	case "keys.nwk_s_enc_key.kek_label":
		return v.GetKeys().FieldIsZero("nwk_s_enc_key.kek_label")
	case "keys.nwk_s_enc_key.key":
		return v.GetKeys().FieldIsZero("nwk_s_enc_key.key")
	case "keys.s_nwk_s_int_key":
		return v.GetKeys().FieldIsZero("s_nwk_s_int_key")
	case "keys.s_nwk_s_int_key.encrypted_key":
		return v.GetKeys().FieldIsZero("s_nwk_s_int_key.encrypted_key")
	case "keys.s_nwk_s_int_key.kek_label":
		return v.GetKeys().FieldIsZero("s_nwk_s_int_key.kek_label")
	case "keys.s_nwk_s_int_key.key":
		return v.GetKeys().FieldIsZero("s_nwk_s_int_key.key")
	case "keys.session_key_id":
		return v.GetKeys().FieldIsZero("session_key_id")
	case "last_a_f_cnt_down":
		return v.LastAFCntDown == 0
	case "last_conf_f_cnt_down":
		return v.LastConfFCntDown == 0
	case "last_f_cnt_up":
		return v.LastFCntUp == 0
	case "last_n_f_cnt_down":
		return v.LastNFCntDown == 0
	case "queued_application_downlinks":
		return v.QueuedApplicationDownlinks == nil
	case "started_at":
		return v.StartedAt == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *EndDeviceVersionIdentifiers) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "brand_id":
		return v.BrandId == ""
	case "firmware_version":
		return v.FirmwareVersion == ""
	case "hardware_version":
		return v.HardwareVersion == ""
	case "model_id":
		return v.ModelId == ""
	case "band_id":
		return v.BandId == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *LoRaAllianceProfileIdentifiers) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "vendor_id":
		return v.VendorId == 0
	case "vendor_profile_id":
		return v.VendorProfileId == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *EndDevice) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "activated_at":
		return v.ActivatedAt == nil
	case "application_server_address":
		return v.ApplicationServerAddress == ""
	case "application_server_id":
		return v.ApplicationServerId == ""
	case "application_server_kek_label":
		return v.ApplicationServerKekLabel == ""
	case "attributes":
		return v.Attributes == nil
	case "battery_percentage":
		return v.BatteryPercentage == nil
	case "claim_authentication_code":
		return v.ClaimAuthenticationCode == nil
	case "claim_authentication_code.valid_from":
		return v.ClaimAuthenticationCode.FieldIsZero("valid_from")
	case "claim_authentication_code.valid_to":
		return v.ClaimAuthenticationCode.FieldIsZero("valid_to")
	case "claim_authentication_code.value":
		return v.ClaimAuthenticationCode.FieldIsZero("value")
	case "created_at":
		return v.CreatedAt == nil
	case "description":
		return v.Description == ""
	case "downlink_margin":
		return v.DownlinkMargin == 0
	case "formatters":
		return v.Formatters == nil
	case "formatters.down_formatter":
		return v.Formatters.FieldIsZero("down_formatter")
	case "formatters.down_formatter_parameter":
		return v.Formatters.FieldIsZero("down_formatter_parameter")
	case "formatters.up_formatter":
		return v.Formatters.FieldIsZero("up_formatter")
	case "formatters.up_formatter_parameter":
		return v.Formatters.FieldIsZero("up_formatter_parameter")
	case "frequency_plan_id":
		return v.FrequencyPlanId == ""
	case "ids":
		return v.Ids == nil
	case "ids.application_ids":
		return v.Ids.FieldIsZero("application_ids")
	case "ids.application_ids.application_id":
		return v.Ids.FieldIsZero("application_ids.application_id")
	case "ids.dev_addr":
		return v.Ids.FieldIsZero("dev_addr")
	case "ids.dev_eui":
		return v.Ids.FieldIsZero("dev_eui")
	case "ids.device_id":
		return v.Ids.FieldIsZero("device_id")
	case "ids.join_eui":
		return v.Ids.FieldIsZero("join_eui")
	case "join_server_address":
		return v.JoinServerAddress == ""
	case "last_dev_nonce":
		return v.LastDevNonce == 0
	case "last_dev_status_received_at":
		return v.LastDevStatusReceivedAt == nil
	case "last_join_nonce":
		return v.LastJoinNonce == 0
	case "last_rj_count_0":
		return v.LastRjCount_0 == 0
	case "last_rj_count_1":
		return v.LastRjCount_1 == 0
	case "last_seen_at":
		return v.LastSeenAt == nil
	case "locations":
		return v.Locations == nil
	case "lorawan_phy_version":
		return v.LorawanPhyVersion == 0
	case "lorawan_version":
		return v.LorawanVersion == 0
	case "mac_settings":
		return v.MacSettings == nil
	case "mac_settings.adr":
		return v.MacSettings.FieldIsZero("adr")
	case "mac_settings.adr.mode":
		return v.MacSettings.FieldIsZero("adr.mode")
	case "mac_settings.adr.mode.static":
		return v.MacSettings.FieldIsZero("adr.mode.static")
	case "mac_settings.adr.mode.static.data_rate_index":
		return v.MacSettings.FieldIsZero("adr.mode.static.data_rate_index")
	case "mac_settings.adr.mode.static.tx_power_index":
		return v.MacSettings.FieldIsZero("adr.mode.static.tx_power_index")
	case "mac_settings.adr.mode.static.nb_trans":
		return v.MacSettings.FieldIsZero("adr.mode.static.nb_trans")
	case "mac_settings.adr.mode.dynamic":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic")
	case "mac_settings.adr.mode.dynamic.channel_steering":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.channel_steering")
	case "mac_settings.adr.mode.dynamic.channel_steering.mode":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.channel_steering.mode")
	case "mac_settings.adr.mode.dynamic.channel_steering.mode.disabled":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.channel_steering.mode.disabled")
	case "mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.channel_steering.mode.lora_narrow")
	case "mac_settings.adr.mode.dynamic.margin":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.margin")
	case "mac_settings.adr.mode.dynamic.min_data_rate_index":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.min_data_rate_index")
	case "mac_settings.adr.mode.dynamic.min_data_rate_index.value":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.min_data_rate_index.value")
	case "mac_settings.adr.mode.dynamic.max_data_rate_index":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.max_data_rate_index")
	case "mac_settings.adr.mode.dynamic.max_data_rate_index.value":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.max_data_rate_index.value")
	case "mac_settings.adr.mode.dynamic.min_tx_power_index":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.min_tx_power_index")
	case "mac_settings.adr.mode.dynamic.max_tx_power_index":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.max_tx_power_index")
	case "mac_settings.adr.mode.dynamic.min_nb_trans":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.min_nb_trans")
	case "mac_settings.adr.mode.dynamic.max_nb_trans":
		return v.MacSettings.FieldIsZero("adr.mode.dynamic.max_nb_trans")
	case "mac_settings.adr.mode.disabled":
		return v.MacSettings.FieldIsZero("adr.mode.disabled")
	case "mac_settings.adr_margin":
		return v.MacSettings.FieldIsZero("adr_margin")
	case "mac_settings.beacon_frequency":
		return v.MacSettings.FieldIsZero("beacon_frequency")
	case "mac_settings.beacon_frequency.value":
		return v.MacSettings.FieldIsZero("beacon_frequency.value")
	case "mac_settings.class_b_timeout":
		return v.MacSettings.FieldIsZero("class_b_timeout")
	case "mac_settings.class_b_c_downlink_interval":
		return v.MacSettings.FieldIsZero("class_b_c_downlink_interval")
	case "mac_settings.class_c_timeout":
		return v.MacSettings.FieldIsZero("class_c_timeout")
	case "mac_settings.desired_adr_ack_delay_exponent":
		return v.MacSettings.FieldIsZero("desired_adr_ack_delay_exponent")
	case "mac_settings.desired_adr_ack_delay_exponent.value":
		return v.MacSettings.FieldIsZero("desired_adr_ack_delay_exponent.value")
	case "mac_settings.desired_adr_ack_limit_exponent":
		return v.MacSettings.FieldIsZero("desired_adr_ack_limit_exponent")
	case "mac_settings.desired_adr_ack_limit_exponent.value":
		return v.MacSettings.FieldIsZero("desired_adr_ack_limit_exponent.value")
	case "mac_settings.desired_beacon_frequency":
		return v.MacSettings.FieldIsZero("desired_beacon_frequency")
	case "mac_settings.desired_beacon_frequency.value":
		return v.MacSettings.FieldIsZero("desired_beacon_frequency.value")
	case "mac_settings.desired_max_duty_cycle":
		return v.MacSettings.FieldIsZero("desired_max_duty_cycle")
	case "mac_settings.desired_max_duty_cycle.value":
		return v.MacSettings.FieldIsZero("desired_max_duty_cycle.value")
	case "mac_settings.desired_max_eirp":
		return v.MacSettings.FieldIsZero("desired_max_eirp")
	case "mac_settings.desired_max_eirp.value":
		return v.MacSettings.FieldIsZero("desired_max_eirp.value")
	case "mac_settings.desired_ping_slot_data_rate_index":
		return v.MacSettings.FieldIsZero("desired_ping_slot_data_rate_index")
	case "mac_settings.desired_ping_slot_data_rate_index.value":
		return v.MacSettings.FieldIsZero("desired_ping_slot_data_rate_index.value")
	case "mac_settings.desired_ping_slot_frequency":
		return v.MacSettings.FieldIsZero("desired_ping_slot_frequency")
	case "mac_settings.desired_ping_slot_frequency.value":
		return v.MacSettings.FieldIsZero("desired_ping_slot_frequency.value")
	case "mac_settings.desired_rx1_data_rate_offset":
		return v.MacSettings.FieldIsZero("desired_rx1_data_rate_offset")
	case "mac_settings.desired_rx1_data_rate_offset.value":
		return v.MacSettings.FieldIsZero("desired_rx1_data_rate_offset.value")
	case "mac_settings.desired_rx1_delay":
		return v.MacSettings.FieldIsZero("desired_rx1_delay")
	case "mac_settings.desired_rx1_delay.value":
		return v.MacSettings.FieldIsZero("desired_rx1_delay.value")
	case "mac_settings.desired_rx2_data_rate_index":
		return v.MacSettings.FieldIsZero("desired_rx2_data_rate_index")
	case "mac_settings.desired_rx2_data_rate_index.value":
		return v.MacSettings.FieldIsZero("desired_rx2_data_rate_index.value")
	case "mac_settings.desired_rx2_frequency":
		return v.MacSettings.FieldIsZero("desired_rx2_frequency")
	case "mac_settings.desired_rx2_frequency.value":
		return v.MacSettings.FieldIsZero("desired_rx2_frequency.value")
	case "mac_settings.factory_preset_frequencies":
		return v.MacSettings.FieldIsZero("factory_preset_frequencies")
	case "mac_settings.max_duty_cycle":
		return v.MacSettings.FieldIsZero("max_duty_cycle")
	case "mac_settings.max_duty_cycle.value":
		return v.MacSettings.FieldIsZero("max_duty_cycle.value")
	case "mac_settings.ping_slot_data_rate_index":
		return v.MacSettings.FieldIsZero("ping_slot_data_rate_index")
	case "mac_settings.ping_slot_data_rate_index.value":
		return v.MacSettings.FieldIsZero("ping_slot_data_rate_index.value")
	case "mac_settings.ping_slot_frequency":
		return v.MacSettings.FieldIsZero("ping_slot_frequency")
	case "mac_settings.ping_slot_frequency.value":
		return v.MacSettings.FieldIsZero("ping_slot_frequency.value")
	case "mac_settings.ping_slot_periodicity":
		return v.MacSettings.FieldIsZero("ping_slot_periodicity")
	case "mac_settings.ping_slot_periodicity.value":
		return v.MacSettings.FieldIsZero("ping_slot_periodicity.value")
	case "mac_settings.relay":
		return v.MacSettings.FieldIsZero("relay")
	case "mac_settings.relay.mode":
		return v.MacSettings.FieldIsZero("relay.mode")
	case "mac_settings.relay.mode.served":
		return v.MacSettings.FieldIsZero("relay.mode.served")
	case "mac_settings.relay.mode.served.backoff":
		return v.MacSettings.FieldIsZero("relay.mode.served.backoff")
	case "mac_settings.relay.mode.served.mode":
		return v.MacSettings.FieldIsZero("relay.mode.served.mode")
	case "mac_settings.relay.mode.served.mode.always":
		return v.MacSettings.FieldIsZero("relay.mode.served.mode.always")
	case "mac_settings.relay.mode.served.mode.dynamic":
		return v.MacSettings.FieldIsZero("relay.mode.served.mode.dynamic")
	case "mac_settings.relay.mode.served.mode.dynamic.smart_enable_level":
		return v.MacSettings.FieldIsZero("relay.mode.served.mode.dynamic.smart_enable_level")
	case "mac_settings.relay.mode.served.mode.end_device_controlled":
		return v.MacSettings.FieldIsZero("relay.mode.served.mode.end_device_controlled")
	case "mac_settings.relay.mode.served.second_channel":
		return v.MacSettings.FieldIsZero("relay.mode.served.second_channel")
	case "mac_settings.relay.mode.served.second_channel.ack_offset":
		return v.MacSettings.FieldIsZero("relay.mode.served.second_channel.ack_offset")
	case "mac_settings.relay.mode.served.second_channel.data_rate_index":
		return v.MacSettings.FieldIsZero("relay.mode.served.second_channel.data_rate_index")
	case "mac_settings.relay.mode.served.second_channel.frequency":
		return v.MacSettings.FieldIsZero("relay.mode.served.second_channel.frequency")
	case "mac_settings.relay.mode.served.serving_device_id":
		return v.MacSettings.FieldIsZero("relay.mode.served.serving_device_id")
	case "mac_settings.relay.mode.serving":
		return v.MacSettings.FieldIsZero("relay.mode.serving")
	case "mac_settings.relay.mode.serving.second_channel":
		return v.MacSettings.FieldIsZero("relay.mode.serving.second_channel")
	case "mac_settings.relay.mode.serving.second_channel.ack_offset":
		return v.MacSettings.FieldIsZero("relay.mode.serving.second_channel.ack_offset")
	case "mac_settings.relay.mode.serving.second_channel.data_rate_index":
		return v.MacSettings.FieldIsZero("relay.mode.serving.second_channel.data_rate_index")
	case "mac_settings.relay.mode.serving.second_channel.frequency":
		return v.MacSettings.FieldIsZero("relay.mode.serving.second_channel.frequency")
	case "mac_settings.relay.mode.serving.default_channel_index":
		return v.MacSettings.FieldIsZero("relay.mode.serving.default_channel_index")
	case "mac_settings.relay.mode.serving.cad_periodicity":
		return v.MacSettings.FieldIsZero("relay.mode.serving.cad_periodicity")
	case "mac_settings.relay.mode.serving.uplink_forwarding_rules":
		return v.MacSettings.FieldIsZero("relay.mode.serving.uplink_forwarding_rules")
	case "mac_settings.relay.mode.serving.limits":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits")
	case "mac_settings.relay.mode.serving.limits.reset_behavior":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.reset_behavior")
	case "mac_settings.relay.mode.serving.limits.join_requests":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.join_requests")
	case "mac_settings.relay.mode.serving.limits.join_requests.bucket_size":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.join_requests.bucket_size")
	case "mac_settings.relay.mode.serving.limits.join_requests.reload_rate":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.join_requests.reload_rate")
	case "mac_settings.relay.mode.serving.limits.notifications":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.notifications")
	case "mac_settings.relay.mode.serving.limits.notifications.bucket_size":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.notifications.bucket_size")
	case "mac_settings.relay.mode.serving.limits.notifications.reload_rate":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.notifications.reload_rate")
	case "mac_settings.relay.mode.serving.limits.uplink_messages":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.uplink_messages")
	case "mac_settings.relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.uplink_messages.bucket_size")
	case "mac_settings.relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.uplink_messages.reload_rate")
	case "mac_settings.relay.mode.serving.limits.overall":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.overall")
	case "mac_settings.relay.mode.serving.limits.overall.bucket_size":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.overall.bucket_size")
	case "mac_settings.relay.mode.serving.limits.overall.reload_rate":
		return v.MacSettings.FieldIsZero("relay.mode.serving.limits.overall.reload_rate")
	case "mac_settings.desired_relay":
		return v.MacSettings.FieldIsZero("desired_relay")
	case "mac_settings.desired_relay.mode":
		return v.MacSettings.FieldIsZero("desired_relay.mode")
	case "mac_settings.desired_relay.mode.served":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served")
	case "mac_settings.desired_relay.mode.served.backoff":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.backoff")
	case "mac_settings.desired_relay.mode.served.mode":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.mode")
	case "mac_settings.desired_relay.mode.served.mode.always":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.mode.always")
	case "mac_settings.desired_relay.mode.served.mode.dynamic":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.mode.dynamic")
	case "mac_settings.desired_relay.mode.served.mode.dynamic.smart_enable_level":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.mode.dynamic.smart_enable_level")
	case "mac_settings.desired_relay.mode.served.mode.end_device_controlled":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.mode.end_device_controlled")
	case "mac_settings.desired_relay.mode.served.second_channel":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.second_channel")
	case "mac_settings.desired_relay.mode.served.second_channel.ack_offset":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.second_channel.ack_offset")
	case "mac_settings.desired_relay.mode.served.second_channel.data_rate_index":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.second_channel.data_rate_index")
	case "mac_settings.desired_relay.mode.served.second_channel.frequency":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.second_channel.frequency")
	case "mac_settings.desired_relay.mode.served.serving_device_id":
		return v.MacSettings.FieldIsZero("desired_relay.mode.served.serving_device_id")
	case "mac_settings.desired_relay.mode.serving":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving")
	case "mac_settings.desired_relay.mode.serving.second_channel":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.second_channel")
	case "mac_settings.desired_relay.mode.serving.second_channel.ack_offset":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.second_channel.ack_offset")
	case "mac_settings.desired_relay.mode.serving.second_channel.data_rate_index":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.second_channel.data_rate_index")
	case "mac_settings.desired_relay.mode.serving.second_channel.frequency":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.second_channel.frequency")
	case "mac_settings.desired_relay.mode.serving.default_channel_index":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.default_channel_index")
	case "mac_settings.desired_relay.mode.serving.cad_periodicity":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.cad_periodicity")
	case "mac_settings.desired_relay.mode.serving.uplink_forwarding_rules":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.uplink_forwarding_rules")
	case "mac_settings.desired_relay.mode.serving.limits":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits")
	case "mac_settings.desired_relay.mode.serving.limits.reset_behavior":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.reset_behavior")
	case "mac_settings.desired_relay.mode.serving.limits.join_requests":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.join_requests")
	case "mac_settings.desired_relay.mode.serving.limits.join_requests.bucket_size":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.join_requests.bucket_size")
	case "mac_settings.desired_relay.mode.serving.limits.join_requests.reload_rate":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.join_requests.reload_rate")
	case "mac_settings.desired_relay.mode.serving.limits.notifications":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.notifications")
	case "mac_settings.desired_relay.mode.serving.limits.notifications.bucket_size":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.notifications.bucket_size")
	case "mac_settings.desired_relay.mode.serving.limits.notifications.reload_rate":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.notifications.reload_rate")
	case "mac_settings.desired_relay.mode.serving.limits.uplink_messages":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.uplink_messages")
	case "mac_settings.desired_relay.mode.serving.limits.uplink_messages.bucket_size":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.uplink_messages.bucket_size")
	case "mac_settings.desired_relay.mode.serving.limits.uplink_messages.reload_rate":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.uplink_messages.reload_rate")
	case "mac_settings.desired_relay.mode.serving.limits.overall":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.overall")
	case "mac_settings.desired_relay.mode.serving.limits.overall.bucket_size":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.overall.bucket_size")
	case "mac_settings.desired_relay.mode.serving.limits.overall.reload_rate":
		return v.MacSettings.FieldIsZero("desired_relay.mode.serving.limits.overall.reload_rate")
	case "mac_settings.resets_f_cnt":
		return v.MacSettings.FieldIsZero("resets_f_cnt")
	case "mac_settings.resets_f_cnt.value":
		return v.MacSettings.FieldIsZero("resets_f_cnt.value")
	case "mac_settings.rx1_data_rate_offset":
		return v.MacSettings.FieldIsZero("rx1_data_rate_offset")
	case "mac_settings.rx1_data_rate_offset.value":
		return v.MacSettings.FieldIsZero("rx1_data_rate_offset.value")
	case "mac_settings.rx1_delay":
		return v.MacSettings.FieldIsZero("rx1_delay")
	case "mac_settings.rx1_delay.value":
		return v.MacSettings.FieldIsZero("rx1_delay.value")
	case "mac_settings.rx2_data_rate_index":
		return v.MacSettings.FieldIsZero("rx2_data_rate_index")
	case "mac_settings.rx2_data_rate_index.value":
		return v.MacSettings.FieldIsZero("rx2_data_rate_index.value")
	case "mac_settings.rx2_frequency":
		return v.MacSettings.FieldIsZero("rx2_frequency")
	case "mac_settings.rx2_frequency.value":
		return v.MacSettings.FieldIsZero("rx2_frequency.value")
	case "mac_settings.schedule_downlinks":
		return v.MacSettings.FieldIsZero("schedule_downlinks")
	case "mac_settings.schedule_downlinks.value":
		return v.MacSettings.FieldIsZero("schedule_downlinks.value")
	case "mac_settings.status_count_periodicity":
		return v.MacSettings.FieldIsZero("status_count_periodicity")
	case "mac_settings.status_time_periodicity":
		return v.MacSettings.FieldIsZero("status_time_periodicity")
	case "mac_settings.supports_32_bit_f_cnt":
		return v.MacSettings.FieldIsZero("supports_32_bit_f_cnt")
	case "mac_settings.supports_32_bit_f_cnt.value":
		return v.MacSettings.FieldIsZero("supports_32_bit_f_cnt.value")
	case "mac_settings.use_adr":
		return v.MacSettings.FieldIsZero("use_adr")
	case "mac_settings.use_adr.value":
		return v.MacSettings.FieldIsZero("use_adr.value")
	case "mac_settings.uplink_dwell_time":
		return v.MacSettings.FieldIsZero("uplink_dwell_time")
	case "mac_settings.uplink_dwell_time.value":
		return v.MacSettings.FieldIsZero("uplink_dwell_time.value")
	case "mac_settings.downlink_dwell_time":
		return v.MacSettings.FieldIsZero("downlink_dwell_time")
	case "mac_settings.downlink_dwell_time.value":
		return v.MacSettings.FieldIsZero("downlink_dwell_time.value")
	case "mac_state":
		return v.MacState == nil
	case "max_frequency":
		return v.MaxFrequency == 0
	case "min_frequency":
		return v.MinFrequency == 0
	case "multicast":
		return !v.Multicast
	case "name":
		return v.Name == ""
	case "net_id":
		return types.MustNetID(v.NetId).OrZero().IsZero()
	case "network_server_address":
		return v.NetworkServerAddress == ""
	case "network_server_kek_label":
		return v.NetworkServerKekLabel == ""
	case "pending_mac_state":
		return v.PendingMacState == nil
	case "pending_session":
		return v.PendingSession == nil
	case "pending_session.dev_addr":
		return v.PendingSession.FieldIsZero("dev_addr")
	case "pending_session.keys":
		return v.PendingSession.FieldIsZero("keys")
	case "pending_session.keys.app_s_key":
		return v.PendingSession.FieldIsZero("keys.app_s_key")
	case "pending_session.keys.app_s_key.encrypted_key":
		return v.PendingSession.FieldIsZero("keys.app_s_key.encrypted_key")
	case "pending_session.keys.app_s_key.kek_label":
		return v.PendingSession.FieldIsZero("keys.app_s_key.kek_label")
	case "pending_session.keys.app_s_key.key":
		return v.PendingSession.FieldIsZero("keys.app_s_key.key")
	case "pending_session.keys.f_nwk_s_int_key":
		return v.PendingSession.FieldIsZero("keys.f_nwk_s_int_key")
	case "pending_session.keys.f_nwk_s_int_key.encrypted_key":
		return v.PendingSession.FieldIsZero("keys.f_nwk_s_int_key.encrypted_key")
	case "pending_session.keys.f_nwk_s_int_key.kek_label":
		return v.PendingSession.FieldIsZero("keys.f_nwk_s_int_key.kek_label")
	case "pending_session.keys.f_nwk_s_int_key.key":
		return v.PendingSession.FieldIsZero("keys.f_nwk_s_int_key.key")
	case "pending_session.keys.nwk_s_enc_key":
		return v.PendingSession.FieldIsZero("keys.nwk_s_enc_key")
	case "pending_session.keys.nwk_s_enc_key.encrypted_key":
		return v.PendingSession.FieldIsZero("keys.nwk_s_enc_key.encrypted_key")
	case "pending_session.keys.nwk_s_enc_key.kek_label":
		return v.PendingSession.FieldIsZero("keys.nwk_s_enc_key.kek_label")
	case "pending_session.keys.nwk_s_enc_key.key":
		return v.PendingSession.FieldIsZero("keys.nwk_s_enc_key.key")
	case "pending_session.keys.s_nwk_s_int_key":
		return v.PendingSession.FieldIsZero("keys.s_nwk_s_int_key")
	case "pending_session.keys.s_nwk_s_int_key.encrypted_key":
		return v.PendingSession.FieldIsZero("keys.s_nwk_s_int_key.encrypted_key")
	case "pending_session.keys.s_nwk_s_int_key.kek_label":
		return v.PendingSession.FieldIsZero("keys.s_nwk_s_int_key.kek_label")
	case "pending_session.keys.s_nwk_s_int_key.key":
		return v.PendingSession.FieldIsZero("keys.s_nwk_s_int_key.key")
	case "pending_session.keys.session_key_id":
		return v.PendingSession.FieldIsZero("keys.session_key_id")
	case "pending_session.last_a_f_cnt_down":
		return v.PendingSession.FieldIsZero("last_a_f_cnt_down")
	case "pending_session.last_conf_f_cnt_down":
		return v.PendingSession.FieldIsZero("last_conf_f_cnt_down")
	case "pending_session.last_f_cnt_up":
		return v.PendingSession.FieldIsZero("last_f_cnt_up")
	case "pending_session.last_n_f_cnt_down":
		return v.PendingSession.FieldIsZero("last_n_f_cnt_down")
	case "pending_session.queued_application_downlinks":
		return v.PendingSession.FieldIsZero("queued_application_downlinks")
	case "pending_session.started_at":
		return v.PendingSession.FieldIsZero("started_at")
	case "picture":
		return v.Picture == nil
	case "picture.embedded":
		return v.Picture.FieldIsZero("embedded")
	case "picture.embedded.data":
		return v.Picture.FieldIsZero("embedded.data")
	case "picture.embedded.mime_type":
		return v.Picture.FieldIsZero("embedded.mime_type")
	case "picture.sizes":
		return v.Picture.FieldIsZero("sizes")
	case "power_state":
		return v.PowerState == 0
	case "provisioner_id":
		return v.ProvisionerId == ""
	case "provisioning_data":
		return v.ProvisioningData == nil
	case "queued_application_downlinks":
		return v.QueuedApplicationDownlinks == nil
	case "resets_join_nonces":
		return !v.ResetsJoinNonces
	case "root_keys":
		return v.RootKeys == nil
	case "root_keys.app_key":
		return v.RootKeys.FieldIsZero("app_key")
	case "root_keys.app_key.encrypted_key":
		return v.RootKeys.FieldIsZero("app_key.encrypted_key")
	case "root_keys.app_key.kek_label":
		return v.RootKeys.FieldIsZero("app_key.kek_label")
	case "root_keys.app_key.key":
		return v.RootKeys.FieldIsZero("app_key.key")
	case "root_keys.nwk_key":
		return v.RootKeys.FieldIsZero("nwk_key")
	case "root_keys.nwk_key.encrypted_key":
		return v.RootKeys.FieldIsZero("nwk_key.encrypted_key")
	case "root_keys.nwk_key.kek_label":
		return v.RootKeys.FieldIsZero("nwk_key.kek_label")
	case "root_keys.nwk_key.key":
		return v.RootKeys.FieldIsZero("nwk_key.key")
	case "root_keys.root_key_id":
		return v.RootKeys.FieldIsZero("root_key_id")
	case "serial_number":
		return v.SerialNumber == ""
	case "service_profile_id":
		return v.ServiceProfileId == ""
	case "session":
		return v.Session == nil
	case "session.dev_addr":
		return v.Session.FieldIsZero("dev_addr")
	case "session.keys":
		return v.Session.FieldIsZero("keys")
	case "session.keys.app_s_key":
		return v.Session.FieldIsZero("keys.app_s_key")
	case "session.keys.app_s_key.encrypted_key":
		return v.Session.FieldIsZero("keys.app_s_key.encrypted_key")
	case "session.keys.app_s_key.kek_label":
		return v.Session.FieldIsZero("keys.app_s_key.kek_label")
	case "session.keys.app_s_key.key":
		return v.Session.FieldIsZero("keys.app_s_key.key")
	case "session.keys.f_nwk_s_int_key":
		return v.Session.FieldIsZero("keys.f_nwk_s_int_key")
	case "session.keys.f_nwk_s_int_key.encrypted_key":
		return v.Session.FieldIsZero("keys.f_nwk_s_int_key.encrypted_key")
	case "session.keys.f_nwk_s_int_key.kek_label":
		return v.Session.FieldIsZero("keys.f_nwk_s_int_key.kek_label")
	case "session.keys.f_nwk_s_int_key.key":
		return v.Session.FieldIsZero("keys.f_nwk_s_int_key.key")
	case "session.keys.nwk_s_enc_key":
		return v.Session.FieldIsZero("keys.nwk_s_enc_key")
	case "session.keys.nwk_s_enc_key.encrypted_key":
		return v.Session.FieldIsZero("keys.nwk_s_enc_key.encrypted_key")
	case "session.keys.nwk_s_enc_key.kek_label":
		return v.Session.FieldIsZero("keys.nwk_s_enc_key.kek_label")
	case "session.keys.nwk_s_enc_key.key":
		return v.Session.FieldIsZero("keys.nwk_s_enc_key.key")
	case "session.keys.s_nwk_s_int_key":
		return v.Session.FieldIsZero("keys.s_nwk_s_int_key")
	case "session.keys.s_nwk_s_int_key.encrypted_key":
		return v.Session.FieldIsZero("keys.s_nwk_s_int_key.encrypted_key")
	case "session.keys.s_nwk_s_int_key.kek_label":
		return v.Session.FieldIsZero("keys.s_nwk_s_int_key.kek_label")
	case "session.keys.s_nwk_s_int_key.key":
		return v.Session.FieldIsZero("keys.s_nwk_s_int_key.key")
	case "session.keys.session_key_id":
		return v.Session.FieldIsZero("keys.session_key_id")
	case "session.last_a_f_cnt_down":
		return v.Session.FieldIsZero("last_a_f_cnt_down")
	case "session.last_conf_f_cnt_down":
		return v.Session.FieldIsZero("last_conf_f_cnt_down")
	case "session.last_f_cnt_up":
		return v.Session.FieldIsZero("last_f_cnt_up")
	case "session.last_n_f_cnt_down":
		return v.Session.FieldIsZero("last_n_f_cnt_down")
	case "session.queued_application_downlinks":
		return v.Session.FieldIsZero("queued_application_downlinks")
	case "session.started_at":
		return v.Session.FieldIsZero("started_at")
	case "skip_payload_crypto":
		return !v.SkipPayloadCrypto
	case "skip_payload_crypto_override":
		return v.SkipPayloadCryptoOverride == nil
	case "supports_class_b":
		return !v.SupportsClassB
	case "supports_class_c":
		return !v.SupportsClassC
	case "supports_join":
		return !v.SupportsJoin
	case "lora_alliance_profile_ids":
		return v.LoraAllianceProfileIds == nil
	case "lora_alliance_profile_ids.vendor_id":
		return v.LoraAllianceProfileIds.FieldIsZero("vendor_id")
	case "lora_alliance_profile_ids.vendor_profile_id":
		return v.LoraAllianceProfileIds.FieldIsZero("vendor_profile_id")
	case "updated_at":
		return v.UpdatedAt == nil
	case "used_dev_nonces":
		return v.UsedDevNonces == nil
	case "version_ids":
		return v.VersionIds == nil
	case "version_ids.brand_id":
		return v.VersionIds.FieldIsZero("brand_id")
	case "version_ids.firmware_version":
		return v.VersionIds.FieldIsZero("firmware_version")
	case "version_ids.hardware_version":
		return v.VersionIds.FieldIsZero("hardware_version")
	case "version_ids.model_id":
		return v.VersionIds.FieldIsZero("model_id")
	case "version_ids.band_id":
		return v.VersionIds.FieldIsZero("band_id")
	}
	switch {
	case strings.HasPrefix(p, "mac_state."):
		return v.MacState.FieldIsZero(strings.TrimPrefix(p, "mac_state."))
	case strings.HasPrefix(p, "pending_mac_state."):
		return v.PendingMacState.FieldIsZero(strings.TrimPrefix(p, "pending_mac_state."))
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (m *UpdateEndDeviceRequest) FieldIsZero(p string) bool {
	if m == nil {
		return true
	}
	return m.EndDevice.FieldIsZero(p)
}

// FieldIsZero returns whether path p is zero.
func (m *SetEndDeviceRequest) FieldIsZero(p string) bool {
	if m == nil {
		return true
	}
	return m.EndDevice.FieldIsZero(p)
}

// All EntityType methods implement the IDStringer interface.

func (m *ResetAndGetEndDeviceRequest) EntityType() string {
	return m.GetEndDeviceIds().EntityType()
}

func (m *CreateEndDeviceRequest) EntityType() string {
	return m.GetEndDevice().EntityType()
}

func (m *UpdateEndDeviceRequest) EntityType() string {
	return m.GetEndDevice().EntityType()
}

func (m *SetEndDeviceRequest) EntityType() string {
	return m.GetEndDevice().EntityType()
}

func (m *EndDeviceTemplate) EntityType() string {
	return m.GetEndDevice().EntityType()
}

func (m *GetEndDeviceRequest) EntityType() string {
	return m.GetEndDeviceIds().EntityType()
}

func (m *EndDevice) EntityType() string {
	return m.GetIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *ResetAndGetEndDeviceRequest) IDString() string {
	return m.GetEndDeviceIds().IDString()
}

func (m *CreateEndDeviceRequest) IDString() string {
	return m.GetEndDevice().IDString()
}

func (m *UpdateEndDeviceRequest) IDString() string {
	return m.GetEndDevice().IDString()
}

func (m *SetEndDeviceRequest) IDString() string {
	return m.GetEndDevice().IDString()
}

func (m *EndDeviceTemplate) IDString() string {
	return m.GetEndDevice().IDString()
}

func (m *GetEndDeviceRequest) IDString() string {
	return m.GetEndDeviceIds().IDString()
}

func (m *EndDevice) IDString() string {
	return m.GetIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *ResetAndGetEndDeviceRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDeviceIds().ExtractRequestFields(dst)
}

func (m *CreateEndDeviceRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDevice().ExtractRequestFields(dst)
}

func (m *UpdateEndDeviceRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDevice().ExtractRequestFields(dst)
}

func (m *SetEndDeviceRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDevice().ExtractRequestFields(dst)
}

func (m *EndDeviceTemplate) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDevice().ExtractRequestFields(dst)
}

func (m *GetEndDeviceRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDeviceIds().ExtractRequestFields(dst)
}

func (m *EndDevice) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

// UpdateTimestamps sets earliest CreatedAt and latest UpdatedAt timestamps for EndDevice based on src device.
func (d *EndDevice) UpdateTimestamps(src *EndDevice) {
	if d.CreatedAt == nil || (src.CreatedAt != nil && StdTime(src.CreatedAt).Before(*StdTime(d.CreatedAt))) {
		d.CreatedAt = src.CreatedAt
	}
	if d.UpdatedAt == nil || (src.UpdatedAt != nil && StdTime(src.UpdatedAt).After(*StdTime(d.UpdatedAt))) {
		d.UpdatedAt = src.UpdatedAt
	}
}

// EndDeviceFieldPathsNestedWithoutWrappers is the set of EndDevice nested paths without the wrapper paths.
var EndDeviceFieldPathsNestedWithoutWrappers = FieldsWithoutWrappers(EndDeviceFieldPathsNested)
