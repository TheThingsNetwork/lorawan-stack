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
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/applicationregistry"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/redis"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Application Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return shared.ErrBaseComponentInitialize.WithCause(err)
			}

			appsRedis := redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "applications"},
				IndexKeys: applicationregistry.Identifiers,
			})
			appsReg := applicationregistry.New(store.NewByteMapStoreClient(appsRedis))
			config.AS.ApplicationRegistry = appsReg

			devsRedis := redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "devices"},
				IndexKeys: deviceregistry.Identifiers,
			})
			devsReg := deviceregistry.New(store.NewByteMapStoreClient(devsRedis))
			config.AS.DeviceRegistry = devsReg

			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return shared.ErrApplicationServerInitialize.WithCause(err)
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
