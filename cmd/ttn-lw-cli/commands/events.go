// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func getEventsAddresses() []string {
	addressMap := make(map[string]bool)
	addressMap[config.IdentityServerGRPCAddress] = true
	if config.GatewayServerEnabled {
		addressMap[config.GatewayServerGRPCAddress] = true
	}
	if config.NetworkServerEnabled {
		addressMap[config.NetworkServerGRPCAddress] = true
	}
	if config.ApplicationServerEnabled {
		addressMap[config.ApplicationServerGRPCAddress] = true
	}
	if config.JoinServerEnabled {
		addressMap[config.JoinServerGRPCAddress] = true
	}
	var addresses []string
	for address := range addressMap {
		addresses = append(addresses, address)
	}
	return addresses
}

var eventsCommand = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event", "evt", "e"},
	Short:   "Subscribe to events",
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := getEntityIdentifiersSlice(cmd.Flags())
		if len(ids) == 0 {
			return errNoIDs
		}
		tail, _ := cmd.Flags().GetUint32("tail")
		req := &ttnpb.StreamEventsRequest{
			Identifiers: ids,
			Tail:        tail,
		}

		g, gCtx := errgroup.WithContext(ctx)

		events := make(chan *ttnpb.Event)
		go func() {
			g.Wait()
			close(events)
		}()

		for _, address := range getEventsAddresses() {
			address := address // shadow loop variable.
			g.Go(func() error {
				conn, err := api.Dial(gCtx, address)
				if err != nil {
					return err
				}
				stream, err := ttnpb.NewEventsClient(conn).Stream(gCtx, req)
				if err != nil {
					return err
				}
				for {
					event, err := stream.Recv()
					if err != nil {
						if !errors.IsCanceled(err) {
							return err
						}
						break
					}
					select {
					case <-gCtx.Done():
						return gCtx.Err()
					case events <- event:
					}
				}
				return nil
			})
		}

		for evt := range events {
			io.Write(os.Stdout, config.OutputFormat, evt)
		}

		return ctx.Err()
	},
}

var eventsFindRelatedCommand = &cobra.Command{
	Use:     "find-related [correlation-id]",
	Aliases: []string{"related"},
	Short:   "Find related events by correlation ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		var correlationID string
		if len(args) > 0 {
			if len(args) > 1 {
				logger.Warn("Multiple IDs found in arguments, considering only the first")
			}
			correlationID = args[0]
		} else {
			correlationID, _ = cmd.Flags().GetString("correlation-id")
		}
		req := &ttnpb.FindRelatedEventsRequest{
			CorrelationID: correlationID,
		}

		g, gCtx := errgroup.WithContext(ctx)

		events := make(chan *ttnpb.Event)
		go func() {
			g.Wait()
			close(events)
		}()

		for _, address := range getEventsAddresses() {
			address := address // shadow loop variable.
			g.Go(func() error {
				conn, err := api.Dial(gCtx, address)
				if err != nil {
					return err
				}
				res, err := ttnpb.NewEventsClient(conn).FindRelated(gCtx, req)
				if err != nil {
					return err
				}
				for _, event := range res.GetEvents() {
					select {
					case <-gCtx.Done():
						return gCtx.Err()
					case events <- event:
					}
				}
				return nil
			})
		}

		for evt := range events {
			io.Write(os.Stdout, config.OutputFormat, evt)
		}

		return ctx.Err()
	},
}

func init() {
	eventsCommand.Flags().AddFlagSet(entityIdentifiersSliceFlags())
	eventsCommand.Flags().Uint32("tail", 0, "")
	Root.AddCommand(eventsCommand)
	eventsFindRelatedCommand.Flags().String("correlation-id", "", "")
	eventsCommand.AddCommand(eventsFindRelatedCommand)
}
