// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

type totalKeyType struct{}

var totalKey totalKeyType

// SetTotalCount instructs the store to set the total count of List operations
// into total.
func SetTotalCount(ctx context.Context, total *uint64) context.Context {
	return context.WithValue(ctx, totalKey, total)
}

func limitAndOffsetFromContext(ctx context.Context) (limit uint64, offset uint64) {
	md := rpcmetadata.FromIncomingContext(ctx)
	offset = (md.Page - 1) * md.Limit
	if offset < 0 {
		offset = 0
	}
	return md.Limit, offset
}

// countTotal counts the total number of results (without limiting) and sets it
// into the destination set by SetTotalCount.
func countTotal(ctx context.Context, db *gorm.DB) {
	if dest, ok := ctx.Value(totalKey).(*uint64); ok {
		db.Count(dest)
	}
}

// setTotal sets the total number of results into the destination set by
// SetTotalCount if not already set.
func setTotal(ctx context.Context, total uint64) {
	if dest, ok := ctx.Value(totalKey).(*uint64); ok {
		if *dest == 0 {
			*dest = total
		}
	}
}
