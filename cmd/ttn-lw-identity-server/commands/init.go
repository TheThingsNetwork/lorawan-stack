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

package commands

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/assets"
	"go.thethings.network/lorawan-stack/pkg/component"
	conf "go.thethings.network/lorawan-stack/pkg/config"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
)

// ErrIdentityServerLoadData is returned when the data couldn't be loaded into the identity server.
var ErrIdentityServerLoadData = errors.Define("identity_server_load_data", "could not load identity server data")

var (
	initConfigName = "ttn-lw-identity-server"
	initMgr        = conf.InitializeWithDefaults(initConfigName, DefaultInitConfig)
	initConfig     = new(InitConfig)

	initCommand = &cobra.Command{
		Use:   "init",
		Short: "Initializes the Identity Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initMgr.ReadInConfig()
			if err != nil {
				return err
			}

			if err = initMgr.Unmarshal(initConfig); err != nil {
				return err
			}

			c, err := component.New(logger, &component.Config{ServiceBase: initConfig.ServiceBase})
			if err != nil {
				return shared.ErrBaseComponentInitialize.WithCause(err)
			}

			assets, err := assets.New(c, initConfig.Assets)
			if err != nil {
				return shared.ErrIdentityServerInitialize.WithCause(err)
			}
			initConfig.IS.OAuth.Assets = assets

			is, err := identityserver.New(c, initConfig.IS)
			if err != nil {
				return shared.ErrIdentityServerInitialize.WithCause(err)
			}

			logger.Info("Initializing Identity Server...")

			err = is.Init(initConfig.InitialData)
			if err != nil {
				return ErrIdentityServerLoadData.WithCause(err)
			}

			logger.Info("Identity Server initialized")

			return nil
		},
	}
)

func init() {
	initCommand.PersistentFlags().AddFlagSet(initMgr.Flags())
	Root.AddCommand(initCommand)
}
