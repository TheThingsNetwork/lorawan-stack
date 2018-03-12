// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/pkg/applicationserver"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Application Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return errors.NewWithCause(err, "Could not initialize base component")
			}

			redis := redis.New(&redis.Config{Redis: config.Redis})
			reg := deviceregistry.New(store.NewByteStoreClient(redis))
			config.AS.Registry = reg

			as := applicationserver.New(c, &config.AS)
			_ = as

			logger.Info("Starting application server...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
