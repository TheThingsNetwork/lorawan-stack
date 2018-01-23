// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/networkserver"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Network Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})

			redis := redis.New(&redis.Config{Redis: config.Redis})
			reg := deviceregistry.New(store.NewByteStoreClient(redis))
			config.NS.Registry = reg

			ns := networkserver.New(c, &config.NS)
			_ = ns

			return c.Start()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
