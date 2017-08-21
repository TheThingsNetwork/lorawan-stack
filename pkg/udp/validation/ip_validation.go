package validation

import "github.com/TheThingsNetwork/ttn/pkg/udp"

// Validator is an object that can validate UDP packets.
type Validator interface {
	// Valid returns true if the sent packet is considered valid.
	Valid(packet udp.Packet) bool
}
