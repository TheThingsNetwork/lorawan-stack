// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errNoValidationReference = errors.DefineInvalidArgument("no_validation_reference", "no validation reference set")
	errNoValidationToken     = errors.DefineInvalidArgument("no_validation_token", "no validation token set")
)

var (
	emailValidations = &cobra.Command{
		Use:     "email-validations",
		Aliases: []string{"ev", "email-validation", "email-validations"},
		Short:   "Email validations commands",
	}

	emailValidationsValidate = &cobra.Command{
		Use:     "validate [reference] [token]",
		Aliases: []string{"v"},
		Short:   "Validate an user's email address",
		RunE: func(cmd *cobra.Command, args []string) error {
			reference, _ := cmd.Flags().GetString("reference")
			token, _ := cmd.Flags().GetString("token")
			switch len(args) {
			case 1:
				reference = args[0]
			case 2:
				reference = args[0]
				token = args[1]
			default:
			}
			if reference == "" {
				return errNoValidationReference.New()
			}
			if token == "" {
				return errNoValidationToken.New()
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEmailValidationRegistryClient(is).Validate(ctx, &ttnpb.ValidateEmailRequest{
				Id:    reference,
				Token: token,
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}

	emailValidationsRequest = &cobra.Command{
		Use:     "request [user-id]",
		Aliases: []string{"r"},
		Short:   "Request validation for an user's email address",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID.New()
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEmailValidationRegistryClient(is).RequestValidation(ctx, usrID)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
)

func init() {
	emailValidationsValidate.Flags().String("reference", "", "Reference of the requested validation")
	emailValidationsValidate.Flags().String("token", "", "Token that you received")
	emailValidationsRequest.PersistentFlags().AddFlagSet(userIDFlags())
	emailValidations.AddCommand(emailValidationsValidate)
	emailValidations.AddCommand(emailValidationsRequest)
	usersCommand.AddCommand(emailValidations)
}
