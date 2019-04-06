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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Config represents the NetworkServer configuration.
type Config struct {
	Devices             DeviceRegistry         `name:"-"`
	DownlinkTasks       DownlinkTaskQueue      `name:"-"`
	NetID               types.NetID            `name:"net-id" description:"NetID of this Network Server"`
	DevAddrPrefixes     []types.DevAddrPrefix  `name:"dev-addr-prefixes" description:"DevAddrPrefixes of this Network Server"`
	DeduplicationWindow time.Duration          `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`
	CooldownWindow      time.Duration          `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"`
	DownlinkPriorities  DownlinkPriorityConfig `name:"downlink-priorities" description:"Downlink message priorities"`
	DefaultMACSettings  MACSettingConfig       `name:"default-mac-settings" description:"Default MAC settings to fallback to if not specified by device, band or frequency plan"`
}

// MACSettingConfig defines MAC-layer configuration.
type MACSettingConfig struct {
	ADRMargin              *float32       `name:"adr-margin" description:"The default margin Network Server should add in ADR requests if not configured in device's MAC settings"`
	DesiredRx1Delay        *ttnpb.RxDelay `name:"desired-rx1-delay" description:"Desired Rx1Delay value Network Server should use if not configured in device's MAC settings"`
	ClassBTimeout          *time.Duration `name:"class-b-timeout" description:"Deadline for a device in class B mode to respond to requests from the Network Server if not configured in device's MAC settings"`
	ClassCTimeout          *time.Duration `name:"class-c-timeout" description:"Deadline for a device in class C mode to respond to requests from the Network Server if not configured in device's MAC settings"`
	StatusTimePeriodicity  *time.Duration `name:"status-time-periodicity" description:"The interval after which a DevStatusReq MACCommand shall be sent by Network Server if not configured in device's MAC settings"`
	StatusCountPeriodicity *uint32        `name:"status-count-periodicity" description:"Number of uplink messages after which a DevStatusReq MACCommand shall be sent by Network Server if not configured in device's MAC settings"`
}

// DownlinkPriorityConfig defines priorities for downlink messages.
type DownlinkPriorityConfig struct {
	// JoinAccept is the downlink priority for join-accept messages.
	JoinAccept string `name:"join-accept" description:"Priority for join-accept messages (lowest, low, below_normal, normal, above_normal, high, highest)"`
	// MACCommands is the downlink priority for downlink messages with MAC commands as FRMPayload (FPort = 0) or as FOpts.
	// If the MAC commands are carried in FOpts, the highest priority of this value and the concerning application
	// downlink message's priority is used.
	MACCommands string `name:"mac-commands" description:"Priority for messages carrying MAC commands (lowest, low, below_normal, normal, above_normal, high, highest)"`
	// MaxApplicationDownlink is the highest priority permitted by the Network Server for application downlink.
	MaxApplicationDownlink string `name:"max-application-downlink" description:"Maximum priority for application downlink messages (lowest, low, below_normal, normal, above_normal, high, highest)"`
}

var downlinkPriorityConfigTable = map[string]ttnpb.TxSchedulePriority{
	"":             ttnpb.TxSchedulePriority_NORMAL,
	"lowest":       ttnpb.TxSchedulePriority_LOWEST,
	"low":          ttnpb.TxSchedulePriority_LOW,
	"below_normal": ttnpb.TxSchedulePriority_BELOW_NORMAL,
	"normal":       ttnpb.TxSchedulePriority_NORMAL,
	"above_normal": ttnpb.TxSchedulePriority_ABOVE_NORMAL,
	"high":         ttnpb.TxSchedulePriority_HIGH,
	"highest":      ttnpb.TxSchedulePriority_HIGHEST,
}

var errDownlinkPriority = errors.DefineInvalidArgument("downlink_priority", "invalid downlink priority `{value}`")

// Parse attempts to parse the configuration and returns a DownlinkPriorities.
func (c DownlinkPriorityConfig) Parse() (DownlinkPriorities, error) {
	var p DownlinkPriorities
	var ok bool
	if p.JoinAccept, ok = downlinkPriorityConfigTable[c.JoinAccept]; !ok {
		return DownlinkPriorities{}, errDownlinkPriority.WithAttributes("value", c.JoinAccept)
	}
	if p.MACCommands, ok = downlinkPriorityConfigTable[c.MACCommands]; !ok {
		return DownlinkPriorities{}, errDownlinkPriority.WithAttributes("value", c.MACCommands)
	}
	if p.MaxApplicationDownlink, ok = downlinkPriorityConfigTable[c.MaxApplicationDownlink]; !ok {
		return DownlinkPriorities{}, errDownlinkPriority.WithAttributes("value", c.MaxApplicationDownlink)
	}
	return p, nil
}
