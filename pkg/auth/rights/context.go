// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rights

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type rightsKey int

const key rightsKey = 1

// NewContext returns ctx with the given rights within.
func NewContext(ctx context.Context, rights []ttnpb.Right) context.Context {
	return context.WithValue(ctx, key, rights)
}

// FromContext returns the rights from ctx, otherwise empty slice if they are not found.
func FromContext(ctx context.Context) []ttnpb.Right {
	if r, ok := ctx.Value(key).([]ttnpb.Right); ok {
		return r
	}
	return make([]ttnpb.Right, 0)
}
