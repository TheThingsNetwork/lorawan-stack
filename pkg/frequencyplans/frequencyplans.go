// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package frequencyplans contains abstractions to fetch and manipulate frequency plans.
package frequencyplans

import (
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	yaml "gopkg.in/yaml.v2"
)

const yamlFetchErrorCache = 1 * time.Minute

// LBT contains the listen-before-talk requirements for a region.
type LBT struct {
	RSSITarget float32       `yaml:"rssi-target"`
	RSSIOffset float32       `yaml:"rssi-offset,omitempty"`
	ScanTime   time.Duration `yaml:"scan-time"`
}

// Clone returns a cloned LBT.
func (lbt *LBT) Clone() *LBT {
	if lbt == nil {
		return nil
	}
	nlbt := *lbt
	return &nlbt
}

// ToConcentratorConfig returns the LBT configuration in the protobuf format.
func (lbt *LBT) ToConcentratorConfig() *ttnpb.ConcentratorConfig_LBTConfiguration {
	return &ttnpb.ConcentratorConfig_LBTConfiguration{
		RSSIOffset: lbt.RSSIOffset,
		RSSITarget: lbt.RSSITarget,
		ScanTime:   lbt.ScanTime,
	}
}

// DwellTime contains the dwell time devices must abide to for a region.
type DwellTime struct {
	Uplinks   *bool          `yaml:"uplinks,omitempty"`
	Downlinks *bool          `yaml:"downlinks,omitempty"`
	Duration  *time.Duration `yaml:"duration,omitempty"`
}

// GetUplinks returns whether the dwell time is enabled on uplinks.
func (dt *DwellTime) GetUplinks() bool {
	return dt != nil && dt.Uplinks != nil && *dt.Uplinks
}

// GetDownlinks returns whether the dwell time is enabled on downlinks.
func (dt *DwellTime) GetDownlinks() bool {
	return dt != nil && dt.Downlinks != nil && *dt.Downlinks
}

// Clone returns a cloned DwellTime.
func (dt *DwellTime) Clone() *DwellTime {
	if dt == nil {
		return nil
	}
	ndt := *dt
	if dt.Uplinks != nil {
		val := *dt.Uplinks
		ndt.Uplinks = &val
	}
	if dt.Downlinks != nil {
		val := *dt.Downlinks
		ndt.Downlinks = &val
	}
	if dt.Duration != nil {
		val := *dt.Duration
		ndt.Duration = &val
	}
	return &ndt
}

type ChannelDwellTime struct {
	Enabled  *bool          `yaml:"enabled,omitempty"`
	Duration *time.Duration `yaml:"duration,omitempty"`
}

// GetEnabled returns whether the dwell time is enabled.
func (dt *ChannelDwellTime) GetEnabled() bool {
	return dt != nil && dt.Enabled != nil && *dt.Enabled
}

// Clone returns a cloned ChannelDwellTime.
func (dt *ChannelDwellTime) Clone() *ChannelDwellTime {
	if dt == nil {
		return nil
	}
	ndt := *dt
	if dt.Enabled != nil {
		val := *dt.Enabled
		ndt.Enabled = &val
	}
	if dt.Duration != nil {
		val := *dt.Duration
		ndt.Duration = &val
	}
	return &ndt
}

// Channel contains the configuration of a channel.
type Channel struct {
	Frequency   uint64            `yaml:"frequency"`
	DwellTime   *ChannelDwellTime `yaml:"dwell-time,omitempty"`
	Radio       uint8             `yaml:"radio"`
	MinDataRate uint8             `yaml:"min-data-rate"`
	MaxDataRate uint8             `yaml:"max-data-rate"`
}

// Copy returns an identical channel configuration.
func (c *Channel) Clone() *Channel {
	if c == nil {
		return nil
	}
	nc := *c
	nc.DwellTime = c.DwellTime.Clone()
	return &nc
}

// ToConcentratorConfig returns the channel configuration in the protobuf format.
func (c *Channel) ToConcentratorConfig() *ttnpb.ConcentratorConfig_Channel {
	return &ttnpb.ConcentratorConfig_Channel{
		Frequency: c.Frequency,
		Radio:     uint32(c.Radio),
	}
}

// LoRaStandardChannel contains the configuration of the LoRa standard channel on a gateway.
type LoRaStandardChannel struct {
	Channel
	Bandwidth       uint32 `yaml:"bandwidth"`
	SpreadingFactor uint32 `yaml:"spreading-factor"`
}

// Clone returns a cloned LoRaStandardChannel.
func (lsc *LoRaStandardChannel) Clone() *LoRaStandardChannel {
	if lsc == nil {
		return nil
	}
	nlsc := *lsc
	nlsc.Channel = *lsc.Channel.Clone()
	return &nlsc
}

// ToConcentratorConfig returns the LoRa standard channel configuration in the protobuf format.
func (lsc *LoRaStandardChannel) ToConcentratorConfig() *ttnpb.ConcentratorConfig_LoRaStandardChannel {
	return &ttnpb.ConcentratorConfig_LoRaStandardChannel{
		ConcentratorConfig_Channel: *lsc.Channel.ToConcentratorConfig(),
		Bandwidth:                  lsc.Bandwidth,
		SpreadingFactor:            lsc.SpreadingFactor,
	}
}

// FSKChannel contains the configuration of the FSKChannel on a gateway.
type FSKChannel struct {
	Channel
	Bandwidth uint32 `yaml:"bandwidth"`
	BitRate   uint32 `yaml:"bit-rate"`
}

// Clone returns a cloned FSKChannel.
func (fskc *FSKChannel) Clone() *FSKChannel {
	if fskc == nil {
		return nil
	}
	nfskc := *fskc
	nfskc.Channel = *fskc.Channel.Clone()
	return &nfskc
}

// ToConcentratorConfig returns the FSK channel configuration in the protobuf format.
func (fskc *FSKChannel) ToConcentratorConfig() *ttnpb.ConcentratorConfig_FSKChannel {
	return &ttnpb.ConcentratorConfig_FSKChannel{
		ConcentratorConfig_Channel: *fskc.Channel.ToConcentratorConfig(),
		Bandwidth:                  fskc.Bandwidth,
		BitRate:                    fskc.BitRate,
	}
}

// TimeOffAir contains the time-off-air regulations that emissions must abide to.
type TimeOffAir struct {
	Fraction float32       `yaml:"fraction,omitempty"`
	Duration time.Duration `yaml:"duration,omitempty"`
}

// Clone returns a cloned TimeOffAir.
func (toa *TimeOffAir) Clone() *TimeOffAir {
	if toa == nil {
		return nil
	}
	ntoa := *toa
	return &ntoa
}

// RadioTxConfiguration contains the Tx emission-configuration of a radio on a gateway.
type RadioTxConfiguration struct {
	MinFrequency   uint64  `yaml:"min-frequency"`
	MaxFrequency   uint64  `yaml:"max-frequency"`
	NotchFrequency *uint64 `yaml:"notch-frequency,omitempty"`
}

func (txc *RadioTxConfiguration) Clone() *RadioTxConfiguration {
	if txc == nil {
		return nil
	}
	ntxc := *txc
	if txc.NotchFrequency != nil {
		val := *txc.NotchFrequency
		ntxc.NotchFrequency = &val
	}
	return &ntxc
}

// Radio contains the configuration of a radio on a gateway.
type Radio struct {
	Enable          bool                  `yaml:"enable"`
	ChipType        string                `yaml:"chip-type,omitempty"`
	Frequency       uint64                `yaml:"frequency,omitempty"`
	RSSIOffset      float32               `yaml:"rssi-offset,omitempty"`
	TxConfiguration *RadioTxConfiguration `yaml:"tx,omitempty"`
}

func (r *Radio) Clone() *Radio {
	if r == nil {
		return nil
	}
	nr := *r
	nr.TxConfiguration = r.TxConfiguration.Clone()
	return &nr
}

// ToConcentratorConfig returns the radio configuration in the protobuf format.
func (r Radio) ToConcentratorConfig() *ttnpb.GatewayRadio {
	ccr := &ttnpb.GatewayRadio{
		Enable:     r.Enable,
		Frequency:  r.Frequency,
		ChipType:   r.ChipType,
		RSSIOffset: r.RSSIOffset,
	}
	if tx := r.TxConfiguration; tx != nil {
		ccr.TxConfiguration = &ttnpb.GatewayRadio_TxConfiguration{
			MinFrequency: tx.MinFrequency,
			MaxFrequency: tx.MaxFrequency,
		}
		if tx.NotchFrequency != nil {
			ccr.TxConfiguration.NotchFrequency = *tx.NotchFrequency
		}
	}
	return ccr
}

// FrequencyPlan contains the local regulations and settings for a region.
type FrequencyPlan struct {
	BandID string `yaml:"band-id,omitempty"`

	UplinkChannels      []Channel            `yaml:"uplink-channels,omitempty"`
	DownlinkChannels    []Channel            `yaml:"downlink-channels,omitempty"`
	LoRaStandardChannel *LoRaStandardChannel `yaml:"lora-standard-channel,omitempty"`
	FSKChannel          *FSKChannel          `yaml:"fsk-channel,omitempty"`

	TimeOffAir TimeOffAir `yaml:"time-off-air,omitempty"`
	DwellTime  DwellTime  `yaml:"dwell-time,omitempty"`
	LBT        *LBT       `yaml:"listen-before-talk,omitempty"`

	Radios      []Radio `yaml:"radios,omitempty"`
	ClockSource uint8   `yaml:"clock-source,omitempty"`

	// PingSlot allows override of default band settings for the class B ping slot.
	PingSlot *Channel `yaml:"ping-slot,omitempty"`
	// Rx2 allows override of default band settings for Rx2.
	Rx2     *Channel `yaml:"rx2,omitempty"`
	MaxEIRP *float32 `yaml:"max-eirp,omitempty"`
}

// Extend returns the same frequency plan, with values overridden by the passed frequency plan.
func (fp FrequencyPlan) Extend(ext FrequencyPlan) FrequencyPlan {
	if ext.BandID != "" {
		fp.BandID = ext.BandID
	}
	if channels := ext.UplinkChannels; len(channels) > 0 {
		fp.UplinkChannels = []Channel{}
		for _, ch := range channels {
			fp.UplinkChannels = append(fp.UplinkChannels, *ch.Clone())
		}
	}
	if channels := ext.DownlinkChannels; len(channels) > 0 {
		fp.DownlinkChannels = []Channel{}
		for _, ch := range channels {
			fp.DownlinkChannels = append(fp.DownlinkChannels, *ch.Clone())
		}
	}
	if ext.LoRaStandardChannel != nil {
		fp.LoRaStandardChannel = ext.LoRaStandardChannel.Clone()
	}
	if ext.FSKChannel != nil {
		fp.FSKChannel = ext.FSKChannel.Clone()
	}
	if ext.TimeOffAir != (TimeOffAir{}) {
		fp.TimeOffAir = *ext.TimeOffAir.Clone()
	}
	if ext.DwellTime != (DwellTime{}) {
		fp.DwellTime = *ext.DwellTime.Clone()
	}
	if ext.LBT != nil {
		fp.LBT = ext.LBT.Clone()
	}
	if radios := ext.Radios; len(radios) > 0 {
		fp.Radios = []Radio{}
		for _, r := range radios {
			fp.Radios = append(fp.Radios, *r.Clone())
		}
		fp.ClockSource = ext.ClockSource
	}
	if ext.PingSlot != nil {
		fp.PingSlot = ext.PingSlot.Clone()
	}
	if ext.Rx2 != nil {
		fp.Rx2 = ext.Rx2.Clone()
	}
	if ext.MaxEIRP != nil {
		val := *ext.MaxEIRP
		fp.MaxEIRP = &val
	}
	return fp
}

var (
	errNoDwellTimeDuration = errors.DefineInvalidArgument("no_dwell_time_duration", "no dwell time duration specified")
	errInvalidChannel      = errors.Define("channel", "invalid frequency plan channel `{index}`")
)

// Validate returns an error if the frequency plan is invalid.
func (fp FrequencyPlan) Validate() error {
	_, err := band.GetByID(fp.BandID)
	if err != nil {
		return err
	}
	fpdt := fp.DwellTime
	if (fpdt.GetUplinks() || fpdt.GetDownlinks()) && fpdt.Duration == nil {
		return errNoDwellTimeDuration
	}
	for _, channels := range [][]Channel{fp.UplinkChannels, fp.DownlinkChannels} {
		for i, channel := range channels {
			chdt := channel.DwellTime
			if chdt == nil || chdt.Enabled == nil {
				continue
			}
			if *chdt.Enabled && fpdt.Duration == nil && chdt.Duration == nil {
				return errInvalidChannel.WithAttributes("index", i).WithCause(errNoDwellTimeDuration)
			}
		}
	}
	return nil
}

// RespectsDwellTime returns whether the transmission respects the frequency plan's dwell time restrictions.
func (fp *FrequencyPlan) RespectsDwellTime(isDownlink bool, frequency uint64, duration time.Duration) bool {
	var channels []Channel
	if isDownlink {
		channels = fp.DownlinkChannels
	} else {
		channels = fp.UplinkChannels
	}
	allChannels := make([]Channel, len(channels), len(channels)+2)
	copy(allChannels, channels)
	if fp.LoRaStandardChannel != nil {
		allChannels = append(allChannels, fp.LoRaStandardChannel.Channel)
	}
	if fp.FSKChannel != nil {
		allChannels = append(allChannels, fp.FSKChannel.Channel)
	}
	fpdtEnabled := isDownlink && fp.DwellTime.GetDownlinks() || !isDownlink && fp.DwellTime.GetUplinks()
	for _, ch := range allChannels {
		if ch.Frequency != frequency {
			continue
		}
		if fpdtEnabled && (ch.DwellTime == nil || ch.DwellTime.Enabled == nil) || ch.DwellTime.GetEnabled() {
			var dwellTime time.Duration
			if ch.DwellTime != nil && ch.DwellTime.Duration != nil {
				dwellTime = *ch.DwellTime.Duration
			} else {
				dwellTime = *fp.DwellTime.Duration
			}
			return duration <= dwellTime
		}
		return true
	}
	if !fpdtEnabled {
		return true
	}
	return duration <= *fp.DwellTime.Duration
}

// ToConcentratorConfig returns the frequency plan in the protobuf format.
func (fp *FrequencyPlan) ToConcentratorConfig() *ttnpb.ConcentratorConfig {
	cc := &ttnpb.ConcentratorConfig{}
	for _, channel := range fp.UplinkChannels {
		cc.Channels = append(cc.Channels, channel.ToConcentratorConfig())
	}
	if fp.LoRaStandardChannel != nil {
		cc.LoRaStandardChannel = fp.LoRaStandardChannel.ToConcentratorConfig()
	}
	if fp.FSKChannel != nil {
		cc.FSKChannel = fp.FSKChannel.ToConcentratorConfig()
	}
	if fp.LBT != nil {
		cc.LBT = fp.LBT.ToConcentratorConfig()
	}
	if fp.PingSlot != nil {
		cc.PingSlot = fp.PingSlot.ToConcentratorConfig()
	}
	for _, radio := range fp.Radios {
		cc.Radios = append(cc.Radios, radio.ToConcentratorConfig())
	}
	cc.ClockSource = uint32(fp.ClockSource)
	return cc
}

// FrequencyPlanDescription describes a frequency plan in the YAML format.
type FrequencyPlanDescription struct {
	// ID to identify the frequency plan.
	ID string `yaml:"id"`
	// Description of the frequency plan.
	Description string `yaml:"description"`
	// BaseFrequency in Mhz.
	BaseFrequency uint16 `yaml:"base-frequency"`
	// Filename of the frequency plan within the repo.
	Filename string `yaml:"file"`
	// BaseID is the ID of the frequency plan that's the basis for this extended frequency plan.
	BaseID string `yaml:"base,omitempty"`
}

var errFetchFailed = errors.Define("fetch", "fetching failed")

func (d FrequencyPlanDescription) content(f fetch.Interface) ([]byte, error) {
	content, err := f.File(d.Filename)
	if err != nil {
		return nil, errFetchFailed.WithCause(err)
	}
	return content, nil
}

var errParseFile = errors.DefineCorruption("parse_file", "could not parse file")

func (d FrequencyPlanDescription) proto(f fetch.Interface) (FrequencyPlan, error) {
	fp := FrequencyPlan{}
	content, err := d.content(f)
	if err != nil {
		return fp, err
	}
	if err := yaml.Unmarshal(content, &fp); err != nil {
		return fp, errParseFile.WithCause(err)
	}
	return fp, nil
}

type frequencyPlanList []FrequencyPlanDescription

func (l frequencyPlanList) get(id string) (FrequencyPlanDescription, bool) {
	for _, f := range l {
		if f.ID == id {
			return f, true
		}
	}

	return FrequencyPlanDescription{}, false
}

type queryResult struct {
	fp   *FrequencyPlan
	err  error
	time time.Time
}

// Store contains frequency plans.
type Store struct {
	Fetcher fetch.Interface

	descriptionsMu             sync.Mutex
	descriptionsCache          frequencyPlanList
	descriptionsFetchErrorTime time.Time
	descriptionsFetchError     error

	frequencyPlansCache map[string]queryResult
	frequencyPlansMu    sync.Mutex
}

// NewStore of frequency plans.
func NewStore(fetcher fetch.Interface) *Store {
	return &Store{
		Fetcher:             fetcher,
		frequencyPlansCache: map[string]queryResult{},
	}
}

func (s *Store) fetchDescriptions() (frequencyPlanList, error) {
	content, err := s.Fetcher.File("frequency-plans.yml")
	if err != nil {
		return nil, errFetchFailed.WithCause(err)
	}
	descriptions := frequencyPlanList{}
	if err = yaml.Unmarshal(content, &descriptions); err != nil {
		return nil, errParseFile.WithCause(err)
	}
	return descriptions, nil
}

func (s *Store) descriptions() (frequencyPlanList, error) {
	s.descriptionsMu.Lock()
	defer s.descriptionsMu.Unlock()
	if s.descriptionsCache != nil {
		return s.descriptionsCache, nil
	}
	if time.Since(s.descriptionsFetchErrorTime) < yamlFetchErrorCache {
		return nil, s.descriptionsFetchError
	}
	descriptions, err := s.fetchDescriptions()
	if err != nil {
		s.descriptionsFetchError = err
		s.descriptionsFetchErrorTime = time.Now()
		return nil, err
	}
	s.descriptionsFetchErrorTime = time.Time{}
	s.descriptionsFetchError = nil
	s.descriptionsCache = descriptions
	return descriptions, nil
}

var (
	errRead     = errors.Define("read", "could not read frequency plan `{id}`")
	errReadBase = errors.Define("read_base", "could not read the base `{base_id}` of frequency plan `{id}`")
	errReadList = errors.Define("read_list", "could not read the list of frequency plans")
	errNotFound = errors.DefineNotFound("not_found", "frequency plan `{id}` not found")
	errInvalid  = errors.DefineCorruption("invalid", "invalid frequency plan")
)

func (s *Store) getByID(id string) (*FrequencyPlan, error) {
	descriptions, err := s.descriptions()
	if err != nil {
		return nil, errReadList.WithCause(err)
	}
	description, ok := descriptions.get(id)
	if !ok {
		return nil, errNotFound.WithAttributes("id", id)
	}
	proto, err := description.proto(s.Fetcher)
	if err != nil {
		return nil, errRead.WithCause(err).WithAttributes("id", id)
	}
	if description.BaseID != "" {
		base, ok := descriptions.get(description.BaseID)
		if !ok {
			return nil, errReadBase.WithCause(errNotFound.WithAttributes("id", description.BaseID)).WithAttributes(
				"id", description.ID,
				"base_id", description.BaseID,
			)
		}
		var baseProto FrequencyPlan
		baseProto, err = base.proto(s.Fetcher)
		if err != nil {
			return nil, errReadBase.WithCause(err).WithAttributes(
				"id", description.ID,
				"base_id", description.BaseID,
			)
		}
		proto = baseProto.Extend(proto)
	}
	if err := proto.Validate(); err != nil {
		return nil, errInvalid.WithCause(err)
	}
	return &proto, nil
}

// GetByID retrieves the frequency plan that has the given ID.
func (s *Store) GetByID(id string) (*FrequencyPlan, error) {
	if id == "" {
		return nil, errNotFound.WithAttributes("id", id)
	}

	s.frequencyPlansMu.Lock()
	defer s.frequencyPlansMu.Unlock()
	if cached, ok := s.frequencyPlansCache[id]; ok && cached.err == nil || time.Since(cached.time) < yamlFetchErrorCache {
		return cached.fp, cached.err
	}
	fp, err := s.getByID(id)
	s.frequencyPlansCache[id] = queryResult{
		time: time.Now(),
		fp:   fp,
		err:  err,
	}
	return fp, err
}

// GetAllIDs returns the list of IDs of the available frequency plans.
func (s *Store) GetAllIDs() ([]string, error) {
	descriptions, err := s.descriptions()
	if err != nil {
		return nil, errReadList.WithCause(err)
	}

	ids := []string{}
	for _, description := range descriptions {
		ids = append(ids, description.ID)
	}

	return ids, nil
}
