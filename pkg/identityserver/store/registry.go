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

package store

import (
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var models []interface{}

func registerModel(m ...interface{}) {
	models = append(models, m...)
}

var errMissingTable = errors.DefineCorruption("database_table", "database table `{table}` does not exist")

// Check that the database contains all tables.
func Check(db *gorm.DB) error {
	for _, model := range models {
		if !db.HasTable(model) {
			tableName := db.NewScope(model).GetModelStruct().TableName(db)
			return errMissingTable.WithAttributes("table", tableName)
		}
	}
	return nil
}

// AutoMigrate automatically migrates the database for the registered models.
func AutoMigrate(db *gorm.DB) *gorm.DB {
	return db.AutoMigrate(models...)
}
