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
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/redis"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Network Stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return errors.NewWithCause(err, "Could not initialize base component")
			}

			nsRedis := redis.New(&redis.Config{Redis: config.Redis, Namespace: []string{"ns", "devices"}})
			config.NS.Registry = deviceregistry.New(store.NewByteMapStoreClient(nsRedis))

			asRedis := redis.New(&redis.Config{Redis: config.Redis, Namespace: []string{"as", "devices"}})
			config.AS.Registry = deviceregistry.New(store.NewByteMapStoreClient(asRedis))

			jsRedis := redis.New(&redis.Config{Redis: config.Redis, Namespace: []string{"js", "devices"}})
			config.JS.Registry = deviceregistry.New(store.NewByteMapStoreClient(jsRedis))

			_, err = identityserver.New(c, config.IS)
			if err != nil {
				return errors.NewWithCause(err, "Could not create Identity Server")
			}

			gs, err := gatewayserver.New(c, config.GS)
			if err != nil {
				return errors.NewWithCause(err, "Could not create Gateway Server")
			}
			_ = gs

			ns, err := networkserver.New(c, &config.NS)
			if err != nil {
				return errors.NewWithCause(err, "Could not create Network Server")
			}
			_ = ns

			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return errors.NewWithCause(err, "Could not create Application Server")
			}
			_ = as

			js, err := joinserver.New(c, &config.JS)
			if err != nil {
				return errors.NewWithCause(err, "Could not create Join Server")
			}
			_ = js

			// TODO: Web UI

			logger.Info("Starting stack...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
