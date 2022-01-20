// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EntityContacts sets the contacts of entities to their first collaborator.
type EntityContacts struct{}

// Name implements Migration.
func (EntityContacts) Name() string {
	return "entity_contacts"
}

//go:embed "entity_contacts.sql.tmpl"
var entityContactsQuery string

// Apply implements Migration.
func (m EntityContacts) Apply(ctx context.Context, db *gorm.DB) error {
	var (
		sqlDB  = db.CommonDB()
		res    sql.Result
		err    error
		rows   int64
		logger = log.FromContext(ctx)
	)

	res, err = sqlDB.Exec(fmt.Sprintf(entityContactsQuery, "application", ttnpb.Right_RIGHT_APPLICATION_ALL))
	if err != nil {
		return err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	logger.WithField("rows", rows).Debug("updated application contacts")

	res, err = sqlDB.Exec(fmt.Sprintf(entityContactsQuery, "client", ttnpb.Right_RIGHT_CLIENT_ALL))
	if err != nil {
		return err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	logger.WithField("rows", rows).Debug("updated client contacts")

	res, err = sqlDB.Exec(fmt.Sprintf(entityContactsQuery, "gateway", ttnpb.Right_RIGHT_GATEWAY_ALL))
	if err != nil {
		return err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	logger.WithField("rows", rows).Debug("updated gateway contacts")

	res, err = sqlDB.Exec(fmt.Sprintf(entityContactsQuery, "organization", ttnpb.Right_RIGHT_ORGANIZATION_ALL))
	if err != nil {
		return err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	logger.WithField("rows", rows).Debug("updated organization contacts")

	return nil
}

// Rollback implements migration. For the EntityContacts migration this is a no-op.
func (m EntityContacts) Rollback(ctx context.Context, db *gorm.DB) error {
	return nil
}

func init() {
	All = append(All, EntityContacts{})
}
