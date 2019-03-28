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

package udp_test

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
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mock"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestWeb(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	t.Run("GetGateway", func(t *testing.T) {
		httpAddress := "0.0.0.0:8098"
		gs := mock.NewServer()
		gtw := &ttnpb.Gateway{
			GatewayIdentifiers:   registeredGatewayID,
			FrequencyPlanID:      "EXAMPLE",
			GatewayServerAddress: "localhost",
		}
		gs.RegisterGateway(ctx, registeredGatewayID, gtw)
		s := StartWeb(newContextWithRightsFetcher(ctx), gs, testConfig)
		conf := &component.Config{
			ServiceBase: config.ServiceBase{
				HTTP: config.HTTP{
					Listen: httpAddress,
				},
			},
		}
		c := component.MustNew(test.GetLogger(t), conf)
		c.RegisterWeb(s)
		test.Must(nil, c.Start())
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
					url := fmt.Sprintf("http://%s/api/v3/gs/gateways/%s/global_conf.json",
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
