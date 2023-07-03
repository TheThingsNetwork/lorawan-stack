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

package gatewayconfigurationserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockGatewayRegistryClient struct {
	ttnpb.GatewayRegistryClient

	ctx context.Context
	in  *ttnpb.GetGatewayRequest

	out *ttnpb.Gateway
	err error
}

func (c *mockGatewayRegistryClient) Get(ctx context.Context, in *ttnpb.GetGatewayRequest, _ ...grpc.CallOption) (*ttnpb.Gateway, error) {
	c.ctx, c.in = ctx, in
	return c.out, c.err
}

type rightsFetcher struct {
	rights.AuthInfoFetcher
	rights.EntityFetcher
}

func newContextWithRightsFetcher(ctx context.Context) context.Context {
	return rights.NewContextWithFetcher(ctx, &rightsFetcher{
		EntityFetcher: rights.EntityFetcherFunc(func(ctx context.Context, ids *ttnpb.EntityIdentifiers) (*ttnpb.Rights, error) {
			md := rpcmetadata.FromIncomingContext(ctx)
			if md.AuthType != "Bearer" {
				return nil, nil
			}
			return ttnpb.RightsFrom(
				ttnpb.Right_RIGHT_GATEWAY_INFO,
			), nil
		}),
	})
}

func TestGetGateway(t *testing.T) {
	for _, tc := range []struct {
		Name              string
		SetupStore        func(*mockGatewayRegistryClient)
		SetupRequest      func(*http.Request)
		ResponseAssertion func(*assertions.Assertion, *httptest.ResponseRecorder) bool
	}{
		{
			Name: "Not Found",
			SetupStore: func(reg *mockGatewayRegistryClient) {
				reg.out, reg.err = nil, status.Error(codes.NotFound, "not found")
			},
			ResponseAssertion: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) bool {
				return a.So(rec.Code, should.Equal, http.StatusNotFound)
			},
		},
		{
			Name: "Any Authenticated Gateway",
			SetupStore: func(reg *mockGatewayRegistryClient) {
				reg.out, reg.err = &ttnpb.Gateway{
					Description: "Gateway Description",
					Attributes: map[string]string{
						"key": "some-key",
					},
					FrequencyPlanId:      "EU_863_870",
					GatewayServerAddress: "gatewayserver",
					Antennas: []*ttnpb.GatewayAntenna{
						{Location: &ttnpb.Location{Latitude: 12.34, Longitude: 56.78, Altitude: 90}},
					},
				}, nil
			},
			ResponseAssertion: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) bool {
				body := rec.Body.String()
				return a.So(rec.Code, should.Equal, http.StatusOK) &&
					a.So(body, assertions.ShouldContainSubstring, `"attributes":{"description":"Gateway Description"}`) &&
					a.So(body, assertions.ShouldContainSubstring, `"frequency_plan":"EU_863_870"`) &&
					a.So(body, assertions.ShouldContainSubstring, `"frequency_plan_url":"http://example.com/api/v2/frequency-plans/EU_863_870"`) &&
					a.So(body, assertions.ShouldContainSubstring, `"router":{"id":"gatewayserver","mqtt_address":"mqtts://gatewayserver:8881"}`) &&
					a.So(body, assertions.ShouldContainSubstring, `"antenna_location":{"latitude":12.34,"longitude":56.78,"altitude":90}`)
			},
		},
		{
			Name: "Authenticated TTKG",
			SetupStore: func(reg *mockGatewayRegistryClient) {
				reg.out, reg.err = &ttnpb.Gateway{
					Description: "Gateway Description",
					Attributes: map[string]string{
						"key": "some-key",
					},
					FrequencyPlanId:      "EU_863_870",
					GatewayServerAddress: "gatewayserver",
					Antennas: []*ttnpb.GatewayAntenna{
						{Location: &ttnpb.Location{Latitude: 12.34, Longitude: 56.78, Altitude: 90}},
					},
				}, nil
			},
			SetupRequest: func(req *http.Request) {
				req.Header.Set("User-Agent", "TTNGateway")
			},
			ResponseAssertion: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) bool {
				body := rec.Body.String()
				return a.So(rec.Code, should.Equal, http.StatusOK) &&
					a.So(body, assertions.ShouldNotContainSubstring, `"attributes"`) &&
					a.So(body, assertions.ShouldContainSubstring, `"router":{"mqtt_address":"mqtts://gatewayserver:8881"}`)
			},
		},
		{
			Name: "Any Unauthenticated Gateway",
			SetupStore: func(reg *mockGatewayRegistryClient) {
				reg.out, reg.err = &ttnpb.Gateway{
					Description: "Gateway Description",
					Attributes: map[string]string{
						"key": "some-key",
					},
					FrequencyPlanId:      "EU_863_870",
					GatewayServerAddress: "gatewayserver",
					Antennas: []*ttnpb.GatewayAntenna{
						{Location: &ttnpb.Location{Latitude: 12.34, Longitude: 56.78, Altitude: 90}},
					},
				}, nil
			},
			SetupRequest: func(req *http.Request) {
				req.Header.Del("Authorization")
			},
			ResponseAssertion: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) bool {
				return a.So(rec.Code, should.Equal, http.StatusOK) &&
					a.So(rec.Body.String(), assertions.ShouldNotContainSubstring, `"router":{"mqtt_address":"mqtts://gatewayserver:8881"}`)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := log.NewContext(test.Context(), test.GetLogger(t))

			reg := &mockGatewayRegistryClient{}
			tc.SetupStore(reg)

			auth := func(ctx context.Context) grpc.CallOption {
				return grpc.PerRPCCredentials(nil)
			}

			conf := &component.Config{
				ServiceBase: config.ServiceBase{
					HTTP: config.HTTP{
						Listen: ":0",
					},
				},
			}
			c := componenttest.NewComponent(t, conf)
			c.AddContextFiller(func(ctx context.Context) context.Context {
				return rights.NewContextWithFetcher(ctx, &rightsFetcher{})
			})

			New(c, WithRegistry(reg), WithAuth(auth))
			componenttest.StartComponent(t, c)
			defer c.Close()

			req := httptest.NewRequest(http.MethodGet, "/api/v2/gateways/foo-gtw", nil).WithContext(ctx)
			req.Header.Set("Authorization", "key some-key")
			if tc.SetupRequest != nil {
				tc.SetupRequest(req)
			}

			rec := httptest.NewRecorder()
			c.ServeHTTP(rec, req)
			tc.ResponseAssertion(a, rec)
		})
	}
}

func TestGetFrequencyPlan(t *testing.T) {
	for _, tc := range []struct {
		Name              string
		SetupRequest      func(*http.Request)
		ErrorAssertion    func(actual any, expected ...any) string
		ResponseAssertion func(*assertions.Assertion, *httptest.ResponseRecorder) bool
	}{
		{
			Name:           "Any Gateway",
			ErrorAssertion: should.BeNil,
			ResponseAssertion: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) bool {
				body := rec.Body.String()
				return a.So(rec.Code, should.Equal, http.StatusOK) &&
					a.So(body, assertions.ShouldContainSubstring, `"SX1301_conf"`) &&
					a.So(body, assertions.ShouldContainSubstring, `"chan_multiSF_0"`) &&
					a.So(body, assertions.ShouldContainSubstring, `"tx_lut_0"`)
			},
		},
		{
			Name: "TTKG",
			SetupRequest: func(req *http.Request) {
				req.Header.Set("User-Agent", "TTNGateway")
			},
			ErrorAssertion: should.BeNil,
			ResponseAssertion: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) bool {
				body := rec.Body.String()
				return a.So(rec.Code, should.Equal, http.StatusOK) &&
					a.So(body, assertions.ShouldNotContainSubstring, `"tx_lut_0"`)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := log.NewContext(test.Context(), test.GetLogger(t))

			conf := &component.Config{
				ServiceBase: config.ServiceBase{
					HTTP: config.HTTP{
						Listen: ":0",
					},
					FrequencyPlans: config.FrequencyPlansConfig{
						ConfigSource: "static",
						Static:       test.StaticFrequencyPlans,
					},
				},
			}
			c := componenttest.NewComponent(t, conf)
			New(c)
			componenttest.StartComponent(t, c)
			defer c.Close()

			req := httptest.NewRequest(http.MethodGet, "/api/v2/frequency-plans/EU_863_870", nil).WithContext(ctx)
			if tc.SetupRequest != nil {
				tc.SetupRequest(req)
			}

			rec := httptest.NewRecorder()
			c.ServeHTTP(rec, req)
			tc.ResponseAssertion(a, rec)
		})
	}
}
