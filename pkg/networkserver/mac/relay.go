// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package mac

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func secondChFields(secondCh *ttnpb.RelaySecondChannel) []any {
	if secondCh == nil {
		return nil
	}
	return []any{
		"relay_second_ch_ack_offset", secondCh.AckOffset,
		"relay_second_ch_data_rate_index", secondCh.DataRateIndex,
		"relay_second_ch_frequency", secondCh.Frequency,
	}
}

func servingRelayFields(serving *ttnpb.ServingRelayParameters) log.Fielder {
	if serving == nil {
		return log.Fields()
	}
	return log.Fields(
		append(
			secondChFields(serving.SecondChannel),
			"relay_default_ch_index", serving.DefaultChannelIndex,
			"relay_cad_periodicity", serving.CadPeriodicity,
		)...,
	)
}

func relayForwardLimitsFields(limits *ttnpb.RelayForwardLimits, prefix string) []any {
	if limits == nil {
		return nil
	}
	return []any{
		fmt.Sprintf("relay_%v_limit_bucket_size", prefix), limits.BucketSize,
		fmt.Sprintf("relay_%v_limit_reload_rate", prefix), limits.ReloadRate,
	}
}

func relayConfigureForwardLimitsFields(limits *ttnpb.ServingRelayParameters_ForwardingLimits) log.Fielder {
	if limits == nil {
		return log.Fields()
	}
	fields := []any{"relay_limit_reset_behavior", limits.ResetBehavior}
	fields = append(fields, relayForwardLimitsFields(limits.JoinRequests, "join_requests")...)
	fields = append(fields, relayForwardLimitsFields(limits.Notifications, "notifications")...)
	fields = append(fields, relayForwardLimitsFields(limits.UplinkMessages, "uplink_messages")...)
	fields = append(fields, relayForwardLimitsFields(limits.Overall, "overall")...)
	return log.Fields(fields...)
}

func servedRelayFields(served *ttnpb.ServedRelayParameters) log.Fielder {
	if served == nil {
		return log.Fields()
	}
	fields := []any{}
	switch {
	case served.GetAlways() != nil:
		fields = append(fields, "relay_mode", "always")
	case served.GetDynamic() != nil:
		fields = append(
			fields,
			"relay_mode", "dynamic",
			"relay_smart_enable_level", served.GetDynamic().SmartEnableLevel,
		)
	case served.GetEndDeviceControlled() != nil:
		fields = append(fields, "relay_mode", "end_device_controlled")
	default:
		panic("unreachable")
	}
	fields = append(fields, "relay_backoff", served.Backoff)
	fields = append(fields, secondChFields(served.SecondChannel)...)
	return log.Fields(fields...)
}
