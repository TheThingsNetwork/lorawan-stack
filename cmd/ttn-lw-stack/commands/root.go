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
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/commands"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared/version"
	conf "go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	logobservability "go.thethings.network/lorawan-stack/v3/pkg/log/middleware/observability"
	logsentry "go.thethings.network/lorawan-stack/v3/pkg/log/middleware/sentry"
	pkgversion "go.thethings.network/lorawan-stack/v3/pkg/version"
	"go.uber.org/automaxprocs/maxprocs"
)

var errMissingFlag = errors.DefineInvalidArgument("missing_flag", "missing CLI flag `{flag}`")

var (
	ctx    = context.Background()
	logger log.Stack
	name   = "ttn-lw-stack"
	mgr    = conf.InitializeWithDefaults(name, "ttn_lw", DefaultConfig,
		conf.WithDeprecatedFlag(
			"interop.sender-client-cas",
			"TLS client authentication with LoRaWAN Backend Interfaces is deprecated",
		),
	)
	config = new(Config)

	versionUpdate       chan pkgversion.Update
	versionCheckTimeout = time.Second

	// Root command is the entrypoint of the program.
	Root = &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "The Things Stack for LoRaWAN",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "__complete" {
				return nil
			}

			// read in config from file
			err := mgr.ReadInConfig()
			if err != nil {
				return err
			}

			// unmarshal config
			if err = mgr.Unmarshal(config); err != nil {
				return err
			}

			// enable configured experimental features
			experimental.EnableFeatures(config.Experimental.Features...)

			// initialize configuration fallbacks
			if err := shared.InitializeFallbacks(&config.ServiceBase); err != nil {
				return err
			}

			// create logger
			logger, err = shared.InitializeLogger(&config.Log)
			if err != nil {
				return err
			}

			logger.Use(logobservability.New())

			if config.Sentry.DSN != "" {
				opts := sentry.ClientOptions{
					Dsn:         config.Sentry.DSN,
					Release:     pkgversion.String(),
					Environment: config.Sentry.Environment,
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

			if _, err := maxprocs.Set(); err != nil {
				logger.WithError(err).Debug("Failed to set GOMAXPROCS")
			}

			ctx = log.NewContext(ctx, logger)

			// check version in background
			versionUpdate = make(chan pkgversion.Update)
			if config.SkipVersionCheck {
				close(versionUpdate)
			} else {
				go func(ctx context.Context) {
					defer close(versionUpdate)
					update, err := pkgversion.CheckUpdate(ctx)
					if err != nil {
						log.FromContext(ctx).WithError(err).Warn("Failed to check version update")
					} else if update != nil {
						versionUpdate <- *update
					} else {
						log.FromContext(ctx).Debug("No new version available")
					}
				}(ctx)
			}

			telemetryConfigFallback(ctx, config)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "__complete" {
				return nil
			}

			select {
			case <-ctx.Done():
			case <-time.After(versionCheckTimeout):
				logger.Warn("Version check timed out")
			case versionUpdate, ok := <-versionUpdate:
				if ok {
					pkgversion.LogUpdate(ctx, &versionUpdate)
				}
			}
			return nil
		},
	}
)

var (
	versionCommand     = version.Print(Root)
	genManPagesCommand = commands.GenManPages(Root)
	genMDDocCommand    = commands.GenMDDoc(Root)
	genJSONTreeCommand = commands.GenJSONTree(Root)
	completeCommand    = commands.Complete()
)

func runNoop(cmd *cobra.Command, args []string) error { return nil }

func init() {
	Root.PersistentFlags().AddFlagSet(mgr.Flags())

	versionCommand.PersistentPreRunE = runNoop
	versionCommand.PersistentPostRunE = runNoop
	Root.AddCommand(versionCommand)

	genManPagesCommand.PersistentPreRunE = runNoop
	genManPagesCommand.PersistentPostRunE = runNoop
	Root.AddCommand(genManPagesCommand)

	genMDDocCommand.PersistentPreRunE = runNoop
	genMDDocCommand.PersistentPostRunE = runNoop
	Root.AddCommand(genMDDocCommand)

	genJSONTreeCommand.PersistentPreRunE = runNoop
	genJSONTreeCommand.PersistentPostRunE = runNoop
	Root.AddCommand(genJSONTreeCommand)

	completeCommand.PersistentPreRunE = runNoop
	completeCommand.PersistentPostRunE = runNoop
	Root.AddCommand(completeCommand)
}
