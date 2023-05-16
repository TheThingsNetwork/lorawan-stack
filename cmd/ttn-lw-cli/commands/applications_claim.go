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
	"fmt"

	"github.com/spf13/cobra"
)

var (
	applicationClaim = &cobra.Command{
		Use:    "claim",
		Short:  "Manage claim settings in applications (DEPRECATED)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(
				"this command is no longer supported. End device claiming is integrated into the device creation flow",
			)
		},
	}
	applicationClaimAuthorize = &cobra.Command{
		Use:    "authorize [application-id]",
		Short:  "Authorize an application for claiming (DEPRECATED)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(
				"this command is no longer supported. End device claiming is integrated into the device creation flow",
			)
		},
	}
	applicationClaimUnauthorize = &cobra.Command{
		Use:    "unauthorize [application-id]",
		Short:  "Unauthorize an application for claiming (DEPRECATED)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(
				"this command is no longer supported. End device claiming is integrated into the device creation flow",
			)
		},
	}
)

func init() {
	applicationClaimAuthorize.Flags().String("api-key", "", "")
	applicationClaimAuthorize.Flags().String("api-key-expiry", "", "API key expiry date (YYYY-MM-DD:HH:mm) - only applicable when creating API Key") //nolint:lll
	applicationClaim.AddCommand(applicationClaimAuthorize)
	applicationClaim.AddCommand(applicationClaimUnauthorize)
	applicationClaim.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationClaim)
}
