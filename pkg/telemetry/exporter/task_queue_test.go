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

package telemetry_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	. "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockTaskQueue struct {
	dispatchCounter map[string]uint
	popCounter      map[string]uint
	sync.Mutex
}

func (mtq *mockTaskQueue) GetDispatchCounter(consumerID string) uint {
	mtq.Lock()
	defer mtq.Unlock()
	return mtq.dispatchCounter[consumerID]
}

func (mtq *mockTaskQueue) GetPopCounter(consumerID string) uint {
	mtq.Lock()
	defer mtq.Unlock()
	return mtq.popCounter[consumerID]
}

func (*mockTaskQueue) Add(context.Context, string, time.Time, bool) error {
	return nil
}

func (*mockTaskQueue) RegisterCallback(string, TaskCallback) {}

func (mtq *mockTaskQueue) Dispatch(_ context.Context, consumerID string) error {
	mtq.Lock()
	defer mtq.Unlock()
	mtq.dispatchCounter[consumerID]++
	return nil
}

func (mtq *mockTaskQueue) Pop(_ context.Context, consumerID string) error {
	mtq.Lock()
	defer mtq.Unlock()
	mtq.popCounter[consumerID]++
	return nil
}

// TestTaskWrappers tests that the task wrappers works as intended.
func TestTaskWrappers(t *testing.T) {
	t.Parallel()
	mtq := &mockTaskQueue{
		dispatchCounter: make(map[string]uint),
		popCounter:      make(map[string]uint),
	}
	t.Run("DispatchTask", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		consumerID := "dispatch_func"

		f := DispatchTask(mtq, consumerID)
		a.So(f(ctx), should.BeNil)
		a.So(mtq.GetDispatchCounter(consumerID), should.Equal, 1)
	})

	t.Run("PopTask", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		consumerID := "pop_func"

		f := PopTask(mtq, consumerID)
		a.So(f(ctx), should.BeNil)
		a.So(mtq.GetPopCounter(consumerID), should.Equal, 1)
	})
}

// TestConcurrentTaskSet adds two distinct tasks within the task queue in order to test the all of its operations.
//
// Involves in creating the dispatch loop, adding tasks, registering callback, popping tasks and validating the amount
// of times the callback have been called.
func TestConcurrentTaskSet(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	callbackDelay := 10 * test.Delay
	testTimeout := 200 * test.Delay

	// Add timeout to context. Garanties that the callbacks will be called a limited amount of times.
	ctx, cancel := context.WithTimeout(ctx, testTimeout)
	defer cancel()

	cl, closer := test.NewRedis(ctx, "telemetry_test")
	defer closer()

	redisConsumerGroup := "tm"
	tq, tqCloser, err := NewRedisTaskQueue(ctx, cl, 100000, redisConsumerGroup, redis.DefaultStreamBlockLimit)
	a.So(err, should.BeNil)
	defer func() {
		a.So(tqCloser(ctx), should.BeNil)
	}()

	// Add tasks.
	a.So(tq.Add(ctx, "test_task_1", time.Now().Add(callbackDelay), false), should.BeNil)
	a.So(tq.Add(ctx, "test_task_2", time.Now().Add(callbackDelay), false), should.BeNil)

	cntA := int64(0)
	cntB := int64(0)

	// Register callbacks.
	tq.RegisterCallback("test_task_1", func(ctx context.Context) (time.Time, error) {
		atomic.AddInt64(&cntA, 1)
		return time.Now().Add(callbackDelay), nil
	})
	tq.RegisterCallback("test_task_2", func(ctx context.Context) (time.Time, error) {
		atomic.AddInt64(&cntB, 1)
		return time.Now().Add(callbackDelay), nil
	})

	hostname, err := os.Hostname()
	a.So(err, should.BeNil)
	consumerIDPrefix := fmt.Sprintf("%s:%d", hostname, os.Getpid())

	errCh := make(chan error, 1)

	// Create the dispatch loop.
	go func(ctx context.Context, consumerID string) {
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
			default:
				if err := tq.Dispatch(ctx, consumerID); err != nil {
					errCh <- err
				}
			}
		}
	}(ctx, consumerIDPrefix)

	// Create the pop loop with 3 consumers.
	for i := 0; i < 3; i++ {
		consumerID := fmt.Sprintf("%s:%d", consumerIDPrefix, i)
		go func(ctx context.Context, consumerID string) {
			for {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
				default:
					if err := tq.Pop(ctx, consumerID); err != nil {
						errCh <- err
					}
				}
			}
		}(ctx, consumerID)
	}

	// Wait for the callbacks to be called a few times.
	time.Sleep(testTimeout / 2)

	select {
	case err := <-errCh:
		t.Fatalf("Timeout %v, Error: %v", testTimeout, err)
	default:
		// Test delay is 200 * test.Delay.
		// The default callback delay is 5 * test.Delay.
		// The time to call the callbacks a few times is 50 * test.Delay.
		// Considering the delay within interacting with redis, it is still expected to have at least 5 calls to each
		// callback.
		//
		a.So(atomic.LoadInt64(&cntA), should.BeGreaterThanOrEqualTo, 5)
		a.So(atomic.LoadInt64(&cntB), should.BeGreaterThanOrEqualTo, 5)
	}
}
