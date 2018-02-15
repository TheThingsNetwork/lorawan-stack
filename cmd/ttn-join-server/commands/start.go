// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/joinserver"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Join Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})

			redis := redis.New(&redis.Config{Redis: config.Redis})
			reg := deviceregistry.New(store.NewByteStoreClient(redis))
			config.JS.Registry = reg

			js := joinserver.New(c, &config.JS)
			_ = js

			logger.Info("Starting join server...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
