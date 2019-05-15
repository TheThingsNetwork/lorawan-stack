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
	"time"

	encoding "go.thethings.network/lorawan-stack/pkg/ttnpb/udp"
)

// Firewall filters packets by tracking addresses and time.
type Firewall interface {
	Filter(packet encoding.Packet) bool
}

type addrTime struct {
	net.UDPAddr
	lastSeen time.Time
}

type memoryFirewall struct {
	push            sync.Map
	pull            sync.Map
	addrChangeBlock time.Duration
}

// NewMemoryFirewall returns an in-memory Firewall.
func NewMemoryFirewall(ctx context.Context, addrChangeBlock time.Duration) Firewall {
	v := &memoryFirewall{
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
				v.gc()
			}
		}
	}()
	return v
}

func (v *memoryFirewall) filter(packet encoding.Packet, store *sync.Map) bool {
	if packet.GatewayEUI == nil || packet.GatewayAddr == nil {
		return false
	}
	now := time.Now().UTC()
	eui := *packet.GatewayEUI
	val, ok := store.Load(eui)
	if ok {
		a := val.(addrTime)
		if a.UDPAddr.String() != packet.GatewayAddr.String() && a.lastSeen.Add(v.addrChangeBlock).After(now) {
			return false
		}
	}
	store.Store(eui, addrTime{
		UDPAddr:  *packet.GatewayAddr,
		lastSeen: now,
	})
	return true
}

func (v *memoryFirewall) Filter(packet encoding.Packet) bool {
	switch packet.PacketType {
	case encoding.PullData:
		return v.filter(packet, &v.pull)
	case encoding.PushData, encoding.TxAck:
		return v.filter(packet, &v.push)
	}
	return false
}

func (v *memoryFirewall) gc() {
	now := time.Now().UTC()
	gcStore := func(store *sync.Map) {
		store.Range(func(k, val interface{}) bool {
			a := val.(addrTime)
			if a.lastSeen.Add(v.addrChangeBlock).Before(now) {
				store.Delete(k)
			}
			return true
		})
	}
	gcStore(&v.pull)
	gcStore(&v.push)
}
