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

package storetest

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	_ "github.com/lib/pq" // PostgreSQL driver.
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func GetDSN(defaultDB string) *url.URL {
	dsn := url.URL{
		Scheme: "postgresql",
		Host:   "localhost:5432",
		Path:   defaultDB,
	}
	dsn.User = url.UserPassword("root", "root")
	query := make(url.Values)
	query.Add("sslmode", "disable")
	if dbAddress := os.Getenv("SQL_DB_ADDRESS"); dbAddress != "" {
		dsn.Host = dbAddress
	}
	if dbName := os.Getenv("TEST_DATABASE_NAME"); dbName != "" {
		dsn.Path = dbName
	}
	if dbAuth := os.Getenv("SQL_DB_AUTH"); dbAuth != "" {
		var username, password string
		idx := strings.Index(dbAuth, ":")
		if idx != -1 {
			username, password = dbAuth[:idx], dbAuth[idx+1:]
		} else {
			username = dbAuth
		}
		dsn.User = url.UserPassword(username, password)
	}
	dsn.RawQuery = query.Encode()
	return &dsn
}

func GetSchemaDSN(base *url.URL, schemaName string) *url.URL {
	dsn := *base
	query := dsn.Query()
	query.Add("search_path", schemaName)
	dsn.RawQuery = query.Encode()
	return &dsn
}

func CreateSchema(db *sql.DB, schemaName string) error {
	_, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName))
	if err != nil {
		return fmt.Errorf("failed to drop old schema: %w", err)
	}
	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

func DropSchema(db *sql.DB, schemaName string) error {
	_, err := db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE;", schemaName))
	if err != nil {
		return fmt.Errorf("failed to drop old schema: %w", err)
	}
	return nil
}

func New(t *testing.T, newStore func(t *testing.T, dsn *url.URL) Store) *StoreTest {
	return &StoreTest{
		t:            t,
		dsn:          GetDSN("ttn_lorawan_is_store_test"),
		newStore:     newStore,
		population:   &Population{},
		schemaSuffix: fmt.Sprintf("%p", newStore),
	}
}

type Store interface {
	Init(ctx context.Context) error
	Close() error
}

type StoreTest struct {
	t            *testing.T
	dsn          *url.URL
	newStore     func(t *testing.T, dsn *url.URL) Store
	population   *Population
	schemaSuffix string
}

func (s *StoreTest) schemaName(t *testing.T) string {
	t.Helper()
	return strcase.ToSnake(t.Name() + s.schemaSuffix)
}

func (s *StoreTest) PrepareDB(t *testing.T) Store {
	_, ctx := test.New(t)

	db, err := sql.Open("postgres", s.dsn.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	start := time.Now()

	schemaName := s.schemaName(t)

	if err = CreateSchema(db, schemaName); err != nil {
		t.Fatal(err)
	}

	store := s.newStore(t, GetSchemaDSN(s.dsn, schemaName))

	if err := store.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if err := s.population.Populate(ctx, store); err != nil {
		t.Fatal(err)
	}

	t.Logf("Prepared schema %s in %s", schemaName, time.Since(start))

	return store
}

func (s *StoreTest) DestroyDB(t *testing.T, assertClean bool, exceptions ...string) {
	schemaName := s.schemaName(t)

	if t.Failed() {
		t.Logf("Keeping database to help debugging: %q", GetSchemaDSN(s.dsn, schemaName))
		return
	}

	db, err := sql.Open("postgres", s.dsn.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	start := time.Now()

	var totalRowCount int
	if assertClean {
		tableNameRows, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = $1", schemaName)
		if err != nil {
			t.Fatal(err)
		}
		var tableNames []string
		for tableNameRows.Next() {
			var tableName string
			if err := tableNameRows.Scan(&tableName); err != nil {
				t.Fatal(err)
			}
		}
	nextTable:
		for _, tableName := range tableNames {
			for _, exception := range exceptions {
				if tableName == exception {
					continue nextTable
				}
			}

			var rowCount int
			row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName))
			err = row.Scan(&rowCount)
			if err != nil {
				t.Fatal(err)
			}
			totalRowCount += rowCount
			if rowCount > 0 {
				t.Errorf("%d rows left in table %s", rowCount, tableName)
			}
		}
	}

	if totalRowCount == 0 {
		if err = DropSchema(db, schemaName); err != nil {
			t.Fatal(err)
		}
		t.Logf("Destroyed schema %s in %s", schemaName, time.Since(start))
	}

	s.population = &Population{}
}
