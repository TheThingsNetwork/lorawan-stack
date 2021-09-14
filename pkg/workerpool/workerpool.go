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
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

const (
	// defaultWorkerIdleTimeout is the duration after which an idle worker stops to save resources.
	defaultWorkerIdleTimeout = 1 * time.Second
	// defaultQueueSize is the default queue size for the worker pool.
	defaultQueueSize = 32
	// defaultMinWorkers is the default number of minimum workers kept in the pool.
	defaultMinWorkers = 4
	// defaultMaxWorkers is the default number of maximum workers kept in the pool.
	defaultMaxWorkers = 64
)

// Component contains a minimal component.Component definition.
type Component interface {
	StartTask(*component.TaskConfig)
	FromRequestContext(context.Context) context.Context
}

// Handler is a function that processes items published to the worker pool.
type Handler func(ctx context.Context, item interface{})

// HandlerFactory is a function that creates a Handler.
type HandlerFactory func() (Handler, error)

// StaticHandlerFactory creates a HandlerFactory that always returns the same handler.
func StaticHandlerFactory(f Handler) HandlerFactory {
	return func() (Handler, error) {
		return f, nil
	}
}

// Config is the configuration of the worker pool.
type Config struct {
	Component
	context.Context                  // The base context of the pool.
	Name              string         // The name of the pool.
	CreateHandler     HandlerFactory // The function that creates handlers.
	MinWorkers        int            // The minimum number of workers in the pool.
	MaxWorkers        int            // The maximum number of workers in the pool.
	QueueSize         int            // The size of the work queue.
	WorkerIdleTimeout time.Duration  // The maximum amount of time a worker will stay idle before closing.
}

// WorkerPool is a dynamic pool of workers to which work items can be published.
// The workers are created on demand and live as long as work is available.
type WorkerPool interface {
	// Publish publishes an item to the worker pool to be processed.
	// Publish may spawn a worker in order to fullfil the work load.
	// Publish does not block.
	Publish(ctx context.Context, item interface{}) error
}

type contextualItem struct {
	ctx  context.Context
	item interface{}
}

type workerPool struct {
	Config

	mainQueue chan *contextualItem // mainQueue allows items to be buffered between publishers and workers.
	fastQueue chan *contextualItem // fastQueue allows direct communication between publishers and idle workers.

	workers int32
}

func (wp *workerPool) handle(ctx context.Context, it *contextualItem, handler Handler) {
	registerWorkerBusy(wp.Name)
	defer registerWorkerIdle(wp.Name)
	defer registerWorkProcessed(it.ctx, wp.Name)
	defer registerWorkLatency(wp.Name, time.Now())
	handler(it.ctx, it.item)
}

func (wp *workerPool) workerBody(handler Handler, initialWork *contextualItem) func(context.Context) error {
	worker := func(ctx context.Context) error {
		var decremented bool
		defer func() {
			if !decremented {
				atomic.AddInt32(&wp.workers, -1)
			}
		}()

		defer registerWorkerStopped(wp.Name)

		registerWorkerIdle(wp.Name)
		defer registerWorkerBusy(wp.Name)

		if initialWork != nil {
			wp.handle(ctx, initialWork, handler)
		}

		for {
			select {
			case <-wp.Done():
				return wp.Err()

			case <-ctx.Done():
				return ctx.Err()

			case <-time.After(wp.WorkerIdleTimeout):
				if decrementIfGreaterThan(&wp.workers, int32(wp.MinWorkers)) {
					decremented = true
					return nil
				}

			case item := <-wp.fastQueue:
				wp.handle(ctx, item, handler)

			case item := <-wp.mainQueue:
				registerWorkDequeued(wp.Name)
				wp.handle(ctx, item, handler)
			}
		}
	}
	return worker
}

// spawnWorker spawns a worker task, if a worker slot is available.
func (wp *workerPool) spawnWorker(initialWork *contextualItem) (bool, error) {
	handler, err := wp.CreateHandler()
	if err != nil {
		return false, err
	}

	if !incrementIfSmallerThan(&wp.workers, int32(wp.MaxWorkers)) {
		return false, nil
	}

	registerWorkerStarted(wp.Name)

	wp.StartTask(&component.TaskConfig{
		Context: wp.Context,
		ID:      wp.Name,
		Func:    wp.workerBody(handler, initialWork),
		Restart: component.TaskRestartNever,
		Backoff: component.DefaultTaskBackoffConfig,
	})

	return true, nil
}

var errPoolFull = errors.DefineResourceExhausted("pool_full", "the worker pool is full")

// enqueueSpawn attempts to enqueue the work item, spawning a worker task if possible.
// If an idle worker can pickup the work, the work is provided to the idle worker.
// If the work item can be enqueued, it will be enqueued, and the pool will attempt
// to spawn an extra worker.
// If the work cannot be enqueued, the pool will attempt to spawn an extra worker
// that will handle the work. If this fails, the work is dropped.
func (wp *workerPool) enqueueSpawn(ctx context.Context, it *contextualItem) error {
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

	spawned, err := wp.spawnWorker(it)
	// err == nil if spawned || it == nil
	// which is fine as the work is either picked up by the new worker
	// or was already placed in the queue.
	if err != nil || spawned || it == nil {
		return err
	}

	registerWorkDropped(it.ctx, wp.Name)
	return errPoolFull.New()
}

// Publish implements WorkerPool.
func (wp *workerPool) Publish(ctx context.Context, item interface{}) error {
	return wp.enqueueSpawn(ctx, &contextualItem{
		ctx:  wp.FromRequestContext(ctx),
		item: item,
	})
}

// NewWorkerPool creates a new WorkerPool with the provided configuration.
func NewWorkerPool(cfg Config) (WorkerPool, error) {
	if cfg.WorkerIdleTimeout == 0 {
		cfg.WorkerIdleTimeout = defaultWorkerIdleTimeout
	}
	if cfg.MinWorkers <= 0 {
		cfg.MinWorkers = defaultMinWorkers
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = defaultMaxWorkers
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = defaultQueueSize
	}
	if cfg.MinWorkers > cfg.MaxWorkers {
		cfg.MaxWorkers = cfg.MinWorkers
	}

	wp := &workerPool{
		Config: cfg,

		mainQueue: make(chan *contextualItem, cfg.QueueSize),
		fastQueue: make(chan *contextualItem, 0),
	}

	for i := 0; i < wp.MinWorkers; i++ {
		if _, err := wp.spawnWorker(nil); err != nil {
			return nil, err
		}
	}

	return wp, nil
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
