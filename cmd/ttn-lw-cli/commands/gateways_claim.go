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
	"fmt"
	"os"
	"time"

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
)

var (
	gatewayClaimCommand = &cobra.Command{
		Use:   "claim [gateway-eui]",
		Short: "Claim a gateway (EXPERIMENTAL)",
		Long: `Claim an gateway (EXPERIMENTAL)
The claiming procedure transfers ownership of gateways using the Device
Claiming Server.

Authentication of gateway claiming is by the Gateway EUI and the claim
authentication code (which is part of the gateway entity itself) stored
in the Identity Server.

Claim authentication code validity is controlled by the owner of the
gateway by setting the value and optionally a time window when the
code is valid. As part of the claiming, the claim authentication code
is not transferred to the new claimed gateway. This prevents further claim
requests.

As part of claiming, you can optionally provide the target Gateway ID and
Gateway Server Address.

Additionally, for LoRa Basics Station gateways, it is required to provide
the current auth key (Ex: A The Things Stack API Key or an auth token)
used by the gateway and the URI of the Target CUPS server to which the
gateway should connect.
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
			authenticationCode, _ := cmd.Flags().GetBytesHex("authentication-code")
			targetGatewayServerAddress, _ := cmd.Flags().GetString("target-gateway-server-address")
			targetGatewayID, _ := cmd.Flags().GetString("target-gateway-id")
			currentGatewayKey, _ := cmd.Flags().GetString("current-gateway-key")
			targetCUPSURI, _ := cmd.Flags().GetString("target-cups-uri")
			req := &ttnpb.ClaimGatewayRequest{
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEUI:         *gtwIDs.EUI,
						AuthenticationCode: authenticationCode,
					},
				},
				Collaborator:               *collaborator,
				TargetGatewayServerAddress: targetGatewayServerAddress,
				TargetGatewayID:            targetGatewayID,
			}
			if currentGatewayKey != "" || targetCUPSURI != "" {
				req.CUPSRedirection = &ttnpb.CUPSRedirection{
					CurrentGatewayKey: currentGatewayKey,
					TargetCUPSURI:     targetCUPSURI,
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

			requiredRights := []ttnpb.Right{
				ttnpb.RIGHT_GATEWAY_READ_SECRETS,
				ttnpb.RIGHT_GATEWAY_DELETE,
				ttnpb.RIGHT_GATEWAY_INFO,
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
					GatewayIdentifiers: *gtwID,
					Name:               fmt.Sprintf("Gateway Claim Authorization Key, generated by %s at %s", name, time.Now().UTC().Format(time.RFC3339)),
					Rights:             requiredRights,
				})
				if err != nil {
					return err
				}

				logger.Infof("Created API Key with ID: %s", res.ID)
				key = res.Key
			}

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayClaimingServerClient(dcs).AuthorizeGateway(ctx, &ttnpb.AuthorizeGatewayRequest{
				GatewayIdentifiers: *gtwID,
				APIKey:             key,
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
	gatewayClaimCommand.AddCommand(gatewayClaimAuthorize)
	gatewayClaimCommand.Flags().AddFlagSet(collaboratorFlags())
	gatewayClaimCommand.AddCommand(gatewayClaimUnauthorize)
	gatewayClaimCommand.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewayClaimCommand.Flags().BytesHex("authentication-code", nil, "(hex)")
	gatewayClaimCommand.Flags().String("target-cups-uri", "", "")
	gatewayClaimCommand.Flags().String("target-gateway-server-address", "", "")
	gatewayClaimCommand.Flags().String("current-gateway-key", "", "")
	gatewayClaimCommand.Flags().String("target-gateway-id", "", "gateway ID for the claimed gateway")
	gatewaysCommand.AddCommand(gatewayClaimCommand)
}
