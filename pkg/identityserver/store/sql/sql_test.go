// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql_test

import (
	"fmt"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
)

const (
	address            = "postgres://root@localhost:26257/%s?sslmode=disable"
	database           = "is_store_tests"
	attributesDatabase = "is_store_attributes_test"
)

// Store instances shared across different tests.
// The key of the map is the database name.
var testingStore map[string]*Store = make(map[string]*Store)

// testStore returns a store instance of the given database. The store is
// initialized only the first time is called on a specific database as it is
// indexed on a map that will be used for the subsequent times this method
// is called.
func testStore(t testing.TB, database string) *Store {
	if _, exists := testingStore[database]; !exists {
		testingStore[database] = cleanStore(t, database)
	}

	return testingStore[database]
}

// cleanStore returns a clean store instance. The database will be dropped
// and recreated again, migrations will be applied and finally will be feed
// with some test users.
func cleanStore(t testing.TB, database string) *Store {
	uri := fmt.Sprintf(address, database)
	logger := test.GetLogger(t).WithFields(log.Fields(
		"namespace", "Identity Server",
		"connection_uri", uri,
	))

	s, err := Open(uri)
	if err != nil {
		logger.WithError(err).Fatal("Failed to open a store with the CockroachDB instance")
	}

	err = s.DropDatabase()
	if err != nil {
		logger.WithError(err).Fatal("Failed to drop database")
	}

	err = s.Init()
	if err != nil {
		logger.WithError(err).Fatalf("Failed to initialize store")
		return nil
	}

	for _, user := range []*ttnpb.User{alice, bob} {
		err := s.Users.Create(user)
		if err != nil {
			logger.WithError(err).Fatal("Failed to feed test store with users")
		}
	}

	return s
}
