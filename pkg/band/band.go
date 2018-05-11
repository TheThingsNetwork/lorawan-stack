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
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
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

// Rx1DataRateFunc computes the RX1 data rate index.
type Rx1DataRateFunc func(idx, offset uint32, dwellTime bool) (uint32, error)

// Rx1ChannelFunc computes the RX1 channel index.
type Rx1ChannelFunc func(idx uint32) (uint32, error)

// Rx2Parameters contains downlink datarate index and channel
type Rx2Parameters struct {
	DataRateIndex uint8
	Frequency     uint32
}

// ID is the ID of band
type ID = string

type versionSwap = func(b Band) Band

func bandIdentity(b Band) Band                      { return b }
func rx1ChannelIdentity(idx uint32) (uint32, error) { return idx, nil }

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
	ID ID

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

	// ReceiveDelay1 is the default JoinAccept window timing in seconds
	JoinAcceptDelay1 time.Duration
	// ReceiveDelay2 is the JoinAccept window timing in seconds
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

	// Rx1Parameters is the default function that determines the settings for a Tx sent during Rx1
	Rx1Channel  Rx1ChannelFunc
	Rx1DataRate Rx1DataRateFunc

	// DefaultRx2Parameters are the default parameters that determine the settings for a Tx sent during Rx2
	DefaultRx2Parameters Rx2Parameters

	regionalParameters1_0   versionSwap
	regionalParameters1_0_1 versionSwap
	regionalParameters1_0_2 versionSwap
}

// DutyCycle for the [MinFrequency;MaxFrequency[ sub-band
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
var All = make([]Band, 0)

// GetByID returns the band if it was found, and returns an error otherwise
func GetByID(id ID) (Band, error) {
	for _, band := range All {
		if band.ID == id {
			return band, nil
		}
	}
	return Band{}, ErrBandNotFound.New(errors.Attributes{
		"band": id,
	})
}

type swap struct {
	version   ttnpb.PHYVersion
	downgrade versionSwap
}

func (b Band) downgrades() []swap {
	return []swap{
		{version: ttnpb.PHY_V1_1, downgrade: bandIdentity},
		{version: ttnpb.PHY_V1_0_2, downgrade: b.regionalParameters1_0_2},
		{version: ttnpb.PHY_V1_0_1, downgrade: b.regionalParameters1_0_1},
		{version: ttnpb.PHY_V1_0, downgrade: b.regionalParameters1_0},
	}
}

// Version returns the band parameters for a given version.
func (b Band) Version(wantedVersion ttnpb.PHYVersion) (Band, error) {
	for _, swap := range b.downgrades() {
		if swap.downgrade == nil {
			return b, ErrUnsupportedLoRaWANRegionalParameters.New(nil)
		}
		b = swap.downgrade(b)
		if swap.version == wantedVersion {
			return b, nil
		}
	}

	return b, ErrUnknownLoRaWANRegionalParameters.New(nil)
}

// Versions supported for this band.
func (b Band) Versions() []ttnpb.PHYVersion {
	versions := []ttnpb.PHYVersion{ttnpb.PHY_V1_1}
	for _, swap := range b.downgrades() {
		if swap.downgrade != nil {
			versions = append(versions, swap.version)
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
