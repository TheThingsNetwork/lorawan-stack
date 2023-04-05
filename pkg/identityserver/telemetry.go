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

package identityserver

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/istelemetry"
)

// initilizeTelemetryTasks starts the telemetry dispatcher, consumers and Identity Server's tasks.
func (is *IdentityServer) initilizeTelemetryTasks(ctx context.Context) error {
	logger := log.FromContext(ctx)

	tmCfg := is.GetBaseConfig(is.Context()).Telemetry
	tq := is.telemetryQueue
	if tq == nil || !tmCfg.Enable {
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	consumerIDPrefix := fmt.Sprintf("%s:%d", hostname, os.Getpid())

	if entityCntTm := tmCfg.EntityCountTelemetry; entityCntTm.Enable {
		cl, err := is.HTTPClient(ctx)
		if err != nil {
			return err
		}

		entityTask := istelemetry.New(
			istelemetry.WithUID(telemetry.GenerateHash(is.Context(), tmCfg.UIDElements...)),
			istelemetry.WithBunDB(bun.NewDB(is.db, pgdialect.New())),
			istelemetry.WithTarget(tmCfg.Target),
			istelemetry.WithHTTPClient(cl),
		)
		if err := entityTask.Validate(ctx); err != nil {
			return err
		}

		logger.WithField("task", istelemetry.EntityCountTaskName).Debug("Add task to queue")
		err = tq.Add(ctx, istelemetry.EntityCountTaskName, time.Now().Add(entityCntTm.Interval), false)
		if err != nil {
			return err
		}
		logger.WithField("task", istelemetry.EntityCountTaskName).Debug("Register task callback")
		tq.RegisterCallback(
			istelemetry.EntityCountTaskName,
			telemetry.CallbackWithInterval(ctx, entityCntTm.Interval, entityTask.CountEntities),
		)
	}

	logger.Debug("Start telemetry queue dispatcher")
	is.RegisterTask(&task.Config{
		Context: ctx,
		ID:      "is_telemetry_dispatcher",
		Backoff: task.DefaultBackoffConfig,
		Restart: task.RestartAlways,
		Func:    telemetry.DispatchTask(tq, consumerIDPrefix),
	})

	logger.Debug("Start telemetry queue consumers")
	for i := uint64(0); i < tmCfg.NumConsumers; i++ {
		consumerID := fmt.Sprintf("%s:%d", consumerIDPrefix, i)
		is.RegisterTask(&task.Config{
			Context: ctx,
			ID:      fmt.Sprintf("%s_%d", "is_telemetry_consumer", i),
			Backoff: task.DefaultBackoffConfig,
			Restart: task.RestartAlways,
			Func:    telemetry.PopTask(tq, consumerID),
		})
	}
	return nil
}
