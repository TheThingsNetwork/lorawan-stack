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

package udp

import (
	"strconv"

	"go.thethings.network/lorawan-stack/pkg/types"
)

// DataRate encodes a LoRa data rate as a string or an FSK bit rate as an uint, and implements marshalling and unmarshalling between JSON
type DataRate struct {
	types.DataRate
}

// MarshalJSON implements the json.Marshaler interface.
func (d DataRate) MarshalJSON() ([]byte, error) {
	if d.DataRate.LoRa != "" {
		return []byte(`"` + d.LoRa + `"`), nil
	}
	return []byte(strconv.FormatUint(uint64(d.FSK), 10)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *DataRate) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		d.LoRa = string(data[1 : len(data)-1])
		return nil
	}
	i, err := strconv.ParseUint(string(data), 10, 32)
	if err != nil {
		return err
	}
	d.DataRate.FSK = uint32(i)
	return nil
}
