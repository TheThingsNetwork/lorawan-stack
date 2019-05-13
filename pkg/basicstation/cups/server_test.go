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
	"bytes"
	"context"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestGetTrust(t *testing.T) {
	a := assertions.New(t)

	s := new(Server)

	for _, addr := range []string{
		"thethingsnetwork.org:443",
		"https://thethingsnetwork.org:443",
		"https://thethingsnetwork.org",
	} {
		cert, err := s.getTrust(addr)
		a.So(err, should.BeNil)
		a.So(cert, should.NotBeNil)
	}
}

type mockGatewayClientData struct {
	ctx struct {
		GetIdentifiersForEUI context.Context
		Create               context.Context
		Get                  context.Context
		Update               context.Context
		CreateAPIKey         context.Context
		UpdateAPIKey         context.Context
	}
	req struct {
		GetIdentifiersForEUI *ttnpb.GetGatewayIdentifiersForEUIRequest
		Create               *ttnpb.CreateGatewayRequest
		Get                  *ttnpb.GetGatewayRequest
		Update               *ttnpb.UpdateGatewayRequest
		CreateAPIKey         *ttnpb.CreateGatewayAPIKeyRequest
		UpdateAPIKey         *ttnpb.UpdateGatewayAPIKeyRequest
	}
	opts struct {
		GetIdentifiersForEUI []grpc.CallOption
		Create               []grpc.CallOption
		Get                  []grpc.CallOption
		Update               []grpc.CallOption
		CreateAPIKey         []grpc.CallOption
		UpdateAPIKey         []grpc.CallOption
	}
	res struct {
		GetIdentifiersForEUI *ttnpb.GatewayIdentifiers
		Create               *ttnpb.Gateway
		Get                  *ttnpb.Gateway
		Update               *ttnpb.Gateway
		CreateAPIKey         *ttnpb.APIKey
		UpdateAPIKey         *ttnpb.APIKey
	}
	err struct {
		GetIdentifiersForEUI error
		Create               error
		Get                  error
		Update               error
		CreateAPIKey         error
		UpdateAPIKey         error
	}
}

type mockGatewayClient struct {
	mockGatewayClientData
	ttnpb.GatewayRegistryClient
	ttnpb.GatewayAccessClient
}

func (m *mockGatewayClient) reset() {
	m.mockGatewayClientData = mockGatewayClientData{}
}

func (m *mockGatewayClient) GetIdentifiersForEUI(ctx context.Context, in *ttnpb.GetGatewayIdentifiersForEUIRequest, opts ...grpc.CallOption) (*ttnpb.GatewayIdentifiers, error) {
	m.ctx.GetIdentifiersForEUI, m.req.GetIdentifiersForEUI, m.opts.GetIdentifiersForEUI = ctx, in, opts
	return m.res.GetIdentifiersForEUI, m.err.GetIdentifiersForEUI
}

func (m *mockGatewayClient) Create(ctx context.Context, in *ttnpb.CreateGatewayRequest, opts ...grpc.CallOption) (*ttnpb.Gateway, error) {
	m.ctx.Create, m.req.Create, m.opts.Create = ctx, in, opts
	return m.res.Create, m.err.Create
}

func (m *mockGatewayClient) Get(ctx context.Context, in *ttnpb.GetGatewayRequest, opts ...grpc.CallOption) (*ttnpb.Gateway, error) {
	m.ctx.Get, m.req.Get, m.opts.Get = ctx, in, opts
	return m.res.Get, m.err.Get
}

func (m *mockGatewayClient) Update(ctx context.Context, in *ttnpb.UpdateGatewayRequest, opts ...grpc.CallOption) (*ttnpb.Gateway, error) {
	m.ctx.Update, m.req.Update, m.opts.Update = ctx, in, opts
	return m.res.Update, m.err.Update
}

func (m *mockGatewayClient) CreateAPIKey(ctx context.Context, in *ttnpb.CreateGatewayAPIKeyRequest, opts ...grpc.CallOption) (*ttnpb.APIKey, error) {
	m.ctx.CreateAPIKey, m.req.CreateAPIKey, m.opts.CreateAPIKey = ctx, in, opts
	return m.res.CreateAPIKey, m.err.CreateAPIKey
}

func (m *mockGatewayClient) UpdateAPIKey(ctx context.Context, in *ttnpb.UpdateGatewayAPIKeyRequest, opts ...grpc.CallOption) (*ttnpb.APIKey, error) {
	m.ctx.UpdateAPIKey, m.req.UpdateAPIKey, m.opts.UpdateAPIKey = ctx, in, opts
	return m.res.UpdateAPIKey, m.err.UpdateAPIKey
}

const updateInfoRequest = `{
  "router": "58a0:cbff:fe80:19",
  "cupsUri": "https://mh.sm.tc:7007",
  "tcUri": "wss://mh.sm.tc:7000",
  "cupsCredCrc": 1398343300,
  "tcCredCrc": 3337464763,
  "station": "2.0.0(minihub/debug) 2018-12-06 09:30:35",
  "model": "minihub",
  "package": "2.0.0",
  "keys": [
    392840017
  ]
}`

func mockGateway() *ttnpb.Gateway {
	return &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{
			GatewayID: "test-gateway",
			EUI:       &mockGatewayEUI,
		},
		Attributes: map[string]string{
			cupsURIAttribute:            "https://mh.sm.tc:7007",
			cupsCredentialsCRCAttribute: "1398343300",
			lnsCredentialsCRCAttribute:  "3337464763",
			cupsStationAttribute:        "2.0.0(minihub/debug) 2018-12-06 09:30:35",
			cupsModelAttribute:          "minihub",
			cupsPackageAttribute:        "2.0.0",
		},
		GatewayServerAddress: "wss://mh.sm.tc:7000",
	}
}

var (
	mockFallbackAuth = grpc.PerRPCCredentials(nil)
	mockAuthFunc     = func(ctx context.Context, gatewayEUI types.EUI64, auth string) grpc.CallOption {
		return mockFallbackAuth
	}
	mockGatewayEUI  = types.EUI64{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x19}
	mockErrNotFound = grpc.Errorf(codes.NotFound, "not found")
)

func TestServer(t *testing.T) {
	e := echo.New()

	for _, tt := range []struct {
		Name           string
		StoreSetup     func(*mockGatewayClient)
		Options        []Option
		RequestSetup   func(*http.Request)
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
			Name: "Not Found",
			StoreSetup: func(c *mockGatewayClient) {
				c.err.GetIdentifiersForEUI = mockErrNotFound
			},
			AssertError: should.NotBeNil,
			AssertStore: func(a *assertions.Assertion, c *mockGatewayClient) {
				a.So(c.req.GetIdentifiersForEUI.EUI, should.Equal, mockGatewayEUI)
			},
		},
		{
			Name: "No Changes",
			StoreSetup: func(c *mockGatewayClient) {
				c.res.Get = mockGateway()
				c.res.GetIdentifiersForEUI = &c.res.Get.GatewayIdentifiers
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				var res UpdateInfoResponse
				err := res.UnmarshalBinary(rec.Body.Bytes())
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, UpdateInfoResponse{})
			},
		},
		{
			Name: "Not Explicitly Enabled",
			StoreSetup: func(c *mockGatewayClient) {
				c.res.Get = mockGateway()
				c.res.GetIdentifiersForEUI = &c.res.Get.GatewayIdentifiers
			},
			Options: []Option{
				WithExplicitEnable(true),
			},
			AssertError: should.NotBeNil,
		},
		{
			Name: "Invalid Credentials",
			StoreSetup: func(c *mockGatewayClient) {
				c.res.Get = mockGateway()
				c.res.GetIdentifiersForEUI = &c.res.Get.GatewayIdentifiers
				c.res.Get.Attributes[cupsCredentialsAttribute] = "other string"
			},
			AssertError: should.NotBeNil,
		},
		{
			Name: "Register New Gateway",
			StoreSetup: func(c *mockGatewayClient) {
				c.err.GetIdentifiersForEUI = mockErrNotFound
				c.res.Create = mockGateway()
			},
			Options: []Option{
				WithRegisterUnknown(&ttnpb.OrganizationOrUserIdentifiers{}),
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				var res UpdateInfoResponse
				err := res.UnmarshalBinary(rec.Body.Bytes())
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, UpdateInfoResponse{})
			},
			AssertStore: func(a *assertions.Assertion, s *mockGatewayClient) {
				a.So(s.req.Update.Gateway.Attributes[cupsCredentialsCRCAttribute], should.NotBeEmpty)
				a.So(s.req.Update.Gateway.Attributes[lnsCredentialsCRCAttribute], should.NotBeEmpty)
			},
		},
		{
			Name: "Updated Config",
			StoreSetup: func(c *mockGatewayClient) {
				c.res.Get = mockGateway()
				c.res.GetIdentifiersForEUI = &c.res.Get.GatewayIdentifiers
				c.res.Get.Attributes[cupsURIAttribute] = "https://thethingsnetwork.org:443"
				c.res.Get.GatewayServerAddress = "wss://thethingsnetwork.org:443"
				c.res.Get.Attributes[cupsCredentialsCRCAttribute] = ""
				c.res.Get.Attributes[lnsCredentialsCRCAttribute] = ""
				c.res.CreateAPIKey = &ttnpb.APIKey{
					ID:  "KEYID",
					Key: "KEYCONTENTS",
				}
			},
			Options: []Option{
				WithTrust(&x509.Certificate{
					Raw: []byte("FAKE CERTIFICATE CONTENTS"),
				}),
				WithAllowCUPSURIUpdate(true),
			},
			RequestSetup: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+auth.JoinToken(auth.APIKey, "ID", "KEY"))
			},
			AssertError: should.BeNil,
			AssertResponse: func(a *assertions.Assertion, rec *httptest.ResponseRecorder) {
				var res UpdateInfoResponse
				err := res.UnmarshalBinary(rec.Body.Bytes())
				a.So(err, should.BeNil)
				a.So(res.CUPSURI, should.Equal, "https://thethingsnetwork.org:443")
				a.So(res.LNSURI, should.Equal, "wss://thethingsnetwork.org:443")
				a.So(bytes.Contains(res.CUPSCredentials, []byte("FAKE CERTIFICATE CONTENTS")), should.BeFalse)
				a.So(bytes.Contains(res.CUPSCredentials, []byte("KEYCONTENTS")), should.BeTrue)
				a.So(bytes.Contains(res.LNSCredentials, []byte("KEYCONTENTS")), should.BeTrue)
				a.So(bytes.Equal(res.CUPSCredentials, res.LNSCredentials), should.BeTrue)
			},
			AssertStore: func(a *assertions.Assertion, s *mockGatewayClient) {
				a.So(s.req.Update.Gateway.Attributes[cupsCredentialsCRCAttribute], should.NotBeEmpty)
				a.So(s.req.Update.Gateway.Attributes[lnsCredentialsCRCAttribute], should.NotBeEmpty)
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
				WithAuth(mockAuthFunc),
				WithRegistries(store, store),
			}, tt.Options...)...)
			req := httptest.NewRequest(http.MethodPost, "/update-info", strings.NewReader(updateInfoRequest))
			req = req.WithContext(test.Context())
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "random string")
			if tt.RequestSetup != nil {
				tt.RequestSetup(req)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			err := s.UpdateInfo(c)
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
