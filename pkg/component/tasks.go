// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
)

// TaskFunc is the task function.
type TaskFunc func(context.Context) error

// TaskRestart defines a task's restart policy.
type TaskRestart int

const (
	TaskRestartNever TaskRestart = iota
	TaskRestartAlways
	TaskRestartOnFailure
)

var defaultTaskBackoff = [...]time.Duration{
	10 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	1 * time.Second,
}

type task struct {
	fn      TaskFunc
	restart TaskRestart
	backoff []time.Duration
}

// RegisterTask registers a task, optionally with automatic restart, to be started after the component started.
func (c *Component) RegisterTask(fn TaskFunc, restart TaskRestart, backoff ...time.Duration) {
	if len(backoff) == 0 {
		backoff = defaultTaskBackoff[:]
	}
	c.tasks = append(c.tasks, task{
		fn:      fn,
		restart: restart,
		backoff: backoff,
	})
}

func (c *Component) startTasks() {
	for _, t := range c.tasks {
		go func(t task) {
			invocation := 0
			for {
				invocation++
				err := t.fn(c.ctx)
				if err != nil {
					c.logger.WithField("invocation", invocation).WithError(err).Warn("Task failed")
				}
				switch t.restart {
				case TaskRestartNever:
					return
				case TaskRestartAlways:
				case TaskRestartOnFailure:
					if err == nil {
						return
					}
				}
				select {
				case <-c.ctx.Done():
					return
				default:
				}
				bi := invocation - 1
				if bi >= len(t.backoff) {
					bi = len(t.backoff) - 1
				}
				time.Sleep(t.backoff[bi])
			}
		}(t)
	}
}
