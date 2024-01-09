// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
	"strconv"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	selectRelaySettingsFlags             = util.NormalizedFlagSet()
	selectRelayUplinkForwardingRuleFlags = util.NormalizedFlagSet()

	setRelaySettingsFlags             = util.NormalizedFlagSet()
	setRelayUplinkForwardingRuleFlags = util.NormalizedFlagSet()
)

func getRelayUplinkForwardingRuleID(args []string) (*ttnpb.EndDeviceIdentifiers, uint32, error) {
	applicationID, deviceID, index := "", "", uint32(0)
	switch len(args) {
	case 0:
	case 1, 2:
		logger.Warn("Partial ID found in arguments, not considering arguments")
	case 3:
		applicationID = args[0]
		deviceID = args[1]
		i, err := strconv.ParseUint(args[2], 10, 32)
		if err != nil {
			return nil, 0, err
		}
		index = uint32(i)
	default:
		logger.Warn("Multiple IDs found in arguments, considering the first")
		applicationID = args[0]
		deviceID = args[1]
		i, err := strconv.ParseUint(args[2], 10, 32)
		if err != nil {
			return nil, 0, err
		}
		index = uint32(i)
	}
	if applicationID == "" {
		return nil, 0, errNoApplicationID.New()
	}
	if deviceID == "" {
		return nil, 0, errNoEndDeviceID.New()
	}
	return &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: applicationID,
		},
		DeviceId: deviceID,
	}, index, nil
}

var (
	relaysCommand = &cobra.Command{
		Use:     "relays",
		Aliases: []string{"r"},
		Short:   "Relay commands (EXPERIMENTAL)",
	}
	relaysGetCommand = &cobra.Command{
		Use:     "get [application-id] [device-id]",
		Aliases: []string{"info"},
		Short:   "Get a relay (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectRelaySettingsFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).GetRelay(
				ctx,
				&ttnpb.GetRelayRequest{
					EndDeviceIds: devID,
					FieldMask:    ttnpb.FieldMask(paths...),
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relaysCreateCommand = &cobra.Command{
		Use:     "create [application-id] [device-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create a relay (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			settings := &ttnpb.RelaySettings{}
			if _, err := settings.SetFromFlags(cmd.Flags(), ""); err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).CreateRelay(
				ctx,
				&ttnpb.CreateRelayRequest{
					EndDeviceIds: devID,
					Settings:     settings,
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relaysUpdateCommand = &cobra.Command{
		Use:     "update [application-id] [device-id]",
		Aliases: []string{"set"},
		Short:   "Update a relay (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			settings := &ttnpb.RelaySettings{}
			paths, err := settings.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).UpdateRelay(
				ctx,
				&ttnpb.UpdateRelayRequest{
					EndDeviceIds: devID,
					Settings:     settings,
					FieldMask:    ttnpb.FieldMask(paths...),
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relaysDeleteCommand = &cobra.Command{
		Use:     "delete [application-id] [device-id]",
		Aliases: []string{"del", "remove"},
		Short:   "Delete a relay (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewNsRelayConfigurationServiceClient(ns).DeleteRelay(
				ctx,
				&ttnpb.DeleteRelayRequest{
					EndDeviceIds: devID,
				},
			)
			if err != nil {
				return err
			}

			return nil
		},
	}
	relaysUplinkForwardingRulesCommand = &cobra.Command{
		Use:     "uplink-forwarding-rules",
		Aliases: []string{"uplink-forwarding", "uf", "ufr"},
		Short:   "Uplink forwarding rules commands (EXPERIMENTAL)",
	}
	relaysListUplinkForwardingRulesCommand = &cobra.Command{
		Use:     "list [application-id] [device-id]",
		Aliases: []string{"ls"},
		Short:   "List uplink forwarding rules (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectRelayUplinkForwardingRuleFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).ListRelayUplinkForwardingRules(
				ctx,
				&ttnpb.ListRelayUplinkForwardingRulesRequest{
					EndDeviceIds: devID,
					FieldMask:    ttnpb.FieldMask(paths...),
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relaysGetUplinkForwardingRuleCommand = &cobra.Command{
		Use:     "get [application-id] [device-id] [index]",
		Aliases: []string{"info"},
		Short:   "Get an uplink forwarding rule (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, index, err := getRelayUplinkForwardingRuleID(args)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectRelayUplinkForwardingRuleFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).GetRelayUplinkForwardingRule(
				ctx,
				&ttnpb.GetRelayUplinkForwardingRuleRequest{
					EndDeviceIds: devID,
					Index:        index,
					FieldMask:    ttnpb.FieldMask(paths...),
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relaysCreateUplinkForwardingRuleCommand = &cobra.Command{
		Use:     "create [application-id] [device-id] [index]",
		Aliases: []string{"add", "register"},
		Short:   "Create an uplink forwarding rule (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, index, err := getRelayUplinkForwardingRuleID(args)
			if err != nil {
				return err
			}

			rule := &ttnpb.RelayUplinkForwardingRule{}
			if _, err := rule.SetFromFlags(cmd.Flags(), ""); err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).CreateRelayUplinkForwardingRule(
				ctx,
				&ttnpb.CreateRelayUplinkForwardingRuleRequest{
					EndDeviceIds: devID,
					Index:        index,
					Rule:         rule,
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relaysUpdateUplinkForwardingRuleCommand = &cobra.Command{
		Use:     "update [application-id] [device-id] [index]",
		Aliases: []string{"set"},
		Short:   "Update an uplink forwarding rule (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, index, err := getRelayUplinkForwardingRuleID(args)
			if err != nil {
				return err
			}

			rule := &ttnpb.RelayUplinkForwardingRule{}
			paths, err := rule.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			resp, err := ttnpb.NewNsRelayConfigurationServiceClient(ns).UpdateRelayUplinkForwardingRule(
				ctx,
				&ttnpb.UpdateRelayUplinkForwardingRuleRequest{
					EndDeviceIds: devID,
					Index:        index,
					Rule:         rule,
					FieldMask:    ttnpb.FieldMask(paths...),
				},
			)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
	relayDeleteUplinkForwardingRuleCommand = &cobra.Command{
		Use:     "delete [application-id] [device-id] [index]",
		Aliases: []string{"del", "remove"},
		Short:   "Delete an uplink forwarding rule (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, index, err := getRelayUplinkForwardingRuleID(args)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewNsRelayConfigurationServiceClient(ns).DeleteRelayUplinkForwardingRule(
				ctx,
				&ttnpb.DeleteRelayUplinkForwardingRuleRequest{
					EndDeviceIds: devID,
					Index:        index,
				},
			)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	ttnpb.AddSelectFlagsForRelaySettings(selectRelaySettingsFlags, "", false)
	ttnpb.AddSelectFlagsForRelayUplinkForwardingRule(selectRelayUplinkForwardingRuleFlags, "", false)
	ttnpb.AddSetFlagsForRelaySettings(setRelaySettingsFlags, "", false)
	ttnpb.AddSetFlagsForRelayUplinkForwardingRule(setRelayUplinkForwardingRuleFlags, "", false)

	relaysGetCommand.Flags().AddFlagSet(selectRelaySettingsFlags)
	relaysCommand.AddCommand(relaysGetCommand)
	relaysCreateCommand.Flags().AddFlagSet(setRelaySettingsFlags)
	relaysCommand.AddCommand(relaysCreateCommand)
	relaysUpdateCommand.Flags().AddFlagSet(setRelaySettingsFlags)
	relaysCommand.AddCommand(relaysUpdateCommand)
	relaysCommand.AddCommand(relaysDeleteCommand)

	relaysListUplinkForwardingRulesCommand.Flags().AddFlagSet(selectRelayUplinkForwardingRuleFlags)
	relaysUplinkForwardingRulesCommand.AddCommand(relaysListUplinkForwardingRulesCommand)
	relaysGetUplinkForwardingRuleCommand.Flags().AddFlagSet(selectRelayUplinkForwardingRuleFlags)
	relaysUplinkForwardingRulesCommand.AddCommand(relaysGetUplinkForwardingRuleCommand)
	relaysCreateUplinkForwardingRuleCommand.Flags().AddFlagSet(setRelayUplinkForwardingRuleFlags)
	relaysUplinkForwardingRulesCommand.AddCommand(relaysCreateUplinkForwardingRuleCommand)
	relaysUpdateUplinkForwardingRuleCommand.Flags().AddFlagSet(setRelayUplinkForwardingRuleFlags)
	relaysUplinkForwardingRulesCommand.AddCommand(relaysUpdateUplinkForwardingRuleCommand)
	relaysUplinkForwardingRulesCommand.AddCommand(relayDeleteUplinkForwardingRuleCommand)
	relaysCommand.AddCommand(relaysUplinkForwardingRulesCommand)

	Root.AddCommand(relaysCommand)
}
