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

package migrations

import (
	"context"

	"github.com/jinzhu/gorm"
)

// EUIBlocksIndex removes the unique index constraint from `type` field.
type EUIBlocksIndex struct{}

// Name implements Migration.
func (EUIBlocksIndex) Name() string {
	return "eui_blocks_index"
}

func (EUIBlocksIndex) table() string {
	return "eui_blocks"
}

// Apply implements Migration.
func (m EUIBlocksIndex) Apply(ctx context.Context, db *gorm.DB) error {
	db = db.Table(m.table())
	for _, indexName := range []string{
		"eui_block_index", // The (type) unique index.
	} {
		if !db.Dialect().HasIndex(m.table(), indexName) {
			continue
		}
		if err := db.RemoveIndex(indexName).Error; err != nil {
			return err
		}
	}
	return nil
}

// Rollback implements migration. It recreates the removed "eui_block_index".
func (m EUIBlocksIndex) Rollback(ctx context.Context, db *gorm.DB) error {
	db = db.Table(m.table())
	for indexName, columns := range map[string][]string{
		"eui_block_index": {"type"},
	} {
		if db.Dialect().HasIndex(m.table(), indexName) {
			continue
		}
		if err := db.AddUniqueIndex(indexName, columns...).Error; err != nil {
			return err
		}
	}
	return nil
}
