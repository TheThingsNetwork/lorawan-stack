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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/cpf"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/semtechudp"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	registeredGatewayID  = &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
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
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()

	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, "Bearer", registeredGatewayKey, testGtw, ttnpb.Right_RIGHT_GATEWAY_INFO)

	fpConf := config.FrequencyPlansConfig{
		ConfigSource: "static",
		Static:       test.StaticFrequencyPlans,
	}
	fps := frequencyplans.NewStore(test.Must(fpConf.Fetcher(ctx, config.BlobConfig{}, test.HTTPClientProvider)).(fetch.Interface))

	conf := &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: ":0",
			},
			FrequencyPlans: fpConf,
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	}
	c := componenttest.NewComponent(t, conf)
	c.AddContextFiller(func(ctx context.Context) context.Context {
		ctx = newContextWithRightsFetcher(ctx)
		return ctx
	})

	test.Must(New(c, testConfig))
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	mustMarshal := func(b []byte, err error) []byte { return test.Must(b, err) }
	marshalJSON := func(v any) string {
		return string(mustMarshal(json.MarshalIndent(v, "", "\t")))
	}
	marshalText := func(v encoding.TextMarshaler) string {
		return string(mustMarshal(v.MarshalText()))
	}
	semtechUDPConfig := func(gtw *ttnpb.Gateway) string {
		return marshalJSON(test.Must(semtechudp.Build(gtw, fps)))
	}
	cpfLoradConfig := func(gtw *ttnpb.Gateway) string {
		return marshalJSON(test.Must(cpf.BuildLorad(gtw, fps)))
	}
	cpfLorafwdConfig := func(gtw *ttnpb.Gateway) string {
		return marshalText(test.Must(cpf.BuildLorafwd(gtw)))
	}

	t.Run("Authorization", func(t *testing.T) {
		for _, tc := range []struct {
			Name       string
			ID         *ttnpb.GatewayIdentifiers
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
				ID:         &ttnpb.GatewayIdentifiers{GatewayId: "--invalid-id"},
				Key:        "invalid key",
				ExpectCode: http.StatusBadRequest,
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				t.Run("semtechudp/global_conf.json", func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf(
						"/api/v3/gcs/gateways/%s/semtechudp/global_conf.json",
						tc.ID.GatewayId,
					)
					body := bytes.NewReader([]byte(`{"downlinks":[]}`))
					req := httptest.NewRequest(http.MethodGet, url, body).WithContext(test.Context())
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
						if !a.So(res.Header.Get("Content-Type"), should.Equal, "application/json") {
							t.FailNow()
						}
						b, err := io.ReadAll(res.Body)
						if err != nil {
							t.Fatalf("Failed to read response body: %s", err)
						}
						gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: tc.ID})
						a.So(err, should.BeNil)

						a.So(string(b), should.Equal, semtechUDPConfig(gtw)+"\n")
					}
				})
				t.Run("cpf/lorad/lorad.json", func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf(
						"/api/v3/gcs/gateways/%s/kerlink-cpf/lorad/lorad.json",
						tc.ID.GatewayId,
					)
					req := httptest.NewRequest(http.MethodGet, url, nil).WithContext(test.Context())
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
					rec := httptest.NewRecorder()
					c.ServeHTTP(rec, req)
					res := rec.Result()
					if !a.So(res.StatusCode, should.Equal, tc.ExpectCode) {
						t.FailNow()
					}
					switch res.StatusCode {
					case http.StatusOK:
						if !a.So(res.Header.Get("Content-Type"), should.Equal, "application/json") {
							t.FailNow()
						}
						b, err := io.ReadAll(res.Body)
						if err != nil {
							t.Fatalf("Failed to read response body: %s", err)
						}
						gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: tc.ID})
						a.So(err, should.BeNil)
						a.So(string(b), should.Equal, cpfLoradConfig(gtw)+"\n")
					}
				})
				t.Run("cpf/lorafwd/lorafwd.toml", func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf(
						"/api/v3/gcs/gateways/%s/kerlink-cpf/lorafwd/lorafwd.toml",
						tc.ID.GatewayId,
					)
					req := httptest.NewRequest(http.MethodGet, url, nil).WithContext(test.Context())
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
						b, err := io.ReadAll(res.Body)
						if err != nil {
							t.Fatalf("Failed to read response body: %s", err)
						}
						gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: tc.ID})
						a.So(err, should.BeNil)
						a.So(string(b), should.Equal, cpfLorafwdConfig(gtw))
					}
				})
			})
		}
	})
}

func TestGRPC(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()

	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, "Bearer", registeredGatewayKey, testGtw, ttnpb.Right_RIGHT_GATEWAY_INFO)

	fpConf := config.FrequencyPlansConfig{
		ConfigSource: "static",
		Static:       test.StaticFrequencyPlans,
	}
	fps := frequencyplans.NewStore(test.Must(fpConf.Fetcher(ctx, config.BlobConfig{}, test.HTTPClientProvider)).(fetch.Interface))

	conf := &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: ":0",
			},
			FrequencyPlans: fpConf,
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	}
	c := componenttest.NewComponent(t, conf)
	c.AddContextFiller(func(ctx context.Context) context.Context {
		ctx = newContextWithRightsFetcher(ctx)
		return ctx
	})

	test.Must(New(c, testConfig))
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	t.Run("Authorization", func(t *testing.T) {
		a := assertions.New(t)
		for _, tc := range []struct {
			Name         string
			ID           *ttnpb.GatewayIdentifiers
			Key          string
			ErrAssertion func(error) bool
		}{
			{
				Name: "Valid",
				ID:   registeredGatewayID,
				Key:  registeredGatewayKey,
			},
			{
				Name:         "InvalidKey",
				ID:           registeredGatewayID,
				Key:          "invalid key",
				ErrAssertion: errors.IsPermissionDenied,
			},
			{
				Name:         "InvalidIDAndKey",
				ID:           &ttnpb.GatewayIdentifiers{GatewayId: "--invalid-id"},
				Key:          "invalid key",
				ErrAssertion: errors.IsInvalidArgument,
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     tc.Key,
					AllowInsecure: true,
				})
				client := ttnpb.NewGatewayConfigurationServiceClient(c.LoopbackConn())
				a.So(client, should.NotBeNil)

				_, err := client.GetGatewayConfiguration(ctx, &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: tc.ID,
					Format:     "semtechudp",
					Filename:   "global_conf.json",
				}, creds)
				if tc.ErrAssertion != nil {
					a.So(tc.ErrAssertion(err), should.BeTrue)
				} else {
					a.So(err, should.BeNil)
				}
			})
		}
	})

	mustMarshal := func(b []byte, err error) []byte { return test.Must(b, err) }
	marshalJSON := func(v any) string {
		return string(mustMarshal(json.Marshal(v)))
	}
	marshalText := func(v encoding.TextMarshaler) string {
		return string(mustMarshal(v.MarshalText()))
	}
	semtechUDPConfig := func(gtw *ttnpb.Gateway) string {
		return marshalJSON(test.Must(semtechudp.Build(gtw, fps)))
	}
	cpfLoradConfig := func(gtw *ttnpb.Gateway) string {
		return marshalJSON(test.Must(cpf.BuildLorad(gtw, fps)))
	}
	cpfLorafwdConfig := func(gtw *ttnpb.Gateway) string {
		return marshalText(test.Must(cpf.BuildLorafwd(gtw)))
	}

	t.Run("Request", func(t *testing.T) {
		a := assertions.New(t)
		creds := grpc.PerRPCCredentials(rpcmetadata.MD{
			AuthType:      "Bearer",
			AuthValue:     registeredGatewayKey,
			AllowInsecure: true,
		})
		client := ttnpb.NewGatewayConfigurationServiceClient(c.LoopbackConn())
		a.So(client, should.NotBeNil)
		for _, tc := range []struct {
			Name          string
			Req           *ttnpb.GetGatewayConfigurationRequest
			ErrAssertion  func(error) bool
			RespAssertion func(*ttnpb.GetGatewayConfigurationResponse) bool
		}{
			{
				Name: "invalid format",
				Req: &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: registeredGatewayID,
					Format:     "invalid",
				},
				ErrAssertion: errors.IsInvalidArgument,
			},
			{
				Name: "invalid type",
				Req: &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: registeredGatewayID,
					Format:     "kerlink-cpf",
					Type:       "invalid",
				},
				ErrAssertion: errors.IsInvalidArgument,
			},
			{
				Name: "invalid filename",
				Req: &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: registeredGatewayID,
					Format:     "semtechudp",
					Filename:   "invalid",
				},
				ErrAssertion: errors.IsInvalidArgument,
			},
			{
				Name: "semtechudp",
				Req: &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: registeredGatewayID,
					Format:     "semtechudp",
					Filename:   "global_conf.json",
				},
				RespAssertion: func(resp *ttnpb.GetGatewayConfigurationResponse) bool {
					return a.So(string(resp.Contents), should.Equal, semtechUDPConfig(testGtw))
				},
			},
			{
				Name: "kerlink-cpf/lorad",
				Req: &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: registeredGatewayID,
					Format:     "kerlink-cpf",
					Type:       "lorad",
					Filename:   "lorad.json",
				},
				RespAssertion: func(resp *ttnpb.GetGatewayConfigurationResponse) bool {
					return a.So(string(resp.Contents), should.Equal, cpfLoradConfig(testGtw))
				},
			},
			{
				Name: "kerlink-cpf/lorafwd",
				Req: &ttnpb.GetGatewayConfigurationRequest{
					GatewayIds: registeredGatewayID,
					Format:     "kerlink-cpf",
					Type:       "lorafwd",
					Filename:   "lorafwd.toml",
				},
				RespAssertion: func(resp *ttnpb.GetGatewayConfigurationResponse) bool {
					return a.So(string(resp.Contents), should.Equal, cpfLorafwdConfig(testGtw))
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				resp, err := client.GetGatewayConfiguration(ctx, tc.Req, creds)
				if tc.ErrAssertion != nil {
					a.So(tc.ErrAssertion(err), should.BeTrue)
				} else {
					a.So(err, should.BeNil)
					if tc.RespAssertion != nil {
						a.So(tc.RespAssertion(resp), should.BeTrue)
					}
				}
			})
		}
	})
}

type rightsFetcher struct {
	rights.AuthInfoFetcher
	rights.EntityFetcher
}

func newContextWithRightsFetcher(ctx context.Context) context.Context {
	return rights.NewContextWithFetcher(ctx, &rightsFetcher{
		EntityFetcher: rights.EntityFetcherFunc(func(ctx context.Context, ids *ttnpb.EntityIdentifiers) (*ttnpb.Rights, error) {
			uid := unique.ID(ctx, ids)
			if uid != registeredGatewayUID {
				return nil, nil
			}
			md := rpcmetadata.FromIncomingContext(ctx)
			if md.AuthType != "Bearer" || md.AuthValue != registeredGatewayKey {
				return nil, nil
			}
			return ttnpb.RightsFrom(
				ttnpb.Right_RIGHT_GATEWAY_INFO,
			), nil
		}),
	})
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
	testConfig.TheThingsKickstarterGateway.Default.FirmwareURL = "http://example.com"
	testConfig.TheThingsKickstarterGateway.Default.UpdateChannel = "stable"
}
