// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package band contains structs to handle regional bands
package band

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

var (
	// ErrBandNotFound describes the errors returned when looking for an unknown band
	ErrBandNotFound = &errors.ErrDescriptor{
		MessageFormat: "Band {band} not found",
		Type:          errors.NotFound,
		Code:          1,
		SafeAttributes: []string{
			"band",
		},
	}
)

// PayloadSize abstracts the acceptable payload size depending on contextual parameters
type PayloadSize interface {
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
	DefaultMaxSize    PayloadSize
	NoRepeaterMaxSize PayloadSize
}

// Channel abstracts a band's channel properties
type Channel struct {
	// Frequency indicates the frequency of the channel
	Frequency uint64
	// DataRateIndexes indicates the data rates accepted on this channel
	DataRateIndexes []int
}

// Rx1Emission takes the uplink's emission parameters, and returns downlink datarate index and channel
type Rx1Emission func(frequency uint64, dataRateIndex, Rx1DROffset int, dwellTime bool) (int, uint64)

// Rx2Parameters contains downlink datarate index and channel
type Rx2Parameters struct {
	DataRateIndex uint8
	Frequency     uint32
}

// ID is the ID of band
type ID = string

// Band contains a band's properties
type Band struct {
	ID ID

	// UplinkChannels by default
	UplinkChannels []Channel
	// DownlinkChannels by default
	DownlinkChannels []Channel

	BandDutyCycles []DutyCycle

	DataRates []DataRate

	ImplementsCFList bool

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
	// AdrAckLimit
	AdrAckLimit uint8
	// AdrAckDelay
	AdrAckDelay   uint8
	MinAckTimeout time.Duration
	MaxAckTimeout time.Duration

	// TxOffset in dB: A Tx's power is computed by taking the MaxEIRP (default: +16dBm) and subtracting the offset
	TxOffset []float32

	// DefaultMaxEIRP in dBm
	DefaultMaxEIRP float32

	// Rx1Parameters is the default function that determines the settings for a Tx sent during Rx1
	Rx1Parameters Rx1Emission
	// DefaultRx2Parameters are the default parameters that determine the settings for a Tx sent during Rx2
	DefaultRx2Parameters Rx2Parameters
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

func init() {
	ErrBandNotFound.Register()
}
