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
	"bytes"
	"context"
	"net"
	"sync"
	"time"

	encoding "go.thethings.network/lorawan-stack/pkg/ttnpb/udp"
)

// Firewall filters packets by tracking addresses and time.
type Firewall interface {
	Filter(packet encoding.Packet) bool
}

type addrTime struct {
	net.IP
	lastSeen time.Time
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
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				f.gc()
			}
		}
	}()
	return f
}

func (f *memoryFirewall) Filter(packet encoding.Packet) bool {
	if packet.GatewayEUI == nil || packet.GatewayAddr == nil {
		return false
	}
	now := time.Now().UTC()
	eui := *packet.GatewayEUI
	val, ok := f.m.Load(eui)
	if ok {
		a := val.(addrTime)
		if !bytes.Equal(a.IP, packet.GatewayAddr.IP) && a.lastSeen.Add(f.addrChangeBlock).After(now) {
			return false
		}
	}
	f.m.Store(eui, addrTime{
		IP:       packet.GatewayAddr.IP,
		lastSeen: now,
	})
	return true
}

func (f *memoryFirewall) gc() {
	now := time.Now().UTC()
	f.m.Range(func(k, val interface{}) bool {
		a := val.(addrTime)
		if a.lastSeen.Add(f.addrChangeBlock).Before(now) {
			f.m.Delete(k)
		}
		return true
	})
}
