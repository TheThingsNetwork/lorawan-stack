// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
