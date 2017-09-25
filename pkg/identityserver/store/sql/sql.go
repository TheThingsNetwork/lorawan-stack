// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
)

// Store is an SQL data store.
type Store struct {
	db *db.DB
	store.Store
}

// Open opens and returns a reference to a new SQL datastore or an error if
// opening the connection failed.
func Open(address string) (*Store, error) {
	db, err := db.Open(context.Background(), address, migrations.Registry)
	if err != nil {
		return nil, err
	}

	return FromDB(db), nil
}

// FromDB creates a new Store from an already existing DB.
func FromDB(db *db.DB) *Store {
	store := &Store{
		db: db,
	}

	store.initSubStores()

	return store
}

// WithContext creates a reference to a new Store that will use the
// provided context for every request. This Store will share its database
// connection with the original store so don't close it if you want to keep
// using the parent database.
func (s *Store) WithContext(context context.Context) *Store {
	store := &Store{
		db: s.db.WithContext(context),
	}

	store.initSubStores()

	return store
}

// Close closes the connection to the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// initSubStores initializes all the substores.
func (s *Store) initSubStores() {
	s.Users = &UserStore{Store: s, factory: factory.DefaultUser{}}
	s.Applications = &ApplicationStore{Store: s, factory: factory.DefaultApplication{}}
	s.Gateways = &GatewayStore{Store: s, factory: factory.DefaultGateway{}}
	s.Clients = &ClientStore{Store: s, factory: factory.DefaultClient{}}
}
