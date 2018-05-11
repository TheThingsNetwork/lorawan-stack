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
	"go.thethings.network/lorawan-stack/pkg/component"
	conf "go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
)

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
				return errors.NewWithCause(err, "Could not initialize base component")
			}

			is, err := identityserver.New(c, initConfig.IS)
			if err != nil {
				return errors.NewWithCause(err, "Could not create identity server")
			}

			logger.Info("Initializing Identity Server...")

			err = is.Init(initConfig.InitialData)
			if err != nil {
				return errors.NewWithCause(err, "Could not initialize identity server")
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
