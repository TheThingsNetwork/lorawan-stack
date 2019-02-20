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

	"go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// DefaultNetworkServerConfig is the default configuration for the NetworkServer
var DefaultNetworkServerConfig = networkserver.Config{
	DeduplicationWindow: 200 * time.Millisecond,
	CooldownWindow:      time.Second,
	DownlinkPriorities: networkserver.DownlinkPriorityConfig{
		JoinAccept:             "highest",
		MACCommands:            "highest",
		MaxApplicationDownlink: "high",
	},
	DefaultMACSettings: networkserver.MACSettingConfig{
		ADRMargin:              func(v float32) *float32 { return &v }(networkserver.DefaultADRMargin),
		Rx1Delay:               func(v ttnpb.RxDelay) *ttnpb.RxDelay { return &v }(ttnpb.RX_DELAY_5),
		ClassBTimeout:          func(v time.Duration) *time.Duration { return &v }(time.Minute),
		ClassCTimeout:          func(v time.Duration) *time.Duration { return &v }(networkserver.DefaultClassCTimeout),
		StatusTimePeriodicity:  func(v time.Duration) *time.Duration { return &v }(networkserver.DefaultStatusTimePeriodicity),
		StatusCountPeriodicity: func(v uint32) *uint32 { return &v }(networkserver.DefaultStatusCountPeriodicity),
	},
}
