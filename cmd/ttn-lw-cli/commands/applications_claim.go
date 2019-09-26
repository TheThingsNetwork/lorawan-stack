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
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
				return errNoApplicationID
			}

			key, _ := cmd.Flags().GetString("api-key")
			if key == "" {
				logger.Info("Creating API key")
				apiKey, err := createApplicationAPIKey(ctx, *appID, "Device Claiming",
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				)
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
				ApplicationIdentifiers: *appID,
				APIKey:                 key,
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
				return errNoApplicationID
			}

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewEndDeviceClaimingServerClient(dcs).UnauthorizeApplication(ctx, appID)
			return err
		},
	}
)

func init() {
	applicationClaimAuthorize.Flags().String("api-key", "", "")
	applicationClaim.AddCommand(applicationClaimAuthorize)
	applicationClaim.AddCommand(applicationClaimUnauthorize)
	applicationClaim.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationClaim)
}
