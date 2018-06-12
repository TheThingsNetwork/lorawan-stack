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

package db

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/db/migrations"
)

const migrationHistorySchema = `
	CREATE TABLE IF NOT EXISTS migration_history (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		"order"     INTEGER NOT NULL,
		name        STRING NOT NULL,
		direction   STRING NOT NULL,
		ran_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
`

func (db *DB) ensureSchema() error {
	_, err := db.Exec(migrationHistorySchema)
	return err
}

func (db *DB) currentState() (int, error) {
	var last struct {
		Order     int
		Direction migrations.Direction
	}
	err := db.SelectOne(
		&last,
		`SELECT "order", direction
			FROM migration_history
			ORDER BY ran_at DESC
			LIMIT 1`)
	// No migration has been applied yet.
	if IsNoRows(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	// If last applied migration was backwards, current state is: order - 1.
	if last.Direction == migrations.DirectionBackwards {
		return last.Order - 1, nil
	}
	return last.Order, nil
}

// logAppliedMigration adds a record in the database about an applied migration.
func (db *DB) logAppliedMigration(q QueryContext, order int, name string, direction migrations.Direction) error {
	_, err := q.Exec(
		`INSERT
			INTO migration_history ("order", name, direction)
			VALUES ($1, $2, $3)`,
		order,
		name,
		direction)
	return err
}

// Migrate migrates the database until reach the target migration.
func (db *DB) Migrate(target int) error {
	err := db.ensureSchema()
	if err != nil {
		return err
	}
	current, err := db.currentState()
	if err != nil {
		return err
	}
	incr := 1
	direction := migrations.DirectionForwards
	if target < current {
		incr = -1
		direction = migrations.DirectionBackwards
	}
	for i := current; i != target; i += incr {
		// If direction is ascendent, current migration to perform is actually: i + incr.
		n := i + incr
		if incr == -1 {
			n = i
		}
		migration, exists := db.migrations.Get(n)
		if !exists {
			return errors.Errorf("Migration with order `%d` does not exist", n)
		}
		next := migration.Forwards
		if incr == -1 {
			next = migration.Backwards
		}
		err := db.Transact(func(tx *Tx) error {
			if _, err := tx.Exec(next); err != nil {
				return err
			}
			return db.logAppliedMigration(tx, migration.Order, migration.Name, direction)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// MigrateAll applies forwards all unapplied migrations.
func (db *DB) MigrateAll() error {
	return db.Migrate(db.migrations.Count())
}
