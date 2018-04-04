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

package udp

import (
	"net"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// DefaultWaitDuration is the recommended duration to block the IP of the gateway.
const DefaultWaitDuration = 5 * time.Minute

type footprint struct {
	DownlinkAddress  net.Addr
	DownlinkLastSeen time.Time
	UplinkAddress    net.Addr
	UplinkLastSeen   time.Time
}

// GatewayStore is a goroutine-safe structure that implements two interfaces:
//
// - It is a Validator that keeps track in-memory of a gateway's IP. It is fast and lightweight, but doesn't store the data persistently. When a gateway sends a packet, this validator blocks the gateway to this IP for <duration>. If a packet is sent with the same gateway EUI but from another IP
// before this duration, it is considered invalid. The duration parameter to the constructor is the time.Duration to wait before a gateway's location expires. The Valid function returns true if the gateway has a valid EUI, a valid IP address, and that if there is a new IP, there has been no message
// since <duration>.
//
// - It is an AddressStore, storing the last IP seen in memory.
type GatewayStore struct {
	mu       sync.Mutex
	duration time.Duration
	lastSeen map[types.EUI64]footprint
}

// SetUplinkAddress implements AddressStore for GatewayStore.
func (s *GatewayStore) SetUplinkAddress(eui types.EUI64, addr net.Addr) {
	s.mu.Lock()
	newFootprint := footprint{UplinkAddress: addr, UplinkLastSeen: time.Now()}
	if gwFootprint, found := s.lastSeen[eui]; found {
		newFootprint.DownlinkAddress = gwFootprint.DownlinkAddress
		newFootprint.DownlinkLastSeen = gwFootprint.DownlinkLastSeen
	}
	s.lastSeen[eui] = newFootprint
	s.mu.Unlock()
}

// SetDownlinkAddress implements AddressStore for GatewayStore.
func (s *GatewayStore) SetDownlinkAddress(eui types.EUI64, addr net.Addr) {
	s.mu.Lock()
	newFootprint := footprint{DownlinkAddress: addr, DownlinkLastSeen: time.Now()}
	if gwFootprint, found := s.lastSeen[eui]; found {
		newFootprint.UplinkAddress = gwFootprint.UplinkAddress
		newFootprint.UplinkLastSeen = gwFootprint.UplinkLastSeen
	}
	s.lastSeen[eui] = newFootprint
	s.mu.Unlock()
}

// GetDownlinkAddress implements AddressStore for GatewayStore.
func (s *GatewayStore) GetDownlinkAddress(eui types.EUI64) (net.Addr, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	footprint, seen := s.lastSeen[eui]
	if !seen {
		return nil, false
	}

	address := footprint.DownlinkAddress
	return address, true
}

func (s *GatewayStore) validPacket(packet Packet) bool {
	if packet.GatewayEUI == nil {
		return false
	}
	if packet.GatewayAddr == nil {
		return false
	}
	return true
}

// ValidUplink implements Validator for GatewayStore.
func (s *GatewayStore) ValidUplink(packet Packet) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.validPacket(packet) {
		return false
	}

	footprint, seen := s.lastSeen[*packet.GatewayEUI]
	if !seen {
		return true
	}

	if footprint.UplinkAddress == nil || packet.GatewayAddr.String() == footprint.UplinkAddress.String() {
		return true
	}
	if footprint.UplinkLastSeen.Add(s.duration).Before(time.Now()) {
		return true
	}
	return false
}

// ValidDownlink implements Validator for GatewayStore.
func (s *GatewayStore) ValidDownlink(packet Packet) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.validPacket(packet) {
		return false
	}

	footprint, seen := s.lastSeen[*packet.GatewayEUI]
	if !seen {
		return true
	}

	if footprint.DownlinkAddress == nil || packet.GatewayAddr.String() == footprint.DownlinkAddress.String() {
		return true
	}
	if footprint.DownlinkLastSeen.Add(s.duration).Before(time.Now()) {
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
