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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func durationPtr(v time.Duration) *time.Duration {
	return &v
}

// DefaultNetworkServerConfig is the default configuration for the NetworkServer
var DefaultNetworkServerConfig = networkserver.Config{
	DeduplicationWindow: 200 * time.Millisecond,
	CooldownWindow:      time.Second,
	DownlinkPriorities: networkserver.DownlinkPriorityConfig{
		JoinAccept:             "highest",
		MACCommands:            "highest",
		MaxApplicationDownlink: "high",
	},
	DefaultMACSettings: ttnpb.MACSettings{
		UseADR:                 &pbtypes.BoolValue{Value: true},
		ADRMargin:              &pbtypes.FloatValue{Value: networkserver.DefaultADRMargin},
		ClassBTimeout:          durationPtr(time.Minute),
		ClassCTimeout:          durationPtr(networkserver.DefaultClassCTimeout),
		StatusTimePeriodicity:  durationPtr(networkserver.DefaultStatusTimePeriodicity),
		StatusCountPeriodicity: &pbtypes.UInt32Value{Value: networkserver.DefaultStatusCountPeriodicity},
		Rx1Delay:               &ttnpb.MACSettings_RxDelayValue{Value: ttnpb.RX_DELAY_5},
	},
}
