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
	organizationRights = &cobra.Command{
		Use:   "rights [organization-id]",
		Short: "List the rights to an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
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
				return errNoOrganizationID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			order := getOrder(cmd.Flags())
			res, err := ttnpb.NewOrganizationAccessClient(is).ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
				OrganizationIds: orgID, Limit: limit, Page: page, Order: order,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	organizationCollaboratorsGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get an organization collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), nil)
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewOrganizationAccessClient(is).GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
				OrganizationIds: orgID,
				Collaborator:    collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	organizationCollaboratorsSet = &cobra.Command{
		Use:     "set",
		Aliases: []string{"update"},
		Short:   "Set an organization collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), nil)
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			collaborator := getUserID(cmd.Flags(), nil).GetOrganizationOrUserIdentifiers()
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
			_, err = ttnpb.NewOrganizationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
				OrganizationIds: orgID,
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
	organizationCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete an organization collaborator",
		RunE: func(cmd *cobra.Command, _ []string) error {
			orgID := getOrganizationID(cmd.Flags(), nil)
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			collaborator := getUserID(cmd.Flags(), nil).GetOrganizationOrUserIdentifiers()
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationAccessClient(is).DeleteCollaborator(
				ctx, &ttnpb.DeleteOrganizationCollaboratorRequest{
					OrganizationIds: orgID,
					CollaboratorIds: collaborator,
				},
			)
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
			req := &ttnpb.ListOrganizationAPIKeysRequest{Limit: 50, Page: 1}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			if len(args) > 0 && req.GetOrganizationIds().GetOrganizationId() == "" {
				if len(args) > 1 {
					logger.Warn("Multiple IDs found in arguments, considering only the first")
				}
				req.OrganizationIds = &ttnpb.OrganizationIdentifiers{OrganizationId: args[0]}
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, _, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewOrganizationAccessClient(is).ListAPIKeys(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.ApiKeys)
		},
	}
	organizationAPIKeysGet = &cobra.Command{
		Use:     "get [organization-id] [api-key-id]",
		Aliases: []string{"info"},
		Short:   "Get an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), firstArgs(1, args...))
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewOrganizationAccessClient(is).GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
				OrganizationIds: orgID,
				KeyId:           id,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	organizationAPIKeysCreate = &cobra.Command{
		Use:     "create [organization-id]",
		Aliases: []string{"add", "generate"},
		Short:   "Create an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
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
			res, err := ttnpb.NewOrganizationAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
				OrganizationIds: orgID,
				Name:            name,
				Rights:          rights,
				ExpiresAt:       ttnpb.ProtoTime(expiryDate),
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
	organizationAPIKeysUpdate = &cobra.Command{
		Use:     "set [organization-id] [api-key-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), firstArgs(1, args...))
			if orgID == nil {
				return errNoOrganizationID.New()
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
			_, err = ttnpb.NewOrganizationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
				OrganizationIds: orgID,
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
	organizationAPIKeysDelete = &cobra.Command{
		Use:     "delete [organization-id] [api-key-id]",
		Aliases: []string{"remove"},
		Short:   "Delete an organization API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), firstArgs(1, args...))
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
				OrganizationIds: orgID,
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
	organizationCollaboratorsList.Flags().AddFlagSet(orderFlags())
	organizationCollaborators.AddCommand(organizationCollaboratorsList)
	organizationCollaboratorsGet.Flags().AddFlagSet(collaboratorFlags())
	organizationCollaborators.AddCommand(organizationCollaboratorsGet)
	organizationCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	organizationCollaboratorsSet.Flags().AddFlagSet(organizationRightsFlags)
	organizationCollaborators.AddCommand(organizationCollaboratorsSet)
	organizationCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	organizationCollaborators.AddCommand(organizationCollaboratorsDelete)
	organizationCollaborators.PersistentFlags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationCollaborators)

	ttnpb.AddSetFlagsForListOrganizationAPIKeysRequest(organizationAPIKeysList.Flags(), "", false)
	organizationAPIKeysList.Flags().Lookup("limit").DefValue = "50"
	organizationAPIKeysList.Flags().Lookup("page").DefValue = "1"
	flagsplugin.AddAlias(organizationAPIKeysList.Flags(), "organization-ids.organization-id", "organization-id")
	organizationAPIKeys.AddCommand(organizationAPIKeysList)
	organizationAPIKeysGet.Flags().String("api-key-id", "", "")
	organizationAPIKeys.AddCommand(organizationAPIKeysGet)
	organizationAPIKeysCreate.Flags().String("name", "", "")
	organizationAPIKeysCreate.Flags().AddFlagSet(organizationRightsFlags)
	organizationAPIKeysCreate.Flags().AddFlagSet(apiKeyExpiryFlag)
	organizationAPIKeys.AddCommand(organizationAPIKeysCreate)
	organizationAPIKeysUpdate.Flags().String("api-key-id", "", "")
	organizationAPIKeysUpdate.Flags().String("name", "", "")
	organizationAPIKeysUpdate.Flags().AddFlagSet(organizationRightsFlags)
	organizationAPIKeysUpdate.Flags().AddFlagSet(apiKeyExpiryFlag)
	organizationAPIKeys.AddCommand(organizationAPIKeysUpdate)
	organizationAPIKeysDelete.Flags().String("api-key-id", "", "")
	organizationAPIKeys.AddCommand(organizationAPIKeysDelete)
	organizationAPIKeys.PersistentFlags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationAPIKeys)
}
