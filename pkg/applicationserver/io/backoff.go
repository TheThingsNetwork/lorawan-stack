// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package io

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// TaskHealthyDuration is the duration after which a task is considered to be in steady state.
const TaskHealthyDuration = 1 * time.Minute

// TaskBackoffConfig derives the component.DefaultTaskBackoffConfig and dynamically determines the backoff duration
// based on recent error codes.
var TaskBackoffConfig = &component.TaskBackoffConfig{
	Jitter: component.DefaultTaskBackoffConfig.Jitter,
	DynamicInterval: func(ctx context.Context, executionTime time.Duration, invocation int, err error) time.Duration {
		defaultIntervals := component.DefaultTaskBackoffConfig.Intervals
		extendedIntervals := append(defaultIntervals,
			1*time.Minute,
			5*time.Minute,
			15*time.Minute,
			30*time.Minute,
		)

		var intervals []time.Duration
		switch {
		case errors.IsFailedPrecondition(err),
			errors.IsUnauthenticated(err),
			errors.IsPermissionDenied(err),
			errors.IsInvalidArgument(err),
			errors.IsAlreadyExists(err),
			errors.IsCanceled(err):
			intervals = extendedIntervals
		default:
			intervals = defaultIntervals
		}

		bi := invocation - 1
		if bi >= len(intervals) {
			bi = len(intervals) - 1
		}
		if executionTime > TaskHealthyDuration {
			bi = 0
		}

		return intervals[bi]
	},
}
