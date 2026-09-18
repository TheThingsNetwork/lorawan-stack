// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

package redis

import (
	"hash/maphash"
	"math/bits"
	"sync/atomic"
	"time"
)

const (
	// Slots when events.redis.store.expire-limiter-size is unset: 8 bytes each, so
	// a flat 32 MiB, holding the ideal PEXPIRE rate to about a million streams.
	defaultExpireLimiterSlots = 1 << 22
	// Slots a key may occupy. Eight of 8 bytes are one cache line, so scanning a
	// whole set is a single fetch.
	expireLimiterWays = 8
)

// Slot layout: a non-zero hash tag, then the tick the key was last allowed at,
// so an all-zero slot reads as unused. Ages are computed modulo the tick
// counter, so it never runs out. A tick is a fraction of the interval, so 24
// bits span a million intervals and the tag takes the rest: two keys sharing a
// set and a tag are limited as one, stranding the quieter stream without a TTL.
const (
	expireTicksPerInterval = 16
	expireTickBits         = 24
	expireTickMask         = 1<<expireTickBits - 1
)

// expireLimiter holds PEXPIRE down to one command per key per interval, in a
// fixed pointer-free table. A key may use any way of its set, so two busy keys
// do not evict each other. Anything uncertain allows a PEXPIRE, never skips one.
type expireLimiter struct {
	slots    []atomic.Uint64
	setMask  uint64
	seed     maphash.Seed
	tickUnit time.Duration
	epoch    time.Time
}

// newExpireLimiter allows one PEXPIRE per key per interval over slots entries,
// rounded up to whole sets. A non-positive size, or an interval too short to
// divide into ticks, returns nil, which allows everything.
func newExpireLimiter(slots int, interval time.Duration, epoch time.Time) *expireLimiter {
	tickUnit := interval / expireTicksPerInterval
	if slots <= 0 || tickUnit <= 0 {
		return nil
	}
	sets := uint64((slots + expireLimiterWays - 1) / expireLimiterWays)
	sets = 1 << bits.Len64(sets-1) // Round up to a power of two, so setMask works.
	return &expireLimiter{
		slots:    make([]atomic.Uint64, sets*expireLimiterWays),
		setMask:  sets - 1,
		seed:     maphash.MakeSeed(),
		tickUnit: tickUnit,
		epoch:    epoch,
	}
}

// allow reports whether key needs a refresh at now, recording it only if so.
func (l *expireLimiter) allow(key string, now time.Time) bool {
	if l == nil {
		return true
	}
	elapsed := now.Sub(l.epoch)
	if elapsed < 0 {
		return true
	}
	tick := uint64(int64(elapsed/l.tickUnit) & expireTickMask)
	h := maphash.String(l.seed, key)
	tag, base := h>>expireTickBits, (h&l.setMask)*expireLimiterWays
	if tag == 0 {
		tag = 1 // Zero is reserved for unused slots, so no key may claim it.
	}

	var target, oldestAge uint64
	for way := range uint64(expireLimiterWays) {
		slot := l.slots[base+way].Load()
		last := slot & expireTickMask
		age := uint64(expireTickMask) // An unused slot is evicted before any live one.
		if slot != 0 {
			age = (tick - last) & expireTickMask
		}
		if slot>>expireTickBits == tag {
			if age < expireTicksPerInterval {
				return false
			}
			target = way
			break
		}
		if age > oldestAge {
			target, oldestAge = way, age
		}
	}
	l.slots[base+target].Store(tag<<expireTickBits | tick)
	return true
}
