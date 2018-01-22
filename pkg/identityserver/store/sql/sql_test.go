// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"context"
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "is_store_tests"
)

// Single store instance shared across all tests.
var testingStore *Store

// testStore returns a single and shared store instance everytime the method is
// called. The first time that is called it creates  a new store instance in a
// newly created database.
func testStore(t testing.TB, database string) *Store {
	if testingStore == nil {
		testingStore = cleanStore(t, database)
	}

	return testingStore
}

// cleanStore returns a new store instance attached to a newly created database
// where all migrations has been applied and also has been feed with some users.
func cleanStore(t testing.TB, database string) *Store {
	logger := test.GetLogger(t).WithField("tag", "Identity Server")

	// open database connection
	db, err := db.Open(context.Background(), fmt.Sprintf(address, database), migrations.Registry)
	if err != nil {
		logger.WithError(err).Fatal("Failed to establish a connection with the CockroachDB instance")
		return nil
	}

	// drop database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s CASCADE", database))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to delete database `%s`", database)
		return nil
	}

	// create it again
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", database))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create database `%s`", database)
		return nil
	}

	// apply all migrations
	err = db.MigrateAll()
	if err != nil {
		logger.WithError(err).Fatal("Failed to apply the migrations from the registry")
		return nil
	}

	// instantiate store
	s := FromDB(db)

	// create some users
	for _, user := range testUsers() {
		err := s.Users.Create(user)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to feed test database `%s` with some users", database)
			return nil
		}
	}

	testClientCreate(t, s)

	return s
}
