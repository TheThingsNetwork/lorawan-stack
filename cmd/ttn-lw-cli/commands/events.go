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
	"os"
	"sync"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var eventsCommand = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event", "evt", "e"},
	Short:   "Subscribe to events",
	RunE: func(cmd *cobra.Command, args []string) error {
		var wg sync.WaitGroup

		addresses := make(map[string]bool)
		addresses[config.IdentityServerGRPCAddress] = true
		if config.GatewayServerEnabled {
			addresses[config.GatewayServerGRPCAddress] = true
		}
		if config.NetworkServerEnabled {
			addresses[config.NetworkServerGRPCAddress] = true
		}
		if config.ApplicationServerEnabled {
			addresses[config.ApplicationServerGRPCAddress] = true
		}
		if config.JoinServerEnabled {
			addresses[config.JoinServerGRPCAddress] = true
		}

		ids := getCombinedIdentifiers(cmd.Flags()).GetEntityIdentifiers()
		if len(ids) == 0 {
			return errNoIDs
		}
		tail, _ := cmd.Flags().GetUint32("tail")
		req := &ttnpb.StreamEventsRequest{
			Identifiers: ids,
			Tail:        tail,
		}

		events := make(chan *ttnpb.Event)
		for address := range addresses {
			conn, err := api.Dial(ctx, address)
			if err != nil {
				return err
			}
			stream, err := ttnpb.NewEventsClient(conn).Stream(ctx, req)
			if err != nil {
				return err
			}
			wg.Add(1)
			go func() {
				for {
					event, err := stream.Recv()
					if err != nil {
						if !errors.IsCanceled(err) {
							logger.WithError(err).Warn("Event stream closed")
						}
						break
					}
					events <- event
				}
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(events)
		}()

		for evt := range events {
			io.Write(os.Stdout, config.OutputFormat, evt)
		}

		return ctx.Err()
	},
}

func init() {
	eventsCommand.Flags().AddFlagSet(combinedIdentifiersFlags())
	eventsCommand.Flags().Uint32("tail", 0, "")
	Root.AddCommand(eventsCommand)
}
