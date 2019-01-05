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

package commands

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	asiowebredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	asredis "go.thethings.network/lorawan-stack/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/redis"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Application Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return shared.ErrInitializeBaseComponent.WithCause(err)
			}

			config.AS.Links = &asredis.LinkRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "links"},
			})}
			config.AS.Devices = &asredis.DeviceRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "devices"},
			})}
			if config.AS.Webhooks.Target != "" {
				config.AS.Webhooks.Registry = &asiowebredis.WebhookRegistry{Redis: redis.New(&redis.Config{
					Redis:     config.Redis,
					Namespace: []string{"as", "io", "webhooks"},
				})}
			}

			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			_ = as

			logger.Info("Starting Application Server...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
