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

// Package sql provides an implementation of the store using SQL databases.
package sql

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store/sql/migrations"
)

// storer is the interface all store implementations have to adhere to.
type storer interface {
	// queryer returns the storers query context.
	queryer() db.QueryContext

	// transact starts a new transaction in the storer.
	transact(fn func(*db.Tx) error, opts ...db.TxOption) error

	// store returns the underlying store.
	store() *store.Store
}

// Implementation of the SQL store. It implements storer.
type impl struct {
	db *db.DB
	store.Store
}

// Open opens a new database connection and attachs it to a new store.
func Open(connectionURI string) (*store.Store, error) {
	db, err := db.Open(context.Background(), connectionURI, migrations.Registry)
	if err != nil {
		return nil, err
	}

	return FromDB(db), nil
}

// FromDB creates a new store from a datababase connection.
func FromDB(db *db.DB) *store.Store {
	s := &impl{
		db: db,
	}
	s.Store.Behaviour = s

	initSubStores(s)

	return s.store()
}

// WithContext creates a reference to a new Store that will use the
// provided context for every request. This Store will share its database
// connection with the original store so don't close it if you want to keep
// using the parent database.
func (s *impl) WithContext(context context.Context) *store.Store {
	store := &impl{
		db: s.db.WithContext(context),
	}
	store.Store.Behaviour = store

	initSubStores(store)

	return store.store()
}

// Transact executes fn inside a transaction and retries it or rollbacks it as
// needed. It returns the error fn returns.
func (s *impl) Transact(fn func(*store.Store) error) error {
	return s.transact(func(tx *db.Tx) error {
		store := &txImpl{
			tx: tx,
		}

		initSubStores(store)

		return fn(store.store())
	})
}

// Init creates the database if it does not exist yet and applies the unapplied migrations.
func (s *impl) Init() error {
	_, err := s.db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", s.db.Database()))
	if err != nil {
		return err
	}

	return s.db.MigrateAll()
}

// Clean deletes the database.
func (s *impl) Clean() error {
	_, err := s.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s CASCADE", s.db.Database()))
	return err
}

// Close closes the connection to the database.
func (s *impl) Close() error {
	return s.db.Close()
}

// queryer returns the global database context.
func (s *impl) queryer() db.QueryContext {
	return s.db
}

// transact starts a new transaction.
func (s *impl) transact(fn func(*db.Tx) error, opts ...db.TxOption) error {
	return s.db.Transact(fn, opts...)
}

// store returns the store.Store.
func (s *impl) store() *store.Store {
	return &s.Store
}

// txImpl is a store that keeps a transaction that is being executed. It implements storer.
type txImpl struct {
	tx *db.Tx
	store.Store
}

// queryer returns the transaction that is already happening.
func (s *txImpl) queryer() db.QueryContext {
	return s.tx
}

// transact works in the same transaction that is already happening.
func (s *txImpl) transact(fn func(*db.Tx) error, opts ...db.TxOption) error {
	return fn(s.tx)
}

// store returns the store.Store.
func (s *txImpl) store() *store.Store {
	return &s.Store
}

// initSubStores initializes the sub-stores of the store.Store.
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
