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

	"github.com/uptrace/bun"
)

func applyContext(ctx context.Context, q *bun.SelectQuery, model interface{ _isModel() }) *bun.SelectQuery {
	q = q.Apply(selectWithContext(ctx))
	if _, ok := model.(interface{ _isSoftDelete() }); ok {
		q = q.Apply(selectWithSoftDeletedFromContext(ctx))
	}
	return q
}

func newSelectModel[X interface{ _isModel() }](ctx context.Context, db bun.IDB, model X) *bun.SelectQuery {
	q := db.NewSelect().Model(model)
	return applyContext(ctx, q, model)
}

func newSelectModels[X interface{ _isModel() }](ctx context.Context, db bun.IDB, models *[]X) *bun.SelectQuery {
	var model X
	q := db.NewSelect().Model(models)
	return applyContext(ctx, q, model)
}

func (s *baseStore) newSelectModel(ctx context.Context, model interface{ _isModel() }) *bun.SelectQuery {
	return newSelectModel(ctx, s.DB, model)
}

// TODO: Once generic methods are possible, implement a `func (s *baseStore) newSelectModels`.
