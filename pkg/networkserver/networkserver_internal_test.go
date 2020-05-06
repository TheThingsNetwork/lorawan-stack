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
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNewDevAddr(t *testing.T) {
	t.Run("From NetID", func(t *testing.T) {
		ns, ctx, _, stop := StartTest(
			t,
			component.Config{},
			Config{
				NetID: types.NetID{0x00, 0x00, 0x13},
				DownlinkTasks: MockDownlinkTaskQueue{
					PopFunc: DownlinkTaskPopBlockFunc,
				},
			},
			(1<<3)*test.Delay,
		)
		defer stop()

		assertions.New(t).So(ns.newDevAddr(ctx, nil).HasPrefix(types.DevAddrPrefix{
			DevAddr: types.DevAddr{0x26, 0, 0, 0},
			Length:  7,
		}), should.BeTrue)
	})

	t.Run("Configured DevAddr prefixes", func(t *testing.T) {
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
			component.Config{},
			Config{
				NetID: types.NetID{0x00, 0x00, 0x13},
				DownlinkTasks: MockDownlinkTaskQueue{
					PopFunc: DownlinkTaskPopBlockFunc,
				},
				DevAddrPrefixes: ps,
			},
			(1<<3)*test.Delay,
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

		a := assertions.New(t)
		a.So(seen[ps[0]], should.BeGreaterThan, 0)
		a.So(seen[ps[1]], should.BeGreaterThan, 0)
		a.So(seen[ps[2]], should.BeGreaterThan, 0)
	})
}
