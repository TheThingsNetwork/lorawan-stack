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
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

var (
	selectApplicationFlags = util.NormalizedFlagSet()

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
			req := &ttnpb.ListApplicationsRequest{}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ApplicationRegistry/List"].Allowed)
			if req.FieldMask == nil {
				req.FieldMask = ttnpb.FieldMask(paths...)
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, _, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewApplicationRegistryClient(is).List(ctx, req, opt)
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
				return errNoApplicationID.New()
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ApplicationRegistry/Get"].Allowed)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).Get(ctx, &ttnpb.GetApplicationRequest{
				ApplicationIds: appID,
				FieldMask:      ttnpb.FieldMask(paths...),
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
			var application ttnpb.Application
			if inputDecoder != nil {
				err := inputDecoder.Decode(&application)
				if err != nil {
					return err
				}
			}
			_, err = application.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			appID := getApplicationID(cmd.Flags(), args)
			if appID.GetApplicationId() != "" {
				application.Ids = appID
			}
			if application.GetIds() == nil {
				return errNoApplicationID.New()
			}
			collaborator := &ttnpb.OrganizationOrUserIdentifiers{}
			_, err = collaborator.SetFromFlags(cmd.Flags(), "collaborator")
			if err != nil {
				return err
			}
			if collaborator.GetIds() == nil {
				return errNoCollaborator.New()
			}

			if application.NetworkServerAddress == "" {
				application.NetworkServerAddress = getHost(config.NetworkServerGRPCAddress)
			}
			if application.ApplicationServerAddress == "" {
				application.ApplicationServerAddress = getHost(config.ApplicationServerGRPCAddress)
			}
			if application.JoinServerAddress == "" {
				application.JoinServerAddress = getHost(config.JoinServerGRPCAddress)
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).Create(ctx, &ttnpb.CreateApplicationRequest{
				Application:  &application,
				Collaborator: collaborator,
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
			app := &ttnpb.Application{}
			paths, err := app.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			rawUnsetPaths, _ := cmd.Flags().GetStringSlice("unset")
			unsetPaths := util.NormalizePaths(rawUnsetPaths)
			if len(paths)+len(unsetPaths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			appID := getApplicationID(cmd.Flags(), args)
			if appID.GetApplicationId() != "" {
				app.Ids = appID
			}
			if app.GetIds() == nil {
				return errNoApplicationID.New()
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationRegistryClient(is).Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: app,
				FieldMask:   ttnpb.FieldMask(append(paths, unsetPaths...)...),
			})
			if err != nil {
				return err
			}
			res.SetFields(app, "ids")
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
				return errNoApplicationID.New()
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
				return errNoApplicationID.New()
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
			return nil, errNoApplicationID.New()
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
				return errNoApplicationID.New()
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			if !confirmChoice(applicationPurgeWarning, force) {
				return errNoConfirmation.New()
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
				return errNoApplicationID.New()
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
	ttnpb.AddSelectFlagsForApplication(selectApplicationFlags, "", false)
	ttnpb.AddSetFlagsForListApplicationsRequest(applicationsListCommand.Flags(), "", false)
	AddCollaboratorFlagAlias(applicationsListCommand.Flags(), "collaborator")
	applicationsListCommand.Flags().AddFlagSet(selectApplicationFlags)
	applicationsListCommand.Flags().AddFlagSet(selectAllApplicationFlags)
	applicationsCommand.AddCommand(applicationsListCommand)
	ttnpb.AddSetFlagsForSearchApplicationsRequest(applicationsSearchCommand.Flags(), "", false)
	applicationsSearchCommand.Flags().AddFlagSet(selectApplicationFlags)
	applicationsSearchCommand.Flags().AddFlagSet(selectAllApplicationFlags)
	applicationsCommand.AddCommand(applicationsSearchCommand)
	applicationsGetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsGetCommand.Flags().AddFlagSet(selectApplicationFlags)
	applicationsGetCommand.Flags().AddFlagSet(selectAllApplicationFlags)
	applicationsCommand.AddCommand(applicationsGetCommand)
	ttnpb.AddSetFlagsForApplication(applicationsCreateCommand.Flags(), "", false)
	flagsplugin.AddAlias(applicationsCreateCommand.Flags(), "ids.application-id", "application-id", flagsplugin.WithHidden(false))
	ttnpb.AddSetFlagsForOrganizationOrUserIdentifiers(applicationsCreateCommand.Flags(), "collaborator", true)
	AddCollaboratorFlagAlias(applicationsCreateCommand.Flags(), "collaborator")
	applicationsCommand.AddCommand(applicationsCreateCommand)
	ttnpb.AddSetFlagsForApplication(applicationsSetCommand.Flags(), "", false)
	flagsplugin.AddAlias(applicationsSetCommand.Flags(), "ids.application-id", "application-id", flagsplugin.WithHidden(false))
	applicationsSetCommand.Flags().AddFlagSet(util.UnsetFlagSet())
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

func compareServerAddressesApplication(application *ttnpb.Application, config *Config) (nsMismatch, asMismatch, jsMismatch bool) {
	nsHost, asHost, jsHost := getHost(config.NetworkServerGRPCAddress), getHost(config.ApplicationServerGRPCAddress), getHost(config.JoinServerGRPCAddress)
	if host := getHost(application.NetworkServerAddress); config.NetworkServerEnabled && host != "" && host != nsHost {
		nsMismatch = true
		logger.WithFields(log.Fields(
			"configured", nsHost,
			"registered", host,
		)).Warnf("Registered Network Server address of Application %q does not match CLI configuration", application.GetIds().GetApplicationId())
	}
	if host := getHost(application.ApplicationServerAddress); config.ApplicationServerEnabled && host != "" && host != asHost {
		asMismatch = true
		logger.WithFields(log.Fields(
			"configured", asHost,
			"registered", host,
		)).Warnf("Registered Application Server address of Application %q does not match CLI configuration", application.GetIds().GetApplicationId())
	}
	if host := getHost(application.JoinServerAddress); config.JoinServerEnabled && host != "" && host != jsHost {
		jsMismatch = true
		logger.WithFields(log.Fields(
			"configured", jsHost,
			"registered", host,
		)).Warnf("Registered Join Server address of Application %q does not match CLI configuration", application.GetIds().GetApplicationId())
	}
	return
}
