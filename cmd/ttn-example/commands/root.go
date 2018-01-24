// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"os"

	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/spf13/cobra"
)

var (
	logger *log.Logger
	name   = "ttn-example"
	mgr    = conf.InitializeWithDefaults(name, shared.DefaultBaseConfig)
	config = new(component.Config)

	// Root command is the entrypoint of the program
	Root = &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Example program",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// read in config from file
			err := mgr.ReadInConfig()
			if err != nil {
				return err
			}

			// unmarshal config
			if err = mgr.Unmarshal(config); err != nil {
				return err
			}

			// create logger
			logger, err = log.NewLogger(
				log.WithLevel(config.Log.Level),
				log.WithHandler(log.NewCLI(os.Stdout)),
			)
			return err
		},
	}
)

func init() {
	Root.PersistentFlags().AddFlagSet(mgr.Flags())
}
