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
)

var errNoSessionID = errors.DefineInvalidArgument("no_session_id", "no session ID set")

func getUserSessionID(flagSet *pflag.FlagSet, args []string) (*ttnpb.UserSessionIdentifiers, error) {
	userID, _ := flagSet.GetString("user-id")
	sessionID, _ := flagSet.GetString("session-id")
	switch len(args) {
	case 0:
	case 1:
		logger.Warn("Only single ID found in arguments, not considering arguments")
	case 2:
		userID = args[0]
		sessionID = args[1]
	default:
		logger.Warn("Multiple IDs found in arguments, considering the first")
		userID = args[0]
		sessionID = args[1]
	}
	if userID == "" {
		return nil, errNoUserID
	}
	if sessionID == "" {
		return nil, errNoSessionID
	}
	return &ttnpb.UserSessionIdentifiers{
		UserIdentifiers: ttnpb.UserIdentifiers{UserId: userID},
		SessionID:       sessionID,
	}, nil
}

var (
	userSessions = &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "Manage user sessions",
	}
	userSessionsList = &cobra.Command{
		Use:     "list [user-id]",
		Aliases: []string{"ls"},
		Short:   "List user sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewUserSessionRegistryClient(is).List(ctx, &ttnpb.ListUserSessionsRequest{
				UserIdentifiers: *usrID,
				Limit:           limit,
				Page:            page,
				Order:           getOrder(cmd.Flags()),
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Sessions)
		},
	}
	userSessionsDelete = &cobra.Command{
		Use:     "delete [user-id] [session-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a user session",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getUserSessionID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserSessionRegistryClient(is).Delete(ctx, id)
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	userSessionsList.Flags().AddFlagSet(userIDFlags())
	userSessionsList.Flags().AddFlagSet(paginationFlags())
	userSessions.AddCommand(userSessionsList)
	userSessionsDelete.Flags().AddFlagSet(userIDFlags())
	userSessionsDelete.Flags().String("session-id", "", "")
	userSessions.AddCommand(userSessionsDelete)
	usersCommand.AddCommand(userSessions)
}
