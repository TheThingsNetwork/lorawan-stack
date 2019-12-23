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

package store

import (
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
)

var models []interface{}

func registerModel(m ...interface{}) {
	models = append(models, m...)
}

var (
	errMissingTable  = errors.DefineCorruption("database_table", "database table `{table}` does not exist")
	errMissingColumn = errors.DefineCorruption("database_table_column", "column `{column}` does not exist in database table `{table}`")
)

// Check that the database contains all tables.
func Check(db *gorm.DB) error {
	db = db.Unscoped()
	db.SetLogger(logger{log.Noop})

	dbKind, ok := db.Get("db:kind")
	if !ok || (dbKind != "CockroachDB" && dbKind != "PostgreSQL") {
		for _, model := range models {
			if !db.HasTable(model) {
				tableName := db.NewScope(model).GetModelStruct().TableName(db)
				return errMissingTable.WithAttributes("table", tableName)
			}
		}
		return nil
	}

	// Get the tables from the database.
	var existingTables []struct {
		TableName string
	}
	err := db.Raw("SELECT table_name FROM INFORMATION_SCHEMA.tables WHERE table_type = 'BASE TABLE' AND table_schema = CURRENT_SCHEMA()").
		Scan(&existingTables).
		Error
	if err != nil {
		return err
	}

	// Check that a table exists for each model.
	for _, model := range models {
		tableName := db.NewScope(model).TableName()
		var tableExists bool
		for _, existingTable := range existingTables {
			if existingTable.TableName == tableName {
				tableExists = true
				break
			}
		}
		if !tableExists {
			return errMissingTable.WithAttributes("table", tableName)
		}
	}

	// Get the columns for all tables from the database.
	var existingColumns []struct {
		TableName  string
		ColumnName string
	}
	err = db.Raw("SELECT table_name, column_name FROM INFORMATION_SCHEMA.columns WHERE table_schema = CURRENT_SCHEMA()").
		Scan(&existingColumns).
		Error

	// Check that columns exist for each field of every model.
	for _, model := range models {
		scope := db.NewScope(model)
		tableName := scope.TableName()
		for _, field := range scope.GetModelStruct().StructFields {
			if !field.IsNormal {
				continue
			}
			columnName := field.DBName
			var columnExists bool
			for _, existingColumn := range existingColumns {
				if existingColumn.TableName == tableName && existingColumn.ColumnName == columnName {
					columnExists = true
					break
				}
			}
			if !columnExists {
				return errMissingColumn.WithAttributes("table", tableName, "column", columnName)
			}
		}
	}

	return nil
}

// AutoMigrate automatically migrates the database for the registered models.
func AutoMigrate(db *gorm.DB) *gorm.DB {
	return db.AutoMigrate(models...)
}

// clear database tables for the given models.
// This should be used with caution.
func clear(db *gorm.DB, models ...interface{}) (err error) {
	if dbKind, ok := db.Get("db:kind"); ok && dbKind == "CockroachDB" {
		if err = db.Exec("SET SQL_SAFE_UPDATES = FALSE").Error; err != nil {
			return err
		}
	}
	for _, model := range models {
		if err = db.Unscoped().Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// Clear database tables for all registered models.
// This should be used with caution.
func Clear(db *gorm.DB) error {
	return clear(db, models...)
}
