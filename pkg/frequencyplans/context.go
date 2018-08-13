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

// Package frequencyplans contains abstractions to fetch and manipulate frequency plans.
package frequencyplans

import "context"

type fallbackKeyType struct{}

var fallbackKey fallbackKeyType

// WithFallbackID returns a derived context with the given frequency plan ID to be used as fallback.
func WithFallbackID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, fallbackKey, id)
}

// FallbackIDFromContext returns the fallback frequency plan ID and whether it's set using WithFallbackID.
func FallbackIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(fallbackKey).(string)
	return id, ok
}
