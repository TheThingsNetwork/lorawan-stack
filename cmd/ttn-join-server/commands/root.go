// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"os"

	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/spf13/cobra"
)

var (
	logger *log.Logger
	name   = "ttn-join-server"
	mgr    = conf.InitializeWithDefaults(name, DefaultConfig)
	config = new(Config)

	// Root command is the entrypoint of the program
	Root = &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "The Things Network Join Server",
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
			if sentry, err := shared.SentryMiddleware(config.ServiceBase); err == nil && sentry != nil {
				logger.Use(sentry)
			}
			return err
		},
	}
)

func init() {
	Root.PersistentFlags().AddFlagSet(mgr.Flags())
}
