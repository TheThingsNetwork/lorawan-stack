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
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errInsufficientSourceGatewayRights = errors.DefineInvalidArgument("insufficient_source_gateway_rights", "API Key has insufficient source gateway rights")
	errInvalidTargetCUPSTrust          = errors.DefineInvalidArgument("invalid_target_cups_trust", "invalid target CUPS trust")
)

var gatewayClaimCommand = &cobra.Command{
	Use:   "claim [gateway-eui]",
	Short: "Claim a gateway (EXPERIMENTAL)",
	Long: `Claim an gateway (EXPERIMENTAL)
The claiming procedure transfers ownership of gateways using the Device
Claiming Server.

Gateways need to be authorized for claiming before they can be claimed.
See: "ttn-lw-cli claim authorize"

Authentication of gateway claiming is by the Gateway EUI and the claim
authentication code.
For UDP gateways, the claim authentication code is placed in the gateway
entity, stored in the Identity Server.

As part of claiming, you can optionally provide the target Gateway ID and
Gateway Server address and a frequency plan ID.
For LoRa Basic Station gateways, the Target CUPS URI must be specified.
A PEM encoded CUPS trust may be included in the claim request.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		gtwIDs, err := getGatewayEUI(cmd.Flags(), args, true)
		if err != nil {
			return err
		}
		collaborator := getCollaborator(cmd.Flags())
		if collaborator == nil {
			return errNoCollaborator.New()
		}
		dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
		if err != nil {
			return err
		}
		authenticationCode, _ := cmd.Flags().GetString("authentication-code")
		targetGatewayServerAddress, _ := cmd.Flags().GetString("target-gateway-server-address")
		targetGatewayID, _ := cmd.Flags().GetString("target-gateway-id")
		targetFrequencyPlanID, _ := cmd.Flags().GetString("target-frequency-plan-id")
		targetFrequencyPlanIDs, _ := cmd.Flags().GetStringSlice("target-frequency-plan-ids")

		if len(targetFrequencyPlanIDs) == 0 && targetFrequencyPlanID != "" {
			targetFrequencyPlanIDs = []string{targetFrequencyPlanID}
		}
		req := &ttnpb.ClaimGatewayRequest{
			SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
				AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
					GatewayEui:         gtwIDs.Eui,
					AuthenticationCode: []byte(authenticationCode),
				},
			},
			Collaborator:               collaborator,
			TargetGatewayServerAddress: targetGatewayServerAddress,
			TargetGatewayId:            targetGatewayID,
			TargetFrequencyPlanId:      targetFrequencyPlanID,
			TargetFrequencyPlanIds:     targetFrequencyPlanIDs,
		}
		ids, err := ttnpb.NewGatewayClaimingServerClient(dcs).Claim(ctx, req)
		if err != nil {
			return err
		}

		return io.Write(os.Stdout, config.OutputFormat, ids)
	},
}

func init() {
	gatewayClaimCommand.Flags().AddFlagSet(collaboratorFlags())
	gatewayClaimCommand.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewayClaimCommand.Flags().String("authentication-code", "", "(hex)")
	gatewayClaimCommand.Flags().String("target-cups-uri", "", "")
	gatewayClaimCommand.Flags().String("target-frequency-plan-id", "", "")
	gatewayClaimCommand.Flags().String("target-frequency-plan-ids", "", "")
	gatewayClaimCommand.Flags().AddFlagSet(dataFlags("target-cups-trust", "(optional) Target CUPS trust in PEM format"))
	gatewayClaimCommand.Flags().String("target-gateway-server-address", "", "")
	gatewayClaimCommand.Flags().String("target-gateway-id", "", "gateway ID for the claimed gateway")
	gatewaysCommand.AddCommand(gatewayClaimCommand)

	// Deprecate unsupported flags.
	util.DeprecateFlagWithoutForwarding(
		gatewayClaimCommand.Flags(),
		"target-cups-uri",
		"this functionality is no longer supported",
	)
	util.DeprecateFlagWithoutForwarding(
		gatewayClaimCommand.Flags(),
		"target-cups-trust-local-file",
		"this functionality is no longer supported",
	)
}
