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
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	createOAuthClient = &cobra.Command{
		Use:   "create-oauth-client",
		Short: "Create an OAuth client in the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(log.NewContext(context.Background(), logger), 10*time.Second)
			defer cancel()

			logger.Info("Connecting to Identity Server database...")
			db, err := store.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()

			clientID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			owner, err := cmd.Flags().GetString("owner")
			if err != nil {
				return err
			}
			secret, err := cmd.Flags().GetString("secret")
			if err != nil {
				return err
			}
			if secret == "" {
				noSecret, err := cmd.Flags().GetBool("no-secret")
				if err != nil {
					return err
				}
				if !noSecret {
					secret, err = auth.GenerateKey(ctx)
					if err != nil {
						return err
					}
				}
			}
			var hashedSecret auth.Password
			if secret != "" {
				hashedSecret, err = auth.Hash(secret)
				if err != nil {
					return err
				}
			}
			redirectURIs, err := cmd.Flags().GetStringSlice("redirect-uri")
			if err != nil {
				return err
			}
			authorized, err := cmd.Flags().GetBool("authorized")
			if err != nil {
				return err
			}
			endorsed, err := cmd.Flags().GetBool("endorsed")
			if err != nil {
				return err
			}

			logger.Info("Creating OAuth client...")
			err = store.Transact(ctx, db, func(db *gorm.DB) error {
				cliStore := store.GetClientStore(db)
				cli, err := cliStore.CreateClient(ctx, &ttnpb.Client{
					ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: clientID},
					Name:              name,
					Secret:            string(hashedSecret),
					RedirectURIs:      redirectURIs,
					State:             ttnpb.STATE_APPROVED,
					SkipAuthorization: authorized,
					Endorsed:          endorsed,
					Grants:            []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_REFRESH_TOKEN},
					Rights:            []ttnpb.Right{ttnpb.RIGHT_ALL},
				})
				if err != nil {
					return err
				}
				if owner != "" {
					memberStore := store.GetMembershipStore(db)
					err = memberStore.SetMember(
						ctx,
						ttnpb.UserIdentifiers{UserID: owner}.OrganizationOrUserIdentifiers(),
						cli.ClientIdentifiers.EntityIdentifiers(),
						ttnpb.RightsFrom(ttnpb.RIGHT_CLIENT_ALL),
					)
					if err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				return err
			}

			logger.WithField("secret", secret).Info("Created OAuth client")
			return nil
		},
	}
)

func init() {
	createOAuthClient.Flags().String("id", "console", "OAuth client ID")
	createOAuthClient.Flags().String("name", "", "Name of the OAuth client")
	createOAuthClient.Flags().String("owner", "", "Owner of the OAuth client")
	createOAuthClient.Flags().String("secret", "", "Secret of the OAuth client")
	createOAuthClient.Flags().Bool("no-secret", false, "Do not generate a secret for the OAuth client")
	createOAuthClient.Flags().StringSlice("redirect-uri", []string{}, "Redirect URIs of the OAuth client")
	createOAuthClient.Flags().Bool("authorized", true, "Mark OAuth client as pre-authorized")
	createOAuthClient.Flags().Bool("endorsed", true, "Mark OAuth client as endorsed ")
	isDBCommand.AddCommand(createOAuthClient)
}
