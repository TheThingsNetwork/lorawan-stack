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

package redis

import (
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	Timeout = 10 * test.Delay
)

func TestRegistry(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()

	cl, flush := test.NewRedis(t, "redis_test")
	defer flush()
	defer cl.Close()

	ids := ttnpb.GatewayIdentifiers{
		GatewayID: "gtw1",
		EUI:       &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
	}
	ids2 := ttnpb.GatewayIdentifiers{
		GatewayID: "gtw2",
		EUI:       &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
	}
	ids3 := ttnpb.GatewayIdentifiers{
		GatewayID: "gtw3",
		EUI:       &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
	}
	registry := &GatewayConnectionStatsRegistry{
		Redis: cl,
	}

	now := time.Now().UTC()
	initialStats := &ttnpb.GatewayConnectionStats{
		ConnectedAt:            &now,
		Protocol:               "dummy",
		LastDownlinkReceivedAt: &now,
		DownlinkCount:          1,
		LastUplinkReceivedAt:   &now,
		UplinkCount:            1,
		LastStatusReceivedAt:   nil,
		LastStatus:             nil,
	}

	t.Run("GetNonExisting", func(t *testing.T) {
		stats, err := registry.Get(ctx, ids)
		a.So(stats, should.BeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})

	t.Run("EmptyStats", func(t *testing.T) {
		err := registry.Set(ctx, ids3, nil)
		a.So(err, should.BeNil)
		retrieved, err := registry.Get(ctx, ids3)
		a.So(retrieved, should.BeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})

	t.Run("SetAndClear", func(t *testing.T) {
		err := registry.Set(ctx, ids, initialStats)
		a.So(err, should.BeNil)
		retrieved, err := registry.Get(ctx, ids)
		a.So(err, should.BeNil)
		a.So(retrieved, should.Resemble, initialStats)

		// Other gateways not affected
		stats, err := registry.Get(ctx, ids2)
		a.So(stats, should.BeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		// Unset
		err = registry.Set(ctx, ids, nil)
		a.So(err, should.BeNil)
		retrieved, err = registry.Get(ctx, ids)
		a.So(errors.IsNotFound(err), should.BeTrue)
		a.So(retrieved, should.BeNil)
	})

	t.Run("ClearManyTimes", func(t *testing.T) {
		a.So(registry.Set(ctx, ids, nil), should.BeNil)
		a.So(registry.Set(ctx, ids, nil), should.BeNil)
	})

	t.Run("UpdateUplink", func(t *testing.T) {
		now := time.Now().UTC().Add(time.Minute)
		stats := deepcopy.Copy(initialStats).(*ttnpb.GatewayConnectionStats)

		// Update uplink stats, make sure they work
		stats.UplinkCount = 10
		stats.LastUplinkReceivedAt = &now
		err := registry.Set(ctx, ids, stats)
		a.So(err, should.BeNil)
		retrieved, err := registry.Get(ctx, ids)
		a.So(err, should.BeNil)
		a.So(retrieved, should.Resemble, stats)

		// Now update downlink also
		stats.LastDownlinkReceivedAt = &now
		err = registry.Set(ctx, ids, stats)
		a.So(err, should.BeNil)
		retrieved, err = registry.Get(ctx, ids)
		a.So(err, should.BeNil)
		a.So(retrieved, should.Resemble, stats)

		// Unset uplink
		stats.LastUplinkReceivedAt = nil
		stats.UplinkCount = 0
		err = registry.Set(ctx, ids, stats)
		a.So(err, should.BeNil)
		retrieved, err = registry.Get(ctx, ids)
		a.So(err, should.BeNil)
		a.So(retrieved, should.Resemble, stats)
	})
}
