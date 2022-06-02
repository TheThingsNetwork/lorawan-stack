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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Account is the account model in the database.
type Account struct {
	bun.BaseModel `bun:"table:accounts,alias:acc"`

	Model
	SoftDelete

	UID string `bun:"uid,notnull"`

	AccountID   string `bun:"account_id,notnull"`
	AccountType string `bun:"account_type,notnull"` // user or organization
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Account) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

// EmbeddedAccount contains the account fields that are embedded into User and Organization
// when loading them through the user_accounts and organization_accounts views.
type EmbeddedAccount struct {
	ID string `bun:"id,notnull,scanonly"`

	CreatedAt time.Time  `bun:"created_at,notnull,scanonly"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,scanonly"`
	DeletedAt *time.Time `bun:"deleted_at,scanonly"`

	UID string `bun:"uid,notnull,scanonly"`
}

func selectWithEmbeddedAccountUID(ctx context.Context, id ttnpb.IDStringer) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		return q.Where("?TableAlias.account_uid = ?", id.IDString())
	}
}

func selectWithEmbeddedAccountUIDs[X ttnpb.IDStringer](ctx context.Context, ids ...X) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		if len(ids) == 0 {
			return q
		}
		return q.Where("?TableAlias.account_uid IN (?)", idStrings(ids...))
	}
}
