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
	"fmt"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
)

func noopSelectModifier(q *bun.SelectQuery) *bun.SelectQuery { return q }

func selectWithLimitAndOffsetFromContext(ctx context.Context) func(*bun.SelectQuery) *bun.SelectQuery {
	limit, offset := store.LimitAndOffsetFromContext(ctx)
	if limit == 0 {
		return noopSelectModifier
	}
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Limit(int(limit)).Offset(int(offset))
	}
}

func selectWithOrderFromContext(
	ctx context.Context, defaultColumn string, fieldToColumn map[string]string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	order := store.OrderOptionsFromContext(ctx)
	if column, ok := fieldToColumn[order.Field]; ok {
		return func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order(
				fmt.Sprintf("%s %s", column, order.Direction),
				fmt.Sprintf("%s %s", defaultColumn, order.Direction),
			)
		}
	}
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order(fmt.Sprintf("%s ASC", defaultColumn))
	}
}
