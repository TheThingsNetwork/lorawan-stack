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
	"sync"
	"sync/atomic"
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

	onPublish := func(evs ...events.Event) {}
	onPublishMu := &sync.Mutex{}

	publisher := events.PublishFunc(func(evs ...events.Event) {
		onPublishMu.Lock()
		defer onPublishMu.Unlock()
		onPublish(evs...)
	})

	ev0 := events.New(ctx, "some.event", "some description")
	ev1 := events.New(ctx, "another.event", "another description")
	ev2 := events.New(ctx, "yet.another.event", "yet another description")

	// Expect a flush due to a full batch.
	called := uint64(0)
	onPublishMu.Lock()
	onPublish = func(evs ...events.Event) {
		atomic.AddUint64(&called, 1)
		if a.So(evs, should.HaveLength, 2) {
			a.So(evs[0], should.ResembleEvent, ev0)
			a.So(evs[1], should.ResembleEvent, ev1)
		}
	}
	onPublishMu.Unlock()

	batcher := batch.NewPublisher(ctx, publisher, ts, 2, test.Delay, 2)
	batcher.Publish(ev0)
	batcher.Publish(ev1)
	time.Sleep(2 * test.Delay)
	a.So(atomic.LoadUint64(&called), should.Equal, 1)

	// Expect no publication after the flush.
	onPublishMu.Lock()
	onPublish = func(...events.Event) {
		atomic.AddUint64(&called, 1)
	}
	onPublishMu.Unlock()
	time.Sleep(2 * test.Delay)
	a.So(atomic.LoadUint64(&called), should.Equal, 1)

	// Expect a time triggered flush after the delay.
	onPublishMu.Lock()
	onPublish = func(evs ...events.Event) {
		atomic.AddUint64(&called, 1)
		if a.So(evs, should.HaveLength, 1) {
			a.So(evs[0], should.ResembleEvent, ev2)
		}
	}
	onPublishMu.Unlock()
	batcher.Publish(ev2)
	time.Sleep(2 * test.Delay)
	a.So(atomic.LoadUint64(&called), should.Equal, 2)

	// Expect a flush due to an overflow.
	onPublishMu.Lock()
	onPublish = func(evs ...events.Event) {
		atomic.AddUint64(&called, 1)
		if a.So(evs, should.HaveLength, 3) {
			a.So(evs[0], should.ResembleEvent, ev0)
			a.So(evs[1], should.ResembleEvent, ev1)
			a.So(evs[2], should.ResembleEvent, ev2)
		}
	}
	onPublishMu.Unlock()
	batcher.Publish(ev0, ev1, ev2)
	time.Sleep(2 * test.Delay)
	a.So(atomic.LoadUint64(&called), should.Equal, 3)
}
