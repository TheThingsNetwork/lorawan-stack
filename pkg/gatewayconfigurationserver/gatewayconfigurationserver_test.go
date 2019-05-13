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

package gatewayconfigurationserver_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	. "go.thethings.network/lorawan-stack/pkg/gatewayconfigurationserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredGatewayID  = ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}
	registeredGatewayUID = unique.ID(test.Context(), registeredGatewayID)
	registeredGatewayKey = "test-key"

	testConfig = &Config{
		RequireAuth: true,
	}
)

func TestWeb(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	is, isAddr := startMockIS(ctx)
	is.res.Get = &ttnpb.Gateway{
		GatewayIdentifiers:   registeredGatewayID,
		FrequencyPlanID:      "EU_863_870",
		GatewayServerAddress: "localhost",
	}

	httpAddress := "0.0.0.0:8098"
	conf := &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: httpAddress,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				URL: "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master",
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	}
	c := component.MustNew(test.GetLogger(t), conf)
	c.AddContextFiller(func(ctx context.Context) context.Context {
		ctx = newContextWithRightsFetcher(ctx)
		return ctx
	})

	gcs, err := New(c, testConfig)
	a.So(err, should.BeNil)
	a.So(gcs, should.NotBeNil)

	err = c.Start()
	a.So(err, should.BeNil)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	t.Run("Authorization", func(t *testing.T) {
		for _, tc := range []struct {
			Name       string
			ID         ttnpb.GatewayIdentifiers
			Key        string
			ExpectCode int
		}{
			{
				Name:       "Valid",
				ID:         registeredGatewayID,
				Key:        registeredGatewayKey,
				ExpectCode: http.StatusOK,
			},
			{
				Name:       "InvalidKey",
				ID:         registeredGatewayID,
				Key:        "invalid key",
				ExpectCode: http.StatusForbidden,
			},
			{
				Name:       "InvalidIDAndKey",
				ID:         ttnpb.GatewayIdentifiers{GatewayID: "--invalid-id"},
				Key:        "invalid key",
				ExpectCode: http.StatusBadRequest,
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)
				url := fmt.Sprintf(
					"/api/v3/gcs/gateways/%s/semtechudp/global_conf.json",
					tc.ID.GatewayID,
				)
				body := bytes.NewReader([]byte(`{"downlinks":[]}`))
				req := httptest.NewRequest(http.MethodGet, url, body)
				req = req.WithContext(test.Context())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
				rec := httptest.NewRecorder()
				c.ServeHTTP(rec, req)
				res := rec.Result()
				a.So(res.StatusCode, should.Equal, tc.ExpectCode)
			})
		}
	})
}

func newContextWithRightsFetcher(ctx context.Context) context.Context {
	return rights.NewContextWithFetcher(
		ctx,
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (set *ttnpb.Rights, err error) {
			uid := unique.ID(ctx, ids)
			if uid != registeredGatewayUID {
				return
			}
			md := rpcmetadata.FromIncomingContext(ctx)
			if md.AuthType != "Bearer" || md.AuthValue != registeredGatewayKey {
				return
			}
			set = ttnpb.RightsFrom(
				ttnpb.RIGHT_GATEWAY_INFO,
			)
			return
		}),
	)
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.PeerInfo_Role) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if peer := c.GetPeer(ctx, role, nil); peer != nil {
			return
		}
	}
	panic("could not connect to peer")
}

func init() {
	testConfig.TheThingsGateway.Default.FirmwareURL = "http://example.com"
	testConfig.TheThingsGateway.Default.UpdateChannel = "stable"
}
