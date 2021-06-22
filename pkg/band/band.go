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

// Package band contains structs to handle regional bands.
package band

import (
	"fmt"
	"math"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// eirpDelta is the delta between EIRP and ERP.
const eirpDelta = 2.15

type MaxMACPayloadSizeFunc func(bool) uint16

func makeConstMaxMACPayloadSizeFunc(v uint16) MaxMACPayloadSizeFunc {
	return func(_ bool) uint16 {
		return v
	}
}

func makeDwellTimeMaxMACPayloadSizeFunc(noDwellTimeSize, dwellTimeSize uint16) MaxMACPayloadSizeFunc {
	return func(dwellTime bool) uint16 {
		if dwellTime {
			return dwellTimeSize
		}
		return noDwellTimeSize
	}
}

// DataRate indicates the properties of a band's data rate.
type DataRate struct {
	Rate              ttnpb.DataRate
	MaxMACPayloadSize MaxMACPayloadSizeFunc
}

func makeLoRaDataRate(spreadingFactor uint8, bandwidth uint32, maximumMACPayloadSize MaxMACPayloadSizeFunc) DataRate {
	return DataRate{
		Rate: (&ttnpb.LoRaDataRate{
			SpreadingFactor: uint32(spreadingFactor),
			Bandwidth:       bandwidth,
		}).DataRate(),
		MaxMACPayloadSize: maximumMACPayloadSize,
	}
}

func makeFSKDataRate(bitRate uint32, maximumMACPayloadSize MaxMACPayloadSizeFunc) DataRate {
	return DataRate{
		Rate: (&ttnpb.FSKDataRate{
			BitRate: bitRate,
		}).DataRate(),
		MaxMACPayloadSize: maximumMACPayloadSize,
	}
}

// Channel abstracts a band's channel properties.
type Channel struct {
	// Frequency indicates the frequency of the channel.
	Frequency uint64
	// MinDataRate indicates the index of the minimal data rates accepted on this channel.
	MinDataRate ttnpb.DataRateIndex
	// MinDataRate indicates the index of the maximal data rates accepted on this channel.
	MaxDataRate ttnpb.DataRateIndex
}

// Rx2Parameters contains downlink datarate index and channel.
type Rx2Parameters struct {
	DataRateIndex ttnpb.DataRateIndex
	Frequency     uint64
}

type versionSwap func(b Band) Band

func bandIdentity(b Band) Band {
	return b
}

func composeSwaps(swaps ...versionSwap) versionSwap {
	return func(b Band) Band {
		for _, swap := range swaps {
			b = swap(b)
		}
		return b
	}
}

func channelIndexIdentity(idx uint8) (uint8, error) {
	return idx, nil
}

func channelIndexModulo(n uint8) func(uint8) (uint8, error) {
	return func(idx uint8) (uint8, error) {
		return idx % n, nil
	}
}

// Beacon parameters of a specific band.
type Beacon struct {
	DataRateIndex    ttnpb.DataRateIndex
	CodingRate       string
	InvertedPolarity bool
	// Channel returns in Hz on which beaconing is performed.
	//
	// beaconTime is the integer value, converted in float64, of the 4 bytes “Time” field of the beacon frame.
	ComputeFrequency func(beaconTime float64) uint64
}

// ChMaskCntlPair pairs a ChMaskCntl with a mask.
type ChMaskCntlPair struct {
	Cntl uint8
	Mask [16]bool
}

// Band contains a band's properties.
type Band struct {
	ID string

	Beacon            Beacon
	PingSlotFrequency *uint64

	// MaxUplinkChannels is the maximum amount of uplink channels that can be defined.
	MaxUplinkChannels uint8
	// UplinkChannels are the default uplink channels.
	UplinkChannels []Channel

	// MaxDownlinkChannels is the maximum amount of downlink channels that can be defined.
	MaxDownlinkChannels uint8
	// DownlinkChannels are the default downlink channels.
	DownlinkChannels []Channel

	// SubBands define the sub-bands, their duty-cycle limit and Tx power. The frequency ranges may not overlap.
	SubBands []SubBandParameters

	DataRates map[ttnpb.DataRateIndex]DataRate

	FreqMultiplier   uint64
	ImplementsCFList bool
	CFListType       ttnpb.CFListType

	// ReceiveDelay1 is the default Rx1 window timing in seconds.
	ReceiveDelay1 time.Duration
	// ReceiveDelay2 is the default Rx2 window timing in seconds (ReceiveDelay1 + 1s).
	ReceiveDelay2 time.Duration

	// ReceiveDelay1 is the default join-accept window timing in seconds.
	JoinAcceptDelay1 time.Duration
	// ReceiveDelay2 is the join-accept window timing in seconds.
	JoinAcceptDelay2 time.Duration
	// MaxFCntGap
	MaxFCntGap uint

	// EnableADR determines whether ADR should be enabled.
	EnableADR bool
	// ADRAckLimit
	ADRAckLimit ttnpb.ADRAckLimitExponent
	// ADRAckDelay
	ADRAckDelay          ttnpb.ADRAckDelayExponent
	MinRetransmitTimeout time.Duration
	MaxRetransmitTimeout time.Duration

	// TxOffset in dB: Tx power is computed by taking the MaxEIRP (default: +16dBm) and subtracting the offset.
	TxOffset []float32
	// MaxADRDataRateIndex represents the maximum non-RFU DataRateIndex suitable for ADR, which can be used according to the band's spec.
	MaxADRDataRateIndex ttnpb.DataRateIndex

	TxParamSetupReqSupport bool

	// DefaultMaxEIRP in dBm
	DefaultMaxEIRP float32

	// LoRaCodingRate is the coding rate used for LoRa modulation.
	LoRaCodingRate string

	// Rx1Channel computes the Rx1 channel index.
	Rx1Channel func(uint8) (uint8, error)
	// Rx1DataRate computes the Rx1 data rate index.
	Rx1DataRate func(ttnpb.DataRateIndex, ttnpb.DataRateOffset, bool) (ttnpb.DataRateIndex, error)

	// GenerateChMasks generates a mapping ChMaskCntl -> ChMask.
	// Length of desiredChs must be equal to length of currentChs.
	// Meaning of desiredChs is as follows: for i in range 0..len(desiredChs) if desiredChs[i] == true,
	// then channel with index i should be enabled, otherwise it should be disabled.
	// Meaning of currentChs is as follows: for i in range 0..len(currentChs) if currentChs[i] == true,
	// then channel with index i is enabled, otherwise it is disabled.
	// In case desiredChs equals currentChs, GenerateChMasks returns a singleton, which repesents a noop.
	GenerateChMasks func(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error)
	// ParseChMask computes the channels that have to be masked given ChMask mask and ChMaskCntl cntl.
	ParseChMask func(mask [16]bool, cntl uint8) (map[uint8]bool, error)

	// DefaultRx2Parameters are the default parameters that determine the settings for a Tx sent during Rx2.
	DefaultRx2Parameters Rx2Parameters

	regionalParameters1_v1_0       versionSwap
	regionalParameters1_v1_0_1     versionSwap
	regionalParameters1_v1_0_2     versionSwap
	regionalParameters1_v1_0_2RevB versionSwap
	regionalParameters1_v1_0_3RevA versionSwap
	regionalParameters1_v1_1RevA   versionSwap
}

func (b Band) MaxTxPowerIndex() uint8 {
	n := len(b.TxOffset)
	if n > math.MaxUint8 {
		panic("length of TxOffset overflows uint8")
	}
	return uint8(n) - 1
}

// SubBandParameters contains the sub-band frequency range, duty cycle and Tx power.
type SubBandParameters struct {
	MinFrequency uint64
	MaxFrequency uint64
	DutyCycle    float32
	MaxEIRP      float32
}

// Comprises returns whether the duty cycle applies to the given frequency.
func (d SubBandParameters) Comprises(frequency uint64) bool {
	return frequency >= d.MinFrequency && frequency <= d.MaxFrequency
}

// MaxEmissionDuring the period passed as parameter, that is allowed by that duty cycle.
func (d SubBandParameters) MaxEmissionDuring(period time.Duration) time.Duration {
	return time.Duration(d.DutyCycle * float32(period))
}

// All contains all the bands available.
var All = make(map[string]Band)

// GetByID returns the band if it was found, and returns an error otherwise.
func GetByID(id string) (Band, error) {
	if band, ok := All[id]; ok {
		return band, nil
	}
	return Band{}, errBandNotFound.WithAttributes("id", id)
}

type swapParameters struct {
	version   ttnpb.PHYVersion
	downgrade versionSwap
}

func (b Band) downgrades() []swapParameters {
	return []swapParameters{
		// TODO: Add Regional Parameters for LoRaWAN version 1.0.4 (https://github.com/TheThingsNetwork/lorawan-stack/issues/3513)
		{version: ttnpb.RP001_V1_1_REV_B, downgrade: bandIdentity},
		{version: ttnpb.RP001_V1_1_REV_A, downgrade: b.regionalParameters1_v1_1RevA},
		{version: ttnpb.RP001_V1_0_3_REV_A, downgrade: b.regionalParameters1_v1_0_3RevA},
		{version: ttnpb.RP001_V1_0_2_REV_B, downgrade: b.regionalParameters1_v1_0_2RevB},
		{version: ttnpb.RP001_V1_0_2, downgrade: b.regionalParameters1_v1_0_2},
		{version: ttnpb.TS001_V1_0_1, downgrade: b.regionalParameters1_v1_0_1},
		{version: ttnpb.TS001_V1_0, downgrade: b.regionalParameters1_v1_0},
	}
}

// Version returns the band parameters for a given version.
func (b Band) Version(wantedVersion ttnpb.PHYVersion) (Band, error) {
	var supportedRegionalParameters []string
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

	return b, errUnknownPHYVersion.WithAttributes("version", wantedVersion)
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

// FindSubBand returns the sub-band by frequency, if any.
func (b Band) FindSubBand(frequency uint64) (SubBandParameters, bool) {
	for _, sb := range b.SubBands {
		if sb.Comprises(frequency) {
			return sb, true
		}
	}
	return SubBandParameters{}, false
}

// FindUplinkDataRate returns the uplink data rate with index by API data rate, if any.
func (b Band) FindUplinkDataRate(dr ttnpb.DataRate) (ttnpb.DataRateIndex, DataRate, bool) {
	for i := ttnpb.DataRateIndex(0); int(i) < len(b.DataRates); i++ {
		// NOTE: Some bands (e.g. US915) contain identical data rates under different indexes.
		// It seems to be a convention in the spec for uplink-only data rates to be at indexes
		// lower than downlink-only indexes, hence match the smallest index.
		// NOTE(2): A more robust solution could be to record a list of uplink-only data rates
		// per band and only match those here, however it is more complex and is not necessary.
		bDR, ok := b.DataRates[i]
		if ok && bDR.Rate.Equal(dr) {
			return i, bDR, true
		}
	}
	return 0, DataRate{}, false
}

func makeBeaconFrequencyFunc(frequencies [8]uint64) func(float64) uint64 {
	return func(beaconTime float64) uint64 {
		floor := math.Floor(beaconTime / float64(128))
		return frequencies[int32(floor)%8]
	}
}

var usAuBeaconFrequencies = func() (freqs [8]uint64) {
	for i := 0; i < 8; i++ {
		freqs[i] = 923300000 + uint64(i*600000)
	}
	return freqs
}()

func parseChMask(offset uint8, mask ...bool) map[uint8]bool {
	if len(mask)-1 > int(math.MaxUint8-offset) {
		panic(fmt.Sprintf("channel mask overflows uint8, offset: %d, mask length: %d", offset, len(mask)))
	}
	m := make(map[uint8]bool, len(mask))
	for i, v := range mask {
		m[offset+uint8(i)] = v
	}
	return m
}

func parseChMask16(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0:
		return parseChMask(0, mask[:]...), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func parseChMask72(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0, 1, 2, 3:
		return parseChMask(cntl*16, mask[:]...), nil
	case 4:
		return parseChMask(64, mask[0:8]...), nil
	case 5:
		return parseChMask(0,
			mask[0], mask[0], mask[0], mask[0], mask[0], mask[0], mask[0], mask[0],
			mask[1], mask[1], mask[1], mask[1], mask[1], mask[1], mask[1], mask[1],
			mask[2], mask[2], mask[2], mask[2], mask[2], mask[2], mask[2], mask[2],
			mask[3], mask[3], mask[3], mask[3], mask[3], mask[3], mask[3], mask[3],
			mask[4], mask[4], mask[4], mask[4], mask[4], mask[4], mask[4], mask[4],
			mask[5], mask[5], mask[5], mask[5], mask[5], mask[5], mask[5], mask[5],
			mask[6], mask[6], mask[6], mask[6], mask[6], mask[6], mask[6], mask[6],
			mask[7], mask[7], mask[7], mask[7], mask[7], mask[7], mask[7], mask[7],
		), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			mask[0], mask[1], mask[2], mask[3], mask[4], mask[5], mask[6], mask[7],
		), nil
	case 7:
		return parseChMask(0,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			mask[0], mask[1], mask[2], mask[3], mask[4], mask[5], mask[6], mask[7],
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func parseChMask96(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0, 1, 2, 3, 4, 5:
		return parseChMask(cntl*16, mask[:]...), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func boolsTo16BoolArray(vs ...bool) [16]bool {
	if len(vs) > 16 {
		panic(fmt.Sprintf("length of vs must be less or equal to 16, got %d", len(vs)))
	}
	var ret [16]bool
	for i, v := range vs {
		ret[i] = v
	}
	return ret
}

func generateChMask16(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 16 || len(desiredChs) != 16 {
		return nil, errInvalidChannelCount.New()
	}
	// NOTE: ChMaskCntl==6 never provides a more optimal ChMask sequence than ChMaskCntl==0.
	return []ChMaskCntlPair{
		{
			Mask: boolsTo16BoolArray(desiredChs...),
		},
	}, nil
}

func EqualChMasks(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateChMaskMatrix(pairs []ChMaskCntlPair, currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	n := len(currentChs)
	if n%16 != 0 || len(desiredChs) != n {
		return nil, errInvalidChannelCount.New()
	}
	for i := 0; i < n/16; i++ {
		for j := 0; j < 16; j++ {
			if currentChs[16*i+j] != desiredChs[16*i+j] {
				pairs = append(pairs, ChMaskCntlPair{
					Cntl: uint8(i),
					Mask: boolsTo16BoolArray(desiredChs[16*i : 16*i+16]...),
				})
				break
			}
		}
	}
	return pairs, nil
}

func trueCount(vs ...bool) int {
	var n int
	for _, v := range vs {
		if v {
			n++
		}
	}
	return n
}

func generateChMask72Generic(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 72 || len(desiredChs) != 72 {
		return nil, errInvalidChannelCount.New()
	}
	if EqualChMasks(currentChs, desiredChs) {
		return []ChMaskCntlPair{
			{
				Mask: boolsTo16BoolArray(desiredChs[0:16]...),
			},
		}, nil
	}

	on125 := trueCount(desiredChs[0:64]...)
	switch on125 {
	case 0:
		return []ChMaskCntlPair{
			{
				Cntl: 7,
				Mask: boolsTo16BoolArray(desiredChs[64:72]...),
			},
		}, nil

	case 64:
		return []ChMaskCntlPair{
			{
				Cntl: 6,
				Mask: boolsTo16BoolArray(desiredChs[64:72]...),
			},
		}, nil
	}

	pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 5), currentChs[0:64], desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	for i := 65; i < 72; i++ {
		if currentChs[i] != desiredChs[i] {
			pairs = append(pairs, ChMaskCntlPair{
				Cntl: 4,
				Mask: boolsTo16BoolArray(desiredChs[64:72]...),
			})
			break
		}
	}
	if len(pairs) <= 2 {
		return pairs, nil
	}
	// Count amount of pairs required assuming either ChMaskCntl==6 or ChMaskCntl==7 is sent first.
	// The minimum amount of pairs required in such case will be 2, hence only attempt this if amount
	// of generated pairs so far is higher than 2.
	cntl6Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}, desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	cntl7Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
	}, desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	switch {
	case len(pairs) <= 1+len(cntl6Pairs) && len(pairs) <= 1+len(cntl7Pairs):
		return pairs, nil

	case len(cntl6Pairs) < len(cntl7Pairs):
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl6Pairs)), ChMaskCntlPair{
			Cntl: 6,
			Mask: boolsTo16BoolArray(desiredChs[64:72]...),
		}), cntl6Pairs...), nil

	default:
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl7Pairs)), ChMaskCntlPair{
			Cntl: 7,
			Mask: boolsTo16BoolArray(desiredChs[64:72]...),
		}), cntl7Pairs...), nil
	}
}

//revive:disable:flag-parameter

func makeGenerateChMask72(supportChMaskCntl5 bool) func([]bool, []bool) ([]ChMaskCntlPair, error) {
	if !supportChMaskCntl5 {
		return generateChMask72Generic
	}
	return func(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
		pairs, err := generateChMask72Generic(currentChs, desiredChs)
		if err != nil {
			return nil, err
		}
		if len(pairs) <= 1 {
			return pairs, nil
		}

		var fsbs [8]bool
		for i := 0; i < 8; i++ {
			if trueCount(desiredChs[8*i:8*i+8]...) == 8 {
				fsbs[i] = true
			}
		}
		if n := trueCount(fsbs[:]...); n == 0 || n == 8 {
			// Since there are either no enabled FSBs, or no disabled FSBs we won't be able to compute a
			// more efficient result that one using ChMaskCntl==6 or ChMaskCntl==7.
			return pairs, nil
		}
		cntl5Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 5), []bool{
			fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0],
			fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1],
			fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2],
			fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3],
			fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4],
			fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5],
			fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6],
			fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7],
		}, desiredChs[0:64])
		if err != nil {
			return nil, err
		}
		for i := 65; i < 72; i++ {
			if currentChs[i] != desiredChs[i] {
				cntl5Pairs = append(cntl5Pairs, ChMaskCntlPair{
					Cntl: 4,
					Mask: boolsTo16BoolArray(desiredChs[64:72]...),
				})
				break
			}
		}
		if len(pairs) <= 1+len(cntl5Pairs) {
			return pairs, nil
		}
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl5Pairs)), ChMaskCntlPair{
			Cntl: 5,
			Mask: boolsTo16BoolArray(fsbs[:]...),
		}), cntl5Pairs...), nil
	}
}

//revive:enable:flag-parameter

func generateChMask96(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 96 || len(desiredChs) != 96 {
		return nil, errInvalidChannelCount.New()
	}
	if EqualChMasks(currentChs, desiredChs) {
		return []ChMaskCntlPair{
			{
				Mask: boolsTo16BoolArray(desiredChs[0:16]...),
			},
		}, nil
	}
	if trueCount(desiredChs...) == 96 {
		return []ChMaskCntlPair{
			{
				Cntl: 6,
			},
		}, nil
	}
	pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 6), currentChs, desiredChs)
	if err != nil {
		return nil, err
	}
	if len(pairs) <= 2 {
		return pairs, nil
	}
	// Count amount of pairs required assuming ChMaskCntl==6 is sent first.
	// The minimum amount of pairs required in such case will be 2, hence only attempt this if amount
	// of generated pairs so far is higher than 2.
	cntl6Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 6), []bool{
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}, desiredChs)
	if err != nil {
		return nil, err
	}
	if len(pairs) <= 1+len(cntl6Pairs) {
		return pairs, nil
	}
	return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl6Pairs)), ChMaskCntlPair{
		Cntl: 6,
	}), cntl6Pairs...), nil
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}
