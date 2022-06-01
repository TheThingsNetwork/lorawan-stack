// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

var _ bun.QueryHook = (*LoggerHook)(nil)

// LoggerHook is a bun.QueryHook that logs the queries.
type LoggerHook struct {
	logger log.Interface
}

// LoggerHookOption is an option for the LoggerHook.
type LoggerHookOption func(*LoggerHook)

// NewLoggerHook returns a new LoggerHook that logs queries to the logger.
func NewLoggerHook(logger log.Interface, opts ...LoggerHookOption) *LoggerHook {
	h := &LoggerHook{
		logger: logger,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// BeforeQuery is the hook that is executed before the query runs.
func (*LoggerHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

// AfterQuery is the hook that is executed after the query runs.
func (h *LoggerHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	operation := event.Operation()
	switch operation {
	case "BEGIN", "COMMIT", "ROLLBACK":
		return
	}
	logFields := log.Fields(
		"operation", event.Operation(),
		"duration", time.Since(event.StartTime).Round(time.Microsecond),
		"query", event.Query,
	)
	if event.Result != nil {
		if rows, err := event.Result.RowsAffected(); err == nil {
			logFields = logFields.WithField("rows", rows)
		}
	}
	if event.Err != nil {
		logFields = logFields.WithError(wrapDriverError(event.Err))
	}

	h.logger.WithFields(logFields).Debug("Run database query")
}
