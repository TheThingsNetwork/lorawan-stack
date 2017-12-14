// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
)

type claimsKey struct{}

// NewContext returns a new context with the given claims.
func NewContext(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// FromContext returns the claims from the context if present, otherwise returns
// empty claims.
func FromContext(ctx context.Context) *auth.Claims {
	c, ok := ctx.Value(claimsKey{}).(*auth.Claims)
	if !ok {
		c = new(auth.Claims)
	}
	return c
}
