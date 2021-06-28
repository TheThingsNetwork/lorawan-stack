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

	pbtypes "github.com/gogo/protobuf/types"
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
	selectApplicationFlags = util.FieldMaskFlags(&ttnpb.Application{})
	setApplicationFlags    = util.FieldFlags(&ttnpb.Application{})

	selectAllApplicationFlags = util.SelectAllFlagSet("application")
)

func applicationIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("application-id", "", "")
	return flagSet
}

var (
	errNoApplicationID = errors.DefineInvalidArgument("no_application_id", "no application ID set")
	errNoConfirmation  = errors.DefineInvalidArgument("no_confirmation", "action not confirmed")
)

func getApplicationID(flagSet *pflag.FlagSet, args []string) *ttnpb.ApplicationIdentifiers {
	var applicationID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("Multiple IDs found in arguments, considering only the first")
		}
		applicationID = args[0]
	} else {
		applicationID, _ = flagSet.GetString("application-id")
	}
	if applicationID == "" {
		return nil
	}
	return &ttnpb.ApplicationIdentifiers{ApplicationId: applicationID}
}

var searchApplicationsFlags = func() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(searchFlags)
	// NOTE: These flags need to be named with underscores, not dashes!
	return flagSet
}()

var (
	applicationsCommand = &cobra.Command{
		Use:     "applications",
		Aliases: []string{"application", "apps", "app", "a"},
		Short:   "Application commands",
	}
	applicationsListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ApplicationRegistry/List"].Allowed)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewApplicationRegistryClient(is).List(ctx, &ttnpb.ListApplicationsRequest{
				Collaborator: getCollaborator(cmd.Flags()),
				FieldMask:    &pbtypes.FieldMask{Paths: paths},
				Limit:        limit,
				Page:         page,
				Order:        getOrder(cmd.Flags()),
				Deleted:      getDeleted(cmd.Flags()),
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Applications)
		},
	}
	applicationsSearchCommand = &cobra.Command{
		Use:   "search",
		Short: "Search for applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EntityRegistrySearch/SearchApplications"].Allowed)

			req := &ttnpb.SearchApplicationsRequest{}
			if err := util.SetFields(req, searchApplicationsFlags); err != nil {
				return err
			}
			var (
				opt      grpc.CallOption
				getTotal func() uint64
			)
			req.Limit, req.Page, opt, getTotal = withPagination(cmd.Flags())
			req.FieldMask = &pbtypes.FieldMask{Paths: paths}
			req.Deleted = getDeleted(cmd.Flags())

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEntityRegistrySearchClient(is).SearchApplications(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Applications)
		},
	}
	applicationsGetCommand = &cobra.Command{
		Use:     "get [application-id]",
		Aliases: []string{"info"},
		Short:   "Get an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ApplicationRegistry/Get"].Allowed)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).Get(ctx, &ttnpb.GetApplicationRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              &pbtypes.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsCreateCommand = &cobra.Command{
		Use:     "create [application-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create an application",
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			appID := getApplicationID(cmd.Flags(), args)
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}
			var application ttnpb.Application
			if inputDecoder != nil {
				_, err := inputDecoder.Decode(&application)
				if err != nil {
					return err
				}
			}
			if err := util.SetFields(&application, setApplicationFlags); err != nil {
				return err
			}
			application.Attributes = mergeAttributes(application.Attributes, cmd.Flags())
			if appID != nil && appID.ApplicationId != "" {
				application.ApplicationId = appID.ApplicationId
			}
			if application.ApplicationId == "" {
				return errNoApplicationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).Create(ctx, &ttnpb.CreateApplicationRequest{
				Application:  application,
				Collaborator: *collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		}),
	}
	applicationsSetCommand = &cobra.Command{
		Use:     "set [application-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setApplicationFlags, attributesFlags())
			if len(paths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			var application ttnpb.Application
			if err := util.SetFields(&application, setApplicationFlags); err != nil {
				return err
			}
			application.Attributes = mergeAttributes(application.Attributes, cmd.Flags())
			application.ApplicationIdentifiers = *appID

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: application,
				FieldMask:   &pbtypes.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			res.SetFields(&application, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsDeleteCommand = &cobra.Command{
		Use:     "delete [application-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationRegistryClient(is).Delete(ctx, appID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsRestoreCommand = &cobra.Command{
		Use:   "restore [application-id]",
		Short: "Restore an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationRegistryClient(is).Restore(ctx, appID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsContactInfoCommand = contactInfoCommands("application", func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error) {
		appID := getApplicationID(cmd.Flags(), args)
		if appID == nil {
			return nil, errNoApplicationID
		}
		return appID.GetEntityIdentifiers(), nil
	})
	applicationsPurgeCommand = &cobra.Command{
		Use:     "purge [application-id]",
		Aliases: []string{"permanent-delete", "hard-delete"},
		Short:   "Purge an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			if !confirmChoice(applicationPurgeWarning, force) {
				return errNoConfirmation
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationRegistryClient(is).Purge(ctx, appID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	applicationsIssueNewDevEUICommand = &cobra.Command{
		Use:     "issue-dev-eui [application-id]",
		Aliases: []string{"dev-eui"},
		Short:   "Issue DevEUI for application",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).IssueDevEUI(ctx, appID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
)

func init() {
	applicationsListCommand.Flags().AddFlagSet(collaboratorFlags())
	applicationsListCommand.Flags().AddFlagSet(deletedFlags)
	applicationsListCommand.Flags().AddFlagSet(selectApplicationFlags)
	applicationsListCommand.Flags().AddFlagSet(paginationFlags())
	applicationsListCommand.Flags().AddFlagSet(orderFlags())
	applicationsListCommand.Flags().AddFlagSet(selectAllApplicationFlags)
	applicationsCommand.AddCommand(applicationsListCommand)
	applicationsSearchCommand.Flags().AddFlagSet(searchApplicationsFlags)
	applicationsSearchCommand.Flags().AddFlagSet(deletedFlags)
	applicationsSearchCommand.Flags().AddFlagSet(selectApplicationFlags)
	applicationsSearchCommand.Flags().AddFlagSet(selectAllApplicationFlags)
	applicationsCommand.AddCommand(applicationsSearchCommand)
	applicationsGetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsGetCommand.Flags().AddFlagSet(selectApplicationFlags)
	applicationsGetCommand.Flags().AddFlagSet(selectAllApplicationFlags)
	applicationsCommand.AddCommand(applicationsGetCommand)
	applicationsCreateCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsCreateCommand.Flags().AddFlagSet(collaboratorFlags())
	applicationsCreateCommand.Flags().AddFlagSet(setApplicationFlags)
	applicationsCreateCommand.Flags().AddFlagSet(attributesFlags())
	applicationsCommand.AddCommand(applicationsCreateCommand)
	applicationsSetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsSetCommand.Flags().AddFlagSet(setApplicationFlags)
	applicationsSetCommand.Flags().AddFlagSet(attributesFlags())
	applicationsCommand.AddCommand(applicationsSetCommand)
	applicationsDeleteCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationsDeleteCommand)
	applicationsRestoreCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationsRestoreCommand)
	applicationsContactInfoCommand.PersistentFlags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationsContactInfoCommand)
	applicationsPurgeCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsPurgeCommand.Flags().AddFlagSet(forceFlags())
	applicationsCommand.AddCommand(applicationsPurgeCommand)
	applicationsIssueNewDevEUICommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsCommand.AddCommand(applicationsIssueNewDevEUICommand)
	Root.AddCommand(applicationsCommand)
}
