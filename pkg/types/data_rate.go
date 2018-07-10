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

package types

import (
	"regexp"
	"strconv"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var invalidDataRate = errors.DefineInvalidArgument("data_rate", "invalid data rate")

// DataRate encodes a LoRa data rate as a string or an FSK bit rate as an uint
type DataRate struct {
	LoRa string
	FSK  uint32
}

var sfRegexp = regexp.MustCompile("^SF(6|7|8|9|10|11|12)")

// SpreadingFactor returns the spreading factor of this data rate, if it is a LoRa data rate. It returns an error otherwise.
func (dr DataRate) SpreadingFactor() (uint8, error) {
	matches := sfRegexp.FindStringSubmatch(dr.LoRa)
	if len(matches) != 2 {
		return 0, invalidDataRate
	}
	sf, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, invalidDataRate
	}
	return uint8(sf), err
}

var drRegexp = regexp.MustCompile("BW(125|250|500)$")

// Bandwidth returns the bandwidth of this data rate in Hz, if it is a LoRa data rate. It returns an error otherwise.
func (dr DataRate) Bandwidth() (uint32, error) {
	matches := drRegexp.FindStringSubmatch(dr.LoRa)
	if len(matches) != 2 {
		return 0, invalidDataRate
	}
	bw, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, invalidDataRate
	}
	return uint32(bw) * 1000, err
}
