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

package rights

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type rightsKeyType struct{}

var rightsKey rightsKeyType

type rightsInfo struct {
	uid    string
	rights []ttnpb.Right
}

func newContext(ctx context.Context, uid string, rights []ttnpb.Right) context.Context {
	return context.WithValue(ctx, rightsKey, rightsInfo{uid: uid, rights: rights})
}

// NewContext returns ctx with the given rights within.
func NewContext(ctx context.Context, rights []ttnpb.Right) context.Context {
	return newContext(ctx, "", rights)
}

func fromContext(ctx context.Context) rightsInfo {
	if r, ok := ctx.Value(rightsKey).(rightsInfo); ok {
		return r
	}
	return rightsInfo{}
}

// FromContext returns the rights from ctx, otherwise empty slice if they are not found.
func FromContext(ctx context.Context) []ttnpb.Right {
	rights := fromContext(ctx).rights
	if rights == nil {
		rights = []ttnpb.Right{}
	}
	return rights
}
