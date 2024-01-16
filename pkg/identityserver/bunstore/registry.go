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
	"context"

	"github.com/uptrace/bun"
)

var models []any

func registerModels(m ...any) {
	models = append(models, m...)
}

func init() {
	registerModels(
		&AccessToken{},
		&Account{},
		&APIKey{},
		&Application{},
		&Attribute{},
		&AuthorizationCode{},
		&ClientAuthorization{},
		&Client{},
		&ContactInfoValidation{},
		&ContactInfo{},
		&EmailValidation{},
		&EndDeviceLocation{},
		&EndDevice{},
		&EUIBlock{},
		&GatewayAntenna{},
		&Gateway{},
		&Invitation{},
		&LoginToken{},
		&Membership{},
		&NotificationReceiver{},
		&Notification{},
		&Organization{},
		&Picture{},
		&UserSession{},
		&User{},
	)
}

// clear database tables for the given models.
// This should be used with caution.
func clear(ctx context.Context, db *bun.DB, models ...any) (err error) {
	md, err := getDBMetadata(ctx, db)
	if err != nil {
		return err
	}

	if md.Type == "CockroachDB" {
		if _, err = db.ExecContext(ctx, "SET SQL_SAFE_UPDATES = FALSE"); err != nil {
			return err
		}
	}
	for _, model := range models {
		if _, err = db.NewDelete().Model(model).Where("1=1").ForceDelete().Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Clear database tables for all registered models.
// This should be used with caution.
func Clear(ctx context.Context, db *bun.DB) error {
	return clear(ctx, db, models...)
}
