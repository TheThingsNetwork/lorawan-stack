// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func packetBrokerNetworkIDFlags(allowDefault bool) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("net-id", "", "(hex)")
	flagSet.String("tenant-id", "", "")
	if allowDefault {
		flagSet.Bool("default", false, "")
	}
	return flagSet
}

var errDefaultPacketBrokerNetworkIdentifier = errors.DefineInvalidArgument("default_packet_broker_network_identifier", "default Packet Broker network identifier")

func packetBrokerCacheInfo() {
	logger.Info("Changes to routing policies take up to 10 minutes to be fully propagated")
}

// getPacketBrokerNetworkID parses the network identifier from arguments and flags or nil if default is specified.
func getPacketBrokerNetworkID(flagSet *pflag.FlagSet, args []string, allowDefault bool) (*ttnpb.PacketBrokerNetworkIdentifier, error) {
	def, _ := flagSet.GetBool("default")
	netIDHex, _ := flagSet.GetString("net-id")
	tenantID, _ := flagSet.GetString("tenant-id")
	switch len(args) {
	case 0:
	case 1:
		if args[0] == "default" {
			def = true
		} else {
			netIDHex = args[0]
		}
	case 2:
		netIDHex = args[0]
		tenantID = args[1]
	default:
		logger.Warn("Multiple IDs found in arguments, considering the first")
		netIDHex = args[0]
		tenantID = args[1]
	}
	if def && (!allowDefault || netIDHex != "" || tenantID != "") {
		return nil, errDefaultPacketBrokerNetworkIdentifier.New()
	}
	if def {
		return nil, nil
	}
	var netID types.NetID
	if err := netID.UnmarshalText([]byte(netIDHex)); err != nil {
		return nil, errInvalidNetID.WithCause(err)
	}
	return &ttnpb.PacketBrokerNetworkIdentifier{
		NetId:    netID.MarshalNumber(),
		TenantId: tenantID,
	}, nil
}

func packetBrokerNetworkSearchFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("tenant-id-contains", "", "")
	flagSet.String("name-contains", "", "")
	return flagSet
}

func getPacketBrokerNetworkSearch(flagSet *pflag.FlagSet) (tenantIDContains, nameContains string) {
	tenantIDContains, _ = flagSet.GetString("tenant-id-contains")
	nameContains, _ = flagSet.GetString("name-contains")
	return
}

func packetBrokerRoutingPolicyFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Bool("join", false, "join-request and join-accept")
	flagSet.Bool("join-request", false, "")
	flagSet.Bool("join-accept", false, "")
	flagSet.Bool("mac-data", false, "MAC data uplink and downlink (FPort = 0)")
	flagSet.Bool("mac-data-up", false, "")
	flagSet.Bool("mac-data-down", false, "")
	flagSet.Bool("application-data", false, "application data uplink and downlink (FPort > 0)")
	flagSet.Bool("application-data-up", false, "")
	flagSet.Bool("application-data-down", false, "")
	flagSet.Bool("signal-quality", false, "")
	flagSet.Bool("localization", false, "")
	flagSet.Bool("all", false, "enable all data")
	return flagSet
}

func getPacketBrokerRoutingPolicy(flagSet *pflag.FlagSet) (up ttnpb.PacketBrokerRoutingPolicyUplink, down ttnpb.PacketBrokerRoutingPolicyDownlink) {
	up.JoinRequest, _ = flagSet.GetBool("join-request")
	up.MacData, _ = flagSet.GetBool("mac-data-up")
	up.ApplicationData, _ = flagSet.GetBool("application-data-up")
	up.SignalQuality, _ = flagSet.GetBool("signal-quality")
	up.Localization, _ = flagSet.GetBool("localization")
	down.JoinAccept, _ = flagSet.GetBool("join-accept")
	down.MacData, _ = flagSet.GetBool("mac-data-down")
	down.ApplicationData, _ = flagSet.GetBool("application-data-down")

	if v, _ := flagSet.GetBool("join"); v {
		up.JoinRequest = true
		down.JoinAccept = true
	}
	if v, _ := flagSet.GetBool("mac-data"); v {
		up.MacData = true
		down.MacData = true
	}
	if v, _ := flagSet.GetBool("application-data"); v {
		up.ApplicationData = true
		down.ApplicationData = true
	}
	if v, _ := flagSet.GetBool("all"); v {
		up.JoinRequest = true
		up.MacData = true
		up.ApplicationData = true
		up.SignalQuality = true
		up.Localization = true
		down.JoinAccept = true
		down.MacData = true
		down.ApplicationData = true
	}
	return
}

var errPacketBrokerNetworkID = errors.DefineInvalidArgument("packet_broker_network_id", "invalid Packet Broker network ID")

var (
	packetBrokerCommand = &cobra.Command{
		Use:     "packetbroker",
		Aliases: []string{"pb"},
		Short:   "Packet Broker commands",
	}
	packetBrokerInfoCommand = &cobra.Command{
		Use:   "info",
		Short: "Show Packet Broker info",
		RunE: func(cmd *cobra.Command, args []string) error {
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			reg, err := ttnpb.NewPbaClient(pba).GetInfo(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, reg)
		},
	}
	packetBrokerRegisterCommand = &cobra.Command{
		Use:   "register",
		Short: "Register with Packet Broker",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &ttnpb.PacketBrokerRegisterRequest{}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			reg, err := ttnpb.NewPbaClient(pba).Register(ctx, req)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, reg)
		},
	}
	packetBrokerDeregisterCommand = &cobra.Command{
		Use:   "deregister",
		Short: "Deregister from Packet Broker",
		RunE: func(cmd *cobra.Command, args []string) error {
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewPbaClient(pba).Deregister(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}
			return nil
		},
	}
	packetBrokerNetworksCommand = &cobra.Command{
		Use:     "networks",
		Aliases: []string{"network", "nwk"},
		Short:   "Network commands",
	}
	packetBrokerNetworksListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			tenantIDContains, nameContains := getPacketBrokerNetworkSearch(cmd.Flags())
			withRoutingPolicy, _ := cmd.Flags().GetBool("with-routing-policy")
			res, err := ttnpb.NewPbaClient(pba).ListNetworks(ctx, &ttnpb.ListPacketBrokerNetworksRequest{
				Limit:             limit,
				Page:              page,
				TenantIdContains:  tenantIDContains,
				NameContains:      nameContains,
				WithRoutingPolicy: withRoutingPolicy,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Networks)
		},
	}
	packetBrokerHomeNetworksCommand = &cobra.Command{
		Use:     "home-networks",
		Aliases: []string{"home-network", "homenetworks", "homenetwork", "hn"},
		Short:   "Home Network commands",
	}
	packetBrokerHomeNetworksPoliciesCommand = &cobra.Command{
		Use:     "policies",
		Aliases: []string{"policy", "po"},
		Short:   "Manage Home Network routing policies",
	}
	packetBrokerHomeNetworksPolicyListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Home Network routing policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewPbaClient(pba).ListHomeNetworkRoutingPolicies(ctx, &ttnpb.ListHomeNetworkRoutingPoliciesRequest{
				Limit: limit,
				Page:  page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Policies)
		},
	}
	packetBrokerHomeNetworksPolicyGetCommand = &cobra.Command{
		Use:   "get [default|[net-id] [tenant-id]]",
		Short: "Get a Home Network routing policy",
		Example: `
  To get the default routing policy:
    $ ttn-lw-cli packetbroker home-network policies get default

  To get the routing policy with NetID 000013:
    $ ttn-lw-cli packetbroker home-network policies get 000013

  To get the routing policy with NetID 000013 and tenant ttn (The Things Network):
    $ ttn-lw-cli packetbroker home-network policies get 000013 ttn`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getPacketBrokerNetworkID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			var res any
			if id == nil {
				res, err = ttnpb.NewPbaClient(pba).GetHomeNetworkDefaultRoutingPolicy(ctx, ttnpb.Empty)
			} else {
				res, err = ttnpb.NewPbaClient(pba).GetHomeNetworkRoutingPolicy(ctx, id)
			}
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	packetBrokerHomeNetworksPolicySetCommand = &cobra.Command{
		Use:   "set [default|[net-id] [tenant-id]]",
		Short: "Set a Home Network routing policy",
		Long: `Set a Home Network routing policy

Specify default to configure the default routing policy. The default routing
policy is a fallback routing policy when no specific policy has been defined
for the Home Network (by NetID and tenant ID).`,
		Example: `
  To set the default routing policy to pass join, MAC and application data,
  signal quality and gateway locations:
    $ ttn-lw-cli packetbroker home-network policy set default \
      --join --mac-data --application-data --signal-quality --localization

  To set the routing policy with NetID 000013 and tenant ID ttn (The Things
  Network), allowing all data:
    $ ttn-lw-cli packetbroker home-network policy set 000013 ttn --all

  To set the routing policy with NetID C00001 to allow only uplink:
    $ ttn-lw-cli packetbroker home-network policy set C00001 \
      --mac-data-up --application-data-up --signal-quality --localization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getPacketBrokerNetworkID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			uplink, downlink := getPacketBrokerRoutingPolicy(cmd.Flags())

			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			if id == nil {
				_, err = ttnpb.NewPbaClient(pba).SetHomeNetworkDefaultRoutingPolicy(ctx, &ttnpb.SetPacketBrokerDefaultRoutingPolicyRequest{
					Uplink:   &uplink,
					Downlink: &downlink,
				})
			} else {
				_, err = ttnpb.NewPbaClient(pba).SetHomeNetworkRoutingPolicy(ctx, &ttnpb.SetPacketBrokerRoutingPolicyRequest{
					HomeNetworkId: id,
					Uplink:        &uplink,
					Downlink:      &downlink,
				})
			}
			if err != nil {
				return err
			}
			packetBrokerCacheInfo()
			return nil
		},
	}
	packetBrokerHomeNetworksPolicyDeleteCommand = &cobra.Command{
		Use:   "delete [default|[net-id] [tenant-id]]",
		Short: "Delete a Home Network routing policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getPacketBrokerNetworkID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			if id == nil {
				_, err = ttnpb.NewPbaClient(pba).DeleteHomeNetworkDefaultRoutingPolicy(ctx, ttnpb.Empty)
			} else {
				_, err = ttnpb.NewPbaClient(pba).DeleteHomeNetworkRoutingPolicy(ctx, id)
			}
			if err != nil {
				return err
			}
			packetBrokerCacheInfo()
			return nil
		},
	}
	packetBrokerHomeNetworksGatewayVisibilitiesCommand = &cobra.Command{
		Use:     "gateway-visibilities",
		Aliases: []string{"gateway-visibility", "gatewayvisibilities", "gatewayvisibility", "gatewayvis"},
		Short:   "Manage Home Network gateway visibilities",
	}
	packetBrokerHomeNetworksGatewayVisibilityGetCommand = &cobra.Command{
		Use:   "get default",
		Short: "Get a Home Network gateway visibility",
		Example: `
  To get the default gateway visibility:
    $ ttn-lw-cli packetbroker home-network gateway-visibilities get default`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getPacketBrokerNetworkID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			// TODO: Support per-network settings (https://github.com/TheThingsNetwork/lorawan-stack/issues/4409)
			if id != nil {
				return errPacketBrokerNetworkID.New()
			}

			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewPbaClient(pba).GetHomeNetworkDefaultGatewayVisibility(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	packetBrokerHomeNetworksGatewayVisibilitySetCommand = &cobra.Command{
		Use:   "set default",
		Short: "Set a Home Network gateway visibility",
		Long: `Set a Home Network gateway visibility

Specify default to configure the default gateway visibility.`,
		Example: `
  To set the default gateway visibility to show location and online status:
    $ ttn-lw-cli packetbroker home-network gateway-visibilities set default \
      --location --status

  To set the default gateway visibility to show all fields:
    $ ttn-lw-cli packetbroker home-network gateway-visibilities set default \
      --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getPacketBrokerNetworkID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			// TODO: Support per-network settings (https://github.com/TheThingsNetwork/lorawan-stack/issues/4409)
			if id != nil {
				return errPacketBrokerNetworkID.New()
			}

			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			visibility := &ttnpb.PacketBrokerGatewayVisibility{}
			_, err = visibility.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			if all, _ := cmd.Flags().GetBool("all"); all {
				visibility.Location = true
				visibility.AntennaPlacement = true
				visibility.AntennaCount = true
				visibility.FineTimestamps = true
				visibility.ContactInfo = true
				visibility.Status = true
				visibility.FrequencyPlan = true
				visibility.PacketRates = true
			}
			_, err = ttnpb.NewPbaClient(pba).SetHomeNetworkDefaultGatewayVisibility(ctx, &ttnpb.SetPacketBrokerDefaultGatewayVisibilityRequest{
				Visibility: visibility,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, visibility)
		},
	}
	packetBrokerHomeNetworksGatewayVisibilityDeleteCommand = &cobra.Command{
		Use:   "delete default",
		Short: "Delete a Home Network gateway visibility",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getPacketBrokerNetworkID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			// TODO: Support per-network settings (https://github.com/TheThingsNetwork/lorawan-stack/issues/4409)
			if id != nil {
				return errPacketBrokerNetworkID.New()
			}

			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewPbaClient(pba).DeleteHomeNetworkDefaultGatewayVisibility(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}
			return nil
		},
	}
	packetBrokerHomeNetworksListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Home Networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			tenantIDContains, nameContains := getPacketBrokerNetworkSearch(cmd.Flags())
			res, err := ttnpb.NewPbaClient(pba).ListHomeNetworks(ctx, &ttnpb.ListPacketBrokerHomeNetworksRequest{
				Limit:            limit,
				Page:             page,
				TenantIdContains: tenantIDContains,
				NameContains:     nameContains,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Networks)
		},
	}
	packetBrokerForwardersCommand = &cobra.Command{
		Use:     "forwarders",
		Aliases: []string{"forwarder", "fwd"},
		Short:   "Forwarder commands",
	}
	packetBrokerForwardersPoliciesCommand = &cobra.Command{
		Use:     "policies",
		Aliases: []string{"policy", "po"},
		Short:   "Manage Forwarder routing policies",
	}
	packetBrokerForwardersPoliciesListCommand = &cobra.Command{
		Use:   "list",
		Short: "List routing policies configured by Forwarders",
		RunE: func(cmd *cobra.Command, args []string) error {
			pba, err := api.Dial(ctx, config.PacketBrokerAgentGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewPbaClient(pba).ListForwarderRoutingPolicies(ctx, &ttnpb.ListForwarderRoutingPoliciesRequest{
				Limit: limit,
				Page:  page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Policies)
		},
	}
)

func init() {
	packetBrokerCommand.AddCommand(packetBrokerInfoCommand)
	ttnpb.AddSetFlagsForPacketBrokerRegisterRequest(packetBrokerRegisterCommand.Flags(), "", false)
	packetBrokerCommand.AddCommand(packetBrokerRegisterCommand)
	packetBrokerCommand.AddCommand(packetBrokerDeregisterCommand)
	packetBrokerNetworksListCommand.Flags().AddFlagSet(paginationFlags())
	packetBrokerNetworksListCommand.Flags().AddFlagSet(packetBrokerNetworkSearchFlags())
	packetBrokerNetworksCommand.AddCommand(packetBrokerNetworksListCommand)
	packetBrokerCommand.AddCommand(packetBrokerNetworksCommand)
	packetBrokerHomeNetworksPolicyListCommand.Flags().AddFlagSet(paginationFlags())
	packetBrokerHomeNetworksPoliciesCommand.AddCommand(packetBrokerHomeNetworksPolicyListCommand)
	packetBrokerHomeNetworksPolicyGetCommand.Flags().AddFlagSet(packetBrokerNetworkIDFlags(true))
	packetBrokerHomeNetworksPoliciesCommand.AddCommand(packetBrokerHomeNetworksPolicyGetCommand)
	packetBrokerHomeNetworksPolicySetCommand.Flags().AddFlagSet(packetBrokerNetworkIDFlags(true))
	packetBrokerHomeNetworksPolicySetCommand.Flags().AddFlagSet(packetBrokerRoutingPolicyFlags())
	packetBrokerHomeNetworksPoliciesCommand.AddCommand(packetBrokerHomeNetworksPolicySetCommand)
	packetBrokerHomeNetworksPolicyDeleteCommand.Flags().AddFlagSet(packetBrokerNetworkIDFlags(true))
	packetBrokerHomeNetworksPoliciesCommand.AddCommand(packetBrokerHomeNetworksPolicyDeleteCommand)
	packetBrokerHomeNetworksCommand.AddCommand(packetBrokerHomeNetworksPoliciesCommand)
	packetBrokerHomeNetworksGatewayVisibilitiesCommand.AddCommand(packetBrokerHomeNetworksGatewayVisibilityGetCommand)
	ttnpb.AddSetFlagsForPacketBrokerGatewayVisibility(packetBrokerHomeNetworksGatewayVisibilitySetCommand.Flags(), "", false)
	packetBrokerHomeNetworksGatewayVisibilitySetCommand.Flags().Bool("all", false, "")
	packetBrokerHomeNetworksGatewayVisibilitiesCommand.AddCommand(packetBrokerHomeNetworksGatewayVisibilitySetCommand)
	packetBrokerHomeNetworksGatewayVisibilitiesCommand.AddCommand(packetBrokerHomeNetworksGatewayVisibilityDeleteCommand)
	packetBrokerHomeNetworksCommand.AddCommand(packetBrokerHomeNetworksGatewayVisibilitiesCommand)
	packetBrokerHomeNetworksListCommand.Flags().AddFlagSet(paginationFlags())
	packetBrokerHomeNetworksListCommand.Flags().AddFlagSet(packetBrokerNetworkSearchFlags())
	packetBrokerHomeNetworksCommand.AddCommand(packetBrokerHomeNetworksListCommand)
	packetBrokerCommand.AddCommand(packetBrokerHomeNetworksCommand)
	packetBrokerForwardersPoliciesListCommand.Flags().AddFlagSet(paginationFlags())
	packetBrokerForwardersPoliciesCommand.AddCommand(packetBrokerForwardersPoliciesListCommand)
	packetBrokerForwardersCommand.AddCommand(packetBrokerForwardersPoliciesCommand)
	packetBrokerCommand.AddCommand(packetBrokerForwardersCommand)

	Root.AddCommand(packetBrokerCommand)
}
