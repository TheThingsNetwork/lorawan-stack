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
	"io"
	"strconv"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

func newStore(ctx context.Context, db bun.IDB) (*baseStore, error) {
	s := &baseStore{DB: db}

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

type baseStore struct {
	DB bun.IDB

	server string
	major  int
}

var (
	errUnavailable = errors.DefineUnavailable("database_unavailable", "database unavailable")
	errNotFound    = errors.DefineNotFound("not_found", "no results found")
	errDriver      = errors.Define("driver", "driver error")
)

// driverErrorcodes maps PostgreSQL error codes to the corresponding error definition.
// See https://www.postgresql.org/docs/current/errcodes-appendix.html for more information.
var driverErrorCodes = map[string]*errors.Definition{
	"23505": errors.DefineAlreadyExists("already_exists", "already exists"), // unique_violation
	"57014": errors.DefineCanceled("query_canceled", "query canceled"),      // query_canceled
	"57P01": errUnavailable,                                                 // admin_shutdown
	"57P02": errUnavailable,                                                 // crash_shutdown
	"57P03": errUnavailable,                                                 // cannot_connect_now
}

// driverErrorDetails maps PostgreSQL error codes to attributes.
// See https://www.postgresql.org/docs/current/protocol-error-fields.html for more information.
var driverErrorDetails = []struct {
	field     byte
	attribute string
}{
	{'C', "driver_code"},
	{'M', "driver_message"},
	{'D', "driver_detail"},
	{'t', "driver_table"},
	{'c', "driver_column"},
	{'n', "driver_constraint"},
}

// wrapDriverError wraps a driver error with the corresponding error definition.
func wrapDriverError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errNotFound.WithCause(err)
	}
	if pgdriverErr := (pgdriver.Error{}); errors.As(err, &pgdriverErr) {
		attributes := make([]interface{}, 0, len(driverErrorDetails)*2)
		for _, detail := range driverErrorDetails {
			if value := pgdriverErr.Field(detail.field); value != "" {
				attributes = append(attributes, detail.attribute, value)
			}
		}
		if def, ok := driverErrorCodes[pgdriverErr.Field('C')]; ok {
			return def.WithAttributes(attributes...)
		}
		return errDriver.WithAttributes(attributes...)
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
		return wrapDriverError(err)
	}
	return nil
}
