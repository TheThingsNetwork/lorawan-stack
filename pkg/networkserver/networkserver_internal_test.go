// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNewDevAddr(t *testing.T) {
	test.RunSubtest(t, test.SubtestConfig{
		Name:     "From NetID",
		Parallel: true,
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			ns, ctx, _, stop := StartTest(
				t,
				TestConfig{
					Context: ctx,
					NetworkServer: Config{
						NetID: types.NetID{0x00, 0x00, 0x13},
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
					),
				},
			)
			defer stop()

			a.So(ns.newDevAddr(ctx, nil).HasPrefix(types.DevAddrPrefix{
				DevAddr: types.DevAddr{0x26, 0, 0, 0},
				Length:  7,
			}), should.BeTrue)
		},
	})

	test.RunSubtest(t, test.SubtestConfig{
		Name:     "Configured DevAddr prefixes",
		Parallel: true,
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			ps := []types.DevAddrPrefix{
				{
					DevAddr: types.DevAddr{0x26, 0x01, 0x00, 0x00},
					Length:  16,
				},
				{
					DevAddr: types.DevAddr{0x26, 0xff, 0x01, 0x00},
					Length:  24,
				},
				{
					DevAddr: types.DevAddr{0x27, 0x00, 0x00, 0x00},
					Length:  8,
				},
			}
			ns, ctx, _, stop := StartTest(
				t,
				TestConfig{
					Context: ctx,
					NetworkServer: Config{
						NetID:           types.NetID{0x00, 0x00, 0x13},
						DevAddrPrefixes: ps,
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
					),
				},
			)
			defer stop()

			seen := map[types.DevAddrPrefix]int{}
			for i := 0; i < 100; i++ {
				devAddr := ns.newDevAddr(ctx, nil)
				for _, p := range ps {
					if devAddr.HasPrefix(p) {
						seen[p]++
						break
					}
				}
			}

			a.So(seen[ps[0]], should.BeGreaterThan, 0)
			a.So(seen[ps[1]], should.BeGreaterThan, 0)
			a.So(seen[ps[2]], should.BeGreaterThan, 0)
		},
	})
}
