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

// Package batch contains a batch publisher implementation of events.Publisher.
package batch

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
)

type batchPublisher struct {
	ctx        context.Context
	publisher  events.Publisher
	targetSize int
	delay      time.Duration
	input      chan []events.Event
	output     chan []events.Event
}

func (bp *batchPublisher) process(ctx context.Context) error {
	batch, lastFlushAt := make([]events.Event, 0, bp.targetSize), time.Now()
	t := time.NewTimer(bp.delay)
	defer t.Stop()
	publish := func(lowerBound int, evs ...events.Event) bool {
		batch = append(batch, evs...)
		flushed := false
		for n := len(batch); n >= lowerBound; n = len(batch) {
			toFlush := n
			if upperBound := 2 * lowerBound; n > upperBound {
				toFlush = upperBound
			}
			select {
			case <-bp.ctx.Done():
			case <-ctx.Done():
			case bp.output <- batch[:toFlush]:
				registerBatchFlush(time.Since(lastFlushAt), toFlush)
				batch, lastFlushAt = batch[toFlush:], time.Now()
				flushed = true
			}
		}
		return flushed
	}
	for {
		select {
		case <-bp.ctx.Done():
			return bp.ctx.Err()
		case <-ctx.Done():
			return ctx.Err()
		case evs := <-bp.input:
			if publish(bp.targetSize, evs...) {
				if !t.Stop() {
					<-t.C
				}
				t.Reset(bp.delay)
			}
		case <-t.C:
			publish(1)
			t.Reset(bp.delay)
		}
	}
}

func (bp *batchPublisher) publish(ctx context.Context) error {
	for {
		select {
		case <-bp.ctx.Done():
			return bp.ctx.Err()
		case <-ctx.Done():
			return ctx.Err()
		case evs := <-bp.output:
			bp.publisher.Publish(evs...)
		}
	}
}

// Publish implements events.Publisher.
func (bp *batchPublisher) Publish(evs ...events.Event) {
	select {
	case <-bp.ctx.Done():
		return
	case bp.input <- evs:
	}
}

var _ events.Publisher = (*batchPublisher)(nil)

// NewPublisher returns a new batch publisher.
func NewPublisher(
	ctx context.Context,
	publisher events.Publisher,
	ts task.Starter,
	targetSize int,
	delay time.Duration,
	concurrency int,
) events.Publisher {
	bp := &batchPublisher{
		ctx:        ctx,
		publisher:  publisher,
		targetSize: targetSize,
		delay:      delay,
		input:      make(chan []events.Event, 1),
		output:     make(chan []events.Event, concurrency),
	}
	for name, t := range map[string]struct {
		f func(context.Context) error
		n int
	}{
		"batch_events_process": {bp.process, 1},
		"batch_events_publish": {bp.publish, concurrency},
	} {
		for i := 0; i < t.n; i++ {
			ts.StartTask(&task.Config{
				Context: bp.ctx,
				ID:      name,
				Func:    t.f,
				Restart: task.RestartOnFailure,
				Backoff: task.DefaultBackoffConfig,
			})
		}
	}
	return bp
}
