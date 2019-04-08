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
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	applicationsSubscribeCommand = &cobra.Command{
		Use:     "subscribe [application-id]",
		Aliases: []string{"sub"},
		Short:   "Subscribe to application uplink",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			stream, err := ttnpb.NewAppAsClient(as).Subscribe(ctx, appID)
			if err != nil {
				return err
			}

			var streamErr error
			go func() {
				defer cancel()
				for {
					up, err := stream.Recv()
					if err != nil {
						streamErr = err
						return
					}
					if err = io.Write(os.Stdout, config.OutputFormat, up); err != nil {
						streamErr = err
						return
					}
				}
			}()

			<-ctx.Done()

			if streamErr != nil {
				return streamErr
			}
			return ctx.Err()
		},
	}
)

func init() {
	applicationsSubscribeCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationsSubscribeCommand)
}
