// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/pkg/applicationserver"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
	"github.com/TheThingsNetwork/ttn/pkg/joinserver"
	"github.com/TheThingsNetwork/ttn/pkg/networkserver"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Network Stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})

			redis := redis.New(&redis.Config{Redis: config.Redis})
			reg := deviceregistry.New(store.NewByteStoreClient(redis))
			config.NS.Registry = reg
			config.AS.Registry = reg
			config.JS.Registry = reg

			is, err := identityserver.New(c, config.IS)
			if err != nil {
				return errors.NewWithCause("Could not create identity server", err)
			}

			err = is.Init()
			if err != nil {
				return errors.NewWithCause("Could not initialize identity server", err)
			}

			gs := gatewayserver.New(c, &config.GS)
			_ = gs

			ns := networkserver.New(c, &config.NS)
			_ = ns

			as := applicationserver.New(c, &config.AS)
			_ = as

			js := joinserver.New(c, &config.JS)
			_ = js

			// TODO: Web UI

			return c.Start()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
