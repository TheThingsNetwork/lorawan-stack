// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/migrate"
	ismigrations "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store/migrations"
)

// Migrate migrates the database.
func Migrate(ctx context.Context, db *sql.DB) error {
	bunDB := bun.NewDB(db, pgdialect.New())
	migrator := migrate.NewMigrator(bunDB, ismigrations.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}
	_, err = migrator.Migrate(ctx)
	return err
}
