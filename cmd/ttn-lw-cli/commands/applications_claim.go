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
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	applicationClaim = &cobra.Command{
		Use:   "claim",
		Short: "Manage claim settings in applications",
	}
	applicationClaimAuthorize = &cobra.Command{
		Use:   "authorize [application-id]",
		Short: "Authorize an application for claiming (EXPERIMENTAL)",
		Long: `Authorize an application for claiming (EXPERIMENTAL)

The given API key must have devices and device keys read/write rights. If no API
key is provided, a new API key will be created.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
			}

			expiryDate, err := getAPIKeyExpiry(cmd.Flags())
			if err != nil {
				return err
			}

			key, _ := cmd.Flags().GetString("api-key")
			if key == "" {
				is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
				if err != nil {
					return err
				}
				logger.Info("Creating API key")
				apiKey, err := ttnpb.NewApplicationAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
					ApplicationIds: appID,
					Name:           "Device Claiming",
					Rights: []ttnpb.Right{
						ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
						ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
					},
					ExpiresAt: ttnpb.ProtoTime(expiryDate),
				})
				if err != nil {
					return err
				}
				key = apiKey.Key
			}

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewEndDeviceClaimingServerClient(dcs).AuthorizeApplication(ctx, &ttnpb.AuthorizeApplicationRequest{
				ApplicationIds: appID,
				ApiKey:         key,
			})
			return err
		},
	}
	applicationClaimUnauthorize = &cobra.Command{
		Use:   "unauthorize [application-id]",
		Short: "Unauthorize an application for claiming (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
			}

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}

			logger.Warn("Make sure to delete the API Key used for authorizing claiming as this is not done automatically")

			_, err = ttnpb.NewEndDeviceClaimingServerClient(dcs).UnauthorizeApplication(ctx, appID)
			return err
		},
	}
)

func init() {
	applicationClaimAuthorize.Flags().String("api-key", "", "")
	applicationClaimAuthorize.Flags().String("api-key-expiry", "", "API key expiry date (YYYY-MM-DD:HH:mm) - only applicable when creating API Key")
	applicationClaim.AddCommand(applicationClaimAuthorize)
	applicationClaim.AddCommand(applicationClaimUnauthorize)
	applicationClaim.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationClaim)
}
