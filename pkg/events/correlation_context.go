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

package events

import (
	"context"
	"crypto/rand"

	"github.com/oklog/ulid"
)

type correlationKeyType struct{}

var correlationKey = &correlationKeyType{}

// ContextWithCorrelationID returns a derived context with the correlation IDs if they were't already in there.
func ContextWithCorrelationID(ctx context.Context, cid ...string) context.Context {
	if v := ctx.Value(correlationKey); v != nil {
		if existing, ok := v.([]string); ok {
			for _, cid := range cid {
				for _, existing := range existing {
					if cid == existing {
						return ctx // Correlation ID already in context, just return the original context.
					}
				}
				return context.WithValue(ctx, correlationKey, append(existing, cid)) // Correlation ID was not yet in the context; add cid.
			}
		}
	}
	return context.WithValue(ctx, correlationKey, cid) // Empty (or invalid) context; add cid.
}

// CorrelationIDsFromContext returns the correlation IDs that are attached to the context.
func CorrelationIDsFromContext(ctx context.Context) []string {
	if v := ctx.Value(correlationKey); v != nil {
		if cids, ok := v.([]string); ok {
			return cids
		}
	}
	return nil
}

// ContextWithEnsuredCorrelationID ensures there is at least one correlation ID set in the context.
// The returned context will either be the original context (if an ID was set)
// or a derived context with a random correlation ID.
func ContextWithEnsuredCorrelationID(ctx context.Context) context.Context {
	if v := ctx.Value(correlationKey); v != nil {
		if cids, ok := v.([]string); ok && len(cids) > 0 {
			return ctx
		}
	}
	return ContextWithCorrelationID(ctx, NewCorrelationID())
}

// NewCorrelationID returns a new random correlation ID.
func NewCorrelationID() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}
