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

package component

import (
	"context"
	"net/http"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

const (
	versionCheckPeriod  = 12 * time.Hour
	versionCheckTimeout = 10 * time.Second
)

type httpClientProvider interface {
	HTTPClient(context.Context) (*http.Client, error)
}

// versionCheckTask returns the task configuration for a periodic version check.
func versionCheckTask(ctx context.Context, clientProvider httpClientProvider) *TaskConfig {
	taskConfig := TaskConfig{
		Context: ctx,
		ID:      "version_check",
		Func: func(ctx context.Context) error {
			checkCtx, cancel := context.WithTimeout(ctx, versionCheckTimeout)
			defer cancel()

			client, err := clientProvider.HTTPClient(checkCtx)
			if err != nil {
				return err
			}
			update, err := version.CheckUpdate(checkCtx, version.WithClient(client))
			if err != nil {
				return err
			}
			version.LogUpdate(checkCtx, update)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(versionCheckPeriod):
				return nil
			}
		},
		Restart: TaskRestartAlways,
		Backoff: DialTaskBackoffConfig,
	}
	return &taskConfig
}
