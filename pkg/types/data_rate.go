// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"regexp"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// DataRate encodes a LoRa data rate as a string or an FSK bit rate as an uint
type DataRate struct {
	LoRa string
	FSK  uint32
}

// SpreadingFactor returns the spreading factor of this data rate, if it is a LoRa data rate. It returns an error otherwise.
func (dr DataRate) SpreadingFactor() (uint8, error) {
	re := regexp.MustCompile("SF(7|8|9|10|11|12)")
	matches := re.FindStringSubmatch(dr.LoRa)
	if len(matches) != 2 {
		return 0, errors.New("Spreading factor not found")
	}

	sf, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, errors.NewWithCause("Failed to parse spreading factor", err)
	}
	return uint8(sf), err
}

// Bandwidth returns the spreading factor of this data rate. It returns an error otherwise.
func (dr DataRate) Bandwidth() (uint32, error) {
	if dr.FSK != 0 {
		return dr.FSK, nil
	}

	re := regexp.MustCompile("BW(125|250|500)")
	matches := re.FindStringSubmatch(dr.LoRa)
	if len(matches) != 2 {
		return 0, errors.New("Bandwidth not found")
	}

	bw, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, errors.NewWithCause("Failed to parse bandwidth", err)
	}
	return uint32(bw), err
}
