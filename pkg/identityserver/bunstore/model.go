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
	"time"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
)

// UUID can be embedded in models that should have a UUID field.
type UUID struct {
	ID string `bun:"id,type:uuid,pk,notnull,default:gen_random_uuid()"`
}

// Model is the base model for most of our types.
type Model struct {
	UUID

	CreatedAt time.Time `bun:"created_at,notnull"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Model) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if m.CreatedAt.IsZero() {
			m.CreatedAt = now()
		}
		if m.UpdatedAt.IsZero() {
			m.UpdatedAt = now()
		}
	case *bun.UpdateQuery:
		m.UpdatedAt = now()
	}
	return nil
}

// SoftDelete can be embedded in models that should have soft delete functionality.
type SoftDelete struct {
	DeletedAt *time.Time `bun:"deleted_at,soft_delete"`
}

func selectWithSoftDeletedFromContext(ctx context.Context) func(*bun.SelectQuery) *bun.SelectQuery {
	opts := store.SoftDeletedFromContext(ctx)
	if opts == nil {
		return noopSelectModifier
	}
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if opts.IncludeDeleted {
			q = q.WhereAllWithDeleted()
		}
		if opts.OnlyDeleted {
			q = q.Where("?TableAlias.deleted_at IS NOT NULL")
		}
		if opts.DeletedBefore != nil || opts.DeletedAfter != nil {
			if opts.DeletedBefore != nil {
				q = q.Where("?TableAlias.deleted_at < ?", opts.DeletedBefore)
			}
			if opts.DeletedAfter != nil {
				q = q.Where("?TableAlias.deleted_at > ?", opts.DeletedAfter)
			}
		}
		return q
	}
}
