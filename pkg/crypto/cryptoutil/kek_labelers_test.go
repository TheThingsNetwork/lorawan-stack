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

func netIDPtr(netID types.NetID) *types.NetID { return &netID }

func TestComponentPrefixKEKLabeler(t *testing.T) {
	for i, tc := range []struct {
		Separator,
		Addr string
		Func     func(context.Context, ComponentPrefixKEKLabeler, string) string
		Expected string
	}{
		{
			Separator: "",
			Addr:      "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, nil, "")
			},
			Expected: "ns",
		},
		{
			Separator: "",
			Addr:      "",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, netIDPtr(types.NetID{0x00, 0x00, 0x42}), addr)
			},
			Expected: "ns:000042",
		},
		{
			Separator: "",
			Addr:      "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, nil, addr)
			},
			Expected: "ns:localhost",
		},
		{
			Separator: "",
			Addr:      "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, netIDPtr(types.NetID{0x00, 0x00, 0x42}), addr)
			},
			Expected: "ns:000042:localhost",
		},
		{
			Separator: "",
			Addr:      "localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.AsKEKLabel(ctx, addr)
			},
			Expected: "as:localhost",
		},
		{
			Separator: "",
			Addr:      "localhost:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, netIDPtr(types.NetID{0x00, 0x00, 0x42}), addr)
			},
			Expected: "ns:000042:localhost",
		},
		{
			Separator: "",
			Addr:      "http://localhost",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, netIDPtr(types.NetID{0x00, 0x00, 0x42}), addr)
			},
			Expected: "ns:000042:localhost",
		},
		{
			Separator: "",
			Addr:      "http://localhost:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, netIDPtr(types.NetID{0x00, 0x00, 0x42}), addr)
			},
			Expected: "ns:000042:localhost",
		},
		{
			Separator: "_",
			Addr:      "http://localhost:1234",
			Func: func(ctx context.Context, labeler ComponentPrefixKEKLabeler, addr string) string {
				return labeler.NsKEKLabel(ctx, netIDPtr(types.NetID{0x00, 0x00, 0x42}), addr)
			},
			Expected: "ns_000042_localhost",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			labeler := ComponentPrefixKEKLabeler{
				Separator: tc.Separator,
			}
			label := tc.Func(test.Context(), labeler, tc.Addr)
			a.So(label, should.Equal, tc.Expected)
		})
	}
}
