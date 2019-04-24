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
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

var (
	setup        sync.Once
	dbConnString string
)

func WithDB(t *testing.T, f func(t *testing.T, db *gorm.DB)) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	setup.Do(func() {
		dbAddress := os.Getenv("SQL_DB_ADDRESS")
		if dbAddress == "" {
			dbAddress = "localhost:26257"
		}
		dbName := os.Getenv("TEST_DATABASE_NAME")
		if dbName == "" {
			dbName = "ttn_lorawan_is_store_test"
		}
		dbConnString = fmt.Sprintf("postgresql://root@%s/%s?sslmode=disable", dbAddress, dbName)
		db, err := Open(test.Context(), dbConnString)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		if err = Initialize(db); err != nil {
			panic(err)
		}
		if err = AutoMigrate(db).Error; err != nil {
			panic(err)
		}
		if err = Clear(db); err != nil {
			panic(err)
		}
	})
	db, err := Open(ctx, dbConnString)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db = db.Debug()
	f(t, db)
}

func prepareTest(db *gorm.DB, models ...interface{}) {
	if err := clear(db, models...); err != nil {
		panic(err)
	}
}
