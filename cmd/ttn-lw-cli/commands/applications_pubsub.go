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
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	selectApplicationPubSubFlags = util.FieldMaskFlags(&ttnpb.ApplicationPubSub{})
	setApplicationPubSubFlags    = util.FieldFlags(&ttnpb.ApplicationPubSub{})
)

func applicationPubSubIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("application-id", "", "")
	flagSet.String("pubsub-id", "", "")
	return flagSet
}

func natsFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Bool("nats", false, "use the NATS provider")
	flagSet.String("nats-server-url", "", "")
	return flagSet
}

var errNoPubSubID = errors.DefineInvalidArgument("no_pub_sub_id", "no pubsub ID set")

func getApplicationPubSubID(flagSet *pflag.FlagSet, args []string) (*ttnpb.ApplicationPubSubIdentifiers, error) {
	applicationID, _ := flagSet.GetString("application-id")
	pubsubID, _ := flagSet.GetString("pubsub-id")
	switch len(args) {
	case 0:
	case 1:
		logger.Warn("Only single ID found in arguments, not considering arguments")
	case 2:
		applicationID = args[0]
		pubsubID = args[1]
	default:
		logger.Warn("multiple IDs found in arguments, considering the first")
		applicationID = args[0]
		pubsubID = args[1]
	}
	if applicationID == "" {
		return nil, errNoApplicationID
	}
	if pubsubID == "" {
		return nil, errNoPubSubID
	}
	return &ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: applicationID},
		PubSubID:               pubsubID,
	}, nil
}

var (
	applicationsPubSubsCommand = &cobra.Command{
		Use:     "pubsubs",
		Aliases: []string{"pubsub", "ps"},
		Short:   "Application pubsub commands",
	}
	applicationsPubSubsGetFormatsCommand = &cobra.Command{
		Use:     "get-formats",
		Aliases: []string{"formats"},
		Short:   "Get the available formats for application pubsubs",
		RunE: func(cmd *cobra.Command, args []string) error {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPubSubRegistryClient(as).GetFormats(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPubSubsGetCommand = &cobra.Command{
		Use:     "get [application-id] [pubsub-id]",
		Aliases: []string{"info"},
		Short:   "Get the properties of an application pubsub",
		RunE: func(cmd *cobra.Command, args []string) error {
			pubsubID, err := getApplicationPubSubID(cmd.Flags(), args)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationPubSubFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationPubSubFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPubSubRegistryClient(as).Get(ctx, &ttnpb.GetApplicationPubSubRequest{
				ApplicationPubSubIdentifiers: *pubsubID,
				FieldMask:                    types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPubSubsListCommand = &cobra.Command{
		Use:     "list [application-id]",
		Aliases: []string{"ls"},
		Short:   "List application pubsubs",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationPubSubFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationPubSubFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPubSubRegistryClient(as).List(ctx, &ttnpb.ListApplicationPubSubsRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPubSubsSetCommand = &cobra.Command{
		Use:     "set [application-id] [pubsub-id]",
		Aliases: []string{"update"},
		Short:   "Set the properties of an application pubsub",
		RunE: func(cmd *cobra.Command, args []string) error {
			pubsubID, err := getApplicationPubSubID(cmd.Flags(), args)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setApplicationPubSubFlags)

			var pubsub ttnpb.ApplicationPubSub
			if err = util.SetFields(&pubsub, setApplicationPubSubFlags); err != nil {
				return err
			}
			pubsub.ApplicationPubSubIdentifiers = *pubsubID

			if nats, _ := cmd.Flags().GetBool("nats"); nats {
				serverURL, _ := cmd.Flags().GetString("nats-server-url")
				pubsub.Provider = &ttnpb.ApplicationPubSub_NATS{
					NATS: &ttnpb.ApplicationPubSub_NATSProvider{
						ServerURL: serverURL,
					},
				}
				paths = append(paths, "provider")
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPubSubRegistryClient(as).Set(ctx, &ttnpb.SetApplicationPubSubRequest{
				ApplicationPubSub: pubsub,
				FieldMask:         types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPubSubsDeleteCommand = &cobra.Command{
		Use:   "delete [application-id] [pubsub-id]",
		Short: "Delete an application pubsub",
		RunE: func(cmd *cobra.Command, args []string) error {
			pubsubID, err := getApplicationPubSubID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationPubSubRegistryClient(as).Delete(ctx, pubsubID)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	applicationsPubSubsCommand.AddCommand(applicationsPubSubsGetFormatsCommand)
	applicationsPubSubsGetCommand.Flags().AddFlagSet(applicationPubSubIDFlags())
	applicationsPubSubsGetCommand.Flags().AddFlagSet(selectApplicationPubSubFlags)
	applicationsPubSubsCommand.AddCommand(applicationsPubSubsGetCommand)
	applicationsPubSubsListCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsPubSubsListCommand.Flags().AddFlagSet(selectApplicationPubSubFlags)
	applicationsPubSubsCommand.AddCommand(applicationsPubSubsListCommand)
	applicationsPubSubsSetCommand.Flags().AddFlagSet(applicationPubSubIDFlags())
	applicationsPubSubsSetCommand.Flags().AddFlagSet(setApplicationPubSubFlags)
	applicationsPubSubsSetCommand.Flags().AddFlagSet(natsFlags())
	applicationsPubSubsCommand.AddCommand(applicationsPubSubsSetCommand)
	applicationsPubSubsDeleteCommand.Flags().AddFlagSet(applicationPubSubIDFlags())
	applicationsPubSubsCommand.AddCommand(applicationsPubSubsDeleteCommand)
	applicationsCommand.AddCommand(applicationsPubSubsCommand)
}
