// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"os"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/spf13/cobra"
)

var (
	name      = "example"
	config    = conf.InitializeWithDefaults(name, shared.DefaultBaseConfig)
	logger, _ = log.NewLogger(log.WithLevel(shared.DefaultBaseConfig.Log.Level), log.WithHandler(log.NewCLI(os.Stdout)))
	Root      = &cobra.Command{
		Use:   name,
		Short: "Example program",
	}
)

func init() {
	cobra.OnInitialize(func() {
		// read in config from file
		err := config.ReadInConfig()
		if err != nil {
			logger.WithError(err).Warn("Could not read config file")
		}

		// unmarshal config
		cfg := new(component.Config)
		if err := config.Unmarshal(cfg); err != nil {
			logger.WithError(err).Fatal("Could not parse config")
		}

		// set log level to correct level
		log.WithLevel(cfg.Log.Level)(logger.(*log.Logger))
	})

	Root.PersistentFlags().AddFlagSet(config.Flags())
}
