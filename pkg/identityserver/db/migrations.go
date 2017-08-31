// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db/migrations"
)

var errMigrationsRegistryNotSet = errors.New("identityserver/server/db: no migrations registry has been attached to the database instance")

func errMigrationDoesNotExist(order int) error {
	return fmt.Errorf("identityserver/server/db: migration with order `%d` does not exist", order)
}

func errMigrationFailed(name string, dir migrations.Direction, err error) error {
	return fmt.Errorf("identityserver/server/db: could not apply migration `%s` %s: %s", name, dir, err)
}

const migrationHistorySchema = `
	CREATE TABLE IF NOT EXISTS migration_history (
		"order"     INTEGER NOT NULL,
		name        STRING NOT NULL,
		direction   STRING NOT NULL,
		ran_at      TIMESTAMP DEFAULT current_timestamp(),
		PRIMARY KEY("order", direction, ran_at)
	);
`

func (db *DB) ensureSchema() error {
	_, err := db.Exec(migrationHistorySchema)
	return err
}

func (db *DB) currentState() (int, error) {
	err := db.ensureSchema()
	if err != nil {
		return 0, err
	}
	var last struct {
		Order     int                  `db:"order"`
		Direction migrations.Direction `db:"direction"`
	}
	err = db.SelectOne(
		&last,
		`SELECT "order", direction
			FROM migration_history
			ORDER BY ran_at DESC
			LIMIT 1`)
	// no migration has been applied yet
	if IsNoRows(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	// if last applied migration was backwards current state is: order - 1
	if last.Direction == migrations.DirectionBackwards {
		return last.Order - 1, nil
	}
	return last.Order, nil
}

// logAppliedMigration adds a record in the database about an applied migration
func (db *DB) logAppliedMigration(order int, name string, direction migrations.Direction) error {
	err := db.ensureSchema()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT
			INTO migration_history ("order", name, direction)
			VALUES ($1, $2, $3)`,
		order,
		name,
		direction)
	return err
}

// Migrate migrates the database to a target migration. Implements Migrator interface
func (db *DB) Migrate(target int) error {
	if db.migrations == nil {
		return errMigrationsRegistryNotSet
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
		// if direction is ascendent current migration
		// to perform is actually: i + incr
		n := i + incr
		if incr == -1 {
			n = i
		}
		migration, exists := db.migrations.Get(n)
		if !exists {
			return errMigrationDoesNotExist(n)
		}
		next := migration.Forwards
		back := migration.Backwards
		if incr == -1 {
			next = migration.Backwards
			back = migration.Forwards
		}
		if _, err := db.Exec(next); err != nil {
			return errMigrationFailed(migration.Name, direction, err)
		}
		defer func() {
			if err != nil {
				db.Exec(back)
			}
		}()
		err := db.logAppliedMigration(migration.Order, migration.Name, direction)
		if err != nil {
			return err
		}
	}
	return nil
}

// MigrateAll applies forwards all unapplied migrations. Implements Migrator interface
func (db *DB) MigrateAll() error {
	return db.Migrate(db.migrations.Count())
}
