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

package gcsv2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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

func TestGetGateway(t *testing.T) {
	a := assertions.New(t)

	reg := &mockGatewayRegistryClient{}
	auth := func(ctx context.Context) grpc.CallOption {
		return grpc.PerRPCCredentials(nil)
	}
	c := componenttest.NewComponent(t, &component.Config{})
	s := New(c, WithRegistry(reg), WithAuth(auth))

	mockRightsFetcher := rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
		md := rpcmetadata.FromIncomingContext(ctx)
		if strings.ToLower(md.AuthType) == "bearer" {
			return ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_INFO), nil
		}
		return nil, rights.ErrNoGatewayRights
	})

	e := echo.New()

	for _, tt := range []struct {
		Name           string
		StoreSetup     func(*mockGatewayRegistryClient)
		RequestSetup   func(*http.Request)
		AssertError    func(actual interface{}, expected ...interface{}) string
		AssertStore    func(*assertions.Assertion, *mockGatewayRegistryClient)
		AssertResponse func(*assertions.Assertion, *httptest.ResponseRecorder)
	}{
		{
			Name: "Gateway Not Found",
			StoreSetup: func(reg *mockGatewayRegistryClient) {
				reg.out, reg.err = nil, status.Error(codes.NotFound, "not found")
			},
			AssertError: should.NotBeNil,
		},
		{
			Name: "Gateway With Key",
			StoreSetup: func(reg *mockGatewayRegistryClient) {
				reg.out, reg.err = &ttnpb.Gateway{
					Description: "Gateway Description",
					Attributes: map[string]string{
						"key": "some-key",
					},
					FrequencyPlanID:      "EU_863_870",
					GatewayServerAddress: "gatewayserver",
					Antennas: []ttnpb.GatewayAntenna{
						{Location: ttnpb.Location{Latitude: 12.34, Longitude: 56.78, Altitude: 90}},
					},
				}, nil
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				body := rec.Body.String()
				a.So(body, assertions.ShouldContainSubstring, `"attributes":{"description":"Gateway Description"}`)
				a.So(body, assertions.ShouldContainSubstring, `"frequency_plan":"EU_863_870"`)
				a.So(body, assertions.ShouldContainSubstring, `"frequency_plan_url":"http://example.com/api/v2/frequency-plans/EU_863_870"`)
				a.So(body, assertions.ShouldContainSubstring, `"router":{"id":"gatewayserver","mqtt_address":"mqtts://gatewayserver:8881"}`)
				a.So(body, assertions.ShouldContainSubstring, `"antenna_location":{"latitude":12.34,"longitude":56.78,"altitude":90}`)
			},
		},
		{
			Name: "Same but as TTKG",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("User-Agent", "TTNGateway")
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				body := rec.Body.String()
				a.So(body, assertions.ShouldNotContainSubstring, `"attributes"`)
				a.So(body, assertions.ShouldContainSubstring, `"router":{"mqtt_address":"mqtts://gatewayserver:8881"}`)
			},
		},
		{
			Name: "Same but without Auth",
			RequestSetup: func(req *http.Request) {
				req.Header.Del(echo.HeaderAuthorization)
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				body := rec.Body.String()
				a.So(body, assertions.ShouldNotContainSubstring, `"router":{"mqtt_address":"mqtts://gatewayserver:8881"}`)
			},
		},
	} {
		if tt.StoreSetup != nil {
			tt.StoreSetup(reg)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/v2/gateways/foo-gtw", nil)
		ctx := test.Context()
		ctx = log.NewContext(ctx, test.GetLogger(t))
		ctx = rights.NewContextWithFetcher(ctx, mockRightsFetcher)
		req = req.WithContext(ctx)
		req.Header.Set(echo.HeaderAuthorization, "key some-key")
		if tt.RequestSetup != nil {
			tt.RequestSetup(req)
		}
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v2/gateways/:gateway_id")
		c.SetParamNames("gateway_id")
		c.SetParamValues("foo-gtw")
		err := s.normalizeAuthorization(s.handleGetGateway)(c)
		if tt.AssertError != nil {
			a.So(err, tt.AssertError)
		}
		if tt.AssertResponse != nil {
			tt.AssertResponse(a, rec)
		}
		if tt.AssertStore != nil {
			tt.AssertStore(a, reg)
		}
	}
}

func TestGetFrequencyPlan(t *testing.T) {
	a := assertions.New(t)

	c := componenttest.NewComponent(t, &component.Config{})
	c.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
	s := New(c)

	e := echo.New()

	for _, tt := range []struct {
		Name           string
		RequestSetup   func(*http.Request)
		AssertError    func(actual interface{}, expected ...interface{}) string
		AssertResponse func(*assertions.Assertion, *httptest.ResponseRecorder)
	}{
		{
			Name:        "Regular Request",
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				body := rec.Body.String()
				a.So(body, assertions.ShouldContainSubstring, `"SX1301_conf"`)
				a.So(body, assertions.ShouldContainSubstring, `"chan_multiSF_0"`)
				a.So(body, assertions.ShouldContainSubstring, `"tx_lut_0"`)
			},
		},
		{
			Name: "Same but as TTKG",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("User-Agent", "TTNGateway")
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				body := rec.Body.String()
				a.So(body, assertions.ShouldNotContainSubstring, `"tx_lut_0"`)
			},
		},
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/frequency-plans/EU_863_870", nil)
		ctx := test.Context()
		ctx = log.NewContext(ctx, test.GetLogger(t))
		req = req.WithContext(ctx)
		if tt.RequestSetup != nil {
			tt.RequestSetup(req)
		}
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v2/frequency-plans/:frequency_plan_id")
		c.SetParamNames("frequency_plan_id")
		c.SetParamValues("EU_863_870")
		err := s.normalizeAuthorization(s.handleGetFrequencyPlan)(c)
		if tt.AssertError != nil {
			a.So(err, tt.AssertError)
		}
		if tt.AssertResponse != nil {
			tt.AssertResponse(a, rec)
		}
	}
}
