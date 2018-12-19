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
	"os"

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
	selectGatewayFlags = util.FieldMaskFlags(&ttnpb.Gateway{})
	setGatewayFlags    = util.FieldFlags(&ttnpb.Gateway{})
)

func gatewayIDFlags() *pflag.FlagSet {
	flagSet := new(pflag.FlagSet)
	flagSet.String("gateway-id", "", "")
	return flagSet
}

var errNoGatewayID = errors.DefineInvalidArgument("no_gateway_id", "no gateway ID set")

func getGatewayID(flagSet *pflag.FlagSet, args []string) *ttnpb.GatewayIdentifiers {
	var gatewayID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("multiple IDs found in arguments, considering only the first")
		}
		gatewayID = args[0]
	} else {
		gatewayID, _ = flagSet.GetString("gateway-id")
	}
	if gatewayID == "" {
		return nil
	}
	return &ttnpb.GatewayIdentifiers{GatewayID: gatewayID}
}

var (
	gatewaysCommand = &cobra.Command{
		Use:     "gateways",
		Aliases: []string{"gateway", "gtw", "g"},
		Short:   "Gateway commands",
	}
	gatewaysListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List gateways",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectGatewayFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectGatewayFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, flag.Name)
				})
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayRegistryClient(is).List(ctx, &ttnpb.ListGatewaysRequest{
				Collaborator: getCollaborator(cmd.Flags()),
				FieldMask:    types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res.Gateways)
		},
	}
	gatewaysSearchCommand = &cobra.Command{
		Use:   "search",
		Short: "Search for gateways",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectGatewayFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectGatewayFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, flag.Name)
				})
			}
			req := getSearchEntitiesRequest(cmd.Flags())
			req.FieldMask.Paths = paths

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEntityRegistrySearchClient(is).SearchGateways(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res.Gateways)
		},
	}
	gatewaysGetCommand = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectGatewayFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectGatewayFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, flag.Name)
				})
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayRegistryClient(is).Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIdentifiers: *gtwID,
				FieldMask:          types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	gatewaysCreateCommand = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "register"},
		Short:   "Create a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}
			var gateway ttnpb.Gateway
			util.SetFields(&gateway, setGatewayFlags)
			gateway.Attributes = mergeAttributes(gateway.Attributes, cmd.Flags())
			gateway.GatewayIdentifiers = *gtwID

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayRegistryClient(is).Create(ctx, &ttnpb.CreateGatewayRequest{
				Gateway:      gateway,
				Collaborator: *collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	gatewaysUpdateCommand = &cobra.Command{
		Use:     "update",
		Aliases: []string{"set"},
		Short:   "Update a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setGatewayFlags, attributesFlags())
			if len(paths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			var gateway ttnpb.Gateway
			util.SetFields(&gateway, setGatewayFlags)
			gateway.Attributes = mergeAttributes(gateway.Attributes, cmd.Flags())
			gateway.GatewayIdentifiers = *gtwID

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayRegistryClient(is).Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway:   gateway,
				FieldMask: types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	gatewaysDeleteCommand = &cobra.Command{
		Use:   "delete",
		Short: "Delete a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayRegistryClient(is).Delete(ctx, gtwID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	gatewaysConnectionStats = &cobra.Command{
		Use:   "connection-stats",
		Short: "Get connection stats for a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}

			gs, err := api.Dial(ctx, config.GatewayServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGsClient(gs).GetGatewayConnectionStats(ctx, gtwID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	gatewaysContactInfoCommand = contactInfoCommands("gateway", func(cmd *cobra.Command) (*ttnpb.EntityIdentifiers, error) {
		gtwID := getGatewayID(cmd.Flags(), nil)
		if gtwID == nil {
			return nil, errNoGatewayID
		}
		return gtwID.EntityIdentifiers(), nil
	})
)

func init() {
	gatewaysListCommand.Flags().AddFlagSet(collaboratorFlags())
	gatewaysListCommand.Flags().AddFlagSet(selectGatewayFlags)
	gatewaysCommand.AddCommand(gatewaysListCommand)
	gatewaysSearchCommand.Flags().AddFlagSet(searchFlags())
	gatewaysSearchCommand.Flags().AddFlagSet(selectGatewayFlags)
	gatewaysCommand.AddCommand(gatewaysSearchCommand)
	gatewaysGetCommand.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysGetCommand.Flags().AddFlagSet(selectGatewayFlags)
	gatewaysCommand.AddCommand(gatewaysGetCommand)
	gatewaysCreateCommand.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysCreateCommand.Flags().AddFlagSet(collaboratorFlags())
	gatewaysCreateCommand.Flags().AddFlagSet(setGatewayFlags)
	gatewaysCreateCommand.Flags().AddFlagSet(attributesFlags())
	gatewaysCommand.AddCommand(gatewaysCreateCommand)
	gatewaysUpdateCommand.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysUpdateCommand.Flags().AddFlagSet(setGatewayFlags)
	gatewaysUpdateCommand.Flags().AddFlagSet(attributesFlags())
	gatewaysCommand.AddCommand(gatewaysUpdateCommand)
	gatewaysDeleteCommand.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewaysDeleteCommand)
	gatewaysConnectionStats.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewaysConnectionStats)
	gatewaysContactInfoCommand.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewaysContactInfoCommand)
	Root.AddCommand(gatewaysCommand)
}
