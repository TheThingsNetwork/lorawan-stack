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
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
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
		return nil, wrapDriverError(err)
	}
	res.Next()
	if err = res.Scan(&version); err != nil {
		return nil, wrapDriverError(err)
	}

	s.server, _, _ = strings.Cut(version, " ")

	var serverVersion string
	res, err = db.QueryContext(ctx, "SHOW SERVER_VERSION")
	if err != nil {
		return nil, wrapDriverError(err)
	}
	res.Next()
	if err = res.Scan(&serverVersion); err != nil {
		return nil, wrapDriverError(err)
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
		return wrapDriverError(err)
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
		return wrapDriverError(err)
	}
	return nil
}

var (
	errUnavailable = errors.DefineUnavailable("database_unavailable", "database unavailable")
	errNotFound    = errors.DefineNotFound("not_found", "no results found")
	errDriver      = errors.Define("driver", "driver error")
)

// driverErrorcodes maps PostgreSQL error codes to the corresponding error definition.
// See https://www.postgresql.org/docs/current/errcodes-appendix.html for more information.
var driverErrorCodes = map[string]*errors.Definition{
	pgerrcode.UniqueViolation:  errors.DefineAlreadyExists("already_exists", "already exists"),
	pgerrcode.QueryCanceled:    errors.DefineCanceled("query_canceled", "query canceled"),
	pgerrcode.AdminShutdown:    errUnavailable,
	pgerrcode.CrashShutdown:    errUnavailable,
	pgerrcode.CannotConnectNow: errUnavailable,
}

type driverError struct {
	Code       string
	Message    string
	Detail     string
	Table      string
	Column     string
	Constraint string
}

func (e *driverError) Attributes() []any {
	var attributes []any
	if e.Code != "" {
		attributes = append(attributes, "driver_code", e.Code)
	}
	if e.Message != "" {
		attributes = append(attributes, "driver_message", e.Message)
	}
	if e.Detail != "" {
		attributes = append(attributes, "driver_detail", e.Detail)
	}
	if e.Table != "" {
		attributes = append(attributes, "driver_table", e.Table)
	}
	if e.Column != "" {
		attributes = append(attributes, "driver_column", e.Column)
	}
	if e.Constraint != "" {
		attributes = append(attributes, "driver_constraint", e.Constraint)
	}
	return attributes
}

func (e *driverError) GetError() error {
	attributes := e.Attributes()
	if def, ok := driverErrorCodes[e.Code]; ok {
		err := def.WithAttributes(attributes...)
		if def.Name() == "already_exists" {
			switch {
			case strings.HasSuffix(e.Constraint, "_id_index"):
				return store.ErrIDTaken.WithCause(err)
			case strings.HasSuffix(e.Constraint, "_eui_index"):
				return store.ErrEUITaken.WithCause(err)
			}
		}
		return err
	}
	return errDriver.WithAttributes(attributes...)
}

// wrapDriverError wraps a driver error with the corresponding error definition.
func wrapDriverError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errNotFound.WithCause(err)
	}
	var driverError driverError
	if pgconnErr := (*pgconn.PgError)(nil); errors.As(err, &pgconnErr) {
		driverError.Code = pgconnErr.Code
		driverError.Message = pgconnErr.Message
		driverError.Detail = pgconnErr.Detail
		driverError.Table = pgconnErr.TableName
		driverError.Column = pgconnErr.ColumnName
		driverError.Constraint = pgconnErr.ConstraintName
		return driverError.GetError()
	}
	if pgdriverErr := (pgdriver.Error{}); errors.As(err, &pgdriverErr) {
		// See https://www.postgresql.org/docs/current/protocol-error-fields.html for more information.
		driverError.Code = pgdriverErr.Field('C')
		driverError.Message = pgdriverErr.Field('M')
		driverError.Detail = pgdriverErr.Field('D')
		driverError.Table = pgdriverErr.Field('t')
		driverError.Column = pgdriverErr.Field('c')
		driverError.Constraint = pgdriverErr.Field('n')
		return driverError.GetError()
	}
	if errors.Is(err, io.EOF) {
		return errUnavailable.WithCause(err)
	}
	if errors.Is(err, driver.ErrBadConn) {
		return errUnavailable.WithCause(err)
	}
	if timeoutError := (interface{ Timeout() bool })(nil); errors.As(err, &timeoutError) && timeoutError.Timeout() {
		return context.DeadlineExceeded
	}
	return err
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
		if !errors.Is(err, errUnavailable) {
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
