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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errNoEmail = errors.DefineInvalidArgument("no_email", "no email set")

func getEmail(flagSet *pflag.FlagSet, args []string) string {
	var email string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("multiple emails found in arguments, considering only the first")
		}
		email = args[0]
	} else {
		email, _ = flagSet.GetString("email")
	}
	return email
}

var (
	userInvitations = &cobra.Command{
		Use:     "invitations",
		Aliases: []string{"invitation"},
		Short:   "Manage user invitations",
	}
	userInvitationsList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List user invitations",
		RunE: func(cmd *cobra.Command, args []string) error {
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewUserInvitationRegistryClient(is).List(ctx, &ttnpb.ListInvitationsRequest{
				Limit: limit,
				Page:  page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Invitations)
		},
	}
	userInvitationsCreate = &cobra.Command{
		Use:   "create [email]",
		Short: "Create a user invitation",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := getEmail(cmd.Flags(), args)
			if email == "" {
				return errNoEmail
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserInvitationRegistryClient(is).Send(ctx, &ttnpb.SendInvitationRequest{
				Email: email,
			})
			if err != nil {
				return err
			}
			return nil
		},
	}
	userInvitationsDelete = &cobra.Command{
		Use:   "delete [email]",
		Short: "Delete a user invitation",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := getEmail(cmd.Flags(), args)
			if email == "" {
				return errNoEmail
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserInvitationRegistryClient(is).Delete(ctx, &ttnpb.DeleteInvitationRequest{
				Email: email,
			})
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	userInvitationsList.Flags().String("email", "", "")
	userInvitations.AddCommand(userInvitationsList)
	userInvitations.AddCommand(userInvitationsCreate)
	userInvitationsDelete.Flags().String("email", "", "")
	userInvitations.AddCommand(userInvitationsDelete)
	usersCommand.AddCommand(userInvitations)
}
