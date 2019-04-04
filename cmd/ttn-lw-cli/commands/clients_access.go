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
	clientRights = &cobra.Command{
		Use:   "rights",
		Short: "List the rights to a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID
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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List client collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), args)
			if cliID == nil {
				return errNoClientID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewClientAccessClient(is).ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
				ClientIdentifiers: *cliID, Limit: limit, Page: page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Collaborators)
		},
	}
	clientCollaboratorsSet = &cobra.Command{
		Use:   "set",
		Short: "Set a client collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), nil)
			if cliID == nil {
				return errNoClientID
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
			_, err = ttnpb.NewClientAccessClient(is).SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIdentifiers: *cliID,
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
	clientCollaboratorsDelete = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"remove"},
		Short:   "Delete a client collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getClientID(cmd.Flags(), nil)
			if cliID == nil {
				return errNoClientID
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewClientAccessClient(is).SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIdentifiers: *cliID,
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
)

var clientRightsFlags = rightsFlags(func(flag string) bool {
	return strings.HasPrefix(flag, "right-client")
})

func init() {
	clientRights.Flags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientRights)

	clientCollaboratorsList.Flags().AddFlagSet(paginationFlags())
	clientCollaborators.AddCommand(clientCollaboratorsList)
	clientCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	clientCollaboratorsSet.Flags().AddFlagSet(clientRightsFlags)
	clientCollaborators.AddCommand(clientCollaboratorsSet)
	clientCollaboratorsDelete.Flags().AddFlagSet(collaboratorFlags())
	clientCollaborators.AddCommand(clientCollaboratorsDelete)
	clientCollaborators.PersistentFlags().AddFlagSet(clientIDFlags())
	clientsCommand.AddCommand(clientCollaborators)
}
