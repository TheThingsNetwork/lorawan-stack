// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
)

type paginationOptionsKeyType struct{}

var paginationOptionsKey paginationOptionsKeyType

type paginationOptions struct {
	limit  uint32
	offset uint32
	total  *uint64
}

// WithPagination instructs the store to paginate the results, and set the total
// number of results into total.
func WithPagination(ctx context.Context, limit, page uint32, total *uint64) context.Context {
	md := rpcmetadata.FromIncomingContext(ctx)
	if limit == 0 && md.Limit != 0 {
		limit = uint32(md.Limit)
	}
	if page == 0 && md.Page != 0 {
		page = uint32(md.Page)
	}
	if page == 0 {
		page = 1
	}
	return context.WithValue(ctx, paginationOptionsKey, paginationOptions{
		limit:  limit,
		offset: (page - 1) * limit,
		total:  total,
	})
}

// countTotal counts the total number of results (without limiting) and sets it
// into the destination set by SetTotalCount.
func countTotal(ctx context.Context, db *gorm.DB) {
	if opts, ok := ctx.Value(paginationOptionsKey).(paginationOptions); ok && opts.total != nil && *opts.total == 0 {
		db.Count(opts.total)
	}
}

// setTotal sets the total number of results into the destination set by
// SetTotalCount if not already set.
func setTotal(ctx context.Context, total uint64) {
	if opts, ok := ctx.Value(paginationOptionsKey).(paginationOptions); ok && opts.total != nil && *opts.total == 0 {
		*opts.total = total
	}
}

func limitAndOffsetFromContext(ctx context.Context) (limit, offset uint32) {
	if opts, ok := ctx.Value(paginationOptionsKey).(paginationOptions); ok {
		return opts.limit, opts.offset
	}
	return
}
