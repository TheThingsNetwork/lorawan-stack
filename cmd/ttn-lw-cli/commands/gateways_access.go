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
			res, err := ttnpb.NewGatewayAccessClient(is).ListCollaborators(ctx, &ttnpb.ListGatewayCollaboratorsRequest{
				GatewayIdentifiers: *gtwID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	gatewayCollaboratorsSet = &cobra.Command{
		Use:   "set",
		Short: "Set a gateway collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), nil, true)
			if err != nil {
				return err
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}
			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				return errNoCollaboratorRights
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIdentifiers: *gtwID,
				Collaborator: ttnpb.Collaborator{
					OrganizationOrUserIdentifiers: *collaborator,
					Rights:                        rights,
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
		Aliases: []string{"remove"},
		Short:   "Delete a gateway collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), nil, true)
			if err != nil {
				return err
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIdentifiers: *gtwID,
				Collaborator: ttnpb.Collaborator{
					OrganizationOrUserIdentifiers: *collaborator,
					Rights:                        nil,
				},
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
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewGatewayAccessClient(is).ListAPIKeys(ctx, &ttnpb.ListGatewayAPIKeysRequest{
				GatewayIdentifiers: *gtwID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.APIKeys)
		},
	}
	gatewayAPIKeysCreate = &cobra.Command{
		Use:     "create [gateway-id]",
		Aliases: []string{"add", "generate"},
		Short:   "Create a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), args, true)
			if err != nil {
				return err
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
			res, err := ttnpb.NewGatewayAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIdentifiers: *gtwID,
				Name:               name,
				Rights:             rights,
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
	gatewayAPIKeysUpdate = &cobra.Command{
		Use:     "update [gateway-id] [api-key-id]",
		Aliases: []string{"set"},
		Short:   "Update a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), firstArgs(1, args...), true)
			if err != nil {
				return err
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
			_, err = ttnpb.NewGatewayAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIdentifiers: *gtwID,
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
	gatewayAPIKeysDelete = &cobra.Command{
		Use:     "delete [gateway-id] [api-key-id]",
		Aliases: []string{"remove"},
		Short:   "Delete a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID, err := getGatewayID(cmd.Flags(), firstArgs(1, args...), true)
			if err != nil {
				return err
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIdentifiers: *gtwID,
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

var gatewayRightsFlags = rightsFlags(func(flag string) bool {
	return strings.HasPrefix(flag, "right-gateway")
})

func init() {
	gatewayRights.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayRights)

	gatewayCollaboratorsList.Flags().AddFlagSet(paginationFlags())
	gatewayCollaborators.AddCommand(gatewayCollaboratorsList)
	gatewayCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	gatewayCollaboratorsSet.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayCollaborators.AddCommand(gatewayCollaboratorsSet)
	gatewayCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	gatewayCollaborators.AddCommand(gatewayCollaboratorsDelete)
	gatewayCollaborators.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayCollaborators)

	gatewayAPIKeysList.Flags().AddFlagSet(paginationFlags())
	gatewayAPIKeys.AddCommand(gatewayAPIKeysList)
	gatewayAPIKeysCreate.Flags().String("name", "", "")
	gatewayAPIKeysCreate.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayAPIKeys.AddCommand(gatewayAPIKeysCreate)
	gatewayAPIKeysUpdate.Flags().String("api-key-id", "", "")
	gatewayAPIKeysUpdate.Flags().String("name", "", "")
	gatewayAPIKeysUpdate.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayAPIKeys.AddCommand(gatewayAPIKeysUpdate)
	gatewayAPIKeysDelete.Flags().String("api-key-id", "", "")
	gatewayAPIKeys.AddCommand(gatewayAPIKeysDelete)
	gatewayAPIKeys.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayAPIKeys)
}
