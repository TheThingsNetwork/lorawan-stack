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
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/migrate"
	bunstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/bunstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	ismigrations "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store/migrations"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/v3/pkg/types"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
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
	isDBStatusCommand = &cobra.Command{
		Use:   "status",
		Short: "Check the migration status of the Identity Server database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger.WithField("URI", config.IS.DatabaseURI).Info("Connecting to Identity Server database...")

			sqlDB, err := storeutil.OpenDB(cmd.Context(), config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			bunDB := bun.NewDB(sqlDB, pgdialect.New())

			migrator := migrate.NewMigrator(bunDB, ismigrations.Migrations, migrate.WithMarkAppliedOnSuccess(true))

			group, err := migrator.MigrationsWithStatus(ctx)
			if err != nil {
				return err
			}

			log.FromContext(ctx).
				WithField("migrations", group).
				WithField("unapplied_migrations", group.Unapplied()).
				WithField("applied_migrations", group.Applied()).
				Info("Status fetched")

			return nil
		},
	}
	isDBMigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Connecting to Identity Server database...")

			sqlDB, err := storeutil.OpenDB(cmd.Context(), config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			bunDB := bun.NewDB(sqlDB, pgdialect.New())

			migrator := migrate.NewMigrator(bunDB, ismigrations.Migrations, migrate.WithMarkAppliedOnSuccess(true))

			err = migrator.Init(cmd.Context())
			if err != nil {
				return err
			}

			status, err := migrator.MigrationsWithStatus(cmd.Context())
			if err != nil {
				return err
			}
			if migrations := status.Applied(); len(migrations) > 0 {
				logger.Infof("Applied: %s", migrations)
			}
			if migrations := status.Unapplied(); len(migrations) > 0 {
				logger.Infof("Unapplied: %s", migrations)
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
			db, err := storeutil.OpenDB(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			bunDB := bun.NewDB(db, pgdialect.New())
			st, err := bunstore.NewStore(ctx, bunDB)
			if err != nil {
				return err
			}
			defer db.Close()

			expiryDate := time.Now().Add(-1 * config.IS.Delete.Restore)
			ctx := store.WithSoftDeletedBetween(ctx, nil, &expiryDate)
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			// Find expired applications.
			expiredApplications, err := st.FindApplications(
				ctx, []*ttnpb.ApplicationIdentifiers{}, []string{"ids", "deleted_at"},
			)
			if err != nil {
				return err
			}
			// Find expired users.
			expiredUsers, err := st.FindUsers(ctx, []*ttnpb.UserIdentifiers{}, []string{"ids", "deleted_at"})
			if err != nil {
				return err
			}
			// Find expired organizations.
			expiredOrganizations, err := st.FindOrganizations(
				ctx, []*ttnpb.OrganizationIdentifiers{}, []string{"ids", "deleted_at"},
			)
			if err != nil {
				return err
			}
			// Find expired gateways.
			expiredGateways, err := st.FindGateways(ctx, []*ttnpb.GatewayIdentifiers{}, []string{"ids", "deleted_at"})
			if err != nil {
				return err
			}
			// Find expired clients.
			expiredClients, err := st.FindClients(ctx, []*ttnpb.ClientIdentifiers{}, []string{"ids", "deleted_at"})
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
				err = st.DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the application.
				err = st.DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the application.
				err = st.DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = st.PurgeApplication(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}

			logger.Info("Purging expired users")
			for _, ids := range expiredUsers {
				err = st.DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related API keys before purging the user.
				err = st.DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				err = st.DeleteAccountMembers(ctx, ids.GetIds().GetOrganizationOrUserIdentifiers())
				if err != nil {
					return err
				}
				err = st.DeleteUserAuthorizations(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = st.DeleteAllUserSessions(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = st.PurgeUser(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			logger.Info("Purging expired organizations")
			for _, ids := range expiredOrganizations {
				err = st.DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related API keys before purging the organization.
				err = st.DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				err = st.DeleteAccountMembers(ctx, ids.GetIds().GetOrganizationOrUserIdentifiers())
				if err != nil {
					return err
				}
				err = st.PurgeOrganization(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			logger.Info("Purging expired gateways")
			for _, ids := range expiredGateways {
				// Delete related API keys before purging the gateway.
				err = st.DeleteEntityAPIKeys(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the gateway.
				err = st.DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the gateway.
				err = st.DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = st.PurgeGateway(ctx, ids.GetIds())
				if err != nil {
					return err
				}
			}
			logger.Info("Purging expired clients")
			for _, ids := range expiredClients {
				// Delete related authorizations before purging the client.
				err = st.DeleteClientAuthorizations(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				// Delete related memberships before purging the client.
				err = st.DeleteEntityMembers(ctx, ids.GetIds().GetEntityIdentifiers())
				if err != nil {
					return err
				}
				// Delete related contact info before purging the client.
				err = st.DeleteEntityContactInfo(ctx, ids.GetIds())
				if err != nil {
					return err
				}
				err = st.PurgeClient(ctx, ids.GetIds())
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

			db, err := storeutil.OpenDB(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			bunDB := bun.NewDB(db, pgdialect.New())
			st, err := bunstore.NewStore(ctx, bunDB)
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
				return st.CreateEUIBlock(ctx, config.IS.DevEUIBlock.Prefix, config.IS.DevEUIBlock.InitCounter, "dev_eui")
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
				if err := st.CreateEUIBlock(ctx, *euiPrefix, counter, euiType); err != nil {
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
	isDBCommand.AddCommand(isDBStatusCommand)
	isDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	isDBCommand.AddCommand(isDBCleanupCommand)
	isDBEUIBlockCreationCommand.Flags().Bool("use-config", false, "Create block using values from config")
	isDBEUIBlockCreationCommand.Flags().String("eui-type", "dev_eui", "EUI block type")
	isDBEUIBlockCreationCommand.Flags().String("prefix", "", "Block prefix (format: 1234567800000000/32)")
	isDBEUIBlockCreationCommand.Flags().Int64("init-counter", 0, "Initial counter (determines first address to be issued from block)")
	isDBCommand.AddCommand(isDBEUIBlockCreationCommand)
}
