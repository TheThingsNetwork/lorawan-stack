// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package batch_test

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/batch"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestBatchPublisher(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)
	ts := task.StartTaskFunc(task.DefaultStartTask)

	ch := make(chan []events.Event, 1)
	publisher := events.PublishFunc(func(evs ...events.Event) {
		select {
		case <-ctx.Done():
		case ch <- evs:
		}
	})

	ev0 := events.New(ctx, "some.event", "some description")
	ev1 := events.New(ctx, "another.event", "another description")
	ev2 := events.New(ctx, "yet.another.event", "yet another description")
	ev3 := events.New(ctx, "and.another.event", "and another description")
	ev4 := events.New(ctx, "and.yet.another.event", "and yet another description")

	batcher := batch.NewPublisher(ctx, publisher, ts, 2, test.Delay, 2)
	batcher.Publish(ev0)
	batcher.Publish(ev1)

	// Expect a flush due to a full batch.
	select {
	case <-ctx.Done():
	case <-time.After(2 * test.Delay):
		t.Fatal("timed out waiting for publication")
	case evs := <-ch:
		if a.So(evs, should.HaveLength, 2) {
			a.So(evs[0], should.ResembleEvent, ev0)
			a.So(evs[1], should.ResembleEvent, ev1)
		}
	}

	// Expect no publication after the flush.
	select {
	case <-ctx.Done():
	case <-time.After(2 * test.Delay):
	case <-ch:
		t.Fatal("unexpected publication")
	}

	// Expect a time triggered flush after the delay.
	batcher.Publish(ev2)
	select {
	case <-ctx.Done():
	case <-time.After(2 * test.Delay):
		t.Fatal("timed out waiting for publication")
	case evs := <-ch:
		if a.So(evs, should.HaveLength, 1) {
			a.So(evs[0], should.ResembleEvent, ev2)
		}
	}

	// Expect a flush due to an overflow.
	batcher.Publish(ev0, ev1, ev2)
	select {
	case <-ctx.Done():
	case <-time.After(2 * test.Delay):
		t.Fatal("timed out waiting for publication")
	case evs := <-ch:
		if a.So(evs, should.HaveLength, 3) {
			a.So(evs[0], should.ResembleEvent, ev0)
			a.So(evs[1], should.ResembleEvent, ev1)
			a.So(evs[2], should.ResembleEvent, ev2)
		}
	}

	// Expect two flushes due to an overflow.
	batcher.Publish(ev0, ev1, ev2, ev3, ev4)
	for i := 0; i < 2; i++ {
		had1, had4 := false, false
		select {
		case <-ctx.Done():
		case <-time.After(2 * test.Delay):
			t.Fatal("timed out waiting for publication")
		case evs := <-ch:
			switch n := len(evs); n {
			case 1:
				if a.So(had1, should.BeFalse) {
					a.So(evs[0], should.ResembleEvent, ev4)
					had1 = true
				}
			case 4:
				if a.So(had4, should.BeFalse) {
					a.So(evs[0], should.ResembleEvent, ev0)
					a.So(evs[1], should.ResembleEvent, ev1)
					a.So(evs[2], should.ResembleEvent, ev2)
					a.So(evs[3], should.ResembleEvent, ev3)
					had4 = true
				}
			default:
				t.Fatalf("unexpected number of events: %d", n)
			}
		}
	}
}
