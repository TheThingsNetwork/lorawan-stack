// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import "context"

type claimsKey struct{}

// NewContext returns a new context but with the provided claims.
func NewContext(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

// FromContext returns the claims from the context. If not found it returns
// empty claims.
func FromContext(ctx context.Context) *Claims {
	c, ok := ctx.Value(claimsKey{}).(*Claims)
	if !ok {
		return new(Claims)
	}
	return c
}
