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

package cups

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	mockGatewayID   = "test-gateway"
	mockErrNotFound = grpc.Errorf(codes.NotFound, "not found")
)

type mockGatewayClientData struct {
	ctx struct {
		Get context.Context
	}
	req struct {
		Get *ttnpb.GetGatewayRequest
	}
	opts struct {
		Get []grpc.CallOption
	}
	res struct {
		Get *ttnpb.Gateway
	}
	err struct {
		Get error
	}
}

type mockGatewayClient struct {
	mockGatewayClientData
	ttnpb.GatewayRegistryClient
}

func (m *mockGatewayClient) reset() {
	m.mockGatewayClientData = mockGatewayClientData{}
}

func (m *mockGatewayClient) Get(ctx context.Context, in *ttnpb.GetGatewayRequest, opts ...grpc.CallOption) (*ttnpb.Gateway, error) {
	m.ctx.Get, m.req.Get, m.opts.Get = ctx, in, opts
	return m.res.Get, m.err.Get
}

func mockGateway() *ttnpb.Gateway {
	return &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{
			GatewayID: mockGatewayID,
		},
		UpdateChannel:        "stable",
		FrequencyPlanID:      "EU_863_870",
		GatewayServerAddress: "mqtts://localhost:8883",
	}
}

func TestServer(t *testing.T) {
	e := echo.New()

	for _, tt := range []struct {
		Name           string
		StoreSetup     func(*mockGatewayClient)
		Options        []Option
		RequestSetup   func(*http.Request)
		ContextSetup   func(echo.Context)
		AssertError    func(actual interface{}, expected ...interface{}) string
		AssertStore    func(*assertions.Assertion, *mockGatewayClient)
		AssertResponse func(*assertions.Assertion, *httptest.ResponseRecorder)
	}{
		{
			Name: "No Auth",
			RequestSetup: func(req *http.Request) {
				req.Header.Del(echo.HeaderAuthorization)
			},
			AssertError: should.NotBeNil,
		},
		{
			Name: "No GatewayID",
			ContextSetup: func(c echo.Context) {
				c.SetParamValues()
			},
			AssertError: should.NotBeNil,
		},
		{
			Name: "Not Found",
			StoreSetup: func(c *mockGatewayClient) {
				c.err.Get = mockErrNotFound
			},
			AssertError: should.NotBeNil,
			AssertStore: func(a *assertions.Assertion, c *mockGatewayClient) {
				a.So(c.req.Get.GatewayID, should.Equal, mockGatewayID)
			},
		},
		{
			Name: "Found",
			StoreSetup: func(c *mockGatewayClient) {
				c.res.Get = mockGateway()
			},
			AssertError: should.BeNil,
			AssertStore: func(a *assertions.Assertion, c *mockGatewayClient) {
				a.So(c.req.Get.GatewayID, should.Equal, mockGatewayID)
			},
			AssertResponse: func(a *assertions.Assertion, rc *httptest.ResponseRecorder) {
				a.So(rc.Code, should.Equal, http.StatusOK)
				body := rc.Body
				a.So(body, should.NotBeEmpty)
				var resp map[string]interface{}
				err := json.NewDecoder(rc.Body).Decode(&resp)
				a.So(err, should.BeNil)
				a.So(resp["frequency_plan"], should.Equal, "EU_863_870")
				a.So(resp["frequency_plan_url"], should.Equal, "http://example.com/api/v2/frequency-plans/EU_863_870")
				a.So(resp["firmware_url"], should.Equal, "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1/stable")
				a.So(resp["router"], should.Resemble, map[string]interface{}{
					"mqtt_address": "mqtts://localhost:8883",
				})
				a.So(resp["auto_update"], should.Equal, false)
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			store := &mockGatewayClient{}
			if tt.StoreSetup != nil {
				tt.StoreSetup(store)
			}

			s := NewServer(component.MustNew(test.GetLogger(t), &component.Config{}), append([]Option{
				WithRegistry(store),
				WithDefaultFirmwareURL(defaultFirmwarePath),
			}, tt.Options...)...)
			req := httptest.NewRequest(http.MethodGet, "/api/v2/gateway/test-gateway", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "random string")
			if tt.RequestSetup != nil {
				tt.RequestSetup(req)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames(gatewayIDKey)
			c.SetParamValues(mockGatewayID)
			if tt.ContextSetup != nil {
				tt.ContextSetup(c)
			}
			middleware := []echo.MiddlewareFunc{
				s.validateAndFillGatewayIDs(),
				s.checkAuthPresence(),
			}

			handler := s.handleGatewayInfo
			for _, m := range middleware {
				handler = m(handler)
			}
			err := handler(c)
			if tt.AssertError != nil {
				a.So(err, tt.AssertError)
			}
			if tt.AssertResponse != nil {
				tt.AssertResponse(a, rec)
			}
			if tt.AssertStore != nil {
				tt.AssertStore(a, store)
			}
		})
	}
}
