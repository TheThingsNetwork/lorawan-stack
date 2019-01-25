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
	"fmt"
	"net/url"
	"strings"

	"github.com/jinzhu/gorm"
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
			dbURI, err := url.Parse(config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			dbName := strings.TrimPrefix(dbURI.Path, "/")

			logger.Info("Connecting to Identity Server database...")
			db, err := gorm.Open("postgres", config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()
			store.SetLogger(db, logger)

			var dbVersion string
			err = db.Raw("SELECT version()").Row().Scan(&dbVersion)
			if err != nil {
				return err
			}
			logger.Infof("Detected database %s", dbVersion)

			logger.Infof("Creating database \"%s\"...", dbName)
			err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName)).Error
			if err != nil {
				return err
			}

			if !strings.Contains(dbVersion, "CockroachDB") {
				logger.Infof("Enabling pgcrypto extension \"%s\"...", dbName)
				err = db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error
				if err != nil {
					return err
				}
			}

			logger.Infof("Creating tables in \"%s\"...", dbName)
			err = store.AutoMigrate(db).Error
			if err != nil {
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
			dbURI, err := url.Parse(config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			dbName := strings.TrimPrefix(dbURI.Path, "/")

			logger.Info("Connecting to Identity Server database...")
			db, err := gorm.Open("postgres", config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()
			store.SetLogger(db, logger)

			logger.Infof("Migrating tables in \"%s\"...", dbName)
			err = store.AutoMigrate(db).Error
			if err != nil {
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
