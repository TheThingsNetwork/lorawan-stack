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
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func WithDB(t *testing.T, f func(t *testing.T, db *gorm.DB)) {
	dbAddress := os.Getenv("SQL_DB_ADDRESS")
	if dbAddress == "" {
		dbAddress = "localhost:26257"
	}
	dbName := os.Getenv("TEST_DB_NAME")
	randomDB := false
	if dbName == "" {
		dbName = fmt.Sprintf("%s_%d", strings.ToLower(t.Name()), time.Now().UnixNano())
		randomDB = true
	}
	dbConnString := fmt.Sprintf("postgresql://root@%s/%s?sslmode=disable", dbAddress, dbName)
	db, err := gorm.Open("postgres", dbConnString)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	SetLogger(db, test.GetLogger(t))
	db = db.Debug()
	if err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)).Error; err != nil {
		panic(err)
	}
	if randomDB {
		defer db.Exec(fmt.Sprintf("DROP DATABASE %s CASCADE;", dbName))
	}
	f(t, db)
}

func prepareTest(db *gorm.DB, models ...interface{}) {
	db.AutoMigrate(models...)
	if err := clear(db, models...); err != nil {
		panic(err)
	}
}
