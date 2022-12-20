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

	"github.com/TheThingsIndustries/protoc-gen-go-flags/flagsplugin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

var (
	selectOrganizationFlags = util.NormalizedFlagSet()

	selectAllOrganizationFlags = util.SelectAllFlagSet("organization")
)

func organizationIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("organization-id", "", "")
	return flagSet
}

var errNoOrganizationID = errors.DefineInvalidArgument("no_organization_id", "no organization ID set")

func getOrganizationID(flagSet *pflag.FlagSet, args []string) *ttnpb.OrganizationIdentifiers {
	var organizationID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("Multiple IDs found in arguments, considering only the first")
		}
		organizationID = args[0]
	} else {
		organizationID, _ = flagSet.GetString("organization-id")
	}
	if organizationID == "" {
		return nil
	}
	return &ttnpb.OrganizationIdentifiers{OrganizationId: organizationID}
}

var (
	organizationsCommand = &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"organization", "org", "o"},
		Short:   "Organization commands",
	}
	organizationsListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &ttnpb.ListOrganizationsRequest{}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectOrganizationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.OrganizationRegistry/List"].Allowed)
			if req.FieldMask == nil {
				req.FieldMask = ttnpb.FieldMask(paths...)
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			userID := getUserID(cmd.Flags(), nil)
			if userID == nil {
				return errNoCollaborator.New()
			}
			req.Collaborator = userID.GetOrganizationOrUserIdentifiers()
			_, _, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewOrganizationRegistryClient(is).List(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Organizations)
		},
	}
	organizationsSearchCommand = &cobra.Command{
		Use:   "search",
		Short: "Search for organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectOrganizationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EntityRegistrySearch/SearchOrganizations"].Allowed)

			req := &ttnpb.SearchOrganizationsRequest{}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			var (
				opt      grpc.CallOption
				getTotal func() uint64
			)
			_, _, opt, getTotal = withPagination(cmd.Flags())
			if req.FieldMask == nil {
				req.FieldMask = ttnpb.FieldMask(paths...)
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEntityRegistrySearchClient(is).SearchOrganizations(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Organizations)
		},
	}
	organizationsGetCommand = &cobra.Command{
		Use:     "get [organization-id]",
		Aliases: []string{"info"},
		Short:   "Get an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectOrganizationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.OrganizationRegistry/Get"].Allowed)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewOrganizationRegistryClient(is).Get(ctx, &ttnpb.GetOrganizationRequest{
				OrganizationIds: orgID,
				FieldMask:       ttnpb.FieldMask(paths...),
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	organizationsCreateCommand = &cobra.Command{
		Use:     "create [organization-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create an organization",
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			collaborator := getUserID(cmd.Flags(), nil).GetOrganizationOrUserIdentifiers()
			if collaborator == nil {
				return errNoCollaborator.New()
			}
			var organization ttnpb.Organization
			if inputDecoder != nil {
				err := inputDecoder.Decode(&organization)
				if err != nil {
					return err
				}
			}
			_, err = organization.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID.GetOrganizationId() != "" {
				organization.Ids = &ttnpb.OrganizationIdentifiers{OrganizationId: orgID.GetOrganizationId()}
			}
			if organization.GetIds().GetOrganizationId() == "" {
				return errNoOrganizationID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewOrganizationRegistryClient(is).Create(ctx, &ttnpb.CreateOrganizationRequest{
				Organization: &organization,
				Collaborator: collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		}),
	}
	organizationsSetCommand = &cobra.Command{
		Use:     "set [organization-id]",
		Aliases: []string{"set"},
		Short:   "Set properties of an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			var organization ttnpb.Organization
			paths, err := organization.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			organization.Ids = orgID
			rawUnsetPaths, _ := cmd.Flags().GetStringSlice("unset")
			unsetPaths := util.NormalizePaths(rawUnsetPaths)
			if len(paths)+len(unsetPaths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewOrganizationRegistryClient(is).Update(ctx, &ttnpb.UpdateOrganizationRequest{
				Organization: &organization,
				FieldMask:    ttnpb.FieldMask(append(paths, unsetPaths...)...),
			})
			if err != nil {
				return err
			}

			res.SetFields(&organization, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	organizationsDeleteCommand = &cobra.Command{
		Use:     "delete [organization-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationRegistryClient(is).Delete(ctx, orgID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	organizationsRestoreCommand = &cobra.Command{
		Use:   "restore [organization-id]",
		Short: "Restore an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationRegistryClient(is).Restore(ctx, orgID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	organizationsPurgeCommand = &cobra.Command{
		Use:     "purge [organization-id]",
		Aliases: []string{"permanent-delete", "hard-delete"},
		Short:   "Purge an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID := getOrganizationID(cmd.Flags(), args)
			if orgID == nil {
				return errNoOrganizationID.New()
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			if !confirmChoice(organizationPurgeWarning, force) {
				return errNoConfirmation.New()
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewOrganizationRegistryClient(is).Purge(ctx, orgID)
			if err != nil {
				return err
			}

			return nil
		},
	}

	organizationsContactInfoCommand = contactInfoCommands("organization", func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error) {
		orgID := getOrganizationID(cmd.Flags(), args)
		if orgID == nil {
			return nil, errNoOrganizationID.New()
		}
		return orgID.GetEntityIdentifiers(), nil
	})
)

func init() {
	ttnpb.AddSelectFlagsForOrganization(selectOrganizationFlags, "", false)
	ttnpb.AddSetFlagsForListOrganizationsRequest(organizationsListCommand.Flags(), "", false)
	AddCollaboratorFlagAlias(organizationsListCommand.Flags(), "collaborator")
	organizationsListCommand.Flags().AddFlagSet(selectOrganizationFlags)
	organizationsListCommand.Flags().AddFlagSet(selectAllOrganizationFlags)
	organizationsListCommand.Flags().AddFlagSet(paginationFlags())
	organizationsListCommand.Flags().AddFlagSet(orderFlags())
	organizationsCommand.AddCommand(organizationsListCommand)
	ttnpb.AddSetFlagsForSearchOrganizationsRequest(organizationsSearchCommand.Flags(), "", false)
	organizationsSearchCommand.Flags().AddFlagSet(selectOrganizationFlags)
	organizationsSearchCommand.Flags().AddFlagSet(selectAllOrganizationFlags)
	organizationsCommand.AddCommand(organizationsSearchCommand)
	organizationsGetCommand.Flags().AddFlagSet(organizationIDFlags())
	organizationsGetCommand.Flags().AddFlagSet(selectOrganizationFlags)
	organizationsGetCommand.Flags().AddFlagSet(selectAllOrganizationFlags)
	organizationsCommand.AddCommand(organizationsGetCommand)
	ttnpb.AddSetFlagsForOrganization(organizationsCreateCommand.Flags(), "", false)
	flagsplugin.AddAlias(organizationsCreateCommand.Flags(), "ids.organization-id", "organization-id", flagsplugin.WithHidden(false))
	organizationsCreateCommand.Flags().AddFlagSet(userIDFlags())
	organizationsCommand.AddCommand(organizationsCreateCommand)
	ttnpb.AddSetFlagsForOrganization(organizationsSetCommand.Flags(), "", false)
	flagsplugin.AddAlias(organizationsSetCommand.Flags(), "ids.organization-id", "organization-id", flagsplugin.WithHidden(false))
	organizationsSetCommand.Flags().AddFlagSet(util.UnsetFlagSet())
	organizationsCommand.AddCommand(organizationsSetCommand)
	organizationsDeleteCommand.Flags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationsDeleteCommand)
	organizationsRestoreCommand.Flags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationsRestoreCommand)
	organizationsContactInfoCommand.PersistentFlags().AddFlagSet(organizationIDFlags())
	organizationsCommand.AddCommand(organizationsContactInfoCommand)
	organizationsPurgeCommand.Flags().AddFlagSet(organizationIDFlags())
	organizationsPurgeCommand.Flags().AddFlagSet(forceFlags())
	organizationsCommand.AddCommand(organizationsPurgeCommand)
	Root.AddCommand(organizationsCommand)
}
