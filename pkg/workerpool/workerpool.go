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

package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
)

const (
	// defaultWorkerIdleTimeout is the duration after which an idle worker stops to save resources.
	defaultWorkerIdleTimeout = 1 * time.Second
	// defaultQueueSize is the default queue size for the worker pool.
	defaultQueueSize = 32
	// defaultMinWorkers is the default number of minimum workers kept in the pool.
	defaultMinWorkers = 4
	// defaultMaxWorkers is the default number of maximum workers kept in the pool.
	defaultMaxWorkers = 1024
)

// Component contains a minimal component.Component definition.
type Component interface {
	task.Starter
	FromRequestContext(context.Context) context.Context
}

// Handler is a function that processes items published to the worker pool.
type Handler[T any] func(ctx context.Context, item T)

// Config is the configuration of the worker pool.
type Config[T any] struct {
	Component
	context.Context                 // The base context of the pool.
	Name              string        // The name of the pool.
	Handler           Handler[T]    // The work handler.
	MinWorkers        int           // The minimum number of workers in the pool. Use -1 to disable.
	MaxWorkers        int           // The maximum number of workers in the pool.
	QueueSize         int           // The size of the work queue. Use -1 to disable.
	WorkerIdleTimeout time.Duration // The maximum amount of time a worker will stay idle before closing.
}

// WorkerPool is a dynamic pool of workers to which work items can be published.
// The workers are created on demand and live as long as work is available.
type WorkerPool[T any] interface {
	// Publish publishes an item to the worker pool to be processed.
	// Publish may spawn a worker in order to fullfil the work load.
	// Publish does not block.
	Publish(ctx context.Context, item T) error

	// Wait blocks until all workers have been closed.
	Wait()
}

type contextualItem[T any] struct {
	ctx      context.Context
	item     T
	queuedAt time.Time
}

type workerPool[T any] struct {
	Config[T]

	mainQueue chan *contextualItem[T] // mainQueue allows items to be buffered between publishers and workers.
	fastQueue chan *contextualItem[T] // fastQueue allows direct communication between publishers and idle workers.

	workers int32
	wg      sync.WaitGroup
}

func (wp *workerPool[T]) handle(it *contextualItem[T]) {
	registerWorkerBusy(wp.Name)
	defer registerWorkerIdle(wp.Name)
	defer registerWorkProcessed(it.ctx, wp.Name)
	defer registerWorkLatency(wp.Name, time.Now())
	wp.Handler(it.ctx, it.item)
}

func (wp *workerPool[T]) workerBody(initialWork *contextualItem[T]) func(context.Context) error {
	worker := func(ctx context.Context) error {
		var timeout bool
		defer func() {
			if timeout {
				return
			}
			atomic.AddInt32(&wp.workers, -1)
			select {
			case <-ctx.Done():
			case <-wp.Done():
			default:
				// Since the task did not finish due to a context cancellation
				// or a timeout, the worker body must have panicked. As such
				// we attempt to spawn a replacement worker in order to avoid
				// stalling the queue indefinitely.
				wp.spawnWorker(nil)
			}
		}()

		defer wp.wg.Done()
		defer registerWorkerStopped(wp.Name)

		registerWorkerIdle(wp.Name)
		defer registerWorkerBusy(wp.Name)

		if initialWork != nil {
			wp.handle(initialWork)
		}

		for {
			select {
			case <-wp.Done():
				return wp.Err()

			case <-ctx.Done():
				return ctx.Err()

			case <-time.After(wp.WorkerIdleTimeout):
				if decrementIfGreaterThan(&wp.workers, int32(wp.MinWorkers)) {
					timeout = true
					return nil
				}

			case item := <-wp.fastQueue:
				wp.handle(item)

			case item := <-wp.mainQueue:
				registerWorkDequeued(wp.Name, item.queuedAt)
				wp.handle(item)
			}
		}
	}
	return worker
}

// spawnWorker spawns a worker task, if a worker slot is available.
func (wp *workerPool[T]) spawnWorker(initialWork *contextualItem[T]) bool {
	if !incrementIfSmallerThan(&wp.workers, int32(wp.MaxWorkers)) {
		return false
	}

	registerWorkerStarted(wp.Name)
	wp.wg.Add(1)

	wp.StartTask(&task.Config{
		Context: wp.Context,
		ID:      wp.Name,
		Func:    wp.workerBody(initialWork),
		Restart: task.RestartNever,
		Backoff: task.DefaultBackoffConfig,
	})

	return true
}

var errPoolFull = errors.DefineResourceExhausted("pool_full", "the worker pool is full")

// enqueueSpawn attempts to enqueue the work item, spawning a worker task if possible.
// If an idle worker can pickup the work, the work is provided to the idle worker.
// If the work item can be enqueued, it will be enqueued, and the pool will attempt
// to spawn an extra worker.
// If the work cannot be enqueued, the pool will attempt to spawn an extra worker
// that will handle the work. If this fails, the work is dropped.
func (wp *workerPool[T]) enqueueSpawn(ctx context.Context, it *contextualItem[T]) error {
	// select is fair, and as such if two possible communication paths are possible
	// (both fastQueue, and mainQueue) the one which will proceed is chosen based
	// on a uniform pseudo-random selection. As such, we initially attempt to submit
	// the work directly via the fast queue, and only if that fails we attempt to use
	// the main buffered queue.
	select {
	case <-wp.Done():
		return wp.Err()

	case <-ctx.Done():
		return ctx.Err()

	case wp.fastQueue <- it:
		return nil

	default:
	}

	select {
	case <-wp.Done():
		return wp.Err()

	case <-ctx.Done():
		return ctx.Err()

	case wp.fastQueue <- it:
		return nil

	case wp.mainQueue <- it:
		registerWorkEnqueued(wp.Name)
		it = nil

	default:
	}

	if spawned := wp.spawnWorker(it); spawned || it == nil {
		return nil
	}

	registerWorkDropped(it.ctx, wp.Name)
	return errPoolFull.New()
}

// Publish implements WorkerPool.
func (wp *workerPool[T]) Publish(ctx context.Context, item T) error {
	return wp.enqueueSpawn(ctx, &contextualItem[T]{
		ctx:      wp.FromRequestContext(ctx),
		item:     item,
		queuedAt: time.Now(),
	})
}

// Wait implements WorkerPool.
func (wp *workerPool[T]) Wait() {
	wp.wg.Wait()
}

// NewWorkerPool creates a new WorkerPool with the provided configuration.
func NewWorkerPool[T any](cfg Config[T]) WorkerPool[T] {
	if cfg.WorkerIdleTimeout == 0 {
		cfg.WorkerIdleTimeout = defaultWorkerIdleTimeout
	}
	// We treat 0 as being default initialized, and use the defaults.
	if cfg.MinWorkers == 0 {
		cfg.MinWorkers = defaultMinWorkers
	}
	// We treat negative values as explicitly disabling the minimum number of workers.
	if cfg.MinWorkers < 0 {
		cfg.MinWorkers = 0
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = defaultMaxWorkers
	}
	// We treat 0 as being default initialized, and use the defaults.
	if cfg.QueueSize == 0 {
		cfg.QueueSize = defaultQueueSize
	}
	// We treat negative values as explicitly disabling the queue.
	if cfg.QueueSize < 0 {
		cfg.QueueSize = 0
	}
	if cfg.MinWorkers > cfg.MaxWorkers {
		cfg.MaxWorkers = cfg.MinWorkers
	}

	wp := &workerPool[T]{
		Config: cfg,

		mainQueue: make(chan *contextualItem[T], cfg.QueueSize),
		fastQueue: make(chan *contextualItem[T]),
	}

	for i := 0; i < wp.MinWorkers; i++ {
		wp.spawnWorker(nil)
	}

	return wp
}

func incrementIfSmallerThan(i *int32, max int32) bool {
	for v := atomic.LoadInt32(i); v < max; v = atomic.LoadInt32(i) {
		if atomic.CompareAndSwapInt32(i, v, v+1) {
			return true
		}
	}
	return false
}

func decrementIfGreaterThan(i *int32, min int32) bool {
	for v := atomic.LoadInt32(i); v > min; v = atomic.LoadInt32(i) {
		if atomic.CompareAndSwapInt32(i, v, v-1) {
			return true
		}
	}
	return false
}
