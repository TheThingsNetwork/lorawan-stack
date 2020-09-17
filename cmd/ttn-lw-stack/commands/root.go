// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package commands implements the commands for the ttn-lw-stack binary.
package commands

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/commands"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared/version"
	conf "go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	logobservability "go.thethings.network/lorawan-stack/v3/pkg/log/middleware/observability"
	logsentry "go.thethings.network/lorawan-stack/v3/pkg/log/middleware/sentry"
	pkgversion "go.thethings.network/lorawan-stack/v3/pkg/version"
)

var errMissingFlag = errors.DefineInvalidArgument("missing_flag", "missing CLI flag `{flag}`")

var (
	ctx    = context.Background()
	logger *log.Logger
	name   = "ttn-lw-stack"
	mgr    = conf.InitializeWithDefaults(name, "ttn_lw", DefaultConfig,
		conf.WithDeprecatedFlag("interop.sender-client-cas", "use interop.sender-client-ca sub-fields instead"),
	)
	config = new(Config)

	// Root command is the entrypoint of the program
	Root = &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "The Things Stack for LoRaWAN",
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

			// initialize configuration fallbacks
			if err := shared.InitializeFallbacks(&config.ServiceBase); err != nil {
				return err
			}

			// create logger
			logger = log.NewLogger(
				log.WithLevel(config.Base.Log.Level),
				log.WithHandler(log.NewCLI(os.Stdout)),
			)

			logger.Use(logobservability.New())

			if config.Sentry.DSN != "" {
				opts := sentry.ClientOptions{
					Dsn:     config.Sentry.DSN,
					Release: pkgversion.String(),
				}
				if hostname, err := os.Hostname(); err == nil {
					opts.ServerName = hostname
				}
				err = sentry.Init(opts)
				if err != nil {
					return err
				}
				logger.Use(logsentry.New())
			}

			ctx = log.NewContext(ctx, logger)

			return nil
		},
	}
)

func init() {
	Root.PersistentFlags().AddFlagSet(mgr.Flags())
	Root.AddCommand(version.Print(Root))
	Root.AddCommand(commands.GenManPages(Root))
	Root.AddCommand(commands.GenMDDoc(Root))
	Root.AddCommand(commands.GenYAMLDoc(Root))
	Root.AddCommand(commands.Complete())
}
