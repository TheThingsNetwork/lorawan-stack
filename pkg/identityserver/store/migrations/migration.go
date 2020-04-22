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

package migrations

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
)

// Migration abstracts away the logic of a migration.
type Migration interface {
	// Name returns the name of the migration.
	Name() string
	// Apply applies the changes imposed by the migration to the given database (or transaction).
	Apply(context.Context, *gorm.DB) error
	// Rollback rolls back the changes imposed by the migration to the given database (or transaction).
	Rollback(context.Context, *gorm.DB) error
}

func Apply(ctx context.Context, db *gorm.DB, migrations ...Migration) error {
	migrationStore := store.GetMigrationStore(db)
	for _, migration := range migrations {
		_, err := migrationStore.GetMigration(ctx, migration.Name())
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if err == nil {
			continue
		}
		err = migration.Apply(ctx, db)
		if err != nil {
			return err
		}
		err = migrationStore.CreateMigration(ctx, &store.Migration{
			Name: migration.Name(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

var All []Migration
