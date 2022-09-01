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

package store

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
)

var pgdriverFeatureFlag = experimental.DefineFeature("is.pgdriver", false)

// OpenDB opens the database connection.
func OpenDB(ctx context.Context, databaseURI string) (*sql.DB, error) {
	if pgdriverFeatureFlag.GetValue(ctx) {
		return sql.OpenDB(pgdriver.NewConnector(
			pgdriver.WithDSN(databaseURI),
			pgdriver.WithResetSessionFunc(func(ctx context.Context, cn *pgdriver.Conn) error {
				return checkConn(cn.Conn())
			}),
		)), nil
	}
	config, err := pgx.ParseConfig(databaseURI)
	if err != nil {
		return nil, err
	}
	config.PreferSimpleProtocol = true
	return stdlib.OpenDB(*config, stdlib.OptionResetSession(func(ctx context.Context, cn *pgx.Conn) error {
		return checkConn(cn.PgConn().Conn())
	})), nil
}
