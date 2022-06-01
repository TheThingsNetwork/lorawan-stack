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
	"strings"
)

type paginationOptionsKeyType struct{}

var paginationOptionsKey paginationOptionsKeyType

// PaginationOptions stores the pagination options that are propagated in the context.
type PaginationOptions struct {
	limit  uint32
	offset uint32
	total  *uint64
}

// WithPagination instructs the store to paginate the results, and set the total
// number of results into total.
func WithPagination(ctx context.Context, limit, page uint32, total *uint64) context.Context {
	if page == 0 {
		page = 1
	}
	return context.WithValue(ctx, paginationOptionsKey, PaginationOptions{
		limit:  limit,
		offset: (page - 1) * limit,
		total:  total,
	})
}

// SetTotal sets the total number of results into the destination set by
// SetTotalCount if not already set.
func SetTotal(ctx context.Context, total uint64) {
	if opts, ok := ctx.Value(paginationOptionsKey).(PaginationOptions); ok && opts.total != nil && *opts.total == 0 {
		*opts.total = total
	}
}

// LimitAndOffsetFromContext gets the limit and offset from the context.
func LimitAndOffsetFromContext(ctx context.Context) (limit, offset uint32) {
	if opts, ok := ctx.Value(paginationOptionsKey).(PaginationOptions); ok {
		return opts.limit, opts.offset
	}
	return 0, 0
}

// WithOrder instructs the store to sort the results by the given field.
// If the field is prefixed with a minus, the order is reversed.
func WithOrder(ctx context.Context, spec string) context.Context {
	if spec == "" {
		return ctx
	}
	field := spec
	direction := "ASC"
	if strings.HasPrefix(spec, "-") {
		field = strings.TrimPrefix(spec, "-")
		direction = "DESC"
	}
	return context.WithValue(ctx, orderOptionsKey, OrderOptions{
		Field:     field,
		Direction: direction,
	})
}

type orderOptionsKeyType struct{}

var orderOptionsKey orderOptionsKeyType

// OrderOptions stores the ordering options that are propagated in the context.
type OrderOptions struct {
	Field     string
	Direction string
}

// OrderOptionsFromContext returns the ordering options for the query.
func OrderOptionsFromContext(ctx context.Context) OrderOptions {
	if opts, ok := ctx.Value(orderOptionsKey).(OrderOptions); ok {
		return opts
	}
	return OrderOptions{}
}

// OrderFromContext returns the ordering string (field and direction) for the query.
// If the context contains ordering options, those are used. Otherwise, the default
// field and order are used.
func OrderFromContext(ctx context.Context, table, defaultTableField, defaultDirection string) string {
	if opts, ok := ctx.Value(orderOptionsKey).(OrderOptions); ok && opts.Field != "" {
		direction := opts.Direction
		if direction == "" {
			direction = "ASC"
		}
		if (table == "organizations" && opts.Field == "organization_id") || (table == "users" && opts.Field == "user_id") {
			table = "accounts"
			opts.Field = "uid"
		}
		tableField := fmt.Sprintf(`"%s"."%s"`, table, opts.Field)
		if tableField != defaultTableField && opts.Field != defaultTableField {
			return fmt.Sprintf(`%s %s, %s %s`, tableField, direction, defaultTableField, direction)
		}
		return fmt.Sprintf(`%s %s`, tableField, direction)
	}
	return fmt.Sprintf("%s %s", defaultTableField, defaultDirection)
}
