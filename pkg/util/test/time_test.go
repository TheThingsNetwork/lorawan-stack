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

package test_test

import (
	"testing"
	"time"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMockTime(t *testing.T) {
	a := assertions.New(t)

	now := time.Unix(0, 42)
	clock := NewMockClock(now)
	a.So(clock.Now(), should.Resemble, now)

	d := 6*time.Hour + 5*time.Minute + 4*time.Second + 3*time.Millisecond + 2*time.Microsecond + 1*time.Nanosecond
	now = now.Add(d)
	a.So(clock.Add(d), should.Resemble, now)
	a.So(clock.Now(), should.Resemble, now)

	oldNow := now
	now = now.Add(d)
	a.So(clock.Set(now), should.Resemble, oldNow)
	a.So(clock.Now(), should.Resemble, now)

	n := 5
	afterCh := clock.After(time.Duration(n) * time.Nanosecond)
	for i := 0; i < n; i++ {
		select {
		case <-afterCh:
			t.Error("After channel read succeeded too soon")
		default:
		}
		now = now.Add(time.Nanosecond)
		a.So(clock.Add(time.Nanosecond), should.Resemble, now)
	}
	// Let the goroutine in After send the time on the channel.
	time.Sleep(Delay)
	select {
	case afterNow := <-afterCh:
		a.So(afterNow, should.Resemble, now)
	default:
		t.Error("After channel read should have succeeded")
	}

	select {
	case afterNow := <-afterCh:
		a.So(afterNow, should.BeZeroValue)
	default:
		t.Error("After channel should have been closed")
	}
}
