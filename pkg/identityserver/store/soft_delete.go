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
)

type deletedOptionsKeyType struct{}

var deletedOptionsKey deletedOptionsKeyType

// DeletedOptions stores the options for selecting deleted entities.
type DeletedOptions struct {
	IncludeDeleted bool
	OnlyDeleted    bool
	DeletedBefore  *time.Time
	DeletedAfter   *time.Time
}

// WithSoftDeleted returns a context that tells the store to include (only) deleted entities.
func WithSoftDeleted(ctx context.Context, onlyDeleted bool) context.Context {
	return context.WithValue(ctx, deletedOptionsKey, &DeletedOptions{
		IncludeDeleted: true,
		OnlyDeleted:    onlyDeleted,
	})
}

// WithSoftDeletedBetween returns a context that tells the store to include deleted entities
// between (exclusive) the given times.
func WithSoftDeletedBetween(ctx context.Context, deletedAfter, deletedBefore *time.Time) context.Context {
	return context.WithValue(ctx, deletedOptionsKey, &DeletedOptions{
		IncludeDeleted: true,
		OnlyDeleted:    deletedBefore != nil || deletedAfter != nil,
		DeletedBefore:  deletedBefore,
		DeletedAfter:   deletedAfter,
	})
}

// SoftDeletedFromContext returns the DeletedOptions from the context if present.
func SoftDeletedFromContext(ctx context.Context) *DeletedOptions {
	if opts, ok := ctx.Value(deletedOptionsKey).(*DeletedOptions); ok {
		return opts
	}
	return nil
}
