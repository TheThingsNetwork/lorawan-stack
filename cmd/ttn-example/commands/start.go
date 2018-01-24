// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/spf13/cobra"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the reference component",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := component.New(logger, config)
			return c.Start()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
	startCommand.Flags().AddFlagSet(mgr.WithConfig(&component.Config{
		ServiceBase: shared.DefaultServiceBase,
	}))
}
