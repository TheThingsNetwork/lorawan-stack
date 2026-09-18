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

// The limiter is unexported, so its tests live in the package under test.
package redis

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

const (
	testExpireInterval = time.Hour
	testExpireTick     = testExpireInterval / expireTicksPerInterval
	expireWrapPeriod   = (expireTickMask + 1) * testExpireTick
)

var testExpireEpoch = time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC)

func BenchmarkExpireLimiter(b *testing.B) {
	l := newExpireLimiter(defaultExpireLimiterSlots, testExpireInterval, testExpireEpoch)
	keys := make([]string, 1024)
	for i := range keys {
		keys[i] = fmt.Sprintf("ttn:v3:event_stream:device:app.dev-%d", i)
	}

	for i := 0; b.Loop(); i++ {
		l.allow(keys[i%len(keys)], testExpireEpoch)
	}
}

// TestExpireLimiterAllocations is not parallel, because testing.AllocsPerRun
// panics while a parallel test is running.
func TestExpireLimiterAllocations(t *testing.T) { //nolint:paralleltest
	a := assertions.New(t)
	l := newExpireLimiter(1024, testExpireInterval, testExpireEpoch)

	a.So(testing.AllocsPerRun(100, func() {
		l.allow("stream", testExpireEpoch)
	}), should.BeZeroValue)
}

// newTestExpireLimiter returns a limiter with a single set, so that every key
// competes for the same expireLimiterWays slots.
func newTestExpireLimiter() *expireLimiter {
	return newExpireLimiter(expireLimiterWays, testExpireInterval, testExpireEpoch)
}

func TestExpireLimiter(t *testing.T) {
	t.Parallel()

	t.Run("AllowsOncePerInterval", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newTestExpireLimiter()

		a.So(l.allow("stream", testExpireEpoch), should.BeTrue)
		a.So(l.allow("stream", testExpireEpoch), should.BeFalse)
		a.So(l.allow("stream", testExpireEpoch.Add(testExpireInterval-testExpireTick)), should.BeFalse)
		a.So(l.allow("stream", testExpireEpoch.Add(testExpireInterval)), should.BeTrue)
		a.So(l.allow("stream", testExpireEpoch.Add(testExpireInterval)), should.BeFalse)
	})

	t.Run("AllowsKeyPublishingWithinEveryInterval", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newTestExpireLimiter()

		// A key publishing twice per interval must still be refreshed once per
		// interval, or its stream expires in Redis.
		allowed := 0
		for i := range 20 {
			if l.allow("stream", testExpireEpoch.Add(time.Duration(i)*testExpireInterval/2)) {
				allowed++
			}
		}
		a.So(allowed, should.Equal, 10)
	})

	t.Run("KeepsDistinctKeysApart", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newTestExpireLimiter()

		for i := range expireLimiterWays {
			a.So(l.allow(fmt.Sprintf("stream-%d", i), testExpireEpoch), should.BeTrue)
		}
		for i := range expireLimiterWays {
			a.So(l.allow(fmt.Sprintf("stream-%d", i), testExpireEpoch), should.BeFalse)
		}
	})

	t.Run("AllowsAfterEviction", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newTestExpireLimiter()

		// Filling every way of the only set evicts the oldest key, which must then
		// be allowed again rather than silently suppressed.
		for i := range expireLimiterWays + 1 {
			at := testExpireEpoch.Add(time.Duration(i) * testExpireTick)
			a.So(l.allow(fmt.Sprintf("stream-%d", i), at), should.BeTrue)
		}
		a.So(l.allow("stream-0", testExpireEpoch.Add(expireLimiterWays*testExpireTick)), should.BeTrue)
	})

	t.Run("RateLimitsAcrossTickWrap", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newTestExpireLimiter()

		// Ages are computed modulo the tick counter, so limiting must survive the
		// counter wrapping rather than degrading to one refresh per event. Each
		// period is added separately: several of them overflow a time.Duration.
		at := testExpireEpoch
		for range 2 {
			at = at.Add(expireWrapPeriod)
			just := at.Add(-testExpireTick) // Just before the counter wraps.
			a.So(l.allow("wrap", just), should.BeTrue)
			a.So(l.allow("wrap", just.Add(testExpireInterval-testExpireTick)), should.BeFalse)
			a.So(l.allow("wrap", just.Add(testExpireInterval)), should.BeTrue)
		}
	})

	t.Run("AllowsBeforeEpoch", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newTestExpireLimiter()

		a.So(l.allow("stream", testExpireEpoch.Add(-testExpireTick)), should.BeTrue)
		a.So(l.allow("stream", testExpireEpoch.Add(-testExpireTick)), should.BeTrue)
	})

	t.Run("AllowsWhenDisabled", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)

		for _, l := range []*expireLimiter{
			nil,
			newExpireLimiter(-1, testExpireInterval, testExpireEpoch),
			newExpireLimiter(0, testExpireInterval, testExpireEpoch),
			newExpireLimiter(expireLimiterWays, 0, testExpireEpoch),
		} {
			a.So(l, should.BeNil)
			a.So(l.allow("stream", testExpireEpoch), should.BeTrue)
			a.So(l.allow("stream", testExpireEpoch), should.BeTrue)
		}
	})

	t.Run("RefreshesWellInsideTheJitteredTTL", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)

		// A tick rounds the next refresh up, so it must land inside publish's
		// jittered PEXPIRE, and no sooner than the interval it holds.
		for _, ttl := range []time.Duration{
			time.Second,
			2*time.Second + time.Millisecond,
			time.Minute,
			24 * time.Hour,
		} {
			l := newExpireLimiter(expireLimiterWays, ttl/2, testExpireEpoch)
			a.So(l, should.NotBeNil)

			// The epoch sits on a tick boundary, the worst phase to be refreshed at.
			a.So(l.allow("stream", testExpireEpoch), should.BeTrue)
			step, next := ttl/1000, ttl
			for at := time.Duration(0); at < ttl; at += step {
				if l.allow("stream", testExpireEpoch.Add(at)) {
					next = at
					break
				}
			}
			a.So(next, should.BeGreaterThanOrEqualTo, ttl/2-step)
			a.So(next, should.BeLessThan, time.Duration(float64(ttl)*(1-ttlJitter)))
		}
	})

	t.Run("DisablesItselfBelowOneTick", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)

		// A tick unit that rounds to zero disables the limiter rather than dividing
		// by it.
		const oneTick = expireTicksPerInterval * time.Nanosecond
		a.So(newExpireLimiter(expireLimiterWays, oneTick-time.Nanosecond, testExpireEpoch), should.BeNil)
		a.So(newExpireLimiter(expireLimiterWays, oneTick, testExpireEpoch), should.NotBeNil)
	})

	t.Run("ClampsSmallSizesToOneSet", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)

		// A positive size below one set limits rather than silently disabling.
		for slots := 1; slots < expireLimiterWays; slots++ {
			l := newExpireLimiter(slots, testExpireInterval, testExpireEpoch)
			a.So(len(l.slots), should.Equal, expireLimiterWays)
			a.So(l.allow("stream", testExpireEpoch), should.BeTrue)
			a.So(l.allow("stream", testExpireEpoch), should.BeFalse)
		}
	})

	t.Run("RoundsTableUpToWholeSets", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)

		for slots, want := range map[int]int{
			expireLimiterWays:         1 * expireLimiterWays,
			expireLimiterWays + 1:     2 * expireLimiterWays,
			3 * expireLimiterWays:     4 * expireLimiterWays,
			3*expireLimiterWays + 1:   4 * expireLimiterWays,
			4 * expireLimiterWays:     4 * expireLimiterWays,
			defaultExpireLimiterSlots: defaultExpireLimiterSlots,
		} {
			l := newExpireLimiter(slots, testExpireInterval, testExpireEpoch)
			a.So(len(l.slots), should.Equal, want)
		}
	})

	t.Run("KeepsTableSizeConstant", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newExpireLimiter(1024, testExpireInterval, testExpireEpoch)
		size := len(l.slots)

		for i := range 100_000 {
			l.allow(fmt.Sprintf("stream-%d", i), testExpireEpoch)
		}
		a.So(len(l.slots), should.Equal, size)
	})

	t.Run("AllowsEveryKeyUnderConcurrency", func(t *testing.T) {
		t.Parallel()
		a := assertions.New(t)
		l := newExpireLimiter(1024, testExpireInterval, testExpireEpoch)

		const keys, publishers = 64, 8
		allowed := make([]atomic.Int64, keys)
		wg := sync.WaitGroup{}
		for range publishers {
			wg.Go(func() {
				for i := range keys {
					if l.allow(fmt.Sprintf("stream-%d", i), testExpireEpoch) {
						allowed[i].Add(1)
					}
				}
			})
		}
		wg.Wait()

		for i := range keys {
			a.So(allowed[i].Load(), should.BeGreaterThanOrEqualTo, int64(1))
		}
	})
}
