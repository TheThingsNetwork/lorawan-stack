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
	selectClientFlags = util.NormalizedFlagSet()

	selectAllClientFlags = util.SelectAllFlagSet("client")
)

func clientIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("client-id", "", "")
	return flagSet
}

var errNoClientID = errors.DefineInvalidArgument("no_client_id", "no client ID set")

func getClientID(flagSet *pflag.FlagSet, args []string) *ttnpb.ClientIdentifiers {
	var clientID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("Multiple IDs found in arguments, considering only the first")
		}
		clientID = args[0]
	} else {
		clientID, _ = flagSet.GetString("client-id")
	}
	if clientID == "" {
		return nil
	}
	return &ttnpb.ClientIdentifiers{ClientId: clientID}
}

var (
	clientsCommand = &cobra.Command{
		Use:     "clients",
		Aliases: []string{"client", "cli", "c"},
		Short:   "Client commands",
	}
	clientsListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectClientFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ClientRegistry/List"].Allowed)
			req := &ttnpb.ListClientsRequest{}
			_, err := req.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			if req.GetFieldMask() == nil {
				req.FieldMask = ttnpb.FieldMask(paths...)
			}
			_, _, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewClientRegistryClient(is).List(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Clients)
		},
	}
	clientsSearchCommand = &cobra.Command{
		Use:   "search",
		Short: "Search for clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectClientFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EntityRegistrySearch/SearchClients"].Allowed)

			req := &ttnpb.SearchClientsRequest{}
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
			res, err := ttnpb.NewEntityRegistrySearchClient(is).SearchClients(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Clients)
		},
	}
	clientsGetCommand = &cobra.Command{
		Use:     "get [client-id]",
		Aliases: []string{"info"},
		Short:   "Get a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID.New()
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectClientFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ClientRegistry/Get"].Allowed)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewClientRegistryClient(is).Get(ctx, &ttnpb.GetClientRequest{
				ClientIds: cliID,
				FieldMask: ttnpb.FieldMask(paths...),
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	clientsCreateCommand = &cobra.Command{
		Use:     "create [client-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create a client",
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			cliID := getClientID(cmd.Flags(), args)
			collaborator := &ttnpb.OrganizationOrUserIdentifiers{}
			_, err = collaborator.SetFromFlags(cmd.Flags(), "collaborator")
			if err != nil {
				return err
			}
			if collaborator.GetIds() == nil {
				return errNoCollaborator.New()
			}
			var client ttnpb.Client
			client.State = ttnpb.State_STATE_APPROVED // This may not be honored by the server.
			client.Grants = []ttnpb.GrantType{
				ttnpb.GrantType_GRANT_AUTHORIZATION_CODE,
			}
			if inputDecoder != nil {
				err := inputDecoder.Decode(&client)
				if err != nil {
					return err
				}
			}
			_, err = client.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			if cliID.GetClientId() != "" {
				client.Ids = &ttnpb.ClientIdentifiers{ClientId: cliID.GetClientId()}
			}
			if client.GetIds().GetClientId() == "" {
				return errNoClientID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewClientRegistryClient(is).Create(ctx, &ttnpb.CreateClientRequest{
				Client:       &client,
				Collaborator: collaborator,
			})
			if err != nil {
				return err
			}

			logger.Infof("Client secret: %s", res.Secret)
			logger.Warn("The Client secret will never be shown again")
			logger.Warn("Make sure to copy it to a safe place")

			return io.Write(os.Stdout, config.OutputFormat, res)
		}),
	}
	clientsSetCommand = &cobra.Command{
		Use:     "set [client-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := &ttnpb.Client{}
			paths, err := client.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			if client.GetIds() == nil {
				cliID := getClientID(cmd.Flags(), args)
				if cliID == nil {
					return errNoClientID.New()
				}
			}
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
			res, err := ttnpb.NewClientRegistryClient(is).Update(ctx, &ttnpb.UpdateClientRequest{
				Client:    client,
				FieldMask: ttnpb.FieldMask(append(paths, unsetPaths...)...),
			})
			if err != nil {
				return err
			}

			res.SetFields(client, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	clientsDeleteCommand = &cobra.Command{
		Use:     "delete [client-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewClientRegistryClient(is).Delete(ctx, cliID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	clientsRestoreCommand = &cobra.Command{
		Use:   "restore [client-id]",
		Short: "Restore a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewClientRegistryClient(is).Restore(ctx, cliID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	clientsPurgeCommand = &cobra.Command{
		Use:     "purge [client-id]",
		Aliases: []string{"permanent-delete", "hard-delete"},
		Short:   "Purge an client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID.New()
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			if !confirmChoice(clientPurgeWarning, force) {
				return errNoConfirmation.New()
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewClientRegistryClient(is).Purge(ctx, cliID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	clientsContactInfoCommand = contactInfoCommands("client", func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error) {
		cliID := getClientID(cmd.Flags(), args)
		if cliID == nil {
			return nil, errNoClientID.New()
		}
		return cliID.GetEntityIdentifiers(), nil
	})
)

func init() {
	ttnpb.AddSelectFlagsForClient(selectClientFlags, "", false)
	ttnpb.AddSetFlagsForListClientsRequest(clientsListCommand.Flags(), "", false)
	AddCollaboratorFlagAlias(clientsListCommand.Flags(), "collaborator")
	clientsListCommand.Flags().AddFlagSet(selectClientFlags)
	clientsListCommand.Flags().AddFlagSet(selectAllClientFlags)
	clientsListCommand.Flags().AddFlagSet(paginationFlags())
	clientsListCommand.Flags().AddFlagSet(orderFlags())
	clientsCommand.AddCommand(clientsListCommand)
	ttnpb.AddSetFlagsForSearchClientsRequest(clientsSearchCommand.Flags(), "", false)
	clientsSearchCommand.Flags().AddFlagSet(selectClientFlags)
	clientsSearchCommand.Flags().AddFlagSet(selectAllClientFlags)
	clientsCommand.AddCommand(clientsSearchCommand)
	clientsGetCommand.Flags().AddFlagSet(clientIDFlags())
	clientsGetCommand.Flags().AddFlagSet(selectClientFlags)
	clientsGetCommand.Flags().AddFlagSet(selectAllClientFlags)
	clientsCommand.AddCommand(clientsGetCommand)
	ttnpb.AddSetFlagsForClient(clientsCreateCommand.Flags(), "", false)
	ttnpb.AddSetFlagsForOrganizationOrUserIdentifiers(clientsCreateCommand.Flags(), "collaborator", true)
	AddCollaboratorFlagAlias(clientsCreateCommand.Flags(), "collaborator")
	flagsplugin.AddAlias(clientsCreateCommand.Flags(), "ids.client-id", "client-id", flagsplugin.WithHidden(false))
	clientsCreateCommand.Flags().Lookup("state").DefValue = ttnpb.State_STATE_APPROVED.String()
	clientsCreateCommand.Flags().Lookup("grants").DefValue = ttnpb.GrantType_GRANT_AUTHORIZATION_CODE.String()
	clientsCommand.AddCommand(clientsCreateCommand)
	ttnpb.AddSetFlagsForClient(clientsSetCommand.Flags(), "", false)
	flagsplugin.AddAlias(clientsSetCommand.Flags(), "ids.client-id", "client-id", flagsplugin.WithHidden(false))
	clientsSetCommand.Flags().AddFlagSet(util.UnsetFlagSet())
	clientsCommand.AddCommand(clientsSetCommand)
	clientsDeleteCommand.Flags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientsDeleteCommand)
	clientsRestoreCommand.Flags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientsRestoreCommand)
	clientsContactInfoCommand.PersistentFlags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientsContactInfoCommand)
	clientsPurgeCommand.Flags().AddFlagSet(clientIDFlags())
	clientsPurgeCommand.Flags().AddFlagSet(forceFlags())
	clientsCommand.AddCommand(clientsPurgeCommand)
	Root.AddCommand(clientsCommand)
}
