// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package band contains structs to handle regional bands
package band

import (
	"errors"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

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

type DataRate struct {
	Rate              types.DataRate
	DefaultMaxSize    PayloadSize
	NoRepeaterMaxSize PayloadSize
}

type Channel struct {
	Frequency int
	DataRates []int
}

// Rx1Emission takes the uplink's emission parameters, and returns downlink datarate index and channel
type Rx1Emission func(dataRateIndex, frequency, RX1DROffset int, dwellTime bool) (int, int)

// Rx2Emission takes the uplink's emission parameters, and returns downlink datarate index and channel
type Rx2Emission func(dataRateIndex, frequency, RX2DataRate int) (int, int)

type BandID string

type Band struct {
	ID BandID

	// UplinkChannels by default
	UplinkChannels []Channel
	// DownlinkChannels by default
	DownlinkChannels []Channel

	BandDutyCycles []DutyCycle

	DataRates []DataRate

	ImplementsCFList bool

	// ReceiveDelay1 is the RX1 window timing in seconds
	ReceiveDelay1 time.Duration
	// ReceiveDelay2 is the RX2 window timing in seconds (ReceiveDelay1 + 1s)
	ReceiveDelay2 time.Duration

	// ReceiveDelay1 is the JoinAccept window timing in seconds
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

	// TXOffset in dB: A TX's power is computed by taking the MaxEIRP (default: +16dBm) and substracting the offset
	TXOffset []float32

	// DefaultMaxEIRP in dBm
	DefaultMaxEIRP float32

	RX1Parameters        Rx1Emission
	DefaultRX2Parameters Rx2Emission
}

// DutyCycle for the [MinFrequency;MaxFrequency[ sub-band
type DutyCycle struct {
	MinFrequency int
	MaxFrequency int
	DutyCycle    float32
}

var All = make([]Band, 0)

func GetByID(id BandID) (Band, error) {
	for _, band := range All {
		if band.ID == id {
			return band, nil
		}
	}
	return Band{}, errors.New("Band not found")
}
