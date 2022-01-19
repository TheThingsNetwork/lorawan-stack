// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package band

import (
	"math"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

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
}

// MaxTxPowerIndex returns the maximum TxPower index for the band.
func (b Band) MaxTxPowerIndex() uint8 {
	n := len(b.TxOffset)
	if n > math.MaxUint8 {
		panic("length of TxOffset overflows uint8")
	}
	return uint8(n) - 1
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
func (b Band) FindUplinkDataRate(dr *ttnpb.DataRate) (ttnpb.DataRateIndex, DataRate, bool) {
	if dr == nil {
		return 0, DataRate{}, false
	}
	for i := ttnpb.DataRateIndex_DATA_RATE_0; i <= ttnpb.DataRateIndex_DATA_RATE_15; i++ {
		// NOTE: Some bands (e.g. US915) contain identical data rates under different indexes.
		// It seems to be a convention in the spec for uplink-only data rates to be at indexes
		// lower than downlink-only indexes, hence match the smallest index.
		// NOTE(2): A more robust solution could be to record a list of uplink-only data rates
		// per band and only match those here, however it is more complex and is not necessary.
		bDR, ok := b.DataRates[i]
		if ok && proto.Equal(bDR.Rate, dr) {
			return i, bDR, true
		}
	}
	return 0, DataRate{}, false
}

// FindDownlinkDataRate returns the downlink data rate with index by API data rate, if any.
func (b Band) FindDownlinkDataRate(dr *ttnpb.DataRate) (ttnpb.DataRateIndex, DataRate, bool) {
	if dr == nil {
		return 0, DataRate{}, false
	}
	// NOTE: See notes in FindUplinkDataRate explaining the order of scanning data rates.
	for i := ttnpb.DataRateIndex_DATA_RATE_15; i >= ttnpb.DataRateIndex_DATA_RATE_0; i-- {
		bDR, ok := b.DataRates[i]
		if ok && proto.Equal(bDR.Rate, dr) {
			return i, bDR, true
		}
	}
	return 0, DataRate{}, false
}
