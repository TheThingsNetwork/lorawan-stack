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

package commands

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	asiowebredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	asredis "go.thethings.network/lorawan-stack/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/console"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/joinserver"
	jsredis "go.thethings.network/lorawan-stack/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/pkg/networkserver"
	nsredis "go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/redis"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Network Stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return shared.ErrInitializeBaseComponent.WithCause(err)
			}

			host, err := os.Hostname()
			if err != nil {
				return err
			}

			config.NS.Devices = &nsredis.DeviceRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"ns", "devices"},
			})}
			nsDownlinkTasks := nsredis.NewDownlinkTaskQueue(redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"ns", "tasks"},
			}), 100000, "ns", redis.Key(host, strconv.Itoa(os.Getpid())))
			config.NS.DownlinkTasks = nsDownlinkTasks

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

			config.JS.Devices = &jsredis.DeviceRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"js", "devices"},
			})}

			config.JS.Keys = &jsredis.KeyRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"js", "keys"},
			})}

			is, err := identityserver.New(c, &config.IS)
			if err != nil {
				return shared.ErrInitializeIdentityServer.WithCause(err)
			}
			_ = is

			is.SetRedisCache(redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"is", "cache"},
			}))

			gs, err := gatewayserver.New(c, &config.GS)
			if err != nil {
				return shared.ErrInitializeGatewayServer.WithCause(err)
			}
			_ = gs

			ns, err := networkserver.New(c, &config.NS)
			if err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			ns.Component.RegisterTask(nsDownlinkTasks.Run, component.TaskRestartOnFailure)

			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			_ = as

			js, err := joinserver.New(c, &config.JS)
			if err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			_ = js

			console, err := console.New(c, config.Console)
			if err != nil {
				return shared.ErrInitializeConsole.WithCause(err)
			}
			_ = console

			logger.Info("Starting stack...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
