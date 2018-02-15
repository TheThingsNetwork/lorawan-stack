// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Identity Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})

			is, err := identityserver.New(c, config.IS)
			if err != nil {
				return errors.NewWithCause("Could not create identity server", err)
			}
			logger.Debug("Initializing identity server...")
			err = is.Init()
			if err != nil {
				return errors.NewWithCause("Could not initialize identity server", err)
			}

			logger.Info("Starting identity server...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
