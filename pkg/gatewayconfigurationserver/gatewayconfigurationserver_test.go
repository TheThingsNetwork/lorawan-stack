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
	"encoding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/pkg/gatewayconfigurationserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/cpf"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/semtechudp"
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

func TestGatewayConfigurationServer(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	conf := &component.Config{}
	c := componenttest.NewComponent(t, conf)

	test.Must(New(c, testConfig))
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER)
}

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
	fpConf := config.FrequencyPlansConfig{
		URL: "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master",
	}
	fps := frequencyplans.NewStore(test.Must(fpConf.Fetcher(ctx, config.BlobConfig{})).(fetch.Interface))

	conf := &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: httpAddress,
			},
			FrequencyPlans: fpConf,
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	}
	c := componenttest.NewComponent(t, conf)
	c.AddContextFiller(func(ctx context.Context) context.Context {
		ctx = newContextWithRightsFetcher(ctx)
		return ctx
	})

	gcs, err := New(c, testConfig)
	a.So(err, should.BeNil)
	a.So(gcs, should.NotBeNil)

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	mustMarshal := func(b []byte, err error) []byte { return test.Must(b, err).([]byte) }
	marshalJSON := func(v interface{}) string {
		return string(mustMarshal(json.MarshalIndent(v, "", "\t")))
	}
	marshalText := func(v encoding.TextMarshaler) string {
		return string(mustMarshal(v.MarshalText()))
	}
	semtechUDPConfig := func(gtw *ttnpb.Gateway) string {
		return marshalJSON(test.Must(semtechudp.Build(gtw, fps)).(*semtechudp.Config))
	}
	cpfLoradConfig := func(gtw *ttnpb.Gateway) string {
		return marshalJSON(test.Must(cpf.BuildLorad(gtw, fps)).(*cpf.LoradConfig))
	}
	cpfLorafwdConfig := func(gtw *ttnpb.Gateway) string {
		return marshalText(test.Must(cpf.BuildLorafwd(gtw)).(*cpf.LorafwdConfig))
	}

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
				t.Run("semtechudp/global_conf.json", func(t *testing.T) {
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
					if !a.So(res.StatusCode, should.Equal, tc.ExpectCode) {
						t.FailNow()
					}
					switch res.StatusCode {
					case http.StatusOK:
						if !a.So(res.Header.Get("Content-Type"), should.Equal, "application/json; charset=UTF-8") {
							t.FailNow()
						}
						b, err := ioutil.ReadAll(res.Body)
						if err != nil {
							t.Fatalf("Failed to read response body: %s", err)
						}
						a.So(string(b), should.Equal, semtechUDPConfig(is.res.Get)+"\n")
					}
				})
				t.Run("cpf/lorad/lorad.json", func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf(
						"/api/v3/gcs/gateways/%s/cpf/lorad/lorad.json",
						tc.ID.GatewayID,
					)
					req := httptest.NewRequest(http.MethodGet, url, nil)
					req = req.WithContext(test.Context())
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
					rec := httptest.NewRecorder()
					c.ServeHTTP(rec, req)
					res := rec.Result()
					if !a.So(res.StatusCode, should.Equal, tc.ExpectCode) {
						t.FailNow()
					}
					switch res.StatusCode {
					case http.StatusOK:
						if !a.So(res.Header.Get("Content-Type"), should.Equal, "application/json; charset=UTF-8") {
							t.FailNow()
						}
						b, err := ioutil.ReadAll(res.Body)
						if err != nil {
							t.Fatalf("Failed to read response body: %s", err)
						}
						a.So(string(b), should.Equal, cpfLoradConfig(is.res.Get)+"\n")
					}
				})
				t.Run("cpf/lorafwd/lorafwd.toml", func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf(
						"/api/v3/gcs/gateways/%s/cpf/lorafwd/lorafwd.toml",
						tc.ID.GatewayID,
					)
					req := httptest.NewRequest(http.MethodGet, url, nil)
					req = req.WithContext(test.Context())
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
					rec := httptest.NewRecorder()
					c.ServeHTTP(rec, req)
					res := rec.Result()
					if !a.So(res.StatusCode, should.Equal, tc.ExpectCode) {
						t.FailNow()
					}
					switch res.StatusCode {
					case http.StatusOK:
						if !a.So(res.Header.Get("Content-Type"), should.Equal, "application/toml") {
							t.FailNow()
						}
						b, err := ioutil.ReadAll(res.Body)
						if err != nil {
							t.Fatalf("Failed to read response body: %s", err)
						}
						a.So(string(b), should.Equal, cpfLorafwdConfig(is.res.Get))
					}
				})
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

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

func init() {
	testConfig.TheThingsGateway.Default.FirmwareURL = "http://example.com"
	testConfig.TheThingsGateway.Default.UpdateChannel = "stable"
}
