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

package redis

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
)

type paginationOptionsKeyType struct{}

var paginationOptionsKey paginationOptionsKeyType

type paginationOptions struct {
	limit  int64
	offset int64
	total  *int64
}

// NewContextWithPagination instructs the store to paginate the results.
func NewContextWithPagination(ctx context.Context, limit, page int64, total *int64) context.Context {
	md := rpcmetadata.FromIncomingContext(ctx)
	if limit == 0 && md.Limit != 0 {
		limit = int64(md.Limit)
	}
	if page == 0 && md.Page != 0 {
		page = int64(md.Page)
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

// SetPaginationTotal sets the total number of results inside the paginated context, if it was not set already.
func SetPaginationTotal(ctx context.Context, total int64) {
	if opts, ok := ctx.Value(paginationOptionsKey).(paginationOptions); ok && opts.total != nil && *opts.total == 0 {
		*opts.total = total
	}
}

// PaginationLimitAndOffsetFromContext returns the pagination limit and the offset if they are present.
func PaginationLimitAndOffsetFromContext(ctx context.Context) (limit, offset int64) {
	if opts, ok := ctx.Value(paginationOptionsKey).(paginationOptions); ok {
		return opts.limit, opts.offset
	}
	return
}
