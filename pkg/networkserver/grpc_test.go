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

package networkserver_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestGenerateDevAddr(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		NetID         types.NetID
		DevAddrPrefix types.DevAddrPrefix
	}{
		{
			Name:  "Prefix from NS NetID 1",
			NetID: types.NetID{0x00, 0x00, 0x13},
			DevAddrPrefix: types.DevAddrPrefix{
				DevAddr: test.Must(types.NewDevAddr(types.NetID{0x00, 0x00, 0x13}, nil)).(types.DevAddr),
				Length:  uint8(32 - types.NwkAddrBits(types.NetID{0x00, 0x00, 0x13})),
			},
		},
		{
			Name:  "Prefix from NS NetID 2",
			NetID: types.NetID{0x00, 0x00, 0x14},
			DevAddrPrefix: types.DevAddrPrefix{
				DevAddr: test.Must(types.NewDevAddr(types.NetID{0x00, 0x00, 0x14}, nil)).(types.DevAddr),
				Length:  uint8(32 - types.NwkAddrBits(types.NetID{0x00, 0x00, 0x14})),
			},
		},
		{
			Name:  "Prefix from NS NetID 3",
			NetID: types.NetID{0x12, 0x34, 0x56},
			DevAddrPrefix: types.DevAddrPrefix{
				DevAddr: test.Must(types.NewDevAddr(types.NetID{0x12, 0x34, 0x56}, nil)).(types.DevAddr),
				Length:  uint8(32 - types.NwkAddrBits(types.NetID{0x12, 0x34, 0x56})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ns, ctx, _, stop := StartTest(ctx, TestConfig{
					NetworkServer: Config{
						NetID: tc.NetID,
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
					),
					Component: component.Config{
						ServiceBase: config.ServiceBase{
							FrequencyPlans: config.FrequencyPlansConfig{
								ConfigSource: "static",
								Static:       test.StaticFrequencyPlans,
							},
						},
					},
				})
				defer stop()

				devAddr, err := ttnpb.NewNsClient(ns.LoopbackConn()).GenerateDevAddr(ctx, ttnpb.Empty)
				if a.So(err, should.BeNil) {
					a.So(devAddr.DevAddr.HasPrefix(tc.DevAddrPrefix), should.BeTrue)
				}
			},
		})
	}
	for _, tc := range []struct {
		Name            string
		DevAddrPrefixes []types.DevAddrPrefix
	}{
		{
			Name: "Defined DevAddrPrefixes Set 1",
			DevAddrPrefixes: []types.DevAddrPrefix{
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
			},
		},
		{
			Name: "Defined DevAddrPrefixes Set 2",
			DevAddrPrefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.DevAddr{0x1f, 0x00, 0x00, 0x00},
					Length:  8,
				},
				{
					DevAddr: types.DevAddr{0xff, 0xff, 0x00, 0x00},
					Length:  16,
				},
				{
					DevAddr: types.DevAddr{0x27, 0x00, 0x00, 0x00},
					Length:  8,
				},
			},
		},
		{
			Name: "Defined DevAddrPrefixes Set 3",
			DevAddrPrefixes: []types.DevAddrPrefix{
				{
					DevAddr: types.DevAddr{0xff, 0xff, 0xff, 0x00},
					Length:  24,
				},
				{
					DevAddr: types.DevAddr{0x00, 0xff, 0xff, 0xff},
					Length:  8,
				},
				{
					DevAddr: types.DevAddr{0x27, 0x072, 0x00, 0x00},
					Length:  16,
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ns, ctx, _, stop := StartTest(ctx, TestConfig{
					NetworkServer: Config{
						NetID:           types.NetID{0x00, 0x00, 0x13},
						DevAddrPrefixes: tc.DevAddrPrefixes,
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
					),
					Component: component.Config{
						ServiceBase: config.ServiceBase{
							FrequencyPlans: config.FrequencyPlansConfig{
								ConfigSource: "static",
								Static:       test.StaticFrequencyPlans,
							},
						},
					},
				})
				defer stop()

				hasOneOfPrefixes := func(devAddr *types.DevAddr, seen map[types.DevAddrPrefix]int, prefixes ...types.DevAddrPrefix) bool {
					for i, p := range prefixes {
						if devAddr.HasPrefix(p) {
							seen[prefixes[i]]++
							return true
						}
					}
					return false
				}

				seen := map[types.DevAddrPrefix]int{}
				for i := 0; i < 100; i++ {
					devAddr, err := ttnpb.NewNsClient(ns.LoopbackConn()).GenerateDevAddr(ctx, ttnpb.Empty)
					if a.So(err, should.BeNil) {
						a.So(hasOneOfPrefixes(devAddr.DevAddr, seen, tc.DevAddrPrefixes[0], tc.DevAddrPrefixes[1], tc.DevAddrPrefixes[2]), should.BeTrue)
					}
				}
				a.So(seen[tc.DevAddrPrefixes[0]], should.BeGreaterThan, 0)
				a.So(seen[tc.DevAddrPrefixes[1]], should.BeGreaterThan, 0)
				a.So(seen[tc.DevAddrPrefixes[2]], should.BeGreaterThan, 0)
			},
		})
	}
}

func TestGetDefaultMACSettings(t *testing.T) {
	for _, tc := range []struct {
		name      string
		assertion func(err error) bool
		req       *ttnpb.GetDefaultMACSettingsRequest
	}{
		{
			name:      "NoFrequencyPlanID",
			assertion: errors.IsNotFound,
			req:       &ttnpb.GetDefaultMACSettingsRequest{},
		},
		{
			name:      "NoLoRaWANVersion",
			assertion: errors.IsInvalidArgument,
			req: &ttnpb.GetDefaultMACSettingsRequest{
				FrequencyPlanId: "EU_863_870",
			},
		},
		{
			name:      "OK",
			assertion: func(err error) bool { return err == nil },
			req: &ttnpb.GetDefaultMACSettingsRequest{
				FrequencyPlanId:   "EU_863_870",
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ns, _, _, stop := StartTest(ctx, TestConfig{
					Component: component.Config{
						ServiceBase: config.ServiceBase{
							FrequencyPlans: config.FrequencyPlansConfig{
								ConfigSource: "static",
								Static:       test.StaticFrequencyPlans,
							},
						},
					},
				})
				defer stop()
				settings, err := ttnpb.NewNsClient(ns.LoopbackConn()).GetDefaultMACSettings(test.Context(), tc.req)
				if tc.assertion != nil {
					a.So(tc.assertion(err), should.BeTrue)
				} else {
					a.So(settings, should.NotBeNil)
				}
			},
		})
	}
}
