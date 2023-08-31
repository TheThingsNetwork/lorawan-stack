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
	applicationRights = &cobra.Command{
		Use:   "rights [application-id]",
		Short: "List the rights to an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
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
				return errNoApplicationID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			order := getOrder(cmd.Flags())
			res, err := ttnpb.NewApplicationAccessClient(is).ListCollaborators(
				ctx, &ttnpb.ListApplicationCollaboratorsRequest{
					ApplicationIds: appID, Limit: limit, Page: page, Order: order,
				}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	applicationCollaboratorsGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get an application collaborator",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID.New()
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationAccessClient(is).GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
				ApplicationIds: appID,
				Collaborator:   collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationCollaboratorsSet = &cobra.Command{
		Use:     "set",
		Aliases: []string{"update"},
		Short:   "Set properties of an application collaborator",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID.New()
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
			_, err = ttnpb.NewApplicationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIds: appID,
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
	applicationCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete an application collaborator",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID.New()
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationAccessClient(is).SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIds: appID,
				Collaborator: &ttnpb.Collaborator{
					Ids:    collaborator,
					Rights: nil,
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
			req := &ttnpb.ListApplicationAPIKeysRequest{Limit: 50, Page: 1}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			if len(args) > 0 && req.GetApplicationIds().GetApplicationId() == "" {
				if len(args) > 1 {
					logger.Warn("Multiple IDs found in arguments, considering only the first")
				}
				req.ApplicationIds = &ttnpb.ApplicationIdentifiers{ApplicationId: args[0]}
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, _, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewApplicationAccessClient(is).ListAPIKeys(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.ApiKeys)
		},
	}
	applicationAPIKeysGet = &cobra.Command{
		Use:     "get [application-id] [api-key-id]",
		Aliases: []string{"info"},
		Short:   "Get an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), firstArgs(1, args...))
			if appID == nil {
				return errNoApplicationID.New()
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationAccessClient(is).GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
				ApplicationIds: appID,
				KeyId:          id,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationAPIKeysCreate = &cobra.Command{
		Use:     "create [application-id]",
		Aliases: []string{"add", "register", "generate"},
		Short:   "Create an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
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
			res, err := ttnpb.NewApplicationAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
				ApplicationIds: appID,
				Name:           name,
				Rights:         rights,
				ExpiresAt:      ttnpb.ProtoTime(expiryDate),
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
	applicationAPIKeysUpdate = &cobra.Command{
		Use:     "set [application-id] [api-key-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), firstArgs(1, args...))
			if appID == nil {
				return errNoApplicationID.New()
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
			_, err = ttnpb.NewApplicationAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
				ApplicationIds: appID,
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
	applicationAPIKeysDelete = &cobra.Command{
		Use:     "delete [application-id] [api-key-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete an application API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), firstArgs(1, args...))
			if appID == nil {
				return errNoApplicationID.New()
			}
			id := getAPIKeyID(cmd.Flags(), args, 1)
			if id == "" {
				return errNoAPIKeyID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationAccessClient(is).DeleteAPIKey(ctx, &ttnpb.DeleteApplicationAPIKeyRequest{
				ApplicationIds: appID,
				KeyId:          id,
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
	applicationCollaboratorsList.Flags().AddFlagSet(orderFlags())
	applicationCollaborators.AddCommand(applicationCollaboratorsList)
	applicationCollaboratorsGet.Flags().AddFlagSet(collaboratorFlags())
	applicationCollaborators.AddCommand(applicationCollaboratorsGet)
	applicationCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	applicationCollaboratorsSet.Flags().AddFlagSet(applicationRightsFlags)
	applicationCollaborators.AddCommand(applicationCollaboratorsSet)
	applicationCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	applicationCollaborators.AddCommand(applicationCollaboratorsDelete)
	applicationCollaborators.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationCollaborators)

	ttnpb.AddSetFlagsForListApplicationAPIKeysRequest(applicationAPIKeysList.Flags(), "", false)
	applicationAPIKeysList.Flags().Lookup("limit").DefValue = "50"
	applicationAPIKeysList.Flags().Lookup("page").DefValue = "1"
	flagsplugin.AddAlias(applicationAPIKeysList.Flags(), "application-ids.application-id", "application-id")
	applicationAPIKeys.AddCommand(applicationAPIKeysList)
	applicationAPIKeysGet.Flags().String("api-key-id", "", "")
	applicationAPIKeys.AddCommand(applicationAPIKeysGet)
	applicationAPIKeysCreate.Flags().String("name", "", "")
	applicationAPIKeysCreate.Flags().AddFlagSet(applicationRightsFlags)
	applicationAPIKeysCreate.Flags().AddFlagSet(apiKeyExpiryFlag)
	applicationAPIKeys.AddCommand(applicationAPIKeysCreate)
	applicationAPIKeysUpdate.Flags().String("api-key-id", "", "")
	applicationAPIKeysUpdate.Flags().String("name", "", "")
	applicationAPIKeysUpdate.Flags().AddFlagSet(applicationRightsFlags)
	applicationAPIKeysUpdate.Flags().AddFlagSet(apiKeyExpiryFlag)
	applicationAPIKeys.AddCommand(applicationAPIKeysUpdate)
	applicationAPIKeysDelete.Flags().String("api-key-id", "", "")
	applicationAPIKeys.AddCommand(applicationAPIKeysDelete)
	applicationAPIKeys.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationAPIKeys)
}
