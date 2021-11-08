// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store/migrations"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	isDBCommand = &cobra.Command{
		Use:   "is-db",
		Short: "Manage the Identity Server database",
	}
	isDBInitCommand = &cobra.Command{
		Use:   "init",
		Short: "Initialize the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")
			db, err := store.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()

			if dbVersion, ok := db.Get("db:version"); ok {
				logger.Infof("Detected database %s", dbVersion)
			}

			logger.Info("Initializing database...")
			if err = store.Initialize(db); err != nil {
				return err
			}

			logger.Info("Creating tables...")
			if err = store.AutoMigrate(db).Error; err != nil {
				return err
			}

			logger.Info("Successfully initialized")
			return nil
		},
	}
	isDBMigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")
			db, err := store.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()

			if dbVersion, ok := db.Get("db:version"); ok {
				logger.Infof("Detected database %s", dbVersion)
			}

			err = store.Transact(ctx, db, func(db *gorm.DB) error {
				logger.Info("Migrating table structure...")
				return store.AutoMigrate(db).Error
			})
			if err != nil {
				return err
			}
			logger.Info("Migrating table contents...")
			err = migrations.Apply(ctx, func(ctx context.Context, f func(db *gorm.DB) error) error {
				return store.Transact(ctx, db, f)
			}, migrations.All...)
			if err != nil {
				return err
			}

			logger.Info("Successfully migrated")
			return nil
		},
	}
	isDBCleanupCommand = &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup expired entities in the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")
			db, err := store.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()
			appStore := store.GetApplicationStore(db)
			userStore := store.GetUserStore(db)
			organizationStore := store.GetOrganizationStore(db)
			gatewayStore := store.GetGatewayStore(db)
			clientStore := store.GetClientStore(db)
			ctx := store.WithSoftDeleted(ctx, true)
			ctx = store.WithExpired(ctx, config.IS.Delete.Restore)
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			// Find expired applications.
			expiredApplications, err := appStore.FindApplications(ctx, []*ttnpb.ApplicationIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
			if err != nil {
				return err
			}
			// Find expired users.
			expiredUsers, err := userStore.FindUsers(ctx, []*ttnpb.UserIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
			if err != nil {
				return err
			}
			// Find expired organizations.
			expiredOrganizations, err := organizationStore.FindOrganizations(ctx, []*ttnpb.OrganizationIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
			if err != nil {
				return err
			}
			// Find expired gateways.
			expiredGateways, err := gatewayStore.FindGateways(ctx, []*ttnpb.GatewayIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
			if err != nil {
				return err
			}
			// Find expired clients.
			expiredClients, err := clientStore.FindClients(ctx, []*ttnpb.ClientIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
			if err != nil {
				return err
			}

			if dryRun {
				logger.Warn("Command is running in dry run mode")
				applicationList := make([]string, len(expiredApplications))
				for i, app := range expiredApplications {
					applicationList[i] = app.GetIds().GetApplicationId()
				}
				logger.Info("Deleting following applications: ", applicationList)
				userList := make([]string, len(expiredUsers))
				for i, usr := range expiredUsers {
					userList[i] = usr.GetIds().GetUserId()
				}
				logger.Info("Deleting following users: ", userList)
				organizationList := make([]string, len(expiredOrganizations))
				for i, org := range expiredOrganizations {
					organizationList[i] = org.GetIds().GetOrganizationId()
				}
				logger.Info("Deleting following organizations: ", organizationList)
				gatewayList := make([]string, len(expiredGateways))
				for i, gtw := range expiredGateways {
					gatewayList[i] = gtw.GetIds().GetGatewayId()
				}
				logger.Info("Deleting following gateways: ", gatewayList)
				clientList := make([]string, len(expiredClients))
				for i, cli := range expiredClients {
					clientList[i] = cli.GetIds().GetClientId()
				}
				logger.Info("Deleting following clients: ", clientList)
				logger.Warn("Dry run finished. No data deleted.")
				return nil
			}

			logger.Info("Purging expired applications")
			for _, ids := range expiredApplications {
				// Delete related API keys before purging the application.
				err = store.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the application.
				err = store.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the application.
				err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = appStore.PurgeApplication(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}

			logger.Info("Purging expired users")
			for _, ids := range expiredUsers {
				err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related API keys before purging the user.
				err = store.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				err = store.GetMembershipStore(db).DeleteAccountMembers(ctx, ids.GetIds().GetOrganizationOrUserIdentifiers())
				if err != nil {
					return err
				}
				err = store.GetOAuthStore(db).DeleteUserAuthorizations(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = store.GetUserSessionStore(db).DeleteAllUserSessions(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = userStore.PurgeUser(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			logger.Info("Purging expired organizations")
			for _, ids := range expiredOrganizations {
				err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related API keys before purging the organization.
				err = store.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				err = store.GetMembershipStore(db).DeleteAccountMembers(ctx, ids.GetIds().GetOrganizationOrUserIdentifiers())
				if err != nil {
					return err
				}
				err = organizationStore.PurgeOrganization(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			logger.Info("Purging expired gateways")
			for _, ids := range expiredGateways {
				// Delete related API keys before purging the gateway.
				err = store.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the gateway.
				err = store.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the gateway.
				err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = gatewayStore.PurgeGateway(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			logger.Info("Purging expired clients")
			for _, ids := range expiredClients {
				// Delete related authorizations before purging the client.
				err = store.GetOAuthStore(db).DeleteClientAuthorizations(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the client.
				err = store.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the client.
				err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = clientStore.PurgeClient(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	Root.AddCommand(isDBCommand)
	isDBCommand.AddCommand(isDBInitCommand)
	isDBCommand.AddCommand(isDBMigrateCommand)
	isDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	isDBCommand.AddCommand(isDBCleanupCommand)
}
