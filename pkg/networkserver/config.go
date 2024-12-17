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
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ApplicationUplinkQueueConfig defines application uplink queue configuration.
type ApplicationUplinkQueueConfig struct {
	Queue        ApplicationUplinkQueue `name:"-"`
	BufferSize   uint64                 `name:"buffer-size"`
	NumConsumers uint64                 `name:"num-consumers"`

	FastBufferSize   uint64 `name:"fast-buffer-size"`
	FastNumConsumers uint64 `name:"fast-num-consumers"`
}

// ApplicationUplinkQueueConfig defines downlink task queue configuration.
type DownlinkTaskQueueConfig struct {
	Queue        DownlinkTaskQueue `name:"-"`
	NumConsumers uint64            `name:"num-consumers"`
}

// MACSettingConfig defines MAC-layer configuration.
type MACSettingConfig struct {
	ADRMargin                  *float32                   `name:"adr-margin" description:"The default margin Network Server should add in ADR requests if not configured in device's MAC settings"`
	DesiredRx1Delay            *ttnpb.RxDelay             `name:"desired-rx1-delay" description:"Desired Rx1Delay value Network Server should use if not configured in device's MAC settings"`
	DesiredMaxDutyCycle        *ttnpb.AggregatedDutyCycle `name:"desired-max-duty-cycle" description:"Desired MaxDutyCycle value Network Server should use if not configured in device's MAC settings"`
	DesiredADRAckLimitExponent *ttnpb.ADRAckLimitExponent `name:"desired-adr-ack-limit-exponent" description:"Desired ADR_ACK_LIMIT value Network Server should use if not configured in device's MAC settings"`
	DesiredADRAckDelayExponent *ttnpb.ADRAckDelayExponent `name:"desired-adr-ack-delay-exponent" description:"Desired ADR_ACK_DELAY value Network Server should use if not configured in device's MAC settings"`
	ClassBTimeout              *time.Duration             `name:"class-b-timeout" description:"Deadline for a device in class B mode to respond to requests from the Network Server if not configured in device's MAC settings"`
	ClassCTimeout              *time.Duration             `name:"class-c-timeout" description:"Deadline for a device in class C mode to respond to requests from the Network Server if not configured in device's MAC settings"`
	StatusTimePeriodicity      *time.Duration             `name:"status-time-periodicity" description:"The interval after which a DevStatusReq MACCommand shall be sent by Network Server if not configured in device's MAC settings"`
	StatusCountPeriodicity     *uint32                    `name:"status-count-periodicity" description:"Number of uplink messages after which a DevStatusReq MACCommand shall be sent by Network Server if not configured in device's MAC settings"`
}

// Parse parses the configuration and returns ttnpb.MACSettings.
func (c MACSettingConfig) Parse() (*ttnpb.MACSettings, error) {
	p := &ttnpb.MACSettings{
		ClassBTimeout:         ttnpb.ProtoDuration(c.ClassBTimeout),
		ClassCTimeout:         ttnpb.ProtoDuration(c.ClassCTimeout),
		StatusTimePeriodicity: ttnpb.ProtoDuration(c.StatusTimePeriodicity),
	}
	if c.ADRMargin != nil {
		p.AdrMargin = &wrapperspb.FloatValue{Value: *c.ADRMargin}
	}
	if c.DesiredRx1Delay != nil {
		p.DesiredRx1Delay = &ttnpb.RxDelayValue{Value: *c.DesiredRx1Delay}
	}
	if c.DesiredMaxDutyCycle != nil {
		p.DesiredMaxDutyCycle = &ttnpb.AggregatedDutyCycleValue{Value: *c.DesiredMaxDutyCycle}
	}
	if c.DesiredADRAckLimitExponent != nil {
		p.DesiredAdrAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: *c.DesiredADRAckLimitExponent}
	}
	if c.DesiredADRAckDelayExponent != nil {
		p.DesiredAdrAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: *c.DesiredADRAckDelayExponent}
	}
	if c.StatusCountPeriodicity != nil {
		p.StatusCountPeriodicity = &wrapperspb.UInt32Value{Value: *c.StatusCountPeriodicity}
	}
	if err := p.ValidateFields(); err != nil {
		return nil, err
	}
	return p, nil
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

// InteropConfig represents interoperability client configuration.
type InteropConfig struct {
	config.InteropClient `name:",squash"`
	ID                   *types.EUI64 `name:"id" description:"NSID of this Network Server (EUI)"`
}

// PaginationConfig represents the configuration for pagination.
type PaginationConfig struct {
	DefaultLimit int64 `name:"default-limit" description:"Default limit for pagination"`
}

// Config represents the NetworkServer configuration.
type Config struct {
	ApplicationUplinkQueue     ApplicationUplinkQueueConfig `name:"application-uplink-queue"`
	Devices                    DeviceRegistry               `name:"-"`
	DownlinkTaskQueue          DownlinkTaskQueueConfig      `name:"downlink-task-queue"`
	UplinkDeduplicator         UplinkDeduplicator           `name:"-"`
	ScheduledDownlinkMatcher   ScheduledDownlinkMatcher     `name:"-"`
	NetID                      types.NetID                  `name:"net-id" description:"NetID of this Network Server"`                                                                                   // nolint: lll
	ClusterID                  string                       `name:"cluster-id" description:"Cluster ID of this Network Server"`                                                                          // nolint: lll
	DevAddrPrefixes            []types.DevAddrPrefix        `name:"dev-addr-prefixes" description:"Device address prefixes of this Network Server"`                                                      // nolint: lll
	DeduplicationWindow        time.Duration                `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`                          // nolint: lll
	CooldownWindow             time.Duration                `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"` // nolint: lll
	DownlinkPriorities         DownlinkPriorityConfig       `name:"downlink-priorities" description:"Downlink message priorities"`                                                                       // nolint: lll
	DefaultMACSettings         MACSettingConfig             `name:"default-mac-settings" description:"Default MAC settings to fallback to if not specified by device, band or frequency plan"`           // nolint: lll
	Interop                    InteropConfig                `name:"interop" description:"Interop client configuration"`                                                                                  // nolint: lll
	DeviceKEKLabel             string                       `name:"device-kek-label" description:"Label of KEK used to encrypt device keys at rest"`                                                     // nolint: lll
	DownlinkQueueCapacity      int                          `name:"downlink-queue-capacity" description:"Maximum downlink queue size per-session"`                                                       // nolint: lll
	MACSettingsProfileRegistry MACSettingsProfileRegistry   `name:"-"`
	Pagination                 PaginationConfig             `name:"pagination" description:"Pagination configuration"`
}

// DefaultConfig is the default Network Server configuration.
var DefaultConfig = Config{
	ApplicationUplinkQueue: ApplicationUplinkQueueConfig{
		BufferSize:   1000,
		NumConsumers: 1,

		FastBufferSize:   16384,
		FastNumConsumers: 128,
	},
	DownlinkTaskQueue: DownlinkTaskQueueConfig{
		NumConsumers: 1,
	},
	DeduplicationWindow: 200 * time.Millisecond,
	CooldownWindow:      time.Second,
	DownlinkPriorities: DownlinkPriorityConfig{
		JoinAccept:             "highest",
		MACCommands:            "highest",
		MaxApplicationDownlink: "high",
	},
	DefaultMACSettings: MACSettingConfig{
		ADRMargin:              func(v float32) *float32 { return &v }(mac.DefaultADRMargin),
		DesiredRx1Delay:        func(v ttnpb.RxDelay) *ttnpb.RxDelay { return &v }(ttnpb.RxDelay_RX_DELAY_5),
		ClassBTimeout:          func(v time.Duration) *time.Duration { return &v }(mac.DefaultClassBTimeout),
		ClassCTimeout:          func(v time.Duration) *time.Duration { return &v }(mac.DefaultClassCTimeout),
		StatusTimePeriodicity:  func(v time.Duration) *time.Duration { return &v }(mac.DefaultStatusTimePeriodicity),
		StatusCountPeriodicity: func(v uint32) *uint32 { return &v }(mac.DefaultStatusCountPeriodicity),
	},
	DownlinkQueueCapacity: 10000,
}
