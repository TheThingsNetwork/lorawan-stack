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

// Package band contains structs to handle regional bands
package band

import (
	"math"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// PayloadSizer abstracts the acceptable payload size depending on contextual parameters
type PayloadSizer interface {
	PayloadSize(emptyFOpt, dwellTime bool) uint16
}

type maxPayloadSize struct {
	EmptyFOpt    uint16
	NotEmptyFOpt uint16
}

func (p maxPayloadSize) PayloadSize(emptyFOpt, _ bool) uint16 {
	if emptyFOpt {
		return p.EmptyFOpt
	}
	return p.NotEmptyFOpt
}

type dwellTimePayloadSize struct {
	NoDwellTime uint16
	DwellTime   uint16
}

func (p dwellTimePayloadSize) PayloadSize(_, dwellTime bool) uint16 {
	if dwellTime {
		return p.DwellTime
	}
	return p.NoDwellTime
}

// DataRate indicates the properties of a band's data rate
type DataRate struct {
	Rate              types.DataRate
	DefaultMaxSize    PayloadSizer
	NoRepeaterMaxSize PayloadSizer
}

// Channel abstracts a band's channel properties
type Channel struct {
	// Frequency indicates the frequency of the channel
	Frequency uint64
	// DataRateIndexes indicates the data rates accepted on this channel
	DataRateIndexes []int
}

// Rx2Parameters contains downlink datarate index and channel
type Rx2Parameters struct {
	DataRateIndex ttnpb.DataRateIndex
	Frequency     uint64
}

type versionSwap func(b Band) Band

func bandIdentity(b Band) Band {
	return b
}

func channelIndexIdentity(idx uint32) (uint32, error) {
	return idx, nil
}

func channelIndexModulo(n uint32) func(uint32) (uint32, error) {
	return func(idx uint32) (uint32, error) {
		return idx % n, nil
	}
}

// Beacon parameters of a specific band.
type Beacon struct {
	DataRateIndex    int
	CodingRate       string
	InvertedPolarity bool
	// Channel returns in Hz on which beaconing is performed.
	//
	// beaconTime is the integer value, converted in float64, of the 4 bytes “Time” field of the beacon frame.
	BroadcastChannel func(beaconTime float64) uint32
	PingSlotChannels []uint32
}

// Band contains a band's properties
type Band struct {
	ID string

	Beacon Beacon

	// UplinkChannels by default
	UplinkChannels []Channel
	// DownlinkChannels by default
	DownlinkChannels []Channel

	BandDutyCycles []DutyCycle

	DataRates [16]DataRate

	ImplementsCFList bool
	CFListType       ttnpb.CFListType

	// ReceiveDelay1 is the default Rx1 window timing in seconds
	ReceiveDelay1 time.Duration
	// ReceiveDelay2 is the default Rx2 window timing in seconds (ReceiveDelay1 + 1s)
	ReceiveDelay2 time.Duration

	// ReceiveDelay1 is the default join-accept window timing in seconds
	JoinAcceptDelay1 time.Duration
	// ReceiveDelay2 is the join-accept window timing in seconds
	JoinAcceptDelay2 time.Duration
	// MaxFCntGap
	MaxFCntGap uint
	// ADRAckLimit
	ADRAckLimit uint8
	// ADRAckDelay
	ADRAckDelay   uint8
	MinAckTimeout time.Duration
	MaxAckTimeout time.Duration

	// TxOffset in dB: A Tx's power is computed by taking the MaxEIRP (default: +16dBm) and subtracting the offset
	TxOffset [16]float32

	TxParamSetupReqSupport bool

	// DefaultMaxEIRP in dBm
	DefaultMaxEIRP float32

	// Rx1Channel computes the Rx1 channel index.
	Rx1Channel func(uint32) (uint32, error)
	// Rx1DataRate computes the Rx1 data rate index.
	Rx1DataRate func(ttnpb.DataRateIndex, uint32, bool) (ttnpb.DataRateIndex, error)

	// GenerateChMasks generates a mapping ChMaskCntl -> ChMask.
	// Length of chs must be equal to the maximum number of channels defined for the particular band.
	// Meaning of chs is as follows: for i in range 0..len(chs) if chs[i] == true,
	// then channel with index i should be enabled, otherwise it should be disabled.
	GenerateChMasks func(chs []bool) (map[uint8][16]bool, error)
	// ParseChMask computes the channels that have to be masked given ChMask mask and ChMaskCntl cntl.
	ParseChMask func(mask [16]bool, cntl uint8) (map[uint8]bool, error)

	// DefaultRx2Parameters are the default parameters that determine the settings for a Tx sent during Rx2
	DefaultRx2Parameters Rx2Parameters

	regionalParameters1_0       versionSwap
	regionalParameters1_0_1     versionSwap
	regionalParameters1_0_2RevA versionSwap
	regionalParameters1_0_2RevB versionSwap
	regionalParameters1_1RevA   versionSwap
}

// DutyCycle for the [MinFrequency;MaxFrequency] sub-band
type DutyCycle struct {
	MinFrequency uint64
	MaxFrequency uint64
	DutyCycle    float32
}

// Comprises returns whether the duty cycle applies to that channel
func (d DutyCycle) Comprises(channel uint64) bool {
	return channel >= d.MinFrequency && channel < d.MaxFrequency
}

// MaxEmissionDuring the period passed as parameter, that is allowed by that duty cycle.
func (d DutyCycle) MaxEmissionDuring(period time.Duration) time.Duration {
	return time.Duration(d.DutyCycle * float32(period))
}

// All contains all the bands available
var All = make(map[string]Band)

// GetByID returns the band if it was found, and returns an error otherwise
func GetByID(id string) (Band, error) {
	if band, ok := All[id]; ok {
		return band, nil
	}
	return Band{}, errBandNotFound.WithAttributes("band_id", id)
}

type swapParameters struct {
	version   ttnpb.PHYVersion
	downgrade versionSwap
}

func (b Band) downgrades() []swapParameters {
	return []swapParameters{
		{version: ttnpb.PHY_V1_1_REV_B, downgrade: bandIdentity},
		{version: ttnpb.PHY_V1_1_REV_A, downgrade: b.regionalParameters1_1RevA},
		{version: ttnpb.PHY_V1_0_2_REV_B, downgrade: b.regionalParameters1_0_2RevB},
		{version: ttnpb.PHY_V1_0_2_REV_A, downgrade: b.regionalParameters1_0_2RevA},
		{version: ttnpb.PHY_V1_0_1, downgrade: b.regionalParameters1_0_1},
		{version: ttnpb.PHY_V1_0, downgrade: b.regionalParameters1_0},
	}
}

// Version returns the band parameters for a given version.
func (b Band) Version(wantedVersion ttnpb.PHYVersion) (Band, error) {
	supportedRegionalParameters := []string{}
	for _, swapParameter := range b.downgrades() {
		if swapParameter.downgrade == nil {
			return b, errUnsupportedLoRaWANRegionalParameters.WithAttributes("supported", strings.Join(supportedRegionalParameters, ", "))
		}
		supportedRegionalParameters = append(supportedRegionalParameters, swapParameter.version.String())
		b = swapParameter.downgrade(b)
		if swapParameter.version == wantedVersion {
			return b, nil
		}
	}

	return b, errUnknownLoRaWANRegionalParameters
}

// Versions supported for this band.
func (b Band) Versions() []ttnpb.PHYVersion {
	var versions []ttnpb.PHYVersion
	for _, swapParameter := range b.downgrades() {
		if swapParameter.downgrade != nil {
			versions = append(versions, swapParameter.version)
		} else {
			break
		}
	}
	return versions
}

func beaconChannelFromFrequencies(frequencies [8]uint32) func(float64) uint32 {
	return func(beaconTime float64) uint32 {
		floor := math.Floor(beaconTime / float64(128))
		return frequencies[int32(floor)%8]
	}
}

var usAuBeaconFrequencies = func() [8]uint32 {
	freqs := [8]uint32{}
	for i := 0; i < 8; i++ {
		freqs[i] = 923300000 + uint32(i*600000)
	}
	return freqs
}()

func parseChMask16(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	chans := make(map[uint8]bool, 16)
	switch cntl {
	case 0:
		for i := uint8(0); i < 16; i++ {
			chans[i] = mask[i]
		}
	case 6:
		for i := uint8(0); i < 16; i++ {
			chans[i] = true
		}
	default:
		return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
	}
	return chans, nil
}

func parseChMask72(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	chans := make(map[uint8]bool, 72)
	switch cntl {
	case 0, 1, 2, 3, 4:
		for i := uint8(0); i < 72; i++ {
			chans[i] = (i >= cntl*16 && i < (cntl+1)*16) && mask[i%16]
		}
	case 5:
		for i := uint8(0); i < 64; i++ {
			chans[i] = mask[i/8]
		}
		for i := uint8(64); i < 72; i++ {
			chans[i] = mask[i-64]
		}
	case 6, 7:
		for i := uint8(0); i < 64; i++ {
			chans[i] = cntl == 6
		}
		for i := uint8(64); i < 72; i++ {
			chans[i] = mask[i-64]
		}
	default:
		return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
	}
	return chans, nil
}

func parseChMask96(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	chans := make(map[uint8]bool, 96)
	switch cntl {
	case 0, 1, 2, 3, 4, 5:
		for i := uint8(0); i < 96; i++ {
			chans[i] = (i >= cntl*16 && i < (cntl+1)*16) && mask[i%16]
		}
	case 6:
		for i := uint8(0); i < 96; i++ {
			chans[i] = true
		}
	default:
		return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
	}
	return chans, nil
}

func generateChMaskBlock(mask []bool) ([16]bool, error) {
	if len(mask) > 16 {
		return [16]bool{}, errInvalidChannelCount
	}

	block := [16]bool{}
	for j, on := range mask {
		block[j] = on
	}
	return block, nil
}

func generateChMaskMatrix(mask []bool) (map[uint8][16]bool, error) {
	if len(mask) > math.MaxUint8 {
		return nil, errInvalidChannelCount
	}

	n := uint8(len(mask))
	if n == 0 {
		return nil, errInvalidChannelCount
	}

	ret := make(map[uint8][16]bool, 1+n/16)
	for i := uint8(0); i <= n/16 && 16*i != n; i++ {
		end := 16*i + 16
		if end > n {
			end = n
		}

		block, err := generateChMaskBlock(mask[16*i : end])
		if err != nil {
			return nil, err
		}
		ret[i] = block
	}
	return ret, nil
}

func generateChMask16(mask []bool) (map[uint8][16]bool, error) {
	if len(mask) != 16 {
		return nil, errInvalidChannelCount
	}

	for _, on := range mask {
		if !on {
			return generateChMaskMatrix(mask)
		}
	}
	return map[uint8][16]bool{6: {}}, nil
}

func generateChMask72(mask []bool) (map[uint8][16]bool, error) {
	if len(mask) != 72 {
		return nil, errInvalidChannelCount
	}

	on125 := uint8(0)
	for i := 0; i < 64; i++ {
		if mask[i] {
			on125++
		}
	}

	if on125 == 0 || on125 == 64 {
		block, err := generateChMaskBlock(mask[64:72])
		if err != nil {
			return nil, err
		}

		idx := uint8(6)
		if on125 == 0 {
			idx = 7
		}
		return map[uint8][16]bool{idx: block}, nil
	}

	// TODO: Support ChMaskCntl 5 if required.  (https://github.com/TheThingsIndustries/lorawan-stack/issues/1264)

	return generateChMaskMatrix(mask)
}

func generateChMask96(mask []bool) (map[uint8][16]bool, error) {
	if len(mask) != 96 {
		return nil, errInvalidChannelCount
	}

	for _, on := range mask {
		if !on {
			return generateChMaskMatrix(mask)
		}
	}
	return map[uint8][16]bool{6: {}}, nil
}
