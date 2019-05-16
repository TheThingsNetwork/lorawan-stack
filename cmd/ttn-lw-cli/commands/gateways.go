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

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/pkg/types"
)

var (
	selectGatewayFlags     = util.FieldMaskFlags(&ttnpb.Gateway{})
	setGatewayFlags        = util.FieldFlags(&ttnpb.Gateway{})
	setGatewayAntennaFlags = util.FieldFlags(&ttnpb.GatewayAntenna{}, "antenna")
)

func gatewayIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("gateway-id", "", "")
	flagSet.String("gateway-eui", "", "")
	return flagSet
}

var errNoGatewayID = errors.DefineInvalidArgument("no_gateway_id", "no gateway ID set")

func getGatewayID(flagSet *pflag.FlagSet, args []string, requireID bool) (*ttnpb.GatewayIdentifiers, error) {
	gatewayID, _ := flagSet.GetString("gateway-id")
	gatewayEUIHex, _ := flagSet.GetString("gateway-eui")
	switch len(args) {
	case 0:
	case 1:
		gatewayID = args[0]
	case 2:
		gatewayID = args[0]
		gatewayEUIHex = args[1]
	default:
		logger.Warn("multiple IDs found in arguments, considering the first")
		gatewayID = args[0]
		gatewayEUIHex = args[1]
	}
	if gatewayID == "" && requireID {
		return nil, errNoGatewayID
	}
	ids := &ttnpb.GatewayIdentifiers{GatewayID: gatewayID}
	if gatewayEUIHex != "" {
		var gatewayEUI ttntypes.EUI64
		if err := gatewayEUI.UnmarshalText([]byte(gatewayEUIHex)); err != nil {
			return nil, err
		}
		ids.EUI = &gatewayEUI
	}
	return ids, nil
}

var (
	gatewaysCommand = &cobra.Command{
		Use:     "gateways",
		Aliases: []string{"gateway", "gtw", "g"},
		Short:   "Gateway commands",
	}
	gatewaysListFrequencyPlans = &cobra.Command{
		Use:               "list-frequency-plans",
		Short:             "List available frequency plans for gateways",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseFrequency, _ := cmd.Flags().GetUint32("base-frequency")
			gs, err := api.Dial(ctx, config.GatewayServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewConfigurationClient(gs).ListFrequencyPlans(ctx, &ttnpb.ListFrequencyPlansRequest{
				BaseFrequency: baseFrequency,
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res.FrequencyPlans)
		},
	}
	gatewaysListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List gateways",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectGatewayFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewGatewayRegistryClient(is).List(ctx, &ttnpb.ListGatewaysRequest{
				Collaborator: getCollaborator(cmd.Flags()),
				FieldMask:    types.FieldMask{Paths: paths},
				Limit:        limit,
				Page:         page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Gateways)
		},
	}
	gatewaysSearchCommand = &cobra.Command{
		Use:   "search",
		Short: "Search for gateways",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectGatewayFlags)

			req := getSearchEntitiesRequest(cmd.Flags())
			req.FieldMask.Paths = paths

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEntityRegistrySearchClient(is).SearchGateways(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Gateways)
		},
	}
	gatewaysGetCommand = &cobra.Command{
		Use:     "get [gateway-id]",
		Aliases: []string{"info"},
		Short:   "Get a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, false)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectGatewayFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}

			cli := ttnpb.NewGatewayRegistryClient(is)

			if gtwID.GatewayID == "" && gtwID.EUI != nil {
				gtwID, err = cli.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
					EUI: *gtwID.EUI,
				})
				if err != nil {
					return err
				}
			}

			res, err := cli.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIdentifiers: *gtwID,
				FieldMask:          types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	gatewaysCreateCommand = &cobra.Command{
		Use:     "create [gateway-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create a gateway",
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			gtwID, err := getGatewayID(cmd.Flags(), args, false)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setGatewayFlags, attributesFlags())

			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}
			var gateway ttnpb.Gateway
			if inputDecoder != nil {
				_, err := inputDecoder.Decode(&gateway)
				if err != nil {
					return err
				}
			}

			setDefaults, _ := cmd.Flags().GetBool("defaults")
			if setDefaults {
				gateway.GatewayServerAddress = getHost(config.GatewayServerGRPCAddress)
				paths = append(paths,
					"gateway_server_address",
				)
			}

			if err = util.SetFields(&gateway, setGatewayFlags); err != nil {
				return err
			}

			gateway.Attributes = mergeAttributes(gateway.Attributes, cmd.Flags())
			if gtwID != nil {
				if gtwID.GatewayID != "" {
					gateway.GatewayID = gtwID.GatewayID
				}
				if gtwID.EUI != nil {
					gateway.EUI = gtwID.EUI
				}
			}

			if gateway.GatewayID == "" {
				return errNoGatewayID
			}

			var antenna ttnpb.GatewayAntenna
			if err = util.SetFields(&antenna, setGatewayAntennaFlags, "antenna"); err != nil {
				return err
			}
			gateway.Antennas = []ttnpb.GatewayAntenna{antenna}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
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

			return io.Write(os.Stdout, config.OutputFormat, res)
		}),
	}
	errAntennaIndex       = errors.DefineInvalidArgument("antenna_index", "index of antenna to update out of bounds")
	gatewaysUpdateCommand = &cobra.Command{
		Use:     "update [gateway-id]",
		Aliases: []string{"set"},
		Short:   "Update a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setGatewayFlags, attributesFlags())
			antennaPaths := util.UpdateFieldMask(cmd.Flags(), setGatewayAntennaFlags)
			if len(paths)+len(antennaPaths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			var gateway ttnpb.Gateway
			if err = util.SetFields(&gateway, setGatewayFlags); err != nil {
				return err
			}
			gateway.Attributes = mergeAttributes(gateway.Attributes, cmd.Flags())
			gateway.GatewayIdentifiers = *gtwID

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}

			antennaAdd, _ := cmd.Flags().GetBool("antenna.add")
			antennaRemove, _ := cmd.Flags().GetBool("antenna.remove")
			if len(antennaPaths) > 0 || antennaAdd || antennaRemove {
				res, err := ttnpb.NewGatewayRegistryClient(is).Get(ctx, &ttnpb.GetGatewayRequest{
					GatewayIdentifiers: gateway.GatewayIdentifiers,
					FieldMask:          types.FieldMask{Paths: []string{"antennas"}},
				})
				if err != nil {
					return err
				}
				antennaIndex, _ := cmd.Flags().GetInt("antenna.index")
				if antennaAdd {
					res.Antennas = append(res.Antennas, ttnpb.GatewayAntenna{})
					antennaIndex = len(res.Antennas) - 1
				} else if antennaIndex > len(res.Antennas) {
					return errAntennaIndex
				}
				if antennaRemove {
					gateway.Antennas = append(res.Antennas[:antennaIndex], res.Antennas[antennaIndex+1:]...)
				} else { // create or update
					if err = util.SetFields(&res.Antennas[antennaIndex], setGatewayAntennaFlags, "antenna"); err != nil {
						return err
					}
					gateway.Antennas = res.Antennas
				}
				paths = append(paths, "antennas")
			}

			res, err := ttnpb.NewGatewayRegistryClient(is).Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway:   gateway,
				FieldMask: types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			res.SetFields(&gateway, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	gatewaysDeleteCommand = &cobra.Command{
		Use:   "delete [gateway-id]",
		Short: "Delete a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
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
		Use:   "connection-stats [gateway-id]",
		Short: "Get connection stats for a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}

			gateway, err := ttnpb.NewGatewayRegistryClient(is).Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIdentifiers: *gtwID,
				FieldMask:          types.FieldMask{Paths: []string{"gateway_server_address"}},
			})
			if err != nil {
				return err
			}

			if gsMismatch := compareServerAddressGateway(gateway, config); gsMismatch {
				return errAddressMismatchGateway
			}

			gs, err := api.Dial(ctx, config.GatewayServerGRPCAddress)
			if err != nil {
				return err
			}

			res, err := ttnpb.NewGsClient(gs).GetGatewayConnectionStats(ctx, gtwID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	gatewaysContactInfoCommand = contactInfoCommands("gateway", func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error) {
		gtwID, err := getGatewayID(cmd.Flags(), args, true)
		if err != nil {
			return nil, err
		}
		return gtwID.EntityIdentifiers(), nil
	})
)

func init() {
	gatewaysListFrequencyPlans.Flags().Uint32("base-frequency", 0, "Base frequency in MHz for hardware support (433, 470, 868 or 915)")
	gatewaysCommand.AddCommand(gatewaysListFrequencyPlans)
	gatewaysListCommand.Flags().AddFlagSet(collaboratorFlags())
	gatewaysListCommand.Flags().AddFlagSet(selectGatewayFlags)
	gatewaysListCommand.Flags().AddFlagSet(paginationFlags())
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
	gatewaysCreateCommand.Flags().AddFlagSet(setGatewayAntennaFlags)
	gatewaysCreateCommand.Flags().AddFlagSet(attributesFlags())
	gatewaysCreateCommand.Flags().Bool("defaults", true, "configure gateway with default gateway server address")
	gatewaysCommand.AddCommand(gatewaysCreateCommand)
	gatewaysUpdateCommand.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysUpdateCommand.Flags().AddFlagSet(setGatewayFlags)
	gatewaysUpdateCommand.Flags().Int("antenna.index", 0, "index of the antenna to update or remove")
	gatewaysUpdateCommand.Flags().Bool("antenna.add", false, "add an extra antenna")
	gatewaysUpdateCommand.Flags().Bool("antenna.remove", false, "remove an antenna")
	gatewaysUpdateCommand.Flags().AddFlagSet(setGatewayAntennaFlags)
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

var errAddressMismatchGateway = errors.DefineAborted("gateway_server_address_mismatch", "gateway server address mismatch")

func compareServerAddressGateway(gateway *ttnpb.Gateway, config *Config) (gsMismatch bool) {
	gsHost := getHost(config.GatewayServerGRPCAddress)
	if host := getHost(gateway.GatewayServerAddress); host != "" && host != gsHost {
		gsMismatch = true
		logger.WithFields(log.Fields(
			"configured", gsHost,
			"registered", host,
		)).Warn("Registered Gateway Server address does not match CLI configuration")
	}
	return
}
