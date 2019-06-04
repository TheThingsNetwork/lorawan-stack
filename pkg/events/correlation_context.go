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

package events

import (
	"context"
	"crypto/rand"
	"sort"

	"github.com/oklog/ulid"
)

type correlationKey struct{}

// ContextWithCorrelationID returns a derived context with the correlation IDs if they were not already in there.
func ContextWithCorrelationID(ctx context.Context, cids ...string) context.Context {
	cids = append(cids[:0:0], cids...)
	sort.Strings(cids)

	existing, ok := ctx.Value(correlationKey{}).([]string)
	if !ok {
		return context.WithValue(ctx, correlationKey{}, cids)
	}
	return context.WithValue(ctx, correlationKey{}, mergeStrings(existing, cids))
}

// CorrelationIDsFromContext returns the correlation IDs that are attached to the context.
func CorrelationIDsFromContext(ctx context.Context) []string {
	cids, ok := ctx.Value(correlationKey{}).([]string)
	if !ok {
		return nil
	}
	return cids
}

// NewCorrelationID returns a new random correlation ID.
func NewCorrelationID() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}

// mergeStrings merges 2 sorted string slices and returns the resulting slice
// See https://en.wikipedia.org/wiki/Merge_sort
func mergeStrings(a, b []string) []string {
	merged := make([]string, 0, len(a)+len(b))
	var i, j int
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			merged = append(merged, a[i])
			i++
		} else if a[i] > b[j] {
			merged = append(merged, b[j])
			j++
		} else {
			merged = append(merged, a[i])
			i++
			j++
		}
	}
	if i < len(a) {
		merged = append(merged, a[i:]...)
	} else if j < len(b) {
		merged = append(merged, b[j:]...)
	}
	return merged
}
