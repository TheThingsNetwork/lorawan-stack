// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// Package store implements the Identity Server store interfaces using GORM.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	createdAt = "created_at"
	updatedAt = "updated_at"
	deletedAt = "deleted_at"
	ids       = "ids"
)

func newStore(db *gorm.DB) *baseStore { return &baseStore{DB: db} }

type baseStore struct {
	DB *gorm.DB
}

func (s *baseStore) query(ctx context.Context, model interface{}, funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	query := s.DB.Model(model).Scopes(withContext(ctx), withSoftDeletedIfRequested(ctx))
	if len(funcs) > 0 {
		query = query.Scopes(funcs...)
	}
	return query
}

func (s *baseStore) findEntity(
	ctx context.Context, entityID ttnpb.IDStringer, fields ...string,
) (modelInterface, error) {
	model := modelForID(entityID)
	query := s.query(ctx, model, withID(entityID))
	if len(fields) == 1 && fields[0] == "id" {
		fields[0] = s.DB.NewScope(model).TableName() + ".id"
	}
	if len(fields) > 0 {
		query = query.Select(fields)
	}
	if err := query.First(model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(entityID)
		}
		return nil, convertError(err)
	}
	return model, nil
}

func (s *baseStore) loadContact(ctx context.Context, contact *Account) (*string, error) {
	if contact == nil || contact.AccountType == "" || contact.UID == "" {
		return nil, nil
	}
	err := s.query(ctx, Account{}).
		Where(contact).
		First(contact).
		Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(contact.OrganizationOrUserIdentifiers())
		}
		return nil, convertError(err)
	}
	return &contact.ID, nil
}

func (s *baseStore) createEntity(ctx context.Context, model interface{}) error {
	if model, ok := model.(modelInterface); ok {
		model.SetContext(ctx)
	}
	err := s.DB.Create(model).Error
	if err != nil {
		return convertError(err)
	}
	return nil
}

func (s *baseStore) updateEntity(ctx context.Context, model interface{}, columns ...string) error {
	query := s.query(ctx, model)
	query = query.Select(append(columns, "updated_at"))
	return query.Save(model).Error
}

func (s *baseStore) deleteEntity(ctx context.Context, entityID ttnpb.IDStringer) error {
	model, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return err
	}
	if err = s.DB.Delete(model).Error; err != nil {
		return err
	}
	switch entityType := entityID.EntityType(); entityType {
	case user, organization:
		err = s.DB.Where(Account{
			AccountType: entityType,
			AccountID:   model.PrimaryKey(),
		}).Delete(Account{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *baseStore) restoreEntity(ctx context.Context, entityID ttnpb.IDStringer) error {
	model, err := s.findEntity(store.WithSoftDeleted(ctx, false), entityID, "id")
	if err != nil {
		return err
	}
	switch entityType := entityID.EntityType(); entityType {
	case user, organization:
		err := s.DB.Unscoped().Model(Account{}).Where(Account{
			AccountType: entityType,
			AccountID:   model.PrimaryKey(),
		}).UpdateColumn("deleted_at", gorm.Expr("NULL")).Error
		if err != nil {
			return err
		}
	}
	return s.DB.Unscoped().Model(model).UpdateColumn("deleted_at", gorm.Expr("NULL")).Error
}

func (s *baseStore) purgeEntity(ctx context.Context, entityID ttnpb.IDStringer) error {
	model, err := s.findEntity(store.WithSoftDeleted(ctx, false), entityID, "id")
	if err != nil {
		return err
	}
	return s.DB.Unscoped().Delete(model).Error
}

var (
	errDatabase      = errors.DefineInternal("database", "database error")
	errAlreadyExists = errors.DefineAlreadyExists("already_exists", "entity already exists")
)

var uniqueViolationRegex = regexp.MustCompile(`duplicate key value( .+)? violates unique constraint "([a-z_]+)"`)

func convertError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	}
	if ttnErr, ok := errors.From(err); ok {
		return ttnErr
	}
	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case "unique_violation":
			if match := uniqueViolationRegex.FindStringSubmatch(pqErr.Message); match != nil {
				switch {
				case strings.HasSuffix(match[2], "_id_index"):
					return store.ErrIDTaken.WithCause(err)
				case strings.HasSuffix(match[2], "_eui_index"):
					return store.ErrEUITaken.WithCause(err)
				default:
					return errAlreadyExists.WithCause(err).WithAttributes("index", match[2])
				}
			}
			return errAlreadyExists.WithCause(err)
		default:
			return errDatabase.WithCause(err).WithAttributes("code", pqErr.Code.Name())
		}
	}
	return errDatabase.WithCause(err)
}

// Open opens a new database connection.
func Open(ctx context.Context, dsn string) (*gorm.DB, error) {
	dbURI, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	dbName := strings.TrimPrefix(dbURI.Path, "/")
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db = db.Set("db:name", dbName)
	var dbVersion string
	err = db.Raw("SELECT version()").Row().Scan(&dbVersion)
	if err != nil {
		return nil, err
	}
	db = db.Set("db:version", dbVersion)
	switch {
	case strings.Contains(dbVersion, "CockroachDB"):
		db = db.Set("db:kind", "CockroachDB")
	case strings.Contains(dbVersion, "PostgreSQL"):
		db = db.Set("db:kind", "PostgreSQL")
	}
	err = db.Raw("SHOW server_version").Row().Scan(&dbVersion)
	if err != nil {
		return nil, err
	}
	switch majorMinorPatch := strings.SplitN(dbVersion, ".", 3); len(majorMinorPatch) {
	case 3:
		patch, _ := strconv.Atoi(majorMinorPatch[2])
		db = db.Set("db:version:patch", patch)
		fallthrough
	case 2:
		minor, _ := strconv.Atoi(majorMinorPatch[1])
		db = db.Set("db:version:minor", minor)
		fallthrough
	case 1:
		major, _ := strconv.Atoi(majorMinorPatch[0])
		db = db.Set("db:version:major", major)
	}
	SetLogger(db, log.FromContext(ctx))
	return db, nil
}

// Initialize initializes the database.
func Initialize(db *gorm.DB) error {
	if dbKind, ok := db.Get("db:kind"); ok {
		switch dbKind {
		case "CockroachDB":
			if dbName, ok := db.Get("db:name"); ok {
				if err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)).Error; err != nil {
					return err
				}
			}
		case "PostgreSQL":
			if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ErrTransactionRecovered is returned when a panic is caught from a SQL transaction.
var ErrTransactionRecovered = errors.DefineInternal("transaction_recovered", "Internal Server Error")

// Transact executes f in a db transaction.
func Transact(ctx context.Context, db *gorm.DB, f func(db *gorm.DB) error) (err error) {
	defer trace.StartRegion(ctx, "database transaction").End()
	tx := db.Begin()
	if tx.Error != nil {
		return convertError(tx.Error)
	}
	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintln(os.Stderr, p)
			os.Stderr.Write(debug.Stack())
			if pErr, ok := p.(error); ok {
				switch {
				case errors.Is(pErr, context.Canceled):
					err = context.Canceled
				case errors.Is(pErr, context.DeadlineExceeded):
					err = context.DeadlineExceeded
				default:
					err = ErrTransactionRecovered.WithCause(pErr)
				}
			} else {
				err = ErrTransactionRecovered.WithAttributes("panic", p)
			}
			log.FromContext(ctx).WithError(err).Error("Transaction panicked")
		}
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
		err = convertError(err)
	}()
	SetLogger(tx, log.FromContext(ctx).WithField("namespace", "db"))
	return f(tx)
}

func entityTypeForID(id ttnpb.IDStringer) string {
	return strings.ReplaceAll(id.EntityType(), " ", "_")
}

func modelForEntityType(entityType string) modelInterface {
	switch entityType {
	case application:
		return &Application{}
	case client:
		return &Client{}
	case "end_device":
		return &EndDevice{}
	case gateway:
		return &Gateway{}
	case organization:
		return &Organization{}
	case user:
		return &User{}
	default:
		panic(fmt.Sprintf("can't find model for entity type %s", entityType))
	}
}

func modelForID(id ttnpb.IDStringer) modelInterface {
	return modelForEntityType(entityTypeForID(id))
}

var errMigrationNotFound = errors.DefineNotFound(
	"migration_not_found", "migration not found",
)

func errNotFoundForID(id ttnpb.IDStringer) error {
	switch t := entityTypeForID(id); t {
	case application:
		return store.ErrApplicationNotFound.WithAttributes("application_id", id.IDString())
	case client:
		return store.ErrClientNotFound.WithAttributes("client_id", id.IDString())
	case "end_device":
		appID, devID := splitEndDeviceIDString(id.IDString())
		return store.ErrEndDeviceNotFound.WithAttributes("application_id", appID, "device_id", devID)
	case gateway:
		return store.ErrGatewayNotFound.WithAttributes("gateway_id", id.IDString())
	case organization:
		return store.ErrOrganizationNotFound.WithAttributes("organization_id", id.IDString())
	case user:
		return store.ErrUserNotFound.WithAttributes("user_id", id.IDString())
	default:
		panic(fmt.Sprintf("can't find errNotFound for entity type %s", t))
	}
}

// SetLogger sets the database logger.
func SetLogger(db *gorm.DB, log log.Interface) {
	db.SetLogger(logger{Interface: log})
}

type logger struct {
	log.Interface
}

// Print implements the gorm.logger interface.
func (l logger) Print(v ...interface{}) {
	if len(v) < 3 {
		l.Error(fmt.Sprint(v...))
		return
	}
	logger := l.Interface
	source, ok := v[1].(string)
	if !ok {
		l.Error(fmt.Sprint(v...))
		return
	}
	logger = logger.WithField("source", filepath.Base(source))
	switch v[0] {
	case "log", "error":
		if err, ok := v[2].(error); ok {
			err = convertError(err)
			if errors.Resemble(err, errDatabase) {
				logger.WithError(err).Error("Database error")
			}
			return
		}
		logger.Error(fmt.Sprint(v[2:]...))
		return
	case "sql":
		if len(v) != 6 {
			return
		}
		duration, _ := v[2].(time.Duration)
		query, _ := v[3].(string)
		values, _ := v[4].([]interface{})
		rows, _ := v[5].(int64)
		logger.WithFields(log.Fields(
			"duration", duration,
			"query", query,
			"values", values,
			"rows", rows,
		)).Debug("Run database query")
	default:
		l.Error(fmt.Sprint(v...))
	}
}

func mergeFields(fields ...[]string) []string {
	var outLen int
	for _, fields := range fields {
		outLen += len(fields)
	}
	out := make([]string, 0, outLen)
	for _, fields := range fields {
		out = append(out, fields...)
	}
	return out
}

func cleanFields(fields ...string) []string {
	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

// NewCombinedStore returns a new store that implements store.TransactionalStore.
func NewCombinedStore(db *gorm.DB) *CombinedStore {
	baseStore := &baseStore{DB: db}
	return &CombinedStore{
		db:                db,
		baseStore:         baseStore,
		applicationStore:  applicationStore{baseStore: baseStore},
		clientStore:       clientStore{baseStore: baseStore},
		deviceStore:       deviceStore{baseStore: baseStore},
		gatewayStore:      gatewayStore{baseStore: baseStore},
		organizationStore: organizationStore{baseStore: baseStore},
		userStore:         userStore{baseStore: baseStore},
		userSessionStore:  userSessionStore{baseStore: baseStore},
		membershipStore:   membershipStore{baseStore: baseStore},
		apiKeyStore:       apiKeyStore{baseStore: baseStore},
		oauthStore:        oauthStore{baseStore: baseStore},
		invitationStore:   invitationStore{baseStore: baseStore},
		loginTokenStore:   loginTokenStore{baseStore: baseStore},
		entitySearch:      entitySearch{baseStore: baseStore},
		contactInfoStore:  contactInfoStore{baseStore: baseStore},
		euiStore:          euiStore{baseStore: baseStore},
		notificationStore: notificationStore{baseStore: baseStore},
	}
}

// CombinedStore combines all stores to implement the store.TransactionalStore interface.
type CombinedStore struct {
	TxOptions sql.TxOptions

	db *gorm.DB
	*baseStore

	applicationStore
	clientStore
	deviceStore
	gatewayStore
	organizationStore
	userStore
	userSessionStore
	membershipStore
	apiKeyStore
	oauthStore
	invitationStore
	loginTokenStore
	entitySearch
	contactInfoStore
	euiStore
	notificationStore
}

// Transact implements the store.TransactionalStore interface.
func (s *CombinedStore) Transact(ctx context.Context, fc func(context.Context, store.Store) error) (err error) {
	defer trace.StartRegion(ctx, "database transaction").End()
	tx := s.db.BeginTx(ctx, &s.TxOptions)
	if tx.Error != nil {
		return convertError(tx.Error)
	}
	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintln(os.Stderr, p)
			os.Stderr.Write(debug.Stack())
			if pErr, ok := p.(error); ok {
				switch {
				case errors.Is(err, context.Canceled):
					err = context.Canceled
				case errors.Is(err, context.DeadlineExceeded):
					err = context.DeadlineExceeded
				default:
					err = ErrTransactionRecovered.WithCause(pErr)
				}
			} else {
				err = ErrTransactionRecovered.WithAttributes("panic", p)
			}
			log.FromContext(ctx).WithError(err).Error("Transaction panicked")
		}
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
		err = convertError(err)
	}()
	SetLogger(tx, log.FromContext(ctx).WithField("namespace", "db"))
	return fc(ctx, NewCombinedStore(tx))
}
