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

package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
)

// Func is the task function.
type Func func(context.Context) error

// Execute executes the task function.
func (f Func) Execute(ctx context.Context, logger log.Interface) (err error) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintln(os.Stderr, p)
			os.Stderr.Write(debug.Stack())
			if pErr, ok := p.(error); ok {
				err = errTaskRecovered.WithCause(pErr)
			} else {
				err = errTaskRecovered.WithAttributes("panic", p)
			}
			logger.WithError(err).Error("Task panicked")
		}
	}()
	return f(ctx)
}

// Restart defines a task's restart policy.
type Restart uint8

const (
	// RestartNever denotes a restart policy that never restarts tasks after success or failure.
	RestartNever Restart = iota
	// RestartAlways denotes a restart policy that always restarts tasks, on success and failure.
	RestartAlways
	// RestartOnFailure denotes a restart policy that restarts tasks on failure.
	RestartOnFailure
)

// BackoffIntervalFunc is a function that decides the backoff interval based on the attempt history.
// invocation is a counter, which starts at 1 and is incremented after each task function invocation.
type BackoffIntervalFunc func(ctx context.Context, executionDuration time.Duration, invocation uint, err error) time.Duration

// BackoffConfig represents task backoff configuration.
type BackoffConfig struct {
	Jitter       float64
	IntervalFunc BackoffIntervalFunc
}

// MakeBackoffIntervalFunc returns a new BackoffIntervalFunc.
func MakeBackoffIntervalFunc(onFailure bool, resetDuration time.Duration, intervals ...time.Duration) BackoffIntervalFunc {
	return func(ctx context.Context, executionDuration time.Duration, invocation uint, err error) time.Duration {
		switch {
		case onFailure && err == nil:
			return 0
		case executionDuration > resetDuration:
			return intervals[0]
		case invocation >= uint(len(intervals)):
			return intervals[len(intervals)-1]
		default:
			return intervals[invocation-1]
		}
	}
}

// Values for DefaultBackoffConfig.
const (
	DefaultBackoffResetDuration = time.Minute
	DefaultBackoffJitter        = 0.1
)

var (
	// DefaultBackoffIntervals are the default task backoff intervals.
	DefaultBackoffIntervals = [...]time.Duration{
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		time.Second,
	}
	// DefaultBackoffIntervalFunc is the default BackoffIntervalFunc.
	DefaultBackoffIntervalFunc = MakeBackoffIntervalFunc(false, DefaultBackoffResetDuration, DefaultBackoffIntervals[:]...)
	// DefaultBackoffConfig is the default task backoff config.
	DefaultBackoffConfig = &BackoffConfig{
		Jitter:       DefaultBackoffJitter,
		IntervalFunc: DefaultBackoffIntervalFunc,
	}

	// DialBackoffIntervals are the default task backoff intervals for tasks using Dial.
	DialBackoffIntervals = [...]time.Duration{
		100 * time.Millisecond,
		time.Second,
		10 * time.Second,
	}
	// DialBackoffIntervalFunc is the default BackoffIntervalFunc for tasks using Dial.
	DialBackoffIntervalFunc = MakeBackoffIntervalFunc(false, DefaultBackoffResetDuration, DialBackoffIntervals[:]...)
	// DialBackoffConfig is the default task backoff config for tasks using Dial.
	DialBackoffConfig = &BackoffConfig{
		Jitter:       DefaultBackoffJitter,
		IntervalFunc: DialBackoffIntervalFunc,
	}
)

// Config represents task configuration.
type Config struct {
	Context context.Context
	ID      string
	Func    Func
	Done    func()
	Restart Restart
	Backoff *BackoffConfig
}

// Starter starts tasks with a TaskConfig.
type Starter interface {
	// StartTask starts the specified task function, optionally with restart policy and backoff.
	StartTask(*Config)
}

// StartTaskFunc is a function that implements the TaskStarter interface.
type StartTaskFunc func(*Config)

// StartTask implements the TaskStarter interface.
func (f StartTaskFunc) StartTask(conf *Config) {
	f(conf)
}

var errTaskRecovered = errors.DefineInternal("task_recovered", "task recovered")

// DefaultStartTask is the default TaskStarter.
func DefaultStartTask(conf *Config) {
	logger := log.FromContext(conf.Context).WithField("task_id", conf.ID)
	go func() {
		defer func() {
			if done := conf.Done; done != nil {
				done()
			}
		}()
		for invocation := uint(1); ; invocation++ {
			if invocation == 0 {
				logger.Warn("Invocation count rollover detected")
				invocation = 1
			}
			logger := logger.WithField("invocation", invocation)
			startTime := time.Now()
			err := conf.Func.Execute(conf.Context, logger)
			executionDuration := time.Since(startTime)
			// NOTE: We discard only the common errors here, instead of checking
			// the error code intentionally. The intent is to drop the commonly
			// met errors without hiding other errors with the canceled or deadline
			// exceeded error code.
			switch {
			case err == nil:
			case errors.Is(err, context.Canceled),
				errors.Is(err, errors.ErrContextCanceled),
				errors.Is(err, context.DeadlineExceeded),
				errors.Is(err, errors.ErrContextDeadlineExceeded):
			case errors.Is(err, io.EOF):
			default:
				logger.WithError(err).Warn("Task failed")
			}
			switch conf.Restart {
			case RestartNever:
				return
			case RestartAlways:
			case RestartOnFailure:
				if err == nil {
					return
				}
			default:
				panic("Invalid Config.Restart value")
			}
			select {
			case <-conf.Context.Done():
				return
			default:
			}
			if conf.Backoff == nil {
				continue
			}
			s := conf.Backoff.IntervalFunc(conf.Context, executionDuration, invocation, err)
			if s == 0 {
				continue
			}
			if conf.Backoff.Jitter != 0 {
				s = random.Jitter(s, conf.Backoff.Jitter)
			}
			select {
			case <-conf.Context.Done():
				return
			case <-time.After(s):
			}
		}
	}()
}
