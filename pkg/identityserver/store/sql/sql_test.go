// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"context"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "is_tests"
)

// Single store instance shared across all tests
var testingStore *Store

func testStore() *Store {
	if testingStore != nil {
		return testingStore
	}

	testingStore = cleanStore(database)

	return testingStore
}

func cleanStore(database string) *Store {
	// open database connection
	db, err := db.Open(
		context.Background(),
		fmt.Sprintf(address, database),
		migrations.Registry)
	if err != nil {
		panic(err)
	}

	// drop database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", database))
	if err != nil {
		panic(err)
	}

	// create it again
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", database))
	if err != nil {
		panic(err)
	}

	// apply all migrations
	err = db.MigrateAll()
	if err != nil {
		panic(err)
	}

	// instantiate store
	s, err := FromDB(db)
	if err != nil {
		panic(err)
	}

	// create some users
	for _, user := range testUsers() {
		_, err := s.Users.Create(user)
		if err != nil {
			panic(err)
		}
	}

	return s
}
