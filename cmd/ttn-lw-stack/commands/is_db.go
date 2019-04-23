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
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
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

			logger.Info("Migrating tables...")
			if err = store.AutoMigrate(db).Error; err != nil {
				return err
			}

			logger.Info("Successfully migrated")
			return nil
		},
	}
)

func init() {
	Root.AddCommand(isDBCommand)
	isDBCommand.AddCommand(isDBInitCommand)
	isDBCommand.AddCommand(isDBMigrateCommand)
}
