// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

// Validator is an object that can validate UDP packets.
type Validator interface {
	// ValidUplink returns true if the sent uplink or status is considered valid.
	ValidUplink(packet Packet) bool
	// ValidDownlink returns true if the downlink request is considered valid.
	ValidDownlink(packet Packet) bool
}
