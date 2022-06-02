// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package workerpool_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
)

func TestAtomicConditionals(t *testing.T) {
	var release sync.WaitGroup
	release.Add(1)

	var wg sync.WaitGroup

	var increaseFailures sync.Map
	var decreaseFailures sync.Map

	value := int32(5_000)
	lowerBound := int32(100)
	upperBound := int32(10_000)
	for k := 0; k < 100_000; k++ {
		k := k

		wg.Add(1)
		go func() {
			release.Wait()
			defer wg.Done()
			if workerpool.IncrementIfSmallerThan(&value, upperBound) {
				if v := atomic.LoadInt32(&value); v > upperBound {
					increaseFailures.Store(k, v)
				}
			}
		}()

		wg.Add(1)
		go func() {
			release.Wait()
			defer wg.Done()
			if workerpool.DecrementIfGreaterThan(&value, lowerBound) {
				if v := atomic.LoadInt32(&value); v < lowerBound {
					decreaseFailures.Store(k, v)
				}
			}
		}()

		wg.Add(1)
		go func() {
			release.Wait()
			defer wg.Done()
			if v := atomic.LoadInt32(&value); v > upperBound {
				increaseFailures.Store(k, v)
			}
			if v := atomic.LoadInt32(&value); v < lowerBound {
				decreaseFailures.Store(k, v)
			}
		}()
	}

	release.Done()
	wg.Wait()

	increaseFailures.Range(func(ki interface{}, vi interface{}) bool {
		k, v := ki.(int), vi.(int32)
		t.Errorf("Value %v exceeded upper bound %v in test %v", v, upperBound, k)
		return true
	})

	decreaseFailures.Range(func(ki interface{}, vi interface{}) bool {
		k, v := ki.(int), vi.(int32)
		t.Errorf("Value %v exceeded lower bound %v in test %v", v, lowerBound, k)
		return true
	})
}

var (
	fastTimeout = test.Delay
	slowTimeout = 10 * test.Delay
	testTimeout = 100 * test.Delay
)

func TestWorkerPool(t *testing.T) {
	for _, workerIdleTimeout := range []time.Duration{slowTimeout, fastTimeout} {
		for _, queueSize := range []int{-1, 0, 1} {
			for _, minWorkers := range []int{-1, 0, 1} {
				for _, maxWorkers := range []int{0, 1} {
					name := fmt.Sprintf(
						"minWorkers/%v/maxWorkers/%v/queueSize/%v/idleTimeout/%v",
						minWorkers,
						maxWorkers,
						queueSize,
						workerIdleTimeout,
					)
					t.Run(name, func(t *testing.T) {
						t.Parallel()
						testWorkerPool(t, minWorkers, maxWorkers, queueSize, workerIdleTimeout)
					})
				}
			}
		}
	}
}

func testWorkerPool(t *testing.T, minWorkers int, maxWorkers int, queueSize int, workerIdleTimeout time.Duration) {
	a, ctx := test.New(t)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workCtx := context.WithValue(ctx, "foo", "bar")

	var workToBeDone, workDone, workFailed, duplicatedWork sync.Map
	handlerCalls := int32(0)
	handler := func(ctx context.Context, item int) {
		atomic.AddInt32(&handlerCalls, 1)
		a.So(ctx, should.HaveParentContextOrEqual, workCtx)
		if item == -1 {
			panic("boom")
		}
		if _, exists := workDone.LoadOrStore(item, 0); exists {
			duplicatedWork.Store(item, 0)
		}
	}

	wp := workerpool.NewWorkerPool(workerpool.Config[int]{
		Component:         &mockComponent{},
		Context:           ctx,
		Handler:           handler,
		MinWorkers:        minWorkers,
		MaxWorkers:        maxWorkers,
		QueueSize:         queueSize,
		WorkerIdleTimeout: workerIdleTimeout,
	})

	totalWork := 100_000
	expectedHandlerCalls := int32(0)
	for i := 0; i < totalWork; i++ {
		if err := wp.Publish(workCtx, i); err != nil {
			workFailed.Store(i, 0)
		} else {
			workToBeDone.Store(i, 0)
			expectedHandlerCalls++
		}

		if rand.Intn(100) < 5 {
			if err := wp.Publish(workCtx, -1); err == nil {
				expectedHandlerCalls++
			}
		}
	}

	time.Sleep(testTimeout)
	cancel()
	wp.Wait()

	var countDone, countToBeDone, countFailed int
	workDone.Range(func(k, v interface{}) bool {
		_, failed := workFailed.Load(k)
		a.So(failed, should.BeFalse)

		_, toBeDone := workToBeDone.Load(k)
		a.So(toBeDone, should.BeTrue)

		countDone++

		return true
	})
	workToBeDone.Range(func(k, v interface{}) bool {
		_, done := workDone.Load(k)
		a.So(done, should.BeTrue)

		countToBeDone++

		return true
	})
	workFailed.Range(func(k, v interface{}) bool {
		countFailed++

		return true
	})
	duplicatedWork.Range(func(k, v interface{}) bool {
		t.Fatalf("Item %v was processed multiple times", k)
		return true
	})

	a.So(countDone, should.Equal, countToBeDone)
	a.So(countFailed, should.Equal, totalWork-countDone)
	a.So(handlerCalls, should.Equal, expectedHandlerCalls)
}

type mockComponent struct{}

func (*mockComponent) StartTask(cfg *task.Config) {
	task.DefaultStartTask(cfg)
}

func (*mockComponent) FromRequestContext(ctx context.Context) context.Context {
	return ctx
}

func benchmarkWorkerPool(b *testing.B, processingDelay time.Duration, publishingDelay time.Duration) {
	_, ctx := test.New(b)

	var totalQueueDelayMS int64
	var totalHandled int64
	var published, dropped int64

	for r := 0; r < b.N; r++ {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handler := func(ctx context.Context, tm time.Time) {
			delay := time.Now().Sub(tm).Milliseconds()
			atomic.AddInt64(&totalQueueDelayMS, delay)
			atomic.AddInt64(&totalHandled, 1)

			time.Sleep(random.Jitter(processingDelay, 0.15))
		}

		wp := workerpool.NewWorkerPool(workerpool.Config[time.Time]{
			Component: &mockComponent{},
			Context:   ctx,
			Handler:   handler,
		})

		var wg sync.WaitGroup
		publisher := func() {
			defer wg.Done()
			for p := 0; p < 1_000; p++ {
				if err := wp.Publish(ctx, time.Now()); err != nil {
					atomic.AddInt64(&dropped, 1)
				} else {
					atomic.AddInt64(&published, 1)
				}
				time.Sleep(random.Jitter(publishingDelay, 0.15))
			}
		}

		for i := 0; i < workerpool.DefaultMaxWorkers; i++ {
			wg.Add(1)
			go publisher()
		}

		wg.Wait()

		time.Sleep(testTimeout)
		cancel()
		wp.Wait()
	}

	b.ReportMetric(float64(totalQueueDelayMS)/float64(totalHandled), "queueDelayMS")
	b.ReportMetric(float64(published), "published")
	b.ReportMetric(float64(dropped), "dropped")
}

func BenchmarkWorkerPool(b *testing.B) {
	delays := []time.Duration{5 * time.Millisecond, 10 * time.Millisecond, 50 * time.Millisecond}
	for _, processingDelay := range delays {
		for _, publishingDelay := range delays {
			name := fmt.Sprintf(
				"processingDelay/%v/publishingDelay/%v",
				processingDelay,
				publishingDelay,
			)
			b.Run(name, func(b *testing.B) {
				benchmarkWorkerPool(b, processingDelay, publishingDelay)
			})
		}
	}
}
