// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package commands implements the commands for the ttn-lw-application-server binary.
package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	conf "go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"golang.org/x/oauth2"
)

var (
	logger       *log.Logger
	name         = "ttn-lw-cli"
	mgr          = conf.InitializeWithDefaults(name, "ttn_lw", DefaultConfig)
	config       = &Config{}
	oauth2Config *oauth2.Config
	ctx          = context.Background()
	cache        util.Cache

	// Root command is the entrypoint of the program
	Root = &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "The Things Network Command-line Interface",
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

			// get cache
			cache, err = util.GetCache()
			if err != nil {
				return err
			}

			// create logger
			logger, err = log.NewLogger(
				log.WithLevel(config.Log.Level),
				log.WithHandler(log.NewCLI(os.Stderr)),
			)
			if err != nil {
				return err
			}
			ctx = log.NewContext(ctx, logger)

			// prepare the API
			api.SetLogger(logger)
			api.SetInsecure(config.Insecure)
			if config.CA != "" {
				if err = api.SetCA(config.CA); err != nil {
					return err
				}
			}

			// OAuth
			oauth2Config = &oauth2.Config{
				ClientID: "cli",
				Endpoint: oauth2.Endpoint{
					AuthURL:  fmt.Sprintf("%s/oauth/authorize", config.OAuthServerAddress),
					TokenURL: fmt.Sprintf("%s/oauth/token", config.OAuthServerAddress),
				},
			}

			// Access
			if apiKey, ok := cache.Get("api_key").(string); ok {
				logger.Debug("Using API key")
				api.SetAuth("bearer", apiKey)
			} else if token, ok := cache.Get("oauth_token").(*oauth2.Token); ok && token != nil {
				freshToken, err := oauth2Config.TokenSource(ctx, token).Token()
				if freshToken != token {
					cache.Set("oauth_token", freshToken)
					err := util.SaveCache(cache)
					if err != nil {
						return err
					}
				}
				if err != nil {
					logger.WithError(err).Warn("No valid access token present")
				} else {
					logger.Debugf("Using access token (valid until %s)", freshToken.Expiry.Truncate(time.Minute).Format(time.Kitchen))
					api.SetAuth(freshToken.TokenType, freshToken.AccessToken)
				}
			} else {
				logger.Warn("No access token present")
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// clean up the API
			api.CloseAll()

			err := util.SaveCache(cache)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	Root.PersistentFlags().AddFlagSet(mgr.Flags())
}
