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

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	selectApplicationLinkFlags  = util.FieldMaskFlags(&ttnpb.ApplicationLink{})
	setApplicationLinkFlags     = util.FieldFlags(&ttnpb.ApplicationLink{})
	setApplicationDownlinkFlags = util.FieldFlags(&ttnpb.ApplicationDownlink{})
)

var (
	applicationsLinkCommand = &cobra.Command{
		Use:   "link",
		Short: "Application link commands",
	}
	applicationsLinkGetCommand = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get the properties of an application link",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationLinkFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationLinkFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, flag.Name)
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewAsClient(as).GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	applicationsLinkSetCommand = &cobra.Command{
		Use:     "set",
		Aliases: []string{"update"},
		Short:   "Set the properties of an application link",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setApplicationLinkFlags)

			var link ttnpb.ApplicationLink
			if err := util.SetFields(&link, setApplicationLinkFlags); err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewAsClient(as).SetLink(ctx, &ttnpb.SetApplicationLinkRequest{
				ApplicationIdentifiers: *appID,
				ApplicationLink:        link,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	applicationsLinkDeleteCommand = &cobra.Command{
		Use:   "delete",
		Short: "Delete an application link",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAsClient(as).DeleteLink(ctx, appID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsSubscribeCommand = &cobra.Command{
		Use:     "subscribe",
		Aliases: []string{"sub"},
		Short:   "Subscribe to application uplink",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
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
					if err = io.Write(os.Stdout, config.Format, up); err != nil {
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
	applicationsDownlinkCommand = &cobra.Command{
		Use:   "downlink",
		Short: "Application downlink commands",
	}
	applicationsDownlinkPushCommand = &cobra.Command{
		Use:   "push",
		Short: "Push to the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			var downlink ttnpb.ApplicationDownlink
			if err = util.SetFields(&downlink, setApplicationDownlinkFlags); err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAppAsClient(as).DownlinkQueuePush(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: *devID,
				Downlinks:            []*ttnpb.ApplicationDownlink{&downlink},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsDownlinkReplaceCommand = &cobra.Command{
		Use:   "replace",
		Short: "Replace the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			var downlink ttnpb.ApplicationDownlink
			if err = util.SetFields(&downlink, setApplicationDownlinkFlags); err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAppAsClient(as).DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: *devID,
				Downlinks:            []*ttnpb.ApplicationDownlink{&downlink},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsDownlinkClearCommand = &cobra.Command{
		Use:   "clear",
		Short: "Clear the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAppAsClient(as).DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: *devID,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsDownlinkListCommand = &cobra.Command{
		Use:   "list",
		Short: "List the application downlink queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewAppAsClient(as).DownlinkQueueList(ctx, devID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res.Downlinks)
		},
	}
)

func init() {
	applicationsLinkGetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsLinkGetCommand.Flags().AddFlagSet(selectApplicationLinkFlags)
	applicationsLinkCommand.AddCommand(applicationsLinkGetCommand)
	applicationsLinkSetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsLinkSetCommand.Flags().AddFlagSet(setApplicationLinkFlags)
	applicationsLinkCommand.AddCommand(applicationsLinkSetCommand)
	applicationsLinkDeleteCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsLinkCommand.AddCommand(applicationsLinkDeleteCommand)
	applicationsCommand.AddCommand(applicationsLinkCommand)
	applicationsSubscribeCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationsSubscribeCommand)
	applicationsDownlinkPushCommand.Flags().AddFlagSet(setApplicationDownlinkFlags)
	applicationsDownlinkPushCommand.Flags().AddFlagSet(endDeviceIDFlags())
	applicationsDownlinkCommand.AddCommand(applicationsDownlinkPushCommand)
	applicationsDownlinkReplaceCommand.Flags().AddFlagSet(setApplicationDownlinkFlags)
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
