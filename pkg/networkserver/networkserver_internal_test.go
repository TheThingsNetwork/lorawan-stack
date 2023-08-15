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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
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
				ctx,
				TestConfig{
					NetworkServer: Config{
						NetID: types.NetID{0x00, 0x00, 0x13},
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
						DownlinkDispatchTaskName,
					),
					Component: component.Config{
						ServiceBase: config.ServiceBase{
							FrequencyPlans: config.FrequencyPlansConfig{
								ConfigSource: "static",
								Static:       test.StaticFrequencyPlans,
							},
						},
					},
				},
			)
			defer stop()

			a.So(ns.newDevAddr(ctx).HasPrefix(types.DevAddrPrefix{
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
				ctx,
				TestConfig{
					NetworkServer: Config{
						NetID:           types.NetID{0x00, 0x00, 0x13},
						DevAddrPrefixes: ps,
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
						DownlinkDispatchTaskName,
					),
					Component: component.Config{
						ServiceBase: config.ServiceBase{
							FrequencyPlans: config.FrequencyPlansConfig{
								ConfigSource: "static",
								Static:       test.StaticFrequencyPlans,
							},
						},
					},
				},
			)
			defer stop()

			seen := map[types.DevAddrPrefix]int{}
			for i := 0; i < 100000; i++ {
				devAddr := ns.newDevAddr(ctx)
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

func TestMakeNewDevAddrFunc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name     string
		Prefixes []types.DevAddrPrefix
		Balance  []float64
	}{
		{
			Name: "single /32",
			Prefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.MinDevAddr,
					Length:  32,
				},
			},
			Balance: []float64{
				1.0,
			},
		},
		{
			Name: "two /32",
			Prefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.MinDevAddr,
					Length:  32,
				},
				{
					DevAddr: types.MaxDevAddr,
					Length:  32,
				},
			},
			Balance: []float64{
				1.0 / 2.0,
				1.0 / 2.0,
			},
		},
		{
			Name: "three /32",
			Prefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.MinDevAddr,
					Length:  32,
				},
				{
					DevAddr: types.DevAddr{0x01, 0x00, 0x00, 0x00},
					Length:  32,
				},
				{
					DevAddr: types.MaxDevAddr,
					Length:  32,
				},
			},
			Balance: []float64{
				1.0 / 3.0,
				1.0 / 3.0,
				1.0 / 3.0,
			},
		},
		{
			Name: "one /24 and one /28",
			Prefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.MinDevAddr,
					Length:  24,
				},
				{
					DevAddr: types.MaxDevAddr,
					Length:  28,
				},
			},
			// There are 2^4=16 more /24 addresses than /28 addresses.
			Balance: []float64{
				1.0 - (1.0 / 16.0),
				1.0 / 16.0,
			},
		},
		{
			Name: "one /24 and two /28",
			Prefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.MinDevAddr,
					Length:  24,
				},
				{
					DevAddr: types.DevAddr{0x01, 0x00, 0x00, 0x00},
					Length:  24,
				},
				{
					DevAddr: types.MaxDevAddr,
					Length:  28,
				},
			},
			// There are 256 /24 possible addresses, and 16 /28 possible addresses.
			Balance: []float64{
				1.0 - 256.0/(256.0+256.0+16.0) - 16.0/(256.0+256.0+16.0),
				1.0 - 256.0/(256.0+256.0+16.0) - 16.0/(256.0+256.0+16.0),
				16.0 / (256.0 + 256.0 + 16.0),
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, ctx := test.New(t)
			newF := makeNewDevAddrFunc(tc.Prefixes...)
			weights, total := make([]int, len(tc.Prefixes)), 0
			for i := 0; i < 100000; i++ {
				devAddr := newF(ctx)
				found := false
				for j, prefix := range tc.Prefixes {
					if prefix.Matches(devAddr) {
						found = true
						weights[j]++
						total++
					}
				}
				a.So(found, should.BeTrue)
			}
			for i, weight := range weights {
				weight := float64(weight) / float64(total)
				a.So(weight, should.AlmostEqual, tc.Balance[i], 1e-2)
			}
		})
	}
}
