// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

package task_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/log/handler/memory"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var errTest = errors.DefineAborted("test", "test error")

func TestDefaultStartTaskLogging(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string
		// Context returns the context the task is started with.
		Context func(context.Context) context.Context
		// Func is the task function.
		Func task.Func
		// Message is the expected log message, empty if nothing is expected to be logged.
		Message string
		Level   log.Level
	}{
		{
			Name:    "Failure",
			Context: func(ctx context.Context) context.Context { return ctx },
			Func:    func(context.Context) error { return errTest.New() },
			Message: "Task failed",
			Level:   log.WarnLevel,
		},
		{
			Name:    "Success",
			Context: func(ctx context.Context) context.Context { return ctx },
			Func:    func(context.Context) error { return nil },
		},
		{
			Name:    "EOF",
			Context: func(ctx context.Context) context.Context { return ctx },
			Func:    func(context.Context) error { return io.EOF },
		},
		{
			Name: "ContextCanceled",
			Context: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx
			},
			Func: func(ctx context.Context) error { return ctx.Err() },
		},
		{
			// Contexts from pkg/errorcontext return the cancelation cause instead of
			// context.Canceled, and the task runner cannot tell that apart from a genuine
			// failure. Tasks that stop because their context is done must return nil instead of
			// the context error, or every cancelation is reported as a failure.
			Name: "ErrorContextCanceled",
			Context: func(ctx context.Context) context.Context {
				ctx, cancel := errorcontext.New(ctx)
				cancel(errTest.New())
				return ctx
			},
			Func:    func(ctx context.Context) error { return ctx.Err() },
			Message: "Task failed",
			Level:   log.WarnLevel,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)

			handler := memory.New()
			ctx := log.NewContext(test.Context(), log.NewLogger(handler, log.WithLevel(log.DebugLevel)))

			done := make(chan struct{})
			task.DefaultStartTask(&task.Config{
				Context: tc.Context(ctx),
				ID:      "test",
				Func:    tc.Func,
				Done:    func() { close(done) },
				Restart: task.RestartNever,
			})

			select {
			case <-done:
			case <-time.After(test.Delay << 10):
				t.Fatal("Timed out waiting for the task to finish")
			}

			var entries []log.Entry
			for _, entry := range handler.Entries {
				if entry.Message() == tc.Message {
					entries = append(entries, entry)
				}
			}
			if tc.Message == "" {
				a.So(handler.Entries, should.BeEmpty)
				return
			}
			if !a.So(entries, should.HaveLength, 1) {
				t.FailNow()
			}
			a.So(entries[0].Level(), should.Equal, tc.Level)
		})
	}
}

// TestDefaultStartTaskRestartOnFailure ensures that a task whose context is done is not restarted,
// even though it stopped with an error.
func TestDefaultStartTaskRestartOnFailure(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ctx, cancel := errorcontext.New(test.Context())
	cancel(errTest.New())

	invocations := 0
	done := make(chan struct{})
	task.DefaultStartTask(&task.Config{
		Context: ctx,
		ID:      "test",
		Func: func(ctx context.Context) error {
			invocations++
			return ctx.Err()
		},
		Done:    func() { close(done) },
		Restart: task.RestartOnFailure,
	})

	select {
	case <-done:
	case <-time.After(test.Delay << 10):
		t.Fatal("Timed out waiting for the task to finish")
	}
	a.So(invocations, should.Equal, 1)
}
