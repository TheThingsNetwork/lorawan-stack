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
	organizationRights = &cobra.Command{
		Use:   "rights [organization-id]",
		Short: "List the rights to an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewOrganizationAccessClient(is).ListRights(ctx, orgID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Rights)
		},
	}
	organizationCollaborators = &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collaborator", "members", "member"},
		Short:   "Manage organization collaborators",
	}
	organizationCollaboratorsList = &cobra.Command{
		Use:     "list [organization-id]",
		Aliases: []string{"ls"},
		Short:   "List organization collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewOrganizationAccessClient(is).ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
				OrganizationIdentifiers: *orgID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	organizationCollaboratorsSet = &cobra.Command{
		Use:   "set",
		Short: "Set an organization collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), nil)
			if orgID == nil {
				return errNoOrganizationID
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
			_, err = ttnpb.NewOrganizationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
				OrganizationIdentifiers: *orgID,
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
	organizationCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"remove"},
		Short:   "Delete an organization collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), nil)
			if orgID == nil {
				return errNoOrganizationID
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
				OrganizationIdentifiers: *orgID,
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
	organizationAPIKeys = &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"api-key"},
		Short:   "Manage organization API keys",
	}
	organizationAPIKeysList = &cobra.Command{
		Use:     "list [organization-id]",
		Aliases: []string{"ls"},
		Short:   "List organization API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewOrganizationAccessClient(is).ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
				OrganizationIdentifiers: *orgID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.APIKeys)
		},
	}
	organizationAPIKeysCreate = &cobra.Command{
		Use:     "create [organization-id]",
		Aliases: []string{"add", "generate"},
		Short:   "Create an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID
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
			res, err := ttnpb.NewOrganizationAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
				OrganizationIdentifiers: *orgID,
				Name:                    name,
				Rights:                  rights,
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
	organizationAPIKeysUpdate = &cobra.Command{
		Use:     "update [organization-id] [api-key-id]",
		Aliases: []string{"set"},
		Short:   "Update an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), firstArgs(1, args...))
			if orgID == nil {
				return errNoOrganizationID
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
			_, err = ttnpb.NewOrganizationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
				OrganizationIdentifiers: *orgID,
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
	organizationAPIKeysDelete = &cobra.Command{
		Use:     "delete [organization-id] [api-key-id]",
		Aliases: []string{"remove"},
		Short:   "Delete an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), firstArgs(1, args...))
			if orgID == nil {
				return errNoOrganizationID
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
				OrganizationIdentifiers: *orgID,
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

var organizationRightsFlags = rightsFlags(func(flag string) bool {
	for _, entity := range []string{"application", "client", "gateway", "organization"} {
		if strings.HasPrefix(flag, "right-"+entity) {
			return true
		}
	}
	return false
})

func init() {
	organizationRights.Flags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationRights)

	organizationCollaboratorsList.Flags().AddFlagSet(paginationFlags())
	organizationCollaborators.AddCommand(organizationCollaboratorsList)
	organizationCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	organizationCollaboratorsSet.Flags().AddFlagSet(organizationRightsFlags)
	organizationCollaborators.AddCommand(organizationCollaboratorsSet)
	organizationCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	organizationCollaborators.AddCommand(organizationCollaboratorsDelete)
	organizationCollaborators.PersistentFlags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationCollaborators)

	organizationAPIKeysList.Flags().AddFlagSet(paginationFlags())
	organizationAPIKeys.AddCommand(organizationAPIKeysList)
	organizationAPIKeysCreate.Flags().String("name", "", "")
	organizationAPIKeysCreate.Flags().AddFlagSet(organizationRightsFlags)
	organizationAPIKeys.AddCommand(organizationAPIKeysCreate)
	organizationAPIKeysUpdate.Flags().String("api-key-id", "", "")
	organizationAPIKeysUpdate.Flags().String("name", "", "")
	organizationAPIKeysUpdate.Flags().AddFlagSet(organizationRightsFlags)
	organizationAPIKeys.AddCommand(organizationAPIKeysUpdate)
	organizationAPIKeysDelete.Flags().String("api-key-id", "", "")
	organizationAPIKeys.AddCommand(organizationAPIKeysDelete)
	organizationAPIKeys.PersistentFlags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationAPIKeys)
}
