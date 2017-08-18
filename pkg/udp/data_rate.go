// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"strconv"
)

// DataRate encodes a LoRa data rate as a string or an FSK bit rate as an uint
type DataRate struct {
	LoRa string
	FSK  uint32
}

// MarshalJSON implements the json.Marshaler interface.
func (d DataRate) MarshalJSON() ([]byte, error) {
	if d.LoRa != "" {
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
	d.FSK = uint32(i)
	return nil
}
