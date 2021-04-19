// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package udp

import (
	"context"
	"testing"
	"time"

	"github.com/bluele/gcache"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockServer struct {
	io.Server
	called map[ttnpb.GatewayIdentifiers]int
}

func (s *mockServer) FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error) {
	s.called[ids]++
	return ctx, ids, nil
}

func TestGatewayIdentifiersCache(t *testing.T) {
	a := assertions.New(t)
	c := componenttest.NewComponent(t, &component.Config{})
	componenttest.StartComponent(t, c)
	defer c.Close()

	gs := mock.NewServer(c)

	ttl := time.Minute
	clock := gcache.NewFakeClock()
	cache := gcache.New(1000).Clock(clock).Expiration(ttl).Build()

	m := &mockServer{gs, make(map[ttnpb.GatewayIdentifiers]int)}

	server := newServerWithIdentifiersCache(m, cache)

	ids := ttnpb.GatewayIdentifiers{EUI: &types.EUI64{1, 1, 1, 1, 1, 1, 1, 1}}
	ids2 := ttnpb.GatewayIdentifiers{EUI: &types.EUI64{1, 1, 1, 1, 1, 1, 1, 2}}
	var fetched ttnpb.GatewayIdentifiers

	// Cold cache, ensure server.FillGatewayContext() was called
	_, fetched, _ = server.FillGatewayContext(test.Context(), ids)
	a.So(m.called[ids], should.Equal, 1)
	a.So(fetched, should.Resemble, ids)

	// Warm cache, ensure cache is used
	_, fetched, _ = server.FillGatewayContext(test.Context(), ids)
	a.So(m.called[ids], should.Equal, 1)

	clock.Advance(ttl + time.Second)

	// Cache expired, ensure server.FillGatewayContext() is called again
	_, fetched, _ = server.FillGatewayContext(test.Context(), ids)
	a.So(m.called[ids], should.Equal, 2)
	a.So(fetched, should.Resemble, ids)

	// Different identifiers
	_, fetched, _ = server.FillGatewayContext(test.Context(), ids2)
	a.So(m.called[ids2], should.Equal, 1)
	a.So(fetched, should.Resemble, ids2)
}
