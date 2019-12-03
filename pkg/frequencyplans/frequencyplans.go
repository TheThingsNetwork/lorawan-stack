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

// SubBandParameters contains duty-cycle and maximum EIRP overrides for a sub-band.
type SubBandParameters struct {
	MinFrequency uint64 `yaml:"min-frequency,omitempty"`
	MaxFrequency uint64 `yaml:"max-frequency,omitempty"`
	// DutyCycle is a fraction. A value of 0 is interpreted as 1, i.e. no duty-cycle limitation.
	DutyCycle float32  `yaml:"duty-cycle,omitempty"`
	MaxEIRP   *float32 `yaml:"max-eirp,omitempty"`
}

// Clone returns a cloned SubBandParameters.
func (sb *SubBandParameters) Clone() *SubBandParameters {
	if sb == nil {
		return nil
	}
	nsb := *sb
	if sb.MaxEIRP != nil {
		val := *sb.MaxEIRP
		nsb.MaxEIRP = &val
	}
	return &nsb
}

// Comprises returns whether the given frequency falls in the sub-band.
func (sb SubBandParameters) Comprises(frequency uint64) bool {
	return frequency >= sb.MinFrequency && frequency <= sb.MaxFrequency
}

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
	if lbt == nil {
		return nil
	}
	return &ttnpb.ConcentratorConfig_LBTConfiguration{
		RSSIOffset: lbt.RSSIOffset,
		RSSITarget: lbt.RSSITarget,
		ScanTime:   lbt.ScanTime,
	}
}

// DwellTime contains dwell time settings.
type DwellTime struct {
	Uplinks   *bool          `yaml:"uplinks,omitempty"`
	Downlinks *bool          `yaml:"downlinks,omitempty"`
	Duration  *time.Duration `yaml:"duration,omitempty"`
}

// GetUplinks returns whether the dwell time is applicable to uplinks.
func (dt *DwellTime) GetUplinks() bool {
	return dt != nil && dt.Uplinks != nil && *dt.Uplinks
}

// GetDownlinks returns whether the dwell time is applicable to downlinks.
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

// ChannelDwellTime contains dwell time settings for a channel.
type ChannelDwellTime struct {
	Enabled  *bool          `yaml:"enabled,omitempty"`
	Duration *time.Duration `yaml:"duration,omitempty"`
}

// GetEnabled returns whether dwell time is enabled.
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

// Clone returns a cloned Channel.
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
	if c == nil {
		return nil
	}
	return &ttnpb.ConcentratorConfig_Channel{
		Frequency: c.Frequency,
		Radio:     uint32(c.Radio),
	}
}

// LoRaStandardChannel contains the configuration of the LoRa standard channel.
type LoRaStandardChannel struct {
	Frequency uint64            `yaml:"frequency"`
	DwellTime *ChannelDwellTime `yaml:"dwell-time,omitempty"`
	Radio     uint8             `yaml:"radio"`
	DataRate  uint8             `yaml:"data-rate"`
}

// Clone returns a cloned LoRaStandardChannel.
func (lsc *LoRaStandardChannel) Clone() *LoRaStandardChannel {
	if lsc == nil {
		return nil
	}
	nlsc := *lsc
	nlsc.DwellTime = lsc.DwellTime.Clone()
	return &nlsc
}

// ToConcentratorConfig returns the LoRa standard channel configuration in the protobuf format.
func (lsc *LoRaStandardChannel) ToConcentratorConfig(band band.Band) *ttnpb.ConcentratorConfig_LoRaStandardChannel {
	if lsc == nil {
		return nil
	}
	dr := band.DataRates[lsc.DataRate].Rate.GetLoRa()
	return &ttnpb.ConcentratorConfig_LoRaStandardChannel{
		Frequency:       lsc.Frequency,
		Radio:           uint32(lsc.Radio),
		SpreadingFactor: dr.SpreadingFactor,
		Bandwidth:       dr.Bandwidth,
	}
}

// FSKChannel contains the configuration of an FSK channel.
type FSKChannel struct {
	Frequency uint64            `yaml:"frequency"`
	DwellTime *ChannelDwellTime `yaml:"dwell-time,omitempty"`
	Radio     uint8             `yaml:"radio"`
	DataRate  uint8             `yaml:"data-rate"`
}

// Clone returns a cloned FSKChannel.
func (fskc *FSKChannel) Clone() *FSKChannel {
	if fskc == nil {
		return nil
	}
	nfskc := *fskc
	nfskc.DwellTime = fskc.DwellTime.Clone()
	return &nfskc
}

// ToConcentratorConfig returns the FSK channel configuration in the protobuf format.
func (fskc *FSKChannel) ToConcentratorConfig() *ttnpb.ConcentratorConfig_FSKChannel {
	if fskc == nil {
		return nil
	}
	return &ttnpb.ConcentratorConfig_FSKChannel{
		Frequency: fskc.Frequency,
		Radio:     uint32(fskc.Radio),
	}
}

// TimeOffAir contains the time-off-air settings.
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

// RadioTxConfiguration contains the gateway radio transmission configuration.
type RadioTxConfiguration struct {
	MinFrequency   uint64  `yaml:"min-frequency"`
	MaxFrequency   uint64  `yaml:"max-frequency"`
	NotchFrequency *uint64 `yaml:"notch-frequency,omitempty"`
}

// Clone returns a cloned RadioTxConfiguration.
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

// Radio contains the gateway configuration of a radio.
type Radio struct {
	Enable          bool                  `yaml:"enable"`
	ChipType        string                `yaml:"chip-type,omitempty"`
	Frequency       uint64                `yaml:"frequency,omitempty"`
	RSSIOffset      float32               `yaml:"rssi-offset,omitempty"`
	TxConfiguration *RadioTxConfiguration `yaml:"tx,omitempty"`
}

// Clone returns a cloned Radio.
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

// FrequencyPlan contains a frequency plan.
type FrequencyPlan struct {
	BandID   string              `yaml:"band-id,omitempty"`
	SubBands []SubBandParameters `yaml:"sub-bands,omitempty"`

	UplinkChannels      []Channel            `yaml:"uplink-channels,omitempty"`
	DownlinkChannels    []Channel            `yaml:"downlink-channels,omitempty"`
	LoRaStandardChannel *LoRaStandardChannel `yaml:"lora-standard-channel,omitempty"`
	FSKChannel          *FSKChannel          `yaml:"fsk-channel,omitempty"`

	TimeOffAir TimeOffAir `yaml:"time-off-air,omitempty"`
	DwellTime  DwellTime  `yaml:"dwell-time,omitempty"`
	LBT        *LBT       `yaml:"listen-before-talk,omitempty"`

	Radios      []Radio `yaml:"radios,omitempty"`
	ClockSource uint8   `yaml:"clock-source,omitempty"`

	// PingSlot overrides the default band settings for the class B ping slot.
	PingSlot                *Channel `yaml:"ping-slot,omitempty"`
	DefaultPingSlotDataRate *uint8   `yaml:"ping-slot-default-data-rate,omitempty"`
	// Rx2Channel overrides the default band settings for Rx2.
	Rx2Channel         *Channel `yaml:"rx2-channel,omitempty"`
	DefaultRx2DataRate *uint8   `yaml:"rx2-default-data-rate,omitempty"`
	// MaxEIRP is the maximum EIRP as ceiling for any (sub-)band value.
	MaxEIRP *float32 `yaml:"max-eirp,omitempty"`
}

// Extend returns the same frequency plan, with values overridden by the passed frequency plan.
func (fp FrequencyPlan) Extend(ext FrequencyPlan) FrequencyPlan {
	if ext.BandID != "" {
		fp.BandID = ext.BandID
	}
	if channels := ext.UplinkChannels; len(channels) > 0 {
		fp.UplinkChannels = make([]Channel, 0, len(channels))
		for _, ch := range channels {
			fp.UplinkChannels = append(fp.UplinkChannels, *ch.Clone())
		}
	}
	if channels := ext.DownlinkChannels; len(channels) > 0 {
		fp.DownlinkChannels = make([]Channel, 0, len(channels))
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
		fp.Radios = make([]Radio, 0, len(radios))
		for _, r := range radios {
			fp.Radios = append(fp.Radios, *r.Clone())
		}
		fp.ClockSource = ext.ClockSource
	}
	if ext.PingSlot != nil {
		fp.PingSlot = ext.PingSlot.Clone()
	}
	if ext.DefaultPingSlotDataRate != nil {
		var i uint8
		i = *ext.DefaultPingSlotDataRate
		fp.DefaultPingSlotDataRate = &i
	}
	if ext.Rx2Channel != nil {
		fp.Rx2Channel = ext.Rx2Channel.Clone()
	}
	if ext.DefaultRx2DataRate != nil {
		var i uint8
		i = *ext.DefaultRx2DataRate
		fp.DefaultRx2DataRate = &i
	}
	if subBands := ext.SubBands; len(subBands) > 0 {
		fp.SubBands = make([]SubBandParameters, 0, len(subBands))
		for _, sb := range subBands {
			fp.SubBands = append(fp.SubBands, *sb.Clone())
		}
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
	var chDwellTime *ChannelDwellTime
	var channels []Channel
	if isDownlink {
		channels = fp.DownlinkChannels
	} else {
		channels = fp.UplinkChannels
	}
	for _, ch := range channels {
		if ch.Frequency == frequency {
			chDwellTime = ch.DwellTime
			break
		}
	}
	if chDwellTime == nil && fp.LoRaStandardChannel != nil && fp.LoRaStandardChannel.Frequency == frequency {
		chDwellTime = fp.LoRaStandardChannel.DwellTime
	}
	if chDwellTime == nil && fp.FSKChannel != nil && fp.FSKChannel.Frequency == frequency {
		chDwellTime = fp.FSKChannel.DwellTime
	}
	fpdtEnabled := isDownlink && fp.DwellTime.GetDownlinks() || !isDownlink && fp.DwellTime.GetUplinks()
	if fpdtEnabled && (chDwellTime == nil || chDwellTime.Enabled == nil) || chDwellTime.GetEnabled() {
		var dwellTime time.Duration
		if chDwellTime != nil && chDwellTime.Duration != nil {
			dwellTime = *chDwellTime.Duration
		} else {
			dwellTime = *fp.DwellTime.Duration
		}
		return duration <= dwellTime
	}
	return true
}

// ToConcentratorConfig returns the frequency plan in the protobuf format.
func (fp *FrequencyPlan) ToConcentratorConfig() (*ttnpb.ConcentratorConfig, error) {
	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, err
	}
	cc := &ttnpb.ConcentratorConfig{}
	for _, channel := range fp.UplinkChannels {
		cc.Channels = append(cc.Channels, channel.ToConcentratorConfig())
	}
	cc.LoRaStandardChannel = fp.LoRaStandardChannel.ToConcentratorConfig(band)
	cc.FSKChannel = fp.FSKChannel.ToConcentratorConfig()
	cc.LBT = fp.LBT.ToConcentratorConfig()
	cc.PingSlot = fp.PingSlot.ToConcentratorConfig()
	for _, radio := range fp.Radios {
		cc.Radios = append(cc.Radios, radio.ToConcentratorConfig())
	}
	cc.ClockSource = uint32(fp.ClockSource)
	return cc, nil
}

// FindSubBand returns the sub-band by frequency, if any.
func (fp *FrequencyPlan) FindSubBand(frequency uint64) (SubBandParameters, bool) {
	for _, sb := range fp.SubBands {
		if sb.Comprises(frequency) {
			return sb, true
		}
	}
	return SubBandParameters{}, false
}

// FrequencyPlanDescription describes a frequency plan in the YAML format.
type FrequencyPlanDescription struct {
	// ID is the unique identifier of the frequency plan.
	ID string `yaml:"id"`
	// BaseID is the ID of the base frequency plan that this frequency plan extends (optional).
	BaseID string `yaml:"base-id,omitempty"`
	// Name is a human readable name of the frequency plan.
	Name string `yaml:"name"`
	// BaseFrequency is the base frequency of the frequency plan (i.e. 868, 915)
	BaseFrequency uint16 `yaml:"base-frequency"`
	// File is the file where the frequency plan is defined.
	File string `yaml:"file"`
}

var errFetchFailed = errors.Define("fetch", "fetching failed")

func (d FrequencyPlanDescription) content(f fetch.Interface) ([]byte, error) {
	content, err := f.File(d.File)
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
	var descriptions []*FrequencyPlanDescription
	if err = yaml.Unmarshal(content, &descriptions); err != nil {
		return nil, errParseFile.WithCause(err)
	}
	descriptionsByID := make(map[string]*FrequencyPlanDescription, len(descriptions))
	for _, description := range descriptions {
		descriptionsByID[description.ID] = description
	}
	for _, description := range descriptions {
		if description.BaseID != "" {
			base := descriptionsByID[description.BaseID]
			if description.BaseFrequency == 0 {
				description.BaseFrequency = base.BaseFrequency
			}
		}
	}
	frequencyPlanList := make(frequencyPlanList, len(descriptions))
	for i, description := range descriptions {
		frequencyPlanList[i] = *description
	}
	return frequencyPlanList, nil
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
	errNotConfigured = errors.Define("not_configured", "frequency plans not configured")
	errRead          = errors.Define("read", "could not read frequency plan `{id}`")
	errReadBase      = errors.Define("read_base", "could not read the base `{base_id}` of frequency plan `{id}`")
	errReadList      = errors.Define("read_list", "could not read the list of frequency plans")
	errNotFound      = errors.DefineNotFound("not_found", "frequency plan `{id}` not found")
	errInvalid       = errors.DefineCorruption("invalid", "invalid frequency plan")
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
	if s == nil {
		return nil, errNotConfigured
	}

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
	if s == nil {
		return nil, errNotConfigured
	}

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
