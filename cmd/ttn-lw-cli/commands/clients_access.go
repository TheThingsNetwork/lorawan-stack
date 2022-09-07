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

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	clientRights = &cobra.Command{
		Use:   "rights [client-id]",
		Short: "List the rights to a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewClientAccessClient(is).ListRights(ctx, cliID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.Rights)
		},
	}
	clientCollaborators = &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collaborator", "members", "member"},
		Short:   "Manage client collaborators",
	}
	clientCollaboratorsList = &cobra.Command{
		Use:     "list [client-id]",
		Aliases: []string{"ls"},
		Short:   "List client collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			order := getOrder(cmd.Flags())
			res, err := ttnpb.NewClientAccessClient(is).ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
				ClientIds: cliID, Limit: limit, Page: page, Order: order,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	clientCollaboratorsGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get an client collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), nil)
			if cliID == nil {
				return errNoClientID.New()
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewClientAccessClient(is).GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
				ClientIds:    cliID,
				Collaborator: collaborator,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	clientCollaboratorsSet = &cobra.Command{
		Use:     "set",
		Aliases: []string{"update"},
		Short:   "Set a client collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), nil)
			if cliID == nil {
				return errNoClientID.New()
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
			_, err = ttnpb.NewClientAccessClient(is).SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIds: cliID,
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
	clientCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a client collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), nil)
			if cliID == nil {
				return errNoClientID.New()
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewClientAccessClient(is).SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIds: cliID,
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
)

var clientRightsFlags = rightsFlags(func(flag string) bool {
	return strings.HasPrefix(flag, "right-client")
})

func init() {
	clientRights.Flags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientRights)

	clientCollaboratorsList.Flags().AddFlagSet(paginationFlags())
	clientCollaboratorsList.Flags().AddFlagSet(orderFlags())
	clientCollaborators.AddCommand(clientCollaboratorsList)
	clientCollaboratorsGet.Flags().AddFlagSet(collaboratorFlags())
	clientCollaborators.AddCommand(clientCollaboratorsGet)
	clientCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	clientCollaboratorsSet.Flags().AddFlagSet(clientRightsFlags)
	clientCollaborators.AddCommand(clientCollaboratorsSet)
	clientCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	clientCollaborators.AddCommand(clientCollaboratorsDelete)
	clientCollaborators.PersistentFlags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientCollaborators)
}
