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

package networkserver

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNewDevAddr(t *testing.T) {
	a := assertions.New(t)

	// Use DevAddr prefix from NetID.
	{
		ns := test.Must(New(
			component.MustNew(test.GetLogger(t), &component.Config{}),
			&Config{
				NetID:               types.NetID{0x00, 0x00, 0x13},
				DeduplicationWindow: 42,
				CooldownWindow:      42,
				DownlinkTasks: &MockDownlinkTaskQueue{
					PopFunc: DownlinkTaskPopBlockFunc,
				},
			})).(*NetworkServer)

		if !a.So(ns.devAddrPrefixes, should.HaveLength, 1) {
			t.FailNow()
		}
		a.So(ns.devAddrPrefixes[0], should.Resemble, types.DevAddrPrefix{
			DevAddr: types.DevAddr{0x26, 0, 0, 0},
			Length:  7,
		})

		devAddr := ns.newDevAddr(test.Context(), nil)
		a.So(devAddr.HasPrefix(ns.devAddrPrefixes[0]), should.BeTrue)
	}

	// Configured DevAddr prefixes.
	{
		ns := test.Must(New(
			component.MustNew(test.GetLogger(t), &component.Config{}),
			&Config{
				NetID: types.NetID{0x00, 0x00, 0x13},
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
				DeduplicationWindow: 42,
				CooldownWindow:      42,
				DownlinkTasks: &MockDownlinkTaskQueue{
					PopFunc: DownlinkTaskPopBlockFunc,
				},
			})).(*NetworkServer)

		seen := map[types.DevAddrPrefix]int{}
		for i := 0; i < 100; i++ {
			devAddr := ns.newDevAddr(test.Context(), nil)
			for _, prefix := range ns.devAddrPrefixes {
				if devAddr.HasPrefix(prefix) {
					seen[prefix]++
					break
				}
			}
		}
		a.So(seen[ns.devAddrPrefixes[0]], should.BeGreaterThan, 0)
		a.So(seen[ns.devAddrPrefixes[1]], should.BeGreaterThan, 0)
		a.So(seen[ns.devAddrPrefixes[2]], should.BeGreaterThan, 0)
	}
}
