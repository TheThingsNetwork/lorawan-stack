// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/spf13/cobra"
)

var (
	referenceCommand = &cobra.Command{
		Use:   "reference",
		Short: "Start the reference component",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := new(component.Config)
			err := config.Unmarshal(cfg)
			if err != nil {
				logger.WithError(err).Fatal("Could not parse config")
			}

			c := component.New(logger, cfg)
			c.Start()
		},
	}
)

func init() {
	Root.AddCommand(referenceCommand)
	referenceCommand.PersistentFlags().AddFlagSet(config.WithConfig(&component.Config{
		ServiceBase: shared.DefaultServiceBase,
	}))
}
