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

package datarate

import (
	"fmt"
	"regexp"
	"strconv"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// DR encodes a LoRa data rate or an FSK data rate, and implements marshalling and unmarshalling between JSON.
type DR struct {
	ttnpb.DataRate
}

// MarshalJSON implements the json.Marshaler interface.
func (dr DR) MarshalJSON() ([]byte, error) {
	if dr.GetLoRa() != nil {
		return []byte(strconv.Quote(dr.String())), nil
	}
	if dr.GetFSK() != nil {
		return []byte(dr.String()), nil
	}
	return nil, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (dr *DR) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		datarate, err := ParseLoRa(string(data[1 : len(data)-1]))
		if err != nil {
			return err
		}
		*dr = datarate
		return nil
	}
	i, err := strconv.ParseUint(string(data), 10, 32)
	if err != nil {
		return err
	}
	*dr = DR{
		DataRate: ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_FSK{
				FSK: &ttnpb.FSKDataRate{
					BitRate: uint32(i),
				},
			},
		},
	}
	return nil
}

var (
	errDataRate = errors.DefineInvalidArgument("data_rate", "invalid data rate")
	sfRegexp    = regexp.MustCompile(`^SF([1-9]|10|11|12)BW`)
	bwRegexp    = regexp.MustCompile(`BW(\d+(?:\.\d+)?)$`)
)

// String implements the Stringer interface.
func (dr DR) String() string {
	if lora := dr.GetLoRa(); lora != nil {
		return fmt.Sprintf("SF%dBW%v", lora.SpreadingFactor, float32(lora.Bandwidth)/1000)
	}
	if fsk := dr.GetFSK(); fsk != nil {
		return fmt.Sprintf("%d", fsk.BitRate)
	}
	return ""
}

// ParseLoRa converts a string of format "SFxxBWxxx" to a LoRaDataRate.
func ParseLoRa(dr string) (DR, error) {
	matches := sfRegexp.FindStringSubmatch(dr)
	if len(matches) != 2 {
		return DR{}, errDataRate
	}
	sf, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return DR{}, errDataRate
	}
	matches = bwRegexp.FindStringSubmatch(dr)
	if len(matches) != 2 {
		return DR{}, errDataRate
	}
	bw, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return DR{}, errDataRate
	}
	return DR{
		DataRate: ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: uint32(sf),
					Bandwidth:       uint32(bw * 1000),
				},
			},
		},
	}, nil
}
