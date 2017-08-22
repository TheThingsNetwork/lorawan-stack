// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// DefaultWaitDuration is the recommended duration to block the IP
// of the gateway.
const DefaultWaitDuration = time.Hour

type footprint struct {
	Address net.Addr
	Time    time.Time
}

// GatewayStore is a multithread-safe structure that
// implements two interfaces:
//
// - It is a Validator that keeps track in-memory of a
// gateway's IP. It is fast and lightweight, but doesn't store
// the data persistently. When a gateway sends a packet, this validator
// blocks the gateway to this IP for <duration>. If a packet is sent
// with the same gateway EUI but from another IP before this duration,
// it is considered invalid.
// The duration parameter to the constructor is the time.Duration to wait
// before a gateway's location expires.
// The Valid function returns true if the gateway has a valid EUI, a valid
// IP address, and that if there is a new IP, there has been no message
// since <duration>.
// - It is an AddressStore.
type GatewayStore struct {
	mu       sync.Mutex
	duration time.Duration
	lastSeen map[types.EUI64]footprint
}

// Set implements AddressStore for GatewayStore.
func (s *GatewayStore) Set(eui types.EUI64, addr net.Addr) {
	s.mu.Lock()
	s.lastSeen[eui] = footprint{
		Address: addr,
		Time:    time.Now(),
	}
	s.mu.Unlock()
}

// Get implements AddressStore for GatewayStore.
func (s *GatewayStore) Get(eui types.EUI64) (net.Addr, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	footprint, seen := s.lastSeen[eui]
	if !seen {
		return nil, false
	}

	address := footprint.Address
	return address, true
}

// Valid implements Validator for GatewayStore.
func (s *GatewayStore) Valid(packet Packet) (valid bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if packet.GatewayEUI == nil {
		return false
	}
	if packet.GatewayAddr == nil {
		return false
	}

	footprint, seen := s.lastSeen[*packet.GatewayEUI]
	if !seen {
		return true
	}

	if packet.GatewayAddr.String() == footprint.Address.String() {
		return true
	}
	if footprint.Time.Add(s.duration).Before(time.Now()) {
		return true
	}
	return false
}

// NewGatewayStore returns a GatewayStore pointer
func NewGatewayStore(duration time.Duration) *GatewayStore {
	return &GatewayStore{
		duration: duration,
		lastSeen: make(map[types.EUI64]footprint),
		mu:       sync.Mutex{},
	}
}
