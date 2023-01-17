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

package storeutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	// ErrUnavailable is an error that indicates an unavailable database.
	ErrUnavailable = errors.DefineUnavailable("database_unavailable", "database unavailable")
	// ErrNotFound is an error that indicates a missing entity.
	ErrNotFound = errors.DefineNotFound("not_found", "no results found")
	// ErrDriver is an error that indicates a driver error.
	ErrDriver = errors.Define("driver", "driver error")
	// ErrIDTaken is returned when an entity can not be created because the ID is already taken.
	ErrIDTaken = errors.DefineAlreadyExists("id_taken", "ID already taken, choose a different one and try again")
	// ErrEUITaken is returned when an entity can not be created because the EUI is already taken.
	ErrEUITaken = errors.DefineAlreadyExists("eui_taken", "EUI already taken")
)

// driverErrorCodes maps PostgreSQL error codes to the corresponding error definition.
// See https://www.postgresql.org/docs/current/errcodes-appendix.html for more information.
var driverErrorCodes = map[string]*errors.Definition{
	pgerrcode.UniqueViolation:  errors.DefineAlreadyExists("already_exists", "already exists"),
	pgerrcode.QueryCanceled:    errors.DefineCanceled("query_canceled", "query canceled"),
	pgerrcode.AdminShutdown:    ErrUnavailable,
	pgerrcode.CrashShutdown:    ErrUnavailable,
	pgerrcode.CannotConnectNow: ErrUnavailable,
}

// DriverError encapsulates the PostgreSQL error data.
type DriverError struct {
	Code       string
	Message    string
	Detail     string
	Table      string
	Column     string
	Constraint string
}

// Attributes gets the DriverError attributes.
func (e *DriverError) Attributes() []any {
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

// GetError gets the corresponding error definition.
func (e *DriverError) GetError() error {
	attributes := e.Attributes()
	if def, ok := driverErrorCodes[e.Code]; ok {
		err := def.WithAttributes(attributes...)
		if def.Name() == "already_exists" {
			switch {
			case strings.HasSuffix(e.Constraint, "_id_index"):
				return ErrIDTaken.WithCause(err)
			case strings.HasSuffix(e.Constraint, "_eui_index"):
				return ErrEUITaken.WithCause(err)
			}
		}
		return err
	}
	return ErrDriver.WithAttributes(attributes...)
}

// WrapDriverError wraps a driver error with the corresponding error definition.
func WrapDriverError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound.WithCause(err)
	}
	var driverError DriverError
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
		return ErrUnavailable.WithCause(err)
	}
	if errors.Is(err, driver.ErrBadConn) {
		return ErrUnavailable.WithCause(err)
	}
	if timeoutError := (interface{ Timeout() bool })(nil); errors.As(err, &timeoutError) && timeoutError.Timeout() {
		return context.DeadlineExceeded
	}
	return err
}
