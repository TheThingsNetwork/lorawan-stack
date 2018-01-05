// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Identity Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			is, err := identityserver.New(logger, config)
			if err != nil {
				return err
			}

			return is.Start()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
	startCommand.Flags().AddFlagSet(mgr.WithConfig(&identityserver.Config{
		ServiceBase: shared.DefaultServiceBase,
	}))
}
