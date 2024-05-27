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
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
)

// Firewall filters packets by tracking addresses and time.
type Firewall interface {
	Filter(packet encoding.Packet) error
}

type noopFirewall struct{}

// Filter implements Firewall.
func (noopFirewall) Filter(encoding.Packet) error { return nil }

type addrTime struct {
	ip       net.IP
	lastSeen atomic.Pointer[time.Time]
}

type memoryFirewall struct {
	m               sync.Map
	addrChangeBlock time.Duration
}

// NewMemoryFirewall returns an in-memory Firewall.
func NewMemoryFirewall(ctx context.Context, addrChangeBlock time.Duration) Firewall {
	f := &memoryFirewall{
		addrChangeBlock: addrChangeBlock,
	}
	go func() {
		ticker := time.NewTicker(addrChangeBlock)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				f.gc()
			}
		}
	}()
	return f
}

var (
	errNoEUI            = errors.DefineInvalidArgument("no_eui", "packet has no gateway EUI")
	errNoAddress        = errors.DefineInvalidArgument("no_address", "packet has no gateway address")
	errAlreadyConnected = errors.DefineFailedPrecondition("already_connected", "gateway is already connected")
)

func (f *memoryFirewall) Filter(packet encoding.Packet) error {
	if packet.GatewayEUI == nil {
		return errNoEUI.New()
	}
	if packet.GatewayAddr == nil {
		return errNoAddress.New()
	}
	now := time.Now().UTC()
	eui := *packet.GatewayEUI
	entry := &addrTime{
		ip: packet.GatewayAddr.IP,
	}
	entry.lastSeen.Store(&now)
	actual, loaded := f.m.LoadOrStore(eui, entry)
	if !loaded {
		// This is a new entry. There are no checks or updates to be done.
		return nil
	}
	a := actual.(*addrTime) // nolint:revive
	lastSeen := a.lastSeen.Load()
	if !a.ip.Equal(packet.GatewayAddr.IP) && lastSeen.Add(f.addrChangeBlock).After(now) {
		return errAlreadyConnected.WithAttributes(
			"connected_ip", a.ip.String(),
			"connecting_ip", packet.GatewayAddr.IP.String(),
		)
	}
	for ; lastSeen.Before(now); lastSeen = a.lastSeen.Load() {
		if a.lastSeen.CompareAndSwap(lastSeen, &now) {
			return nil
		}
	}
	return nil
}

func (f *memoryFirewall) gc() {
	now := time.Now().UTC()
	f.m.Range(func(k, val any) bool {
		a := val.(*addrTime) // nolint:revive
		if a.lastSeen.Load().Add(f.addrChangeBlock).Before(now) {
			f.m.Delete(k)
		}
		return true
	})
}
