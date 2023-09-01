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

	"github.com/TheThingsIndustries/protoc-gen-go-flags/flagsplugin"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	gatewayRights = &cobra.Command{
		Use:   "rights [gateway-id]",
		Short: "List the rights to a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).ListRights(ctx, gtwID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Rights)
		},
	}
	gatewayCollaborators = &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collaborator", "members", "member"},
		Short:   "Manage gateway collaborators",
	}
	gatewayCollaboratorsList = &cobra.Command{
		Use:     "list [gateway-id]",
		Aliases: []string{"ls"},
		Short:   "List gateway collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			order := getOrder(cmd.Flags())
			res, err := ttnpb.NewGatewayAccessClient(is).ListCollaborators(ctx, &ttnpb.ListGatewayCollaboratorsRequest{
				GatewayIds: gtwID, Limit: limit, Page: page, Order: order,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	gatewayCollaboratorsGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get an gateway collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), nil, true)
			if err != nil {
				return err
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).GetCollaborator(ctx, &ttnpb.GetGatewayCollaboratorRequest{
				GatewayIds:   gtwID,
				Collaborator: collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	gatewayCollaboratorsSet = &cobra.Command{
		Use:     "set",
		Aliases: []string{"update"},
		Short:   "Set a gateway collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), nil, true)
			if err != nil {
				return err
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}
			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				return errNoCollaboratorRights.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIds: gtwID,
				Collaborator: &ttnpb.Collaborator{
					Ids:    collaborator,
					Rights: rights,
				},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	gatewayCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a gateway collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), nil, true)
			if err != nil {
				return err
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).DeleteCollaborator(ctx, &ttnpb.DeleteGatewayCollaboratorRequest{
				GatewayIds:      gtwID,
				CollaboratorIds: collaborator,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	gatewayAPIKeys = &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"api-key"},
		Short:   "Manage gateway API keys",
	}
	gatewayAPIKeysList = &cobra.Command{
		Use:     "list [gateway-id]",
		Aliases: []string{"ls"},
		Short:   "List gateway API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &ttnpb.ListGatewayAPIKeysRequest{Limit: 50, Page: 1}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			if len(args) > 0 && req.GetGatewayIds().GetGatewayId() == "" {
				if len(args) > 1 {
					logger.Warn("Multiple IDs found in arguments, considering only the first")
				}
				req.GatewayIds = &ttnpb.GatewayIdentifiers{GatewayId: args[0]}
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, _, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewGatewayAccessClient(is).ListAPIKeys(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.ApiKeys)
		},
	}
	gatewayAPIKeysGet = &cobra.Command{
		Use:     "get [gateway-id] [api-key-id]",
		Aliases: []string{"info"},
		Short:   "Get an gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), firstArgs(1, args...), true)
			if err != nil {
				return err
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
				GatewayIds: gtwID,
				KeyId:      id,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	gatewayAPIKeysCreate = &cobra.Command{
		Use:     "create [gateway-id]",
		Aliases: []string{"add", "register", "generate"},
		Short:   "Create a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			name, _ := cmd.Flags().GetString("name")

			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				return errNoAPIKeyRights.New()
			}

			expiryDate, err := getAPIKeyExpiry(cmd.Flags())
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIds: gtwID,
				Name:       name,
				Rights:     rights,
				ExpiresAt:  ttnpb.ProtoTime(expiryDate),
			})
			if err != nil {
				return err
			}

			logger.Infof("API key ID: %s", res.Id)
			logger.Infof("API key value: %s", res.Key)
			logger.Warn("The API key value will never be shown again")
			logger.Warn("Make sure to copy it to a safe place")

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	gatewayAPIKeysUpdate = &cobra.Command{
		Use:     "set [gateway-id] [api-key-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), firstArgs(1, args...), true)
			if err != nil {
				return err
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
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
			_, err = ttnpb.NewGatewayAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIds: gtwID,
				ApiKey: &ttnpb.APIKey{
					Id:        id,
					Name:      name,
					Rights:    rights,
					ExpiresAt: ttnpb.ProtoTime(expiryDate),
				},
				FieldMask: ttnpb.FieldMask(paths...),
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	gatewayAPIKeysDelete = &cobra.Command{
		Use:     "delete [gateway-id] [api-key-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), firstArgs(1, args...), true)
			if err != nil {
				return err
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIds: gtwID,
				ApiKey: &ttnpb.APIKey{
					Id:     id,
					Rights: nil,
				},
				FieldMask: ttnpb.FieldMask("rights"),
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
)

var gatewayRightsFlags = rightsFlags(func(flag string) bool {
	return strings.HasPrefix(flag, "right-gateway")
})

func init() {
	gatewayRights.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayRights)

	gatewayCollaboratorsList.Flags().AddFlagSet(paginationFlags())
	gatewayCollaboratorsList.Flags().AddFlagSet(orderFlags())
	gatewayCollaborators.AddCommand(gatewayCollaboratorsList)
	gatewayCollaboratorsGet.Flags().AddFlagSet(collaboratorFlags())
	gatewayCollaborators.AddCommand(gatewayCollaboratorsGet)
	gatewayCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	gatewayCollaboratorsSet.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayCollaborators.AddCommand(gatewayCollaboratorsSet)
	gatewayCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	gatewayCollaborators.AddCommand(gatewayCollaboratorsDelete)
	gatewayCollaborators.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayCollaborators)

	ttnpb.AddSetFlagsForListGatewayAPIKeysRequest(gatewayAPIKeysList.Flags(), "", false)
	gatewayAPIKeysList.Flags().Lookup("limit").DefValue = "50"
	gatewayAPIKeysList.Flags().Lookup("page").DefValue = "1"
	flagsplugin.AddAlias(gatewayAPIKeysList.Flags(), "gateway-ids.gateway-id", "gateway-id")
	gatewayAPIKeys.AddCommand(gatewayAPIKeysList)
	gatewayAPIKeysGet.Flags().String("api-key-id", "", "")
	gatewayAPIKeys.AddCommand(gatewayAPIKeysGet)
	gatewayAPIKeysCreate.Flags().String("name", "", "")
	gatewayAPIKeysCreate.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayAPIKeysCreate.Flags().AddFlagSet(apiKeyExpiryFlag)
	gatewayAPIKeys.AddCommand(gatewayAPIKeysCreate)
	gatewayAPIKeysUpdate.Flags().String("api-key-id", "", "")
	gatewayAPIKeysUpdate.Flags().String("name", "", "")
	gatewayAPIKeysUpdate.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayAPIKeysUpdate.Flags().AddFlagSet(apiKeyExpiryFlag)
	gatewayAPIKeys.AddCommand(gatewayAPIKeysUpdate)
	gatewayAPIKeysDelete.Flags().String("api-key-id", "", "")
	gatewayAPIKeys.AddCommand(gatewayAPIKeysDelete)
	gatewayAPIKeys.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayAPIKeys)
}
