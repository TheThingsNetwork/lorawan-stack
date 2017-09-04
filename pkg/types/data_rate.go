// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

// DataRate encodes a LoRa data rate as a string or an FSK bit rate as an uint
type DataRate struct {
	LoRa string
	FSK  uint32
}
