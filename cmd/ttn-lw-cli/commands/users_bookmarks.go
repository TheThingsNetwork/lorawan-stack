// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	entityIDFlags = util.NormalizedFlagSet()

	errFlagsUnsupported = errors.DefineInvalidArgument("flags_unsupported", "flags are not supported for this command")
)

func getEntityID(flags *pflag.FlagSet) *ttnpb.EntityIdentifiers {
	if s, err := flags.GetString("entity-ids.application-id"); s != "" && err == nil {
		if devID, err := flags.GetString("entity-ids.device-id"); devID != "" && err == nil {
			return (&ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s},
				DeviceId:       devID,
			}).GetEntityIdentifiers()
		}

		return (&ttnpb.ApplicationIdentifiers{ApplicationId: s}).GetEntityIdentifiers()
	}
	if s, err := flags.GetString("entity-ids.client-id"); s != "" && err == nil {
		return (&ttnpb.ClientIdentifiers{ClientId: s}).GetEntityIdentifiers()
	}
	if s, err := flags.GetString("entity-ids.gateway-id"); s != "" && err == nil {
		return (&ttnpb.GatewayIdentifiers{GatewayId: s}).GetEntityIdentifiers()
	}
	if s, err := flags.GetString("entity-ids.organization-id"); s != "" && err == nil {
		return (&ttnpb.OrganizationIdentifiers{OrganizationId: s}).GetEntityIdentifiers()
	}
	if s, err := flags.GetString("entity-ids.user-id"); s != "" && err == nil {
		return (&ttnpb.UserIdentifiers{UserId: s}).GetEntityIdentifiers()
	}
	return nil
}

var (
	userBookmarks = &cobra.Command{
		Use:     "bookmarks",
		Aliases: []string{"bm", "bookmark"},
		Short:   "Manage user bookmarks",
	}

	userBookmarksCreate = &cobra.Command{
		Use:   "create",
		Short: "Create a new user bookmark",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID.New()
			}

			entityID := getEntityID(cmd.Flags())

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserBookmarkRegistryClient(is).Create(ctx, &ttnpb.CreateUserBookmarkRequest{
				UserIds:   usrID,
				EntityIds: entityID,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}

	userBookmarksList = &cobra.Command{
		Use:   "list",
		Short: "List user's bookmarks",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID.New()
			}
			req := &ttnpb.ListUserBookmarksRequest{UserIds: usrID}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserBookmarkRegistryClient(is).List(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}

	userBookmarksDelete = &cobra.Command{
		Use:   "delete",
		Short: "Delete an user's bookmarks",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID.New()
			}

			entityID := getEntityID(cmd.Flags())

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserBookmarkRegistryClient(is).Delete(ctx, &ttnpb.DeleteUserBookmarkRequest{
				UserIds:   usrID,
				EntityIds: entityID,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}

	userBookmarksBatchDelete = &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete a batch an user's bookmarks. It does not support input from flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputDecoder == nil || len(args) > 0 {
				return errFlagsUnsupported.New()
			}

			req := &ttnpb.BatchDeleteUserBookmarksRequest{}
			err := inputDecoder.Decode(req)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserBookmarkRegistryClient(is).BatchDelete(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
)

func init() {
	entityIDFlags.String("entity-ids.application-id", "", "Application ID")
	entityIDFlags.String("entity-ids.client-id", "", "Client ID")
	entityIDFlags.String("entity-ids.gateway-id", "", "Gateway ID")
	entityIDFlags.String("entity-ids.organization-id", "", "Organization ID")
	entityIDFlags.String("entity-ids.user-id", "", "User ID")
	entityIDFlags.String("entity-ids.device-id", "", "Device ID")
	userBookmarksCreate.Flags().AddFlagSet(userIDFlags())
	userBookmarksCreate.Flags().AddFlagSet(entityIDFlags)
	userBookmarks.AddCommand(userBookmarksCreate)
	ttnpb.AddSetFlagsForListUserBookmarksRequest(userBookmarksList.Flags(), "", false)
	userBookmarksList.Flags().AddFlagSet(userIDFlags())
	userBookmarks.AddCommand(userBookmarksList)
	userBookmarksDelete.Flags().AddFlagSet(entityIDFlags)
	userBookmarksDelete.Flags().AddFlagSet(userIDFlags())
	userBookmarks.AddCommand(userBookmarksDelete)
	userBookmarks.AddCommand(userBookmarksBatchDelete)
	usersCommand.AddCommand(userBookmarks)
}
