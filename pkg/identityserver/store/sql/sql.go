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

package sql

import (
	"context"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
)

// storer is the interface all stores have to adhere to.
type storer interface {
	// queryer returns the storers query context.
	queryer() db.QueryContext

	// transact starts a new transaction in the storer.
	transact(fn func(*db.Tx) error, opts ...db.TxOption) error

	// store returns the underlying store.
	store() *store.Store
}

// Store is a SQL data store.
type Store struct {
	db *db.DB
	store.Store
}

// Open opens a new database connection and attachs it to a new store.
func Open(connectionURI string) (*Store, error) {
	db, err := db.Open(context.Background(), connectionURI, migrations.Registry)
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

	initSubStores(s)
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

	initSubStores(store)

	return store
}

// Transact executes fn inside a transaction and retries it or rollbacks it as
// needed. It returns the error fn returns.
func (s *Store) Transact(fn func(*store.Store) error) error {
	return s.transact(func(tx *db.Tx) error {
		store := &txStore{
			tx: tx,
		}

		initSubStores(store)

		return fn(store.store())
	})
}

// Close closes the connection to the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// queryer returns the global database context.
func (s *Store) queryer() db.QueryContext {
	return s.db
}

// transact starts a new transaction.
func (s *Store) transact(fn func(*db.Tx) error, opts ...db.TxOption) error {
	return s.db.Transact(fn, opts...)
}

// store returns the store.Store.
func (s *Store) store() *store.Store {
	return &s.Store
}

// txStore is a store that keeps a transaction that is being executed.
type txStore struct {
	tx *db.Tx
	store.Store
}

// queryer returns the transaction that is already happening.
func (s *txStore) queryer() db.QueryContext {
	return s.tx
}

// transact works in the same transaction that is already happening.
func (s *txStore) transact(fn func(*db.Tx) error, opts ...db.TxOption) error {
	return fn(s.tx)
}

// store returns the store.Store.
func (s *txStore) store() *store.Store {
	return &s.Store
}

// initSubStores initializes the sub stores of the store.
func initSubStores(s storer) {
	store := s.store()
	store.Users = NewUserStore(s)
	store.Applications = NewApplicationStore(s)
	store.Gateways = NewGatewayStore(s)
	store.Clients = NewClientStore(s)
	store.OAuth = NewOAuthStore(s)
	store.Settings = NewSettingStore(s)
	store.Invitations = NewInvitationStore(s)
	store.Organizations = NewOrganizationStore(s)
}

// Init creates the database if it does not exist yet and applies the unapplied migrations.
func (s *Store) Init() error {
	_, err := s.db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", s.db.Database()))
	if err != nil {
		return err
	}

	return s.MigrateAll()
}

// CreateDatabase creates the database.
func (s *Store) CreateDatabase() error {
	_, err := s.db.Exec(fmt.Sprintf("CREATE DATABASE %s", s.db.Database()))
	return err
}

// DropDatabase deletes the database.
func (s *Store) DropDatabase() error {
	_, err := s.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s CASCADE", s.db.Database()))
	return err
}

// MigrateAll applies all unapplied migrations.
func (s *Store) MigrateAll() error {
	return s.db.MigrateAll()
}
