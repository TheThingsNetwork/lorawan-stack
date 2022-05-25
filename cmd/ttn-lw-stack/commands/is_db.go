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
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	gormstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	ismigrations "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store/migrations"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/v3/pkg/types"
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
			return fmt.Errorf("init command deprecated, use migrate instead")
		},
	}
	isDBMigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")

			sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(config.IS.DatabaseURI)))
			defer sqlDB.Close()

			bunDB := bun.NewDB(sqlDB, pgdialect.New())

			migrator := migrate.NewMigrator(bunDB, ismigrations.Migrations)

			err := migrator.Init(cmd.Context())
			if err != nil {
				return err
			}

			status, err := migrator.MigrationsWithStatus(cmd.Context())
			if err != nil {
				return err
			}
			if migrations := status.Applied(); len(migrations) > 0 {
				logger.Infof("Applied: %s", status)
			}
			if migrations := status.Unapplied(); len(migrations) > 0 {
				logger.Infof("Unapplied: %s", status)
			}

			var group *migrate.MigrationGroup

			rollback, _ := cmd.Flags().GetBool("rollback")

			if rollback {
				group, err = migrator.Rollback(cmd.Context())
			} else {
				group, err = migrator.Migrate(cmd.Context())
			}
			if err != nil {
				return err
			}

			if group.IsZero() {
				logger.Info("Database is up to date")
				return nil
			}

			if rollback {
				logger.WithField("group", group.ID).Info("Database rollback done")
			} else {
				logger.WithField("group", group.ID).Info("Database migration done")
			}

			status, err = migrator.MigrationsWithStatus(cmd.Context())
			if err != nil {
				return err
			}
			if migrations := status.Applied(); len(migrations) > 0 {
				logger.Infof("Applied: %s", status)
			}
			if migrations := status.Unapplied(); len(migrations) > 0 {
				logger.Infof("Unapplied: %s", status)
			}

			return nil
		},
	}
	isDBCleanupCommand = &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup expired entities in the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")
			db, err := gormstore.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()
			appStore := gormstore.GetApplicationStore(db)
			userStore := gormstore.GetUserStore(db)
			organizationStore := gormstore.GetOrganizationStore(db)
			gatewayStore := gormstore.GetGatewayStore(db)
			clientStore := gormstore.GetClientStore(db)
			expiryDate := time.Now().Add(-1 * config.IS.Delete.Restore)
			ctx := store.WithSoftDeletedBetween(ctx, nil, &expiryDate)
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			// Find expired applications.
			expiredApplications, err := appStore.FindApplications(ctx, []*ttnpb.ApplicationIdentifiers{}, []string{"ids", "deleted_at"})
			if err != nil {
				return err
			}
			// Find expired users.
			expiredUsers, err := userStore.FindUsers(ctx, []*ttnpb.UserIdentifiers{}, []string{"ids", "deleted_at"})
			if err != nil {
				return err
			}
			// Find expired organizations.
			expiredOrganizations, err := organizationStore.FindOrganizations(ctx, []*ttnpb.OrganizationIdentifiers{}, []string{"ids", "deleted_at"})
			if err != nil {
				return err
			}
			// Find expired gateways.
			expiredGateways, err := gatewayStore.FindGateways(ctx, []*ttnpb.GatewayIdentifiers{}, []string{"ids", "deleted_at"})
			if err != nil {
				return err
			}
			// Find expired clients.
			expiredClients, err := clientStore.FindClients(ctx, []*ttnpb.ClientIdentifiers{}, []string{"ids", "deleted_at"})
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
				err = gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the application.
				err = gormstore.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the application.
				err = gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
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
				err = gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related API keys before purging the user.
				err = gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				err = gormstore.GetMembershipStore(db).DeleteAccountMembers(ctx, ids.GetIds().GetOrganizationOrUserIdentifiers())
				if err != nil {
					return err
				}
				err = gormstore.GetOAuthStore(db).DeleteUserAuthorizations(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = gormstore.GetUserSessionStore(db).DeleteAllUserSessions(ctx, ids.GetIds())
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
				err = gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related API keys before purging the organization.
				err = gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				err = gormstore.GetMembershipStore(db).DeleteAccountMembers(ctx, ids.GetIds().GetOrganizationOrUserIdentifiers())
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
				err = gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the gateway.
				err = gormstore.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the gateway.
				err = gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
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
				err = gormstore.GetOAuthStore(db).DeleteClientAuthorizations(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the client.
				err = gormstore.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the client.
				err = gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids.GetIds())
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
	isDBEUIBlockCreationCommand = &cobra.Command{
		Use:   "create-eui-block",
		Short: "Create an EUI block in IS db (currently only DevEUI block supported)",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")
			db, err := gormstore.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()
			useConfig, err := cmd.Flags().GetBool("use-config")
			if err != nil {
				return err
			}
			if useConfig {
				logger.Info("Using config values...")
				return gormstore.GetEUIStore(db).CreateEUIBlock(ctx, config.IS.DevEUIBlock.Prefix, config.IS.DevEUIBlock.InitCounter, "dev_eui")
			}
			prefix, err := cmd.Flags().GetString("prefix")
			if err != nil {
				return err
			}
			euiPrefix := &ttntypes.EUI64Prefix{}
			if err := euiPrefix.UnmarshalConfigString(prefix); err != nil {
				return err
			}
			counter, err := cmd.Flags().GetInt64("init-counter")
			if err != nil {
				return err
			}
			euiType, err := cmd.Flags().GetString("eui-type")
			if err != nil {
				return err
			}
			switch euiType {
			case "dev_eui":
				if err := gormstore.GetEUIStore(db).CreateEUIBlock(ctx, *euiPrefix, counter, euiType); err != nil {
					return err
				}
				logger.Info("Block created successfully")
			default:
				logger.Error("Unsupported eui type")
			}
			return nil
		},
	}
)

func init() {
	Root.AddCommand(isDBCommand)
	isDBCommand.AddCommand(isDBInitCommand)
	isDBMigrateCommand.Flags().Bool("rollback", false, "Rollback most recent migration group")
	isDBCommand.AddCommand(isDBMigrateCommand)
	isDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	isDBCommand.AddCommand(isDBCleanupCommand)
	isDBEUIBlockCreationCommand.Flags().Bool("use-config", false, "Create block using values from config")
	isDBEUIBlockCreationCommand.Flags().String("eui-type", "dev_eui", "EUI block type")
	isDBEUIBlockCreationCommand.Flags().String("prefix", "", "Block prefix (format: 1234567800000000/32)")
	isDBEUIBlockCreationCommand.Flags().Int64("init-counter", 0, "Initial counter (determines first address to be issued from block)")
	isDBCommand.AddCommand(isDBEUIBlockCreationCommand)
}
