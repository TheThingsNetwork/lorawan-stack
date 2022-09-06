// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package internal

import (
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	ErrCorruptedMACState = errors.DefineCorruption("corrupted_mac_state", "MAC state is corrupted")
	ErrInvalidDataRate   = errors.DefineInvalidArgument("data_rate", "invalid data rate")
	ErrInvalidPayload    = errors.DefineInvalidArgument("payload", "invalid payload")
	ErrUnknownChannel    = errors.Define("unknown_chanel", "channel is unknown")

	ErrNetworkDownlinkSlot  = errors.DefineCorruption("network_downlink_slot", "could not generate network initiated downlink slot")
	ErrUplinkChannel        = errors.DefineCorruption("uplink_channel", "channel does not allow downlinks")
	ErrSession              = errors.DefineCorruption("session", "no device session")
	ErrMACHandler           = errors.DefineCorruption("mac_handler", "missing MAC handler")
	ErrChannelDataRateRange = errors.DefineCorruption("channel_data_rate_range", "could not generate channel datarate range")
	ErrChannelMask          = errors.DefineCorruption("channel_mask", "could not generate channel mask")

	ErrTooManyChannels = errors.DefineResourceExhausted(
		"too_many_channels",
		"too many channels",
		"parameters",
		"channels_len",
		"phy_max_uplink_channels",
	)
	ErrNoUplinkFrequency = errors.DefineInvalidArgument(
		"no_uplink_frequency",
		"uplink channel with no uplink frequency",
		"parameters",
		"i",
	)
	ErrTxPowerIndexTooHigh = errors.DefineInvalidArgument(
		"tx_power_index_too_high",
		"tx power index too high",
		"desired_adr_tx_power_index",
		"phy_max_tx_power_index",
	)
	ErrInvalidDataRateIndex = errors.DefineInvalidArgument(
		"invalid_data_rate_index",
		"invalid data rate index",
		"desired_adr_data_rate_index",
	)
	ErrRejectedParameters = errors.DefineUnavailable(
		"rejected_parameters",
		"the parameters have been rejected in this session",
		"parameters",
		"data_rate_index",
		"tx_power_index",
	)
	ErrIncompatibleChannelMask = errors.DefineUnavailable(
		"incompatible_channel_mask",
		"the channel mask is not compatible with the current parameters",
		"data_rate_index",
	)
)
