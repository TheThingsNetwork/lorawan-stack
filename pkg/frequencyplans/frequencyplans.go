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
	"fmt"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	yaml "gopkg.in/yaml.v2"
)

const yamlFetchErrorCache = 1 * time.Minute

// LBT contains the listen-before-talk requirements for a region.
type LBT struct {
	RSSITarget float32 `yaml:"rssi-target"`
	RSSIOffset float32 `yaml:"rssi-offset,omitempty"`

	// ScanTime in microseconds.
	ScanTime int32 `yaml:"scan-time"`
}

// Copy returns an identical LBT configuration.
func (lbt *LBT) Copy() *LBT {
	return &LBT{
		RSSIOffset: lbt.RSSIOffset,
		RSSITarget: lbt.RSSITarget,
		ScanTime:   lbt.ScanTime,
	}
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
	Uplinks   *bool `yaml:"uplinks,omitempty"`
	Downlinks *bool `yaml:"downlinks,omitempty"`

	Duration *time.Duration `yaml:"duration,omitempty"`
}

// OnUplinks returns if the dwell time configuration applies on uplinks.
func (dt DwellTime) OnUplinks() bool {
	return dt.Uplinks != nil && *dt.Uplinks
}

// OnDownlinks returns if the dwell time configuration applies on uplinks.
func (dt DwellTime) OnDownlinks() bool {
	return dt.Downlinks != nil && *dt.Downlinks
}

// Copy returns an identical dwell time configuration.
func (dt *DwellTime) Copy() *DwellTime {
	ndt := &DwellTime{}
	if dt.Uplinks != nil {
		ndt.Uplinks = (&(*dt.Uplinks))
	}
	if dt.Downlinks != nil {
		ndt.Downlinks = (&(*dt.Downlinks))
	}
	if dt.Duration != nil {
		ndt.Duration = (&(*dt.Duration))
	}
	return ndt
}

// Extend returns the same dwell time, with values overridden by the passed dwell time.
func (dt DwellTime) Extend(ext DwellTime) DwellTime {
	if ext.Uplinks != nil {
		dt.Uplinks = (&(*ext.Uplinks))
	}
	if ext.Downlinks != nil {
		dt.Downlinks = (&(*ext.Downlinks))
	}
	if ext.Duration != nil {
		dt.Duration = (&(*ext.Duration))
	}
	return dt
}

// Channel contains the configuration of a channel to emit on.
type Channel struct {
	Frequency uint64     `yaml:"frequency"`
	DwellTime *DwellTime `yaml:"dwell-time,omitempty"`
	Radio     uint8      `yaml:"radio"`
}

// Copy returns an identical channel configuration.
func (c *Channel) Copy() *Channel {
	nc := &Channel{
		Frequency: c.Frequency,
		Radio:     c.Radio,
	}
	if c.DwellTime != nil {
		nc.DwellTime = c.DwellTime.Copy()
	}
	return nc
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

// Copy returns an identical LoRa standard channel configuration.
func (lsc *LoRaStandardChannel) Copy() *LoRaStandardChannel {
	return &LoRaStandardChannel{
		Channel:         *lsc.Channel.Copy(),
		Bandwidth:       lsc.Bandwidth,
		SpreadingFactor: lsc.SpreadingFactor,
	}
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

// Copy returns an identical FSK channel configuration.
func (fskc *FSKChannel) Copy() *FSKChannel {
	return &FSKChannel{
		Channel:   *fskc.Channel.Copy(),
		Bandwidth: fskc.Bandwidth,
		BitRate:   fskc.BitRate,
	}
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

// Extend returns the same time-off-air configuration, with values overridden by the passed configuration.
func (toa TimeOffAir) Extend(ext TimeOffAir) TimeOffAir {
	if ext.Duration > time.Duration(0) {
		toa.Duration = ext.Duration
	}
	if ext.Fraction > 0.0 {
		toa.Fraction = ext.Fraction
	}
	return toa
}

// RadioTxConfiguration contains the Tx emission-configuration of a radio on a gateway.
type RadioTxConfiguration struct {
	MinFrequency   uint64  `yaml:"min-frequency"`
	MaxFrequency   uint64  `yaml:"max-frequency"`
	NotchFrequency *uint64 `yaml:"notch-frequency,omitempty"`
}

// Copy returns an identical Tx configuration.
func (txc *RadioTxConfiguration) Copy() *RadioTxConfiguration {
	ntxc := *txc
	if txc.NotchFrequency != nil {
		ntxc.NotchFrequency = (&(*txc.NotchFrequency))
	}
	return &ntxc
}

// Radio contains the configuration of a radio on a gateway.
type Radio struct {
	Enable     bool    `yaml:"enable"`
	ChipType   string  `yaml:"chip-type,omitempty"`
	Frequency  uint64  `yaml:"frequency,omitempty"`
	RSSIOffset float32 `yaml:"rssi-offset,omitempty"`

	TxConfiguration *RadioTxConfiguration `yaml:"tx,omitempty"`
}

// Copy returns an identical radio configuration.
func (r Radio) Copy() Radio {
	nr := r
	if r.TxConfiguration != nil {
		nr.TxConfiguration = r.TxConfiguration.Copy()
	}
	return nr
}

// ToConcentratorConfig returns the radio configuration in the protobuf format.
func (r Radio) ToConcentratorConfig() *ttnpb.ConcentratorConfig_Radio {
	ccr := &ttnpb.ConcentratorConfig_Radio{
		Enable:     r.Enable,
		Frequency:  r.Frequency,
		ChipType:   r.ChipType,
		RSSIOffset: r.RSSIOffset,
	}
	if tx := r.TxConfiguration; tx != nil {
		ccr.TxConfiguration = &ttnpb.ConcentratorConfig_Radio_TxConfiguration{
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

	Channels            []Channel            `yaml:"channels,omitempty"`
	LoRaStandardChannel *LoRaStandardChannel `yaml:"lora-standard-channel,omitempty"`
	FSKChannel          *FSKChannel          `yaml:"fsk-channel,omitempty"`

	TimeOffAir TimeOffAir `yaml:"time-off-air,omitempty"`
	DwellTime  DwellTime  `yaml:"dwell-time,omitempty"`
	LBT        *LBT       `yaml:"listen-before-talk,omitempty"`

	Radios []Radio `yaml:"radios,omitempty"`

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
	if channels := ext.Channels; len(channels) > 0 {
		fp.Channels = []Channel{}
		for _, channel := range channels {
			fp.Channels = append(fp.Channels, *channel.Copy())
		}
	}
	if ext.LoRaStandardChannel != nil {
		fp.LoRaStandardChannel = ext.LoRaStandardChannel.Copy()
	}
	if ext.FSKChannel != nil {
		fp.FSKChannel = ext.FSKChannel.Copy()
	}
	fp.TimeOffAir = fp.TimeOffAir.Extend(ext.TimeOffAir)
	fp.DwellTime = fp.DwellTime.Extend(ext.DwellTime)
	if ext.LBT != nil {
		fp.LBT = ext.LBT.Copy()
	}
	if radios := ext.Radios; len(radios) > 0 {
		fp.Radios = []Radio{}
		for _, radio := range radios {
			fp.Radios = append(fp.Radios, radio.Copy())
		}
	}
	if ext.PingSlot != nil {
		fp.PingSlot = ext.PingSlot.Copy()
	}
	if ext.Rx2 != nil {
		fp.Rx2 = ext.Rx2.Copy()
	}
	if ext.MaxEIRP != nil {
		fp.MaxEIRP = (&(*ext.MaxEIRP))
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
	if fpdt.Duration == nil && (fpdt.OnUplinks() || fpdt.OnDownlinks()) {
		return errNoDwellTimeDuration
	}
	for i, channel := range fp.Channels {
		if channel.DwellTime == nil {
			continue
		}
		dt := channel.DwellTime
		if fpdt.Duration == nil && dt.Duration == nil &&
			(channel.DwellTime.OnUplinks() || channel.DwellTime.OnDownlinks()) {
			return errInvalidChannel.WithAttributes("index", i).WithCause(errNoDwellTimeDuration)
		}
	}
	return nil
}

// RespectsDwellTime returns whether the transmission respects the frequency plan's dwell time restrictions.
func (fp *FrequencyPlan) RespectsDwellTime(isDownlink bool, frequency uint64, duration time.Duration) bool {
	isUplink := !isDownlink
	fpdt := fp.DwellTime
	channels := append(fp.Channels)
	if fp.LoRaStandardChannel != nil {
		channels = append(channels, fp.LoRaStandardChannel.Channel)
	}
	if fp.FSKChannel != nil {
		channels = append(channels, fp.FSKChannel.Channel)
	}
	for index, ch := range channels {
		if ch.Frequency != frequency {
			continue
		}
		chdt := ch.DwellTime
		uplinks := chdt == nil && fpdt.OnUplinks() || chdt.OnUplinks()
		downlinks := chdt == nil && fpdt.OnDownlinks() || chdt.OnDownlinks()
		if isDownlink && !downlinks || isUplink && !uplinks {
			return true
		}
		var dtDuration time.Duration
		switch {
		case chdt != nil && chdt.Duration != nil:
			dtDuration = *chdt.Duration
		case fpdt.Duration != nil:
			dtDuration = *fpdt.Duration
		default:
			panic(fmt.Sprintf("frequency plan has dwell time enabled, but no dwell time duration set for channel %d with frequency %d", index, frequency))
		}
		return duration <= dtDuration
	}
	if isDownlink && fpdt.OnDownlinks() || isUplink && fpdt.OnUplinks() {
		return duration <= *fpdt.Duration
	}
	return true
}

// ToConcentratorConfig returns the frequency plan in the protobuf format.
func (fp *FrequencyPlan) ToConcentratorConfig() *ttnpb.ConcentratorConfig {
	cc := &ttnpb.ConcentratorConfig{}
	for _, channel := range fp.Channels {
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
	return cc
}

// FrequencyPlanDescription describes a frequency plan in the YAML format.
type FrequencyPlanDescription struct {
	// ID to identify the frequency plan.
	ID string `yaml:"id"`
	// Description of the frequency plan.
	Description string `yaml:"description"`
	// BaseFrequency in Mhz.
	BaseFrequency uint16 `yaml:"base_freq"`
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
	fp   FrequencyPlan
	err  error
	time time.Time
}

// Store of frequency plans.
type Store struct {
	// Fetcher is the fetch.Interface used to retrieve data.
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
		Fetcher: fetcher,

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

// getByID returns the frequency plan associated to that ID.
func (s *Store) getByID(id string) (proto FrequencyPlan, err error) {
	descriptions, err := s.descriptions()
	if err != nil {
		return FrequencyPlan{}, errReadList.WithCause(err)
	}

	description, ok := descriptions.get(id)
	if !ok {
		return FrequencyPlan{}, errNotFound.WithAttributes("id", id)
	}

	proto, err = description.proto(s.Fetcher)
	if err != nil {
		return proto, errRead.WithCause(err).WithAttributes("id", id)
	}

	if description.BaseID != "" {
		base, ok := descriptions.get(description.BaseID)
		if !ok {
			return FrequencyPlan{}, errReadBase.WithCause(
				errNotFound.WithAttributes("id", description.BaseID),
			).WithAttributes(
				"id", description.ID,
				"base_id", description.BaseID,
			)
		}

		var baseProto FrequencyPlan
		baseProto, err = base.proto(s.Fetcher)
		if err != nil {
			return FrequencyPlan{}, errReadBase.WithCause(err).WithAttributes(
				"id", description.ID,
				"base_id", description.BaseID,
			)
		}

		proto = baseProto.Extend(proto)
	}

	err = proto.Validate()
	if err != nil {
		return proto, errInvalid.WithCause(err)
	}
	return proto, nil
}

// GetByID tries to retrieve the frequency plan that has the given ID, and returns an error otherwise.
func (s *Store) GetByID(id string) (FrequencyPlan, error) {
	if id == "" {
		return FrequencyPlan{}, errNotFound.WithAttributes("id", id)
	}

	s.frequencyPlansMu.Lock()
	defer s.frequencyPlansMu.Unlock()
	if cached, ok := s.frequencyPlansCache[id]; ok && cached.err == nil || time.Since(cached.time) < yamlFetchErrorCache {
		return cached.fp, cached.err
	}
	proto, err := s.getByID(id)
	s.frequencyPlansCache[id] = queryResult{
		time: time.Now(),
		fp:   proto,
		err:  err,
	}

	return proto, err
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
