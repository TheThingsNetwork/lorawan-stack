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

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
)

// SoftDelete makes a Delete operation set a DeletedAt instead of actually deleting the model.
type SoftDelete struct {
	DeletedAt *time.Time `gorm:"index"`
}

func withSoftDeleted() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}
}

func withSoftDeletedIfRequested(ctx context.Context) func(*gorm.DB) *gorm.DB {
	if opts := store.SoftDeletedFromContext(ctx); opts != nil {
		return func(db *gorm.DB) *gorm.DB {
			scope := db.NewScope(db.Value)
			if !scope.HasColumn(deletedAt) {
				return db
			}
			if opts.IncludeDeleted {
				db = db.Unscoped()
			}
			if opts.OnlyDeleted {
				db = db.Where(fmt.Sprintf("%s.deleted_at IS NOT NULL", scope.TableName()))
			}
			if opts.DeletedBefore != nil || opts.DeletedAfter != nil && scope.HasColumn(deletedAt) {
				if opts.DeletedBefore != nil {
					db = db.Where(fmt.Sprintf("%s.deleted_at < ?", scope.TableName()), opts.DeletedBefore)
				}
				if opts.DeletedAfter != nil {
					db = db.Where(fmt.Sprintf("%s.deleted_at > ?", scope.TableName()), opts.DeletedAfter)
				}
			}
			return db
		}
	}
	return func(db *gorm.DB) *gorm.DB { return db }
}
