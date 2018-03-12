// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Gateway Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return errors.NewWithCause(err, "Could not initialize")
			}

			gs, err := gatewayserver.New(c, &config.GS)
			if err != nil {
				return err
			}
			_ = gs

			logger.Info("Starting gateway server...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
