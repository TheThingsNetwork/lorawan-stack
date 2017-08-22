// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validation

import (
	"net"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/udp"
)

// DefaultWaitDuration is the recommended duration to block the IP
// of the gateway.
const DefaultWaitDuration = time.Hour

type footprint struct {
	Address net.UDPAddr
	Time    time.Time
}

type inMemoryValidator struct {
	duration time.Duration
	lastSeen map[types.EUI64]footprint
}

func (v *inMemoryValidator) Valid(packet udp.Packet) (valid bool) {
	defer func() {
		if valid {
			v.lastSeen[*packet.GatewayEUI] = footprint{
				Address: *packet.GatewayAddr,
				Time:    time.Now(),
			}
		}
	}()
	if packet.GatewayEUI == nil {
		return false
	}
	if packet.GatewayAddr == nil {
		return false
	}

	footprint, seen := v.lastSeen[*packet.GatewayEUI]
	if !seen {
		return true
	}

	if packet.GatewayAddr.String() == footprint.Address.String() {
		return true
	}
	if footprint.Time.Add(v.duration).Before(time.Now()) {
		return true
	}
	return false
}

// InMemoryValidator is a validator that keeps track in-memory of a
// gateway's IP. It is fast-to-access and lightweight, but doesn't store
// the data persistently. When a gateway sends a packet, this validator
// blocks the gateway to this IP for <duration>. If a packet is sent
// with the same gateway EUI but from another IP before this duration,
// it is considered invalid.
//
// The duration parameter to the constructor is the time.Duration to wait
// before a gateway's location expires.
//
// The Valid function returns true if the gateway has a valid EUI, a valid
// IP address, and that if there is a new IP, there has been no message
// since <duration>.
func InMemoryValidator(duration time.Duration) Validator {
	return &inMemoryValidator{
		duration: duration,
		lastSeen: make(map[types.EUI64]footprint),
	}
}
