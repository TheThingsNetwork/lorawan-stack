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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

func WithDB(t *testing.T, f func(t *testing.T, db *gorm.DB)) {
	dbName := fmt.Sprintf("%s_%d", strings.ToLower(t.Name()), time.Now().UnixNano())
	db, err := gorm.Open("postgres", fmt.Sprintf("postgresql://root@localhost:26257/%s?sslmode=disable", dbName))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db = db.Debug()
	if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName)).Error; err != nil {
		panic(err)
	}
	defer db.Exec(fmt.Sprintf("DROP DATABASE %s CASCADE;", dbName))
	f(t, db)
}
