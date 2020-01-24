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

package component

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
)

// TaskFunc is the task function.
type TaskFunc func(context.Context) error

// TaskRestart defines a task's restart policy.
type TaskRestart int

const (
	// TaskRestartNever denotes a restart policy that never restarts tasks after success or failure.
	TaskRestartNever TaskRestart = iota
	// TaskRestartAlways denotes a restart policy that always restarts tasks, on success and failure.
	TaskRestartAlways
	// TaskRestartOnFailure denotes a restart policy that restarts tasks on failure.
	TaskRestartOnFailure
)

var defaultTaskBackoff = [...]time.Duration{
	10 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	1 * time.Second,
}

// TaskBackoffDial is a backoff to use for tasks that are dialing services.
var TaskBackoffDial = []time.Duration{100 * time.Millisecond, 1 * time.Second, 10 * time.Second}

type task struct {
	ctx     context.Context
	id      string
	fn      TaskFunc
	restart TaskRestart
	backoff []time.Duration
}

// RegisterTask registers a task, optionally with restart policy and backoff, to be started after the component started.
func (c *Component) RegisterTask(ctx context.Context, id string, fn TaskFunc, restart TaskRestart, backoff ...time.Duration) {
	c.tasks = append(c.tasks, task{
		ctx:     ctx,
		id:      id,
		fn:      fn,
		restart: restart,
		backoff: backoff,
	})
}

// StartTask starts the specified task function, optionally with restart policy and backoff.
func (c *Component) StartTask(ctx context.Context, id string, fn TaskFunc, restart TaskRestart, jitter float64, backoff ...time.Duration) {
	logger := log.FromContext(ctx).WithField("task_id", id)
	if len(backoff) == 0 {
		backoff = defaultTaskBackoff[:]
	}
	go func() {
		invocation := 0
		for {
			invocation++
			err := fn(ctx)
			if err != nil {
				logger.WithField("invocation", invocation).WithError(err).Warn("Task failed")
			}
			switch restart {
			case TaskRestartNever:
				return
			case TaskRestartAlways:
			case TaskRestartOnFailure:
				if err == nil {
					return
				}
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			bi := invocation - 1
			if bi >= len(backoff) {
				bi = len(backoff) - 1
			}
			s := backoff[bi]
			if jitter != 0 {
				s = random.Jitter(backoff[bi], jitter)
			}
			time.Sleep(s)
		}
	}()
}

func (c *Component) startTasks() {
	for _, t := range c.tasks {
		c.StartTask(t.ctx, t.id, t.fn, t.restart, 0.1, t.backoff...)
	}
}
