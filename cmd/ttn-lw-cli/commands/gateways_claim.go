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
	"encoding/pem"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

var (
	errInsufficientSourceGatewayRights = errors.DefineInvalidArgument("insufficient_source_gateway_rights", "API Key has insufficient source gateway rights")
	errInvalidTargetCUPSTrust          = errors.DefineInvalidArgument("invalid_target_cups_trust", "invalid target CUPS trust")
)

var (
	gatewayClaimCommand = &cobra.Command{
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
			targetCUPSURI, _ := cmd.Flags().GetString("target-cups-uri")
			targetCUPSTrustLocalFile, _ := cmd.Flags().GetString("target-cups-trust-local-file")
			targetFrequencyPlanId, _ := cmd.Flags().GetString("target-frequency-plan-id")

			var targetCUPSTrust []byte
			if targetCUPSTrustLocalFile != "" {
				raw, err := getDataBytes("target-cups-trust", cmd.Flags())
				if err != nil {
					return err
				}
				block, _ := pem.Decode(raw)
				if block == nil || block.Type != "CERTIFICATE" {
					return errInvalidTargetCUPSTrust.New()
				}
				targetCUPSTrust = block.Bytes
			}
			req := &ttnpb.ClaimGatewayRequest{
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         *gtwIDs.Eui,
						AuthenticationCode: []byte(authenticationCode),
					},
				},
				Collaborator:               collaborator,
				TargetGatewayServerAddress: targetGatewayServerAddress,
				TargetGatewayId:            targetGatewayID,
				TargetFrequencyPlanId:      targetFrequencyPlanId,
			}
			if targetCUPSURI != "" {
				req.CupsRedirection = &ttnpb.CUPSRedirection{
					TargetCupsTrust: targetCUPSTrust,
					TargetCupsUri:   targetCUPSURI,
				}
			}
			ids, err := ttnpb.NewGatewayClaimingServerClient(dcs).Claim(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, ids)
		},
	}
	gatewayClaimAuthorize = &cobra.Command{
		Use:   "authorize [gateway-id]",
		Short: "Authorize an gateway for claiming (EXPERIMENTAL)",
		Long: `Authorize an gateway for claiming (EXPERIMENTAL)

The given API key must have the right to
- read gateway information
- read secrets
- delete the gateway.
If no API key is provided, a new one will be created.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return errNoGatewayID.New()
			}

			expiryDate, err := getAPIKeyExpiry(cmd.Flags())
			if err != nil {
				return err
			}

			requiredRights := []ttnpb.Right{
				ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS,
				ttnpb.Right_RIGHT_GATEWAY_DELETE,
				ttnpb.Right_RIGHT_GATEWAY_INFO,
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}

			key, _ := cmd.Flags().GetString("api-key")
			if key != "" {
				retrievedRights, err := ttnpb.NewGatewayAccessClient(is).ListRights(ctx, gtwID, grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:  "Bearer",
					AuthValue: key,
				}))
				if err != nil {
					return err
				}
				if !retrievedRights.IncludesAll(requiredRights...) {
					return errInsufficientSourceGatewayRights.New()
				}
			} else {
				logger.Info("No API Key provided. Creating one")
				res, err := ttnpb.NewGatewayAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
					GatewayIds: gtwID,
					Name:       "Gateway Claim Authorization Key", // This field can only have 50 chars.
					Rights:     requiredRights,
					ExpiresAt:  ttnpb.ProtoTime(expiryDate),
				})
				if err != nil {
					return err
				}

				logger.Infof("Created API Key with ID: %s", res.Id)
				key = res.Key
			}

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayClaimingServerClient(dcs).AuthorizeGateway(ctx, &ttnpb.AuthorizeGatewayRequest{
				GatewayIds: gtwID,
				ApiKey:     key,
			})
			return err
		},
	}
	gatewayClaimUnauthorize = &cobra.Command{
		Use:   "unauthorize [gateway-id]",
		Short: "Unauthorize an gateway for claiming (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return errNoGatewayID.New()
			}

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}

			logger.Warn("Make sure to delete the API Key used for authorizing claiming as this is not done automatically")

			_, err = ttnpb.NewGatewayClaimingServerClient(dcs).UnauthorizeGateway(ctx, gtwID)
			return err
		},
	}
)

func init() {
	gatewayClaimAuthorize.Flags().String("api-key", "", "")
	gatewayClaimAuthorize.Flags().AddFlagSet(apiKeyExpiryFlag)
	gatewayClaimCommand.AddCommand(gatewayClaimAuthorize)
	gatewayClaimCommand.Flags().AddFlagSet(collaboratorFlags())
	gatewayClaimCommand.AddCommand(gatewayClaimUnauthorize)
	gatewayClaimCommand.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewayClaimCommand.Flags().String("authentication-code", "", "(hex)")
	gatewayClaimCommand.Flags().String("target-cups-uri", "", "")
	gatewayClaimCommand.Flags().String("target-frequency-plan-id", "", "")
	gatewayClaimCommand.Flags().AddFlagSet(dataFlags("target-cups-trust", "(optional) Target CUPS trust in PEM format"))
	gatewayClaimCommand.Flags().String("target-gateway-server-address", "", "")
	gatewayClaimCommand.Flags().String("target-gateway-id", "", "gateway ID for the claimed gateway")
	gatewaysCommand.AddCommand(gatewayClaimCommand)
}
