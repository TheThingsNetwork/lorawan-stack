// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package gateways_test

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/gateways/ttgc"
	dcstypes "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestUpstream(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)

	// Invalid ranges.
	ranges := map[string][]string{"ttgc": {"&S(FU*)"}}
	euiPrefixes, err := gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.NotBeNil)
	a.So(euiPrefixes, should.BeEmpty)

	ranges = map[string][]string{"ttgc": {"58A0CBFFFE800000"}}
	euiPrefixes, err = gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.NotBeNil)
	a.So(euiPrefixes, should.BeEmpty)

	ranges = map[string][]string{"ttgc": {"58A0CBFFFE800000/123456"}}
	euiPrefixes, err = gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.NotBeNil)
	a.So(euiPrefixes, should.BeEmpty)

	ranges = map[string][]string{"ttgc": {"58A0CBFFFE800000-58A0CBFFFE800000-58A0CBFFFE800000"}}
	euiPrefixes, err = gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.NotBeNil)
	a.So(euiPrefixes, should.BeEmpty)

	ranges = map[string][]string{"ttgc": {"001616FFFEWXUSD-001616FFFETGENDE"}}
	euiPrefixes, err = gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.NotBeNil)
	a.So(euiPrefixes, should.BeEmpty)

	ranges = map[string][]string{"ttgc": {"001616FFFE42DFAD-001616FFFETGENDE"}}
	euiPrefixes, err = gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.NotBeNil)
	a.So(euiPrefixes, should.BeEmpty)

	// Valid Configuration
	ranges = map[string][]string{
		"ttgc": {
			"58A0CBFFFE800000/48",
			"001616FFFE42DFAD-001616FFFE42E395",
		},
	}
	euiPrefixes, err = gateways.ParseGatewayEUIRanges(ranges)
	a.So(err, should.BeNil)
	a.So(euiPrefixes, should.Resemble, map[string][]dcstypes.EUI64Range{
		"ttgc": {
			dcstypes.RangeFromEUI64Prefix(types.EUI64Prefix{
				EUI64:  types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x00},
				Length: 48,
			}),
			dcstypes.RangeFromEUI64Range(
				types.EUI64{0x00, 0x16, 0x16, 0xff, 0xfe, 0x42, 0xdf, 0xad},
				types.EUI64{0x00, 0x16, 0x16, 0xff, 0xfe, 0x42, 0xe3, 0x95},
			),
		},
	})

	// Invalid configurations
	config := gateways.Config{
		Upstreams: map[string][]string{"ttgc": {"&S(FU*)"}},
		TTGC:      ttgc.Config{},
	}
	upstream, err := gateways.NewUpstream(ctx, config)
	a.So(errors.IsInvalidArgument(err), should.BeTrue)
	a.So(upstream, should.BeNil)

	config = gateways.Config{
		Upstreams: map[string][]string{"unsupported": {"58A0CBFFFE800000/48"}},
		TTGC:      ttgc.Config{},
	}
	upstream, err = gateways.NewUpstream(ctx, config)
	a.So(errors.IsInvalidArgument(err), should.BeTrue)
	a.So(upstream, should.BeNil)

	// Valid Configuration
	config = gateways.Config{
		Upstreams: map[string][]string{"ttgc": {"58A0CBFFFE800000/48"}},
		TTGC:      ttgc.Config{},
	}
	upstream, err = gateways.NewUpstream(ctx, config)
	a.So(err, should.BeNil)
	a.So(upstream, should.NotBeNil)

	// Invalid EUI
	claimer := upstream.Claimer(types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x81, 0x00, 0x00})
	a.So(claimer, should.BeNil)

	// Valid EUI
	claimer = upstream.Claimer(types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x1B})
	a.So(claimer, should.NotBeNil)
}
