// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func equalTime(a, b *time.Time) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return false
	}
	return a.Equal(*b)
}

func idStrings[X ttnpb.IDStringer](in ...X) []string {
	out := make([]string, len(in))
	for i, x := range in {
		out[i] = x.IDString()
	}
	return out
}

func noopSelectModifier(q *bun.SelectQuery) *bun.SelectQuery { return q }

func selectWithLimitAndOffsetFromContext(ctx context.Context) func(*bun.SelectQuery) *bun.SelectQuery {
	limit, offset := store.LimitAndOffsetFromContext(ctx)
	if limit == 0 {
		return noopSelectModifier
	}
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Limit(int(limit)).Offset(int(offset))
	}
}

func selectWithOrderFromContext(
	ctx context.Context, defaultColumn string, fieldToColumn map[string]string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	order := store.OrderOptionsFromContext(ctx)
	if column, ok := fieldToColumn[order.Field]; ok {
		return func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order(
				fmt.Sprintf("%s %s", column, order.Direction),
				fmt.Sprintf("%s %s", defaultColumn, order.Direction),
			)
		}
	}
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order(fmt.Sprintf("%s ASC", defaultColumn))
	}
}

func convertIntSlice[A, B int | ~int32](in []A) []B {
	out := make([]B, len(in))
	for i, el := range in {
		out[i] = B(el)
	}
	return out
}

func secretFromBytes(b []byte) *ttnpb.Secret {
	if len(b) == 0 {
		return nil
	}
	blocks := bytes.SplitN(b, []byte(":"), 2)
	if len(blocks) == 2 {
		return &ttnpb.Secret{
			KeyId: string(blocks[0]),
			Value: blocks[1],
		}
	}
	return nil
}

func secretToBytes(secret *ttnpb.Secret) []byte {
	if secret == nil {
		return nil
	}
	var buf bytes.Buffer
	buf.WriteString(secret.KeyId)
	buf.WriteByte(':')
	buf.Write(secret.Value)
	return buf.Bytes()
}

func eui64FromString(s *string) *types.EUI64 {
	if s == nil {
		return nil
	}
	var eui types.EUI64
	if err := eui.UnmarshalText([]byte(*s)); err != nil {
		return nil
	}
	return &eui
}

func eui64ToString(eui *types.EUI64) *string {
	if eui == nil {
		return nil
	}
	s := eui.String()
	return &s
}
