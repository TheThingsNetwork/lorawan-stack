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
	"strings"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	userRights = &cobra.Command{
		Use:   "rights [user-id]",
		Short: "List the rights to a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserAccessClient(is).ListRights(ctx, usrID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Rights)
		},
	}
	userAPIKeys = &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"api-key"},
		Short:   "Manage user API keys",
	}
	userAPIKeysList = &cobra.Command{
		Use:     "list [user-id]",
		Aliases: []string{"ls"},
		Short:   "List user API keys",
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
			res, err := ttnpb.NewUserAccessClient(is).ListAPIKeys(ctx, &ttnpb.ListUserAPIKeysRequest{
				UserIdentifiers: *usrID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.APIKeys)
		},
	}
	userAPIKeysCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "generate"},
		Short:   "Create a user API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), nil)
			if usrID == nil {
				return errNoUserID
			}
			name, _ := cmd.Flags().GetString("name")

			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				return errNoAPIKeyRights
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
				UserIdentifiers: *usrID,
				Name:            name,
				Rights:          rights,
			})
			if err != nil {
				return err
			}

			logger.Infof("API key ID: %s", res.ID)
			logger.Infof("API key value: %s", res.Key)
			logger.Warn("The API key value will never be shown again")
			logger.Warn("Make sure to copy it to a safe place")

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	userAPIKeysUpdate = &cobra.Command{
		Use:     "update [user-id] [api-key-id]",
		Aliases: []string{"set"},
		Short:   "Update a user API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), firstArgs(1, args...))
			if usrID == nil {
				return errNoUserID
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID
			}
			name, _ := cmd.Flags().GetString("name")

			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				return errNoAPIKeyRights
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
				UserIdentifiers: *usrID,
				APIKey: ttnpb.APIKey{
					ID:     id,
					Name:   name,
					Rights: rights,
				},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	userAPIKeysDelete = &cobra.Command{
		Use:     "delete [user-id] [api-key-id]",
		Aliases: []string{"remove"},
		Short:   "Delete a user API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), firstArgs(1, args...))
			if usrID == nil {
				return errNoUserID
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
				UserIdentifiers: *usrID,
				APIKey: ttnpb.APIKey{
					ID:     id,
					Rights: nil,
				},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
)

var userRightsFlags = rightsFlags(func(flag string) bool {
	for _, entity := range []string{"application", "client", "gateway", "organization", "user"} {
		if strings.HasPrefix(flag, "right-"+entity) {
			return true
		}
	}
	return false
})

func init() {
	userRights.Flags().AddFlagSet(userIDFlags())
	usersCommand.AddCommand(userRights)

	userAPIKeysList.Flags().AddFlagSet(paginationFlags())
	userAPIKeys.AddCommand(userAPIKeysList)
	userAPIKeysCreate.Flags().String("name", "", "")
	userAPIKeysCreate.Flags().AddFlagSet(userRightsFlags)
	userAPIKeys.AddCommand(userAPIKeysCreate)
	userAPIKeysUpdate.Flags().String("api-key-id", "", "")
	userAPIKeysUpdate.Flags().String("name", "", "")
	userAPIKeysUpdate.Flags().AddFlagSet(userRightsFlags)
	userAPIKeys.AddCommand(userAPIKeysUpdate)
	userAPIKeysDelete.Flags().String("api-key-id", "", "")
	userAPIKeys.AddCommand(userAPIKeysDelete)
	userAPIKeys.PersistentFlags().AddFlagSet(userIDFlags())
	usersCommand.AddCommand(userAPIKeys)
}
