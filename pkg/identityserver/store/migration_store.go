// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"runtime/trace"

	"github.com/jinzhu/gorm"
)

// GetMigrationStore returns a MigrationStore on the given db (or transaction).
func GetMigrationStore(db *gorm.DB) MigrationStore {
	return &migrationStore{store: newStore(db)}
}

type migrationStore struct {
	*store
}

func (s *migrationStore) CreateMigration(ctx context.Context, migration *Migration) error {
	defer trace.StartRegion(ctx, "create migration").End()
	return s.createEntity(ctx, migration)
}

func (s *migrationStore) FindMigrations(ctx context.Context) ([]*Migration, error) {
	defer trace.StartRegion(ctx, "find migrations").End()
	query := s.query(ctx, Migration{}).Order(orderFromContext(ctx, "migrations", "created_at", "ASC"))
	var models []*Migration
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *migrationStore) GetMigration(ctx context.Context, name string) (*Migration, error) {
	defer trace.StartRegion(ctx, "get migration").End()
	query := s.query(ctx, Migration{})
	var model Migration
	if err := query.Where(Migration{Name: name}).First(&model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errMigrationNotFound.New()
		}
		return nil, err
	}
	return &model, nil
}

func (s *migrationStore) DeleteMigration(ctx context.Context, name string) error {
	defer trace.StartRegion(ctx, "delete migration").End()
	var model Migration
	if err := s.query(ctx, Migration{}).Where(Migration{Name: name}).First(&model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errMigrationNotFound.New()
		}
		return err
	}
	return s.query(ctx, Migration{}).Delete(&model).Error
}
