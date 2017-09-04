// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
	}
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
	Frequency       int
	DataRateIndexes []int
}

// Rx1Emission takes the uplink's emission parameters, and returns downlink datarate index and channel
type Rx1Emission func(dataRateIndex, frequency, RX1DROffset int, dwellTime bool) (int, int)

// Rx2Parameters contains downlink datarate index and channel
type Rx2Parameters struct {
	DataRateIndex uint8
	Frequency     uint32
}

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

	// ReceiveDelay1 is the default RX1 window timing in seconds
	ReceiveDelay1 time.Duration
	// ReceiveDelay2 is the default RX2 window timing in seconds (ReceiveDelay1 + 1s)
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

	// TXOffset in dB: A TX's power is computed by taking the MaxEIRP (default: +16dBm) and substracting the offset
	TXOffset []float32

	// DefaultMaxEIRP in dBm
	DefaultMaxEIRP float32

	// RX1Parameters is the default function that determines the settings for a TX sent during RX1
	RX1Parameters Rx1Emission
	// RX1Parameters are the default parameters that determine the settings for a TX sent during RX2
	DefaultRX2Parameters Rx2Parameters
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
	return Band{}, ErrBandNotFound.New(errors.Attributes{
		"band": string(id),
	})
}

func init() {
	ErrBandNotFound.Register()
}
