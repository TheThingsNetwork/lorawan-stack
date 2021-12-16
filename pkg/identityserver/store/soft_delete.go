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
	"time"
)

type deletedOptionsKeyType struct{}

var deletedOptionsKey deletedOptionsKeyType

type DeletedOptions struct {
	IncludeDeleted bool
	OnlyDeleted    bool
}

// WithSoftDeleted returns a context that tells the store to include (only) deleted entities.
func WithSoftDeleted(ctx context.Context, onlyDeleted bool) context.Context {
	return context.WithValue(ctx, deletedOptionsKey, &DeletedOptions{
		IncludeDeleted: true,
		OnlyDeleted:    onlyDeleted,
	})
}

// WithoutSoftDeleted returns a context that tells the store not to query for deleted entities.
func WithoutSoftDeleted(ctx context.Context) context.Context {
	return context.WithValue(ctx, deletedOptionsKey, &DeletedOptions{})
}

func SoftDeletedFromContext(ctx context.Context) *DeletedOptions {
	if opts, ok := ctx.Value(deletedOptionsKey).(*DeletedOptions); ok {
		return opts
	}
	return nil
}

type expiredOptionsKeyType struct{}

var expiredOptionsKey expiredOptionsKeyType

type ExpiredOptions struct {
	OnlyExpired      bool
	RestoreThreshold time.Duration
}

// WithExpired returns a context that tells the store to only query expired entities.
func WithExpired(ctx context.Context, threshold time.Duration) context.Context {
	return context.WithValue(ctx, expiredOptionsKey, ExpiredOptions{
		OnlyExpired:      true,
		RestoreThreshold: threshold,
	})
}

func ExpiredFromContext(ctx context.Context) (onlyExpired bool, restoreThreshold time.Duration) {
	if opts, ok := ctx.Value(expiredOptionsKey).(ExpiredOptions); ok {
		return opts.OnlyExpired, opts.RestoreThreshold
	}
	return
}
