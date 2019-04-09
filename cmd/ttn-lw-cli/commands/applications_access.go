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
	applicationRights = &cobra.Command{
		Use:   "rights [application-id]",
		Short: "List the rights to an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationAccessClient(is).ListRights(ctx, appID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Rights)
		},
	}
	applicationCollaborators = &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collaborator", "members", "member"},
		Short:   "Manage application collaborators",
	}
	applicationCollaboratorsList = &cobra.Command{
		Use:     "list [application-id]",
		Aliases: []string{"ls"},
		Short:   "List application collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewApplicationAccessClient(is).ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
				ApplicationIdentifiers: *appID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	applicationCollaboratorsSet = &cobra.Command{
		Use:   "set",
		Short: "Set an application collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID
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
			_, err = ttnpb.NewApplicationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIdentifiers: *appID,
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
	applicationCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"remove"},
		Short:   "Delete an application collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIdentifiers: *appID,
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
	applicationAPIKeys = &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"api-key"},
		Short:   "Manage application API keys",
	}
	applicationAPIKeysList = &cobra.Command{
		Use:     "list [application-id]",
		Aliases: []string{"ls"},
		Short:   "List application API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewApplicationAccessClient(is).ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
				ApplicationIdentifiers: *appID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.APIKeys)
		},
	}
	applicationAPIKeysCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "generate"},
		Short:   "Create an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID
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
			res, err := ttnpb.NewApplicationAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
				ApplicationIdentifiers: *appID,
				Name:                   name,
				Rights:                 rights,
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
	applicationAPIKeysUpdate = &cobra.Command{
		Use:     "update [application-id] [api-key-id]",
		Aliases: []string{"set"},
		Short:   "Update an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), firstArgs(1, args...))
			if appID == nil {
				return errNoApplicationID
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
			_, err = ttnpb.NewApplicationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
				ApplicationIdentifiers: *appID,
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
	applicationAPIKeysDelete = &cobra.Command{
		Use:     "delete [application-id] [api-key-id]",
		Aliases: []string{"remove"},
		Short:   "Delete an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), firstArgs(1, args...))
			if appID == nil {
				return errNoApplicationID
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
				ApplicationIdentifiers: *appID,
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

var applicationRightsFlags = rightsFlags(func(flag string) bool {
	return strings.HasPrefix(flag, "right-application")
})

func init() {
	applicationRights.Flags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationRights)

	applicationCollaboratorsList.Flags().AddFlagSet(paginationFlags())
	applicationCollaborators.AddCommand(applicationCollaboratorsList)
	applicationCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	applicationCollaboratorsSet.Flags().AddFlagSet(applicationRightsFlags)
	applicationCollaborators.AddCommand(applicationCollaboratorsSet)
	applicationCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	applicationCollaborators.AddCommand(applicationCollaboratorsDelete)
	applicationCollaborators.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationCollaborators)

	applicationAPIKeysList.Flags().AddFlagSet(paginationFlags())
	applicationAPIKeys.AddCommand(applicationAPIKeysList)
	applicationAPIKeysCreate.Flags().String("name", "", "")
	applicationAPIKeysCreate.Flags().AddFlagSet(applicationRightsFlags)
	applicationAPIKeys.AddCommand(applicationAPIKeysCreate)
	applicationAPIKeysUpdate.Flags().String("api-key-id", "", "")
	applicationAPIKeysUpdate.Flags().String("name", "", "")
	applicationAPIKeysUpdate.Flags().AddFlagSet(applicationRightsFlags)
	applicationAPIKeys.AddCommand(applicationAPIKeysUpdate)
	applicationAPIKeysDelete.Flags().String("api-key-id", "", "")
	applicationAPIKeys.AddCommand(applicationAPIKeysDelete)
	applicationAPIKeys.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationAPIKeys)
}
