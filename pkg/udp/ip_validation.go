// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

// Validator is an object that can validate UDP packets.
type Validator interface {
	// ValidUplink returns true if the sent uplink or status is considered valid.
	ValidUplink(packet Packet) bool
	// ValidDownlink returns true if the downlink request is considered valid.
	ValidDownlink(packet Packet) bool
}

type alwaysValid struct{}

func (a alwaysValid) ValidUplink(packet Packet) bool   { return true }
func (a alwaysValid) ValidDownlink(packet Packet) bool { return true }

// AlwaysValid returns a Validator that considers all packets valid.
func AlwaysValid() Validator { return alwaysValid{} }
