// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	gatewayRights = &cobra.Command{
		Use:   "rights",
		Short: "List the rights to a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).ListRights(ctx, gtwID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res.Rights)
		},
	}
	gatewayCollaborators = &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collaborator", "members", "member"},
		Short:   "Manage gateway collaborators",
	}
	gatewayCollaboratorsList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List gateway collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).ListCollaborators(ctx, gtwID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res.Collaborators)
		},
	}
	gatewayCollaboratorsSet = &cobra.Command{
		Use:   "set",
		Short: "Set a gateway collaborator",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), nil)
			if gtwID == nil {
				return errNoGatewayID
			}
			collaborator := getCollaborator(cmd.Flags())
			if collaborator == nil {
				return errNoCollaborator
			}
			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				logger.Info("No rights selected, will remove collaborator")
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIdentifiers: *gtwID,
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
	gatewayAPIKeys = &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"api-key"},
		Short:   "Manage gateway API keys",
	}
	gatewayAPIKeysList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List gateway API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), args)
			if gtwID == nil {
				return errNoGatewayID
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).ListAPIKeys(ctx, gtwID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res.APIKeys)
		},
	}
	gatewayAPIKeysCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "generate"},
		Short:   "Create a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			gtwID := getGatewayID(cmd.Flags(), nil)
			if gtwID == nil {
				return errNoGatewayID
			}
			name, _ := cmd.Flags().GetString("name")

			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				logger.Info("No rights selected, won't create API key")
				return nil
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewGatewayAccessClient(is).CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIdentifiers: *gtwID,
				Name:               name,
				Rights:             rights,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.Format, res)
		},
	}
	gatewayAPIKeysUpdate = &cobra.Command{
		Use:     "update",
		Aliases: []string{"set"},
		Short:   "Update a gateway API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			id := getAPIKeyID(cmd.Flags(), args)
			if id == "" {
				return errNoAPIKeyID
			}
			gtwID := getGatewayID(cmd.Flags(), nil)
			if gtwID == nil {
				return errNoGatewayID
			}
			name, _ := cmd.Flags().GetString("name")

			rights := getRights(cmd.Flags())
			if len(rights) == 0 {
				logger.Info("No rights selected, will remove API key")
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewGatewayAccessClient(is).UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIdentifiers: *gtwID,
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
)

var gatewayRightsFlags = rightsFlags(func(flag string) bool {
	return strings.HasPrefix(flag, "right-gateway")
})

func init() {
	gatewayRights.Flags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayRights)

	gatewayCollaborators.AddCommand(gatewayCollaboratorsList)
	gatewayCollaboratorsSet.Flags().AddFlagSet(collaboratorFlags())
	gatewayCollaboratorsSet.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayCollaborators.AddCommand(gatewayCollaboratorsSet)
	gatewayCollaborators.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayCollaborators)

	gatewayAPIKeys.AddCommand(gatewayAPIKeysList)
	gatewayAPIKeysCreate.Flags().String("name", "", "")
	gatewayAPIKeysCreate.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayAPIKeys.AddCommand(gatewayAPIKeysCreate)
	gatewayAPIKeysUpdate.Flags().String("api-key-id", "", "")
	gatewayAPIKeysUpdate.Flags().String("name", "", "")
	gatewayAPIKeysUpdate.Flags().AddFlagSet(gatewayRightsFlags)
	gatewayAPIKeys.AddCommand(gatewayAPIKeysUpdate)
	gatewayAPIKeys.PersistentFlags().AddFlagSet(gatewayIDFlags())
	gatewaysCommand.AddCommand(gatewayAPIKeys)
}
