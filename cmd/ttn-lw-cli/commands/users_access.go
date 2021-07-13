// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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
	userAPIKeysGet = &cobra.Command{
		Use:     "get [user-id] [api-key-id]",
		Aliases: []string{"info"},
		Short:   "Get an user API key",
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
			res, err := ttnpb.NewUserAccessClient(is).GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
				UserIdentifiers: *usrID,
				KeyId:           id,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	userAPIKeysCreate = &cobra.Command{
		Use:     "create [user-id]",
		Aliases: []string{"add", "generate", "register"},
		Short:   "Create a user API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}
			name, _ := cmd.Flags().GetString("name")

			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				return errNoAPIKeyRights
			}

			expiryDate, err := getAPIKeyExpiry(cmd.Flags())
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
				UserIdentifiers: *usrID,
				Name:            name,
				Rights:          rights,
				ExpiresAt:       expiryDate,
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
		Use:     "set [user-id] [api-key-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of a user API key",
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

			rights, expiryDate, paths, err := getAPIKeyFields(cmd.Flags())
			if err != nil {
				return err
			}
			if len(paths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
				UserIdentifiers: *usrID,
				APIKey: ttnpb.APIKey{
					ID:        id,
					Name:      name,
					Rights:    rights,
					ExpiresAt: expiryDate,
				},
				FieldMask: &pbtypes.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	userAPIKeysDelete = &cobra.Command{
		Use:     "delete [user-id] [api-key-id]",
		Aliases: []string{"del", "remove", "rm"},
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
				FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	usersCreateLoginToken = &cobra.Command{
		Use:               "create-login-token [user-id]",
		Short:             "Create a user login token",
		PersistentPreRunE: preRun(optionalAuth),
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserAccessClient(is).CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
				UserIdentifiers: *usrID,
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
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
	userAPIKeysGet.Flags().String("api-key-id", "", "")
	userAPIKeys.AddCommand(userAPIKeysGet)
	userAPIKeysCreate.Flags().String("name", "", "")
	userAPIKeysCreate.Flags().AddFlagSet(userRightsFlags)
	userAPIKeysCreate.Flags().AddFlagSet(apiKeyExpiryFlag)
	userAPIKeys.AddCommand(userAPIKeysCreate)
	userAPIKeysUpdate.Flags().String("api-key-id", "", "")
	userAPIKeysUpdate.Flags().String("name", "", "")
	userAPIKeysUpdate.Flags().AddFlagSet(userRightsFlags)
	userAPIKeysUpdate.Flags().AddFlagSet(apiKeyExpiryFlag)
	userAPIKeys.AddCommand(userAPIKeysUpdate)
	userAPIKeysDelete.Flags().String("api-key-id", "", "")
	userAPIKeys.AddCommand(userAPIKeysDelete)
	userAPIKeys.PersistentFlags().AddFlagSet(userIDFlags())
	usersCommand.AddCommand(userAPIKeys)
	usersCreateLoginToken.Flags().AddFlagSet(userIDFlags())
	usersCommand.AddCommand(usersCreateLoginToken)
}
