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

type deletedOptionsKeyType struct{}

var deletedOptionsKey deletedOptionsKeyType

type deletedOptions struct {
	IncludeDeleted bool
	OnlyDeleted    bool
}

func withSoftDeletedIfRequested(ctx context.Context) func(*gorm.DB) *gorm.DB {
	if opts, ok := ctx.Value(deletedOptionsKey).(*deletedOptions); ok {
		return func(db *gorm.DB) *gorm.DB {
			if opts.IncludeDeleted || opts.OnlyDeleted {
				db = db.Unscoped()
			}
			scope := db.NewScope(db.Value)
			if opts.OnlyDeleted && scope.HasColumn("deleted_at") {
				db = db.Where(fmt.Sprintf("%s.deleted_at IS NOT NULL", scope.TableName()))
			}
			return db
		}
	}
	return func(db *gorm.DB) *gorm.DB { return db }
}

// WithSoftDeleted returns a context that tells the store to include (only) deleted entities.
func WithSoftDeleted(ctx context.Context, onlyDeleted bool) context.Context {
	return context.WithValue(ctx, deletedOptionsKey, &deletedOptions{
		IncludeDeleted: true,
		OnlyDeleted:    onlyDeleted,
	})
}

// WithoutSoftDeleted returns a context that tells the store not to query for deleted entities.
func WithoutSoftDeleted(ctx context.Context) context.Context {
	return context.WithValue(ctx, deletedOptionsKey, &deletedOptions{})
}
