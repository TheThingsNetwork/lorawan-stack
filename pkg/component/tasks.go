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

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
)

// TaskFunc is the task function.
type TaskFunc func(context.Context) error

// TaskRestart defines a task's restart policy.
type TaskRestart uint8

const (
	// TaskRestartNever denotes a restart policy that never restarts tasks after success or failure.
	TaskRestartNever TaskRestart = iota
	// TaskRestartAlways denotes a restart policy that always restarts tasks, on success and failure.
	TaskRestartAlways
	// TaskRestartOnFailure denotes a restart policy that restarts tasks on failure.
	TaskRestartOnFailure
)

var backoffResetTime = time.Minute

type TaskBackoffConfig struct {
	Jitter    float64
	Intervals []time.Duration
}

const DefaultBackoffJitter = 0.1

// DefaultTaskBackoffConfig is a default task backoff config to use.
var DefaultTaskBackoffConfig = &TaskBackoffConfig{
	Jitter: DefaultBackoffJitter,
	Intervals: []time.Duration{
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		time.Second,
	},
}

// DialTaskBackoffConfig is a default task backoff config to use.
var DialTaskBackoffConfig = &TaskBackoffConfig{
	Jitter: DefaultBackoffJitter,
	Intervals: []time.Duration{
		100 * time.Millisecond,
		time.Second,
		10 * time.Second,
	},
}

type TaskConfig struct {
	Context context.Context
	ID      string
	Func    TaskFunc
	Restart TaskRestart
	Backoff *TaskBackoffConfig
}

// RegisterTask registers a task, optionally with restart policy and backoff, to be started after the component started.
func (c *Component) RegisterTask(conf *TaskConfig) {
	c.taskConfigs = append(c.taskConfigs, conf)
}

type TaskStarter interface {
	// StartTask starts the specified task function, optionally with restart policy and backoff.
	StartTask(*TaskConfig)
}

type StartTaskFunc func(*TaskConfig)

func (f StartTaskFunc) StartTask(conf *TaskConfig) {
	f(conf)
}

func DefaultStartTask(conf *TaskConfig) {
	logger := log.FromContext(conf.Context).WithField("task_id", conf.ID)
	go func() {
		invocation := 0
		for {
			invocation++
			startTime := time.Now()
			err := conf.Func(conf.Context)
			executionTime := time.Since(startTime)
			if err != nil && err != context.Canceled {
				logger.WithField("invocation", invocation).WithError(err).Warn("Task failed")
			}
			switch conf.Restart {
			case TaskRestartNever:
				return
			case TaskRestartAlways:
			case TaskRestartOnFailure:
				if err == nil {
					return
				}
			default:
				panic("Invalid TaskConfig.Restart value")
			}
			select {
			case <-conf.Context.Done():
				return
			default:
			}
			if conf.Backoff == nil {
				continue
			}
			bi := invocation - 1
			if bi >= len(conf.Backoff.Intervals) {
				bi = len(conf.Backoff.Intervals) - 1
			}
			if executionTime > backoffResetTime {
				bi = 0
			}
			s := conf.Backoff.Intervals[bi]
			if conf.Backoff.Jitter != 0 {
				s = random.Jitter(conf.Backoff.Intervals[bi], conf.Backoff.Jitter)
			}
			time.Sleep(s)
		}
	}()
}

func (c *Component) StartTask(conf *TaskConfig) {
	c.taskStarter.StartTask(conf)
}

func (c *Component) startTasks() {
	for _, conf := range c.taskConfigs {
		c.taskStarter.StartTask(conf)
	}
}
