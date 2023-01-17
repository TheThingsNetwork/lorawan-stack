// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// Package store implements the Identity Server store interfaces using the bun library.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

type baseDB struct {
	DB *bun.DB

	server string
	major  int
}

func newDB(ctx context.Context, db *bun.DB) (*baseDB, error) {
	s := &baseDB{DB: db}

	var version string
	res, err := db.QueryContext(ctx, "SELECT VERSION()")
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	res.Next()
	if err = res.Scan(&version); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	s.server, _, _ = strings.Cut(version, " ")

	var serverVersion string
	res, err = db.QueryContext(ctx, "SHOW SERVER_VERSION")
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	res.Next()
	if err = res.Scan(&serverVersion); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	major, _, _ := strings.Cut(serverVersion, ".")
	s.major, _ = strconv.Atoi(major)

	return s, nil
}

func (db *baseDB) baseStore() *baseStore {
	return &baseStore{DB: db.DB, baseDB: db}
}

type baseStore struct {
	DB bun.IDB
	*baseDB
}

func (*baseStore) now() time.Time { return now() }

func (s *baseStore) transact(ctx context.Context, fc func(context.Context, bun.IDB) error) error {
	db, ok := s.DB.(*bun.DB)
	if !ok { // Probably already in a transaction, just call the func.
		return fc(ctx, s.DB)
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return storeutil.WrapDriverError(err)
	}
	var done bool
	defer func() {
		if !done {
			tx.Rollback() //nolint:errcheck
		}
	}()
	err = fc(ctx, tx)
	if err != nil {
		return err
	}
	done = true
	err = tx.Commit()
	if err != nil {
		// tx.Commit returns an error if the context provided to BeginTx is canceled.
		if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, sql.ErrTxDone) {
			return ctxErr
		}
		return storeutil.WrapDriverError(err)
	}
	return nil
}

func newStore(baseStore *baseStore) *Store {
	return &Store{
		baseStore: baseStore,

		applicationStore:  newApplicationStore(baseStore),
		clientStore:       newClientStore(baseStore),
		endDeviceStore:    newEndDeviceStore(baseStore),
		gatewayStore:      newGatewayStore(baseStore),
		organizationStore: newOrganizationStore(baseStore),
		userStore:         newUserStore(baseStore),
		userSessionStore:  newUserSessionStore(baseStore),
		apiKeyStore:       newAPIKeyStore(baseStore),
		membershipStore:   newMembershipStore(baseStore),
		contactInfoStore:  newContactInfoStore(baseStore),
		invitationStore:   newInvitationStore(baseStore),
		loginTokenStore:   newLoginTokenStore(baseStore),
		oauthStore:        newOAuthStore(baseStore),
		euiStore:          newEUIStore(baseStore),
		entitySearch:      newEntitySearch(baseStore),
		notificationStore: newNotificationStore(baseStore),
	}
}

// NewStore returns a new store that implements store.TransactionalStore.
func NewStore(ctx context.Context, db *bun.DB) (*Store, error) {
	baseDB, err := newDB(ctx, db)
	if err != nil {
		return nil, err
	}
	return newStore(baseDB.baseStore()), nil
}

// Store is the combined store of all the individual stores.
type Store struct {
	TxOptions sql.TxOptions

	*baseStore

	*applicationStore
	*clientStore
	*endDeviceStore
	*gatewayStore
	*organizationStore
	*userStore
	*userSessionStore
	*apiKeyStore
	*membershipStore
	*contactInfoStore
	*invitationStore
	*loginTokenStore
	*oauthStore
	*euiStore
	*entitySearch
	*notificationStore
}

const (
	initialDelayOnUnavailable = 50 * time.Millisecond
	maxAttempts               = 3
)

// Transact implements the store.TransactionalStore interface.
func (s *Store) Transact(ctx context.Context, fc func(context.Context, store.Store) error) (err error) {
	delayOnUnavailable := initialDelayOnUnavailable
	for i := 0; i < maxAttempts; i++ {
		err = s.baseStore.transact(ctx, func(ctx context.Context, idb bun.IDB) error {
			baseStore := s.baseDB.baseStore()
			baseStore.DB = idb
			return fc(ctx, newStore(baseStore))
		})
		if !errors.Is(err, storeutil.ErrUnavailable) {
			break
		}
		log.FromContext(ctx).WithError(err).WithFields(log.Fields(
			"delay", delayOnUnavailable,
			"attempt", i+1,
		)).Warn("Database unavailable, retrying transaction...")
		time.Sleep(delayOnUnavailable)
		delayOnUnavailable *= 2
	}
	return err
}

// DBMetadata wraps the database metadata
// needed to distinguish between PostgreSQL and CockroachDB.
type DBMetadata struct {
	// Version is the database version.
	Version string

	// Type is the database type (PostgreSQL, CockroachDB).
	Type string
}

// Open a new database connection.
func Open(ctx context.Context, dsn string) (*bun.DB, error) {
	sqlDB, err := storeutil.OpenDB(ctx, dsn)
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqlDB, pgdialect.New())
	return db, nil
}

func getDBMetadata(ctx context.Context, db *bun.DB) (*DBMetadata, error) {
	var dbVersion, serverVersion string
	metadata := &DBMetadata{}
	if err := db.QueryRowContext(ctx, "SELECT version();").Scan(&dbVersion); err != nil {
		return nil, err
	}
	if strings.Contains(dbVersion, "PostgreSQL") {
		metadata.Type = "PostgreSQL"
	} else if strings.Contains(dbVersion, "CockroachDB") {
		metadata.Type = "CockroachDB"
	} else {
		panic(fmt.Sprintf("unknown database type: %s", dbVersion))
	}
	if err := db.QueryRowContext(ctx, "SHOW server_version;").Scan(&serverVersion); err != nil {
		return nil, err
	}
	metadata.Version = serverVersion
	return metadata, nil
}

func initializePostgreSQL(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto;"); err != nil {
		return err
	}
	return nil
}

func initializeCockroachDB(ctx context.Context, db *bun.DB, dbName string) error {
	if _, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS ?;", dbName); err != nil {
		return err
	}
	return nil
}

// Initialize initializes the database.
func Initialize(ctx context.Context, db *bun.DB, dbName string) error {
	md, err := getDBMetadata(ctx, db)
	if err != nil {
		return err
	}
	switch md.Type {
	case "PostgreSQL":
		return initializePostgreSQL(ctx, db)
	case "CockroachDB":
		return initializeCockroachDB(ctx, db, dbName)
	default:
		panic(fmt.Sprintf("unknown database type: %s", md.Type))
	}
}
