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

package ws_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

func withServer(t *testing.T, wsConfig ws.Config, rateLimitConf config.RateLimiting, f func(t *testing.T, is *mock.IdentityServer, serverAddress string)) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			RateLimiting: rateLimitConf,
		},
	})
	c.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	gs := mock.NewServer(c)

	bsWebServer := ws.New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), wsConfig)
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		t.FailNow()
	}
	defer lis.Close()
	go func() error {
		return http.Serve(lis, bsWebServer)
	}()
	servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

	f(t, is, servAddr)
}
