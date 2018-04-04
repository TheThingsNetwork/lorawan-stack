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

package identityserver

import "context"

type claimsKey struct{}

// newContextWithClaims returns a new context but with the provided claims.
func newContextWithClaims(ctx context.Context, c *claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

// claimsFromContext returns the claims from the context.
// If not found it returns empty claims.
func claimsFromContext(ctx context.Context) *claims {
	c, ok := ctx.Value(claimsKey{}).(*claims)
	if !ok {
		return new(claims)
	}
	return c
}
