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
	"testing"

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
	registeredGatewayUID = "test-gateway"
	registeredGatewayID  = ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}
	registeredGatewayKey = "test-key"

	testConfig = &Config{
		RequireAuth: true,
	}
)

func TestWeb(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	gs := &mockGatewayClient{}
	gs.res.Get = &ttnpb.Gateway{
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
		},
	}
	c := component.MustNew(test.GetLogger(t), conf)
	ctx = c.FillContext(ctx)

	gcs, err := New(c, testConfig, []Option{
		WithRegistry(gs),
		WithContext(newContextWithRightsFetcher(ctx)),
	}...)
	a.So(err, should.BeNil)
	a.So(gcs, should.NotBeNil)

	err = c.Start()
	a.So(err, should.BeNil)
	defer c.Close()

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
				url := fmt.Sprintf("http://%s/api/v3/gcs/gateways/%s/global_conf.json",
					httpAddress, tc.ID.GatewayID,
				)
				body := bytes.NewReader([]byte(`{"downlinks":[]}`))
				req, err := http.NewRequest(http.MethodGet, url, body)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
				res, err := http.DefaultClient.Do(req)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
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
