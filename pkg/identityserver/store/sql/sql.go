// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
)

type storer interface {
	queryer() db.QueryContext
	transact(fn func(*db.Tx) error, opts ...db.TxOption) error
}

// Store is a SQL data store.
type Store struct {
	db *db.DB
	store.Store
}

// Open openes a new database connection and attachs it to a new store.
func Open(dsn string) (*Store, error) {
	db, err := db.Open(context.Background(), dsn, migrations.Registry)
	if err != nil {
		return nil, err
	}

	return FromDB(db), nil
}

// FromDB creates a new store from a datababase connection.
func FromDB(db *db.DB) *Store {
	s := &Store{
		db: db,
	}
	s.initSubStores()
	return s
}

// WithContext creates a reference to a new Store that will use the
// provided context for every request. This Store will share its database
// connection with the original store so don't close it if you want to keep
// using the parent database.
func (s *Store) WithContext(context context.Context) *Store {
	store := &Store{
		db: s.db.WithContext(context),
	}

	s.initSubStores()

	return store
}

// Transact executes fn inside a transaction and retries it or rollbacks it as
// needed. It returns the error fn returns.
func (s *Store) Transact(fn func(store.Store) error) error {
	return s.transact(func(tx *db.Tx) error {
		store := &txStore{
			tx: tx,
		}
		store.initSubStores()

		return fn(store.Store)
	})
}

// Close closes the connection to the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) queryer() db.QueryContext {
	return s.db
}

func (s *Store) transact(fn func(*db.Tx) error, opts ...db.TxOption) error {
	return s.db.Transact(fn, opts...)
}

func (s *Store) initSubStores() {
	s.Users = NewUserStore(s, factory.DefaultUser{})
	s.Applications = NewApplicationStore(s, factory.DefaultApplication{})
	s.Gateways = NewGatewayStore(s, factory.DefaultGateway{})
	s.Clients = NewClientStore(s, factory.DefaultClient{})
}

// txStore is a store that holds a transaction that is being executed.
type txStore struct {
	tx *db.Tx
	store.Store
}

// Transact returns error as a transaction is being executed already.
func (s *txStore) Transact(fn func(store.Store) error) error {
	return errors.New("Failed to execute transaction. There is already a transaction in progress.")
}

func (s *txStore) queryer() db.QueryContext {
	return s.tx
}

func (s *txStore) transact(fn func(*db.Tx) error, opts ...db.TxOption) error {
	return fn(s.tx)
}

func (s *txStore) initSubStores() {
	s.Users = NewUserStore(s, factory.DefaultUser{})
	s.Applications = NewApplicationStore(s, factory.DefaultApplication{})
	s.Gateways = NewGatewayStore(s, factory.DefaultGateway{})
	s.Clients = NewClientStore(s, factory.DefaultClient{})
}
