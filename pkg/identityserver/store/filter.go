// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
)

type filterOptionsKeyType struct{}

var filterOptionsKey filterOptionsKeyType

// FilterOptions stores the filter options that are propagated in the context.
type FilterOptions struct {
	Field     string
	Threshold string
}

// WithFilter instructs the store to filter the results by the given field and threshold.
func WithFilter(ctx context.Context, field string, threshold string) context.Context {
	return context.WithValue(ctx, filterOptionsKey, FilterOptions{
		Field:     field,
		Threshold: threshold,
	})
}

// FilterOptionsFromContext returns the filtering options for the query.
func FilterOptionsFromContext(ctx context.Context) (FilterOptions, bool) {
	if opts, ok := ctx.Value(filterOptionsKey).(FilterOptions); ok {
		return opts, true
	}
	return FilterOptions{}, false
}
