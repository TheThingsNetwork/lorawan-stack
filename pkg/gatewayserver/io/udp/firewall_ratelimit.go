// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
)

type rateLimitingFirewall struct {
	f Firewall

	m sync.Map // string to timestamps

	messages  int
	threshold time.Duration
}

// NewRateLimitingFirewall returns a Firewall with rate limiting capabilities.
func NewRateLimitingFirewall(firewall Firewall, messages int, threshold time.Duration) Firewall {
	return &rateLimitingFirewall{
		f:         firewall,
		messages:  messages,
		threshold: threshold,
	}
}

var (
	errRateExceeded = errors.DefineResourceExhausted("rate_exceeded", "gateway traffic exceeded allowed rate")
)

func (f *rateLimitingFirewall) Filter(packet encoding.Packet) error {
	if packet.GatewayEUI == nil {
		return errNoEUI.New()
	}
	if packet.GatewayAddr == nil {
		return errNoAddress.New()
	}
	now := time.Now().UTC()
	eui := *packet.GatewayEUI
	val, ok := f.m.Load(eui)
	var ts *timestamps
	if ok {
		ts = val.(*timestamps)
	} else {
		ts = newTimestamps(f.messages)
		f.m.Store(eui, ts)
	}

	oldestTimestamp := ts.Append(now)
	if !oldestTimestamp.IsZero() && now.Sub(oldestTimestamp) < f.threshold {
		return errRateExceeded.New()
	}

	// Continue filtering
	if f.f != nil {
		return f.f.Filter(packet)
	}
	return nil
}
