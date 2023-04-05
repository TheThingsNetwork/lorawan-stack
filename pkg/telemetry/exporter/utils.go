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

package telemetry

import (
	"context"
	"runtime"

	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/models"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

// DispatchTask returns a task that calls Dispatch with the provided consumerID.
func DispatchTask(q TaskQueue, consumerID string) task.Func {
	return func(ctx context.Context) error {
		return q.Dispatch(ctx, consumerID)
	}
}

// PopTask returns a task that calls Pop with the provided consumerID.
func PopTask(q TaskQueue, consumerID string) task.Func {
	return func(ctx context.Context) error {
		return q.Pop(ctx, consumerID)
	}
}

// OSTelemetryData returns the OS telemetry data which is attached to telemetry messages.
func OSTelemetryData() *models.OSTelemetry {
	return &models.OSTelemetry{
		OperatingSystem: runtime.GOOS,
		Arch:            runtime.GOARCH,
		BinaryVersion:   version.String(),
		GolangVersion:   runtime.Version(),
	}
}
