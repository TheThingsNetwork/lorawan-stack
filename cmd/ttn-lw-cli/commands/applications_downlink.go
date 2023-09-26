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

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	applicationsDownlinkCommand = &cobra.Command{
		Use:   "downlink",
		Short: "Application downlink commands",
	}
	applicationsDownlinkPushCommand = &cobra.Command{
		Use:   "push [application-id] [device-id]",
		Short: "Push to the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			downlink := &ttnpb.ApplicationDownlink{}
			_, err = downlink.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			gateways, err := GetClassBCGatewayIdentifiers(cmd.Flags(), "class-b-c")
			if err != nil {
				return err
			}
			if len(gateways) > 0 {
				downlink.ClassBC.Gateways = gateways
			}
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAppAsClient(as).DownlinkQueuePush(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: devID,
				Downlinks:    []*ttnpb.ApplicationDownlink{downlink},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsDownlinkReplaceCommand = &cobra.Command{
		Use:   "replace [application-id] [device-id]",
		Short: "Replace the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			downlink := &ttnpb.ApplicationDownlink{}
			_, err = downlink.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			gateways, err := GetClassBCGatewayIdentifiers(cmd.Flags(), "class-b-c")
			if err != nil {
				return err
			}
			if len(gateways) > 0 {
				downlink.ClassBC.Gateways = gateways
			}
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAppAsClient(as).DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: devID,
				Downlinks:    []*ttnpb.ApplicationDownlink{downlink},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsDownlinkClearCommand = &cobra.Command{
		Use:   "clear [application-id] [device-id]",
		Short: "Clear the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAppAsClient(as).DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsDownlinkListCommand = &cobra.Command{
		Use:   "list [application-id] [device-id]",
		Short: "List the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewAppAsClient(as).DownlinkQueueList(ctx, devID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Downlinks)
		},
	}
)

func init() {
	ttnpb.AddSetFlagsForApplicationDownlink(applicationsDownlinkPushCommand.Flags(), "", false)
	AddGatewayAntennaIdentifierFlags(applicationsDownlinkPushCommand.Flags(), "class-b-c")
	applicationsDownlinkPushCommand.Flags().AddFlagSet(endDeviceIDFlags())
	applicationsDownlinkCommand.AddCommand(applicationsDownlinkPushCommand)
	ttnpb.AddSetFlagsForApplicationDownlink(applicationsDownlinkReplaceCommand.Flags(), "", false)
	AddGatewayAntennaIdentifierFlags(applicationsDownlinkReplaceCommand.Flags(), "class-b-c")
	applicationsDownlinkReplaceCommand.Flags().AddFlagSet(endDeviceIDFlags())
	applicationsDownlinkCommand.AddCommand(applicationsDownlinkReplaceCommand)
	applicationsDownlinkClearCommand.Flags().AddFlagSet(endDeviceIDFlags())
	applicationsDownlinkCommand.AddCommand(applicationsDownlinkClearCommand)
	applicationsDownlinkListCommand.Flags().AddFlagSet(endDeviceIDFlags())
	applicationsDownlinkCommand.AddCommand(applicationsDownlinkListCommand)

	// The applicationsDownlinkCommand is placed under the end device command
	// It's aliased here, but hidden from the documentation.
	hiddenDownlink := *applicationsDownlinkCommand
	hiddenDownlink.Hidden = true
	applicationsCommand.AddCommand(&hiddenDownlink)
}
