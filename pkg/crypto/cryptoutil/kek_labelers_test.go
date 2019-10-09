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

package cryptoutil_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestComponentPrefixKEKLabeler(t *testing.T) {
	for i, tc := range []struct {
		Separator     string
		ReplaceOldNew []string
		Addr          string
		Func          func(context.Context, ComponentPrefixKEKLabeler, string) string
		Expected      string
	}{
		{
			Addr: "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, nil, "")
			},
			Expected: "ns",
		},
		{
			Addr: "",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns/000042",
		},
		{
			Addr: "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, nil, addr)
			},
			Expected: "ns/localhost",
		},
		{
			Addr: "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns/000042/localhost",
		},
		{
			Addr: "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.AsKEKLabel(ctx, addr)
			},
			Expected: "as/localhost",
		},
		{
			Addr: "localhost:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns/000042/localhost",
		},
		{
			Addr: "http://localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns/000042/localhost",
		},
		{
			Addr: "http://localhost:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns/000042/localhost",
		},
		{
			ReplaceOldNew: []string{":", "_"},
			Addr:          "http://[::1]:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns/000042/__1",
		},
		{
			Separator: "_",
			Addr:      "http://localhost:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, &types.NetID{0x00, 0x00, 0x42}, addr)
			},
			Expected: "ns_000042_localhost",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			labeler := ComponentPrefixKEKLabeler{
				Separator:     tc.Separator,
				ReplaceOldNew: tc.ReplaceOldNew,
			}
			label := tc.Func(test.Context(), labeler, tc.Addr)
			a.So(label, should.Equal, tc.Expected)
		})
	}
}
