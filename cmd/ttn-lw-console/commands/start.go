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

package commands

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/console"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Console",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{
				ServiceBase: config.ServiceBase,
			})
			if err != nil {
				return shared.ErrInitializeBaseComponent.WithCause(err)
			}

			_, err = console.New(c, config.Console)
			if err != nil {
				return shared.ErrInitializeConsole.WithCause(err)
			}

			logger.Info("Starting Console...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
