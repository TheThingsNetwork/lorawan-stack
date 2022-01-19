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
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/gorilla/websocket"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/basicstation"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	pfconfig "go.thethings.network/lorawan-stack/v3/pkg/pfconfig/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	serverAddress          = "127.0.0.1:0"
	registeredGatewayUID   = "eui-0101010101010101"
	registeredGatewayID    = ttnpb.GatewayIdentifiers{GatewayId: registeredGatewayUID}
	registeredGateway      = ttnpb.Gateway{Ids: &registeredGatewayID, FrequencyPlanId: "EU_863_870"}
	registeredGatewayToken = "secrettoken"

	discoveryEndPoint      = "/router-info"
	connectionRootEndPoint = "/traffic/"

	testTrafficEndPoint = "/traffic/eui-0101010101010101"

	timeout       = (1 << 7) * test.Delay
	defaultConfig = Config{
		WSPingInterval:       (1 << 3) * test.Delay,
		AllowUnauthenticated: true,
		UseTrafficTLSAddress: false,
	}

	maxValidRoundTripDelay = (1 << 4) * test.Delay
)

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestClientTokenAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayToken)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	gs := mock.NewServer(c)

	for _, ttc := range []struct {
		Name                 string
		AllowUnauthenticated bool
	}{
		{
			Name:                 "ServerAllowUnauthenticated",
			AllowUnauthenticated: true,
		},
		{
			Name:                 "ServerDontAllowUnauthenticated",
			AllowUnauthenticated: false,
		},
	} {
		cfg := defaultConfig
		cfg.AllowUnauthenticated = ttc.AllowUnauthenticated
		web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), cfg)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		lis, err := net.Listen("tcp", serverAddress)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		defer lis.Close()
		go func() error {
			return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				web.ServeHTTP(w, r)
			}))
		}()
		servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

		for _, tc := range []struct {
			Name           string
			GatewayID      string
			AuthToken      string
			TokenPrefix    string
			ErrorAssertion func(err error) bool
		}{
			{
				Name:           "RegisteredGatewayAndValidKey",
				GatewayID:      registeredGatewayID.GatewayId,
				AuthToken:      registeredGatewayToken,
				ErrorAssertion: nil,
			},
			{
				Name:           "RegisteredGatewayAndValidKey",
				GatewayID:      registeredGatewayID.GatewayId,
				AuthToken:      registeredGatewayToken,
				TokenPrefix:    "Bearer ",
				ErrorAssertion: nil,
			},
			{
				Name:      "RegisteredGatewayAndInValidKey",
				GatewayID: registeredGatewayID.GatewayId,
				AuthToken: "invalidToken",
				ErrorAssertion: func(err error) bool {
					if err == nil {
						return false
					}
					return err == websocket.ErrBadHandshake
				},
			},
			{
				Name:      "RegisteredGatewayAndNoKey",
				GatewayID: registeredGatewayID.GatewayId,
				ErrorAssertion: func(err error) bool {
					if ttc.AllowUnauthenticated && err == nil {
						return true
					}
					return err == websocket.ErrBadHandshake
				},
			},
			{
				Name:      "UnregisteredGateway",
				GatewayID: "eui-1122334455667788",
				AuthToken: registeredGatewayToken,
				ErrorAssertion: func(err error) bool {
					if err == nil {
						return false
					}
					return err == websocket.ErrBadHandshake
				},
			},
		} {
			t.Run(fmt.Sprintf("%s/%s", ttc.Name, tc.Name), func(t *testing.T) {
				a := assertions.New(t)
				h := http.Header{}
				h.Set("Authorization", fmt.Sprintf("%s%s", tc.TokenPrefix, tc.AuthToken))
				conn, _, err := websocket.DefaultDialer.Dial(servAddr+connectionRootEndPoint+tc.GatewayID, h)
				if err != nil {
					if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
						t.Fatalf("Unexpected error: %v", err)
					}
				} else if tc.ErrorAssertion != nil {
					a.So(tc.ErrorAssertion(err), should.BeTrue)
				}
				if conn != nil {
					conn.Close()
				}
			})
		}
	}
}

func TestClientSideTLS(t *testing.T) {
	// TODO: https://github.com/TheThingsNetwork/lorawan-stack/issues/558
}

func TestDiscover(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayToken)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	gs := mock.NewServer(c)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go func() error {
		return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			web.ServeHTTP(w, r)
		}))
	}()
	servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

	// Invalid Endpoints
	for i, tc := range []struct {
		URL string
	}{
		{
			URL: servAddr + "/api/v3/gs/io/basicstation/discover",
		},
		{
			URL: servAddr + discoveryEndPoint + "/eui-0101010101010101",
		},
	} {
		t.Run(fmt.Sprintf("InvalidDiscoveryEndPoint/%d", i), func(t *testing.T) {
			a := assertions.New(t)
			_, res, err := websocket.DefaultDialer.Dial(tc.URL, nil)
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("Unexpected response received: %v", res.Status)
			}
			if !a.So(err, should.Equal, websocket.ErrBadHandshake) {
				t.Fatalf("Connection failed: %v", err)
			}
		})
	}

	// Test Queries
	for _, tc := range []struct {
		Name     string
		Query    interface{}
		Response lbslns.DiscoverResponse
	}{
		{
			Name:     "EmptyEUI",
			Query:    lbslns.DiscoverQuery{},
			Response: lbslns.DiscoverResponse{Error: "Empty router EUI provided"},
		},
		{
			Name:     "EmptyStruct",
			Query:    struct{}{},
			Response: lbslns.DiscoverResponse{Error: "Empty router EUI provided"},
		},
		{
			Name: "InvalidJSONKey",
			Query: struct {
				EUI string `json:"route"`
			}{EUI: `"01-02-03-04-05-06-07-08"`},
			Response: lbslns.DiscoverResponse{Error: "Empty router EUI provided"},
		},
	} {
		t.Run(fmt.Sprintf("InvalidQuery/%s", tc.Name), func(t *testing.T) {
			a := assertions.New(t)
			conn, _, err := websocket.DefaultDialer.Dial(servAddr+discoveryEndPoint, nil)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Connection failed: %v", err)
			}
			defer conn.Close()
			req, err := json.Marshal(tc.Query)
			if err != nil {
				panic(err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, req); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}

			var readErr error
			resCh := make(chan []byte)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, data, err := conn.ReadMessage()
				if err != nil {
					close(resCh)
					if err == websocket.ErrBadHandshake {
						return
					}
					readErr = err
					return
				}
				resCh <- data
			}()
			select {
			case res := <-resCh:
				var response lbslns.DiscoverResponse
				if err := json.Unmarshal(res, &response); err != nil {
					t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
				}
				a.So(response, should.Resemble, tc.Response)
			case <-time.After(timeout):
				t.Fatal("Read message timeout")
			}
			wg.Wait()
			if readErr != nil {
				t.Fatalf("Failed to read message: %v", readErr)
			}
		})
	}

	for _, tc := range []struct {
		Name  string
		Query interface{}
	}{
		{
			Name: "InvalidLength",
			Query: struct {
				EUI string `json:"router"`
			}{EUI: `"01-02-03-04-05-06-07-08-09"`},
		},
		{
			Name: "InvalidLength",
			Query: struct {
				EUI string `json:"router"`
			}{EUI: `"01:02:03:04:05:06:07:08:09"`},
		},
		{
			Name: "InvalidEUIFormat",
			Query: struct {
				EUI string `json:"router"`
			}{EUI: `"01:02:03:04:05:06:07-08"`},
		},
	} {
		t.Run(fmt.Sprintf("InvalidQuery/%s", tc.Name), func(t *testing.T) {
			a := assertions.New(t)
			conn, _, err := websocket.DefaultDialer.Dial(servAddr+discoveryEndPoint, nil)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Connection failed: %v", err)
			}
			defer conn.Close()
			req, err := json.Marshal(tc.Query)
			if err != nil {
				panic(err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, req); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err = conn.ReadMessage()
			}()
			wg.Wait()
			if err != nil {
				t.Fatalf("Failed to read message: %v", err)
			}
		})
	}

	// Valid
	for i, tc := range []struct {
		EndPointEUI string
		EUI         types.EUI64
		Query       interface{}
	}{
		{
			EndPointEUI: "1111111111111111",
			EUI:         types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			Query: lbslns.DiscoverQuery{
				EUI: basicstation.EUI{
					Prefix: "router",
					EUI64:  types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("ValidQuery/%d", i), func(t *testing.T) {
			a := assertions.New(t)
			conn, _, err := websocket.DefaultDialer.Dial(servAddr+discoveryEndPoint, nil)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Connection failed: %v", err)
			}
			defer conn.Close()
			req, err := json.Marshal(tc.Query)
			if err != nil {
				panic(err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, req); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}
			var readErr error
			resCh := make(chan []byte)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, data, err := conn.ReadMessage()
				if err != nil {
					close(resCh)
					if err == websocket.ErrBadHandshake {
						return
					}
					readErr = err
					return
				}
				resCh <- data
			}()
			select {
			case res := <-resCh:
				var response lbslns.DiscoverResponse
				if err := json.Unmarshal(res, &response); err != nil {
					t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
				}
				a.So(response, should.Resemble, lbslns.DiscoverResponse{
					EUI: basicstation.EUI{Prefix: "router", EUI64: tc.EUI},
					Muxs: basicstation.EUI{
						Prefix: "muxs",
					},
					URI: servAddr + connectionRootEndPoint + "eui-" + tc.EndPointEUI,
				})
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
			wg.Wait()
			if readErr != nil {
				t.Fatalf("Failed to read message: %v", readErr)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	gs := mock.NewServer(c)

	gs.RegisterGateway(ctx, registeredGatewayID, &ttnpb.Gateway{
		Ids:             &registeredGatewayID,
		FrequencyPlanId: test.EUFrequencyPlanID,
		Antennas: []*ttnpb.GatewayAntenna{
			{
				Gain: 3,
			},
		},
	})

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go func() error {
		return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			web.ServeHTTP(w, r)
		}))
	}()
	servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

	conn, _, err := websocket.DefaultDialer.Dial(servAddr+testTrafficEndPoint, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	var gsConn *io.Connection
	select {
	case gsConn = <-gs.Connections():
	case <-time.After(timeout):
		t.Fatal("Connection timeout")
	}

	for _, tc := range []struct {
		Name                  string
		VersionQuery          interface{}
		ExpectedRouterConfig  interface{}
		ExpectedStatusMessage ttnpb.GatewayStatus
	}{
		{
			Name: "VersionProd",
			VersionQuery: lbslns.Version{
				Station:  "test-station",
				Firmware: "1.0.0",
				Package:  "test-package",
				Model:    "test-model",
				Protocol: 2,
				Features: "prod gps",
			},
			ExpectedRouterConfig: pfconfig.RouterConfig{
				Region:         "EU863",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{863000000, 870000000},
				DataRates: [16][3]int{
					{12, 125, 0},
					{11, 125, 0},
					{10, 125, 0},
					{9, 125, 0},
					{8, 125, 0},
					{7, 125, 0},
					{7, 250, 0},
					{0, 0, 0},
				},
				SX1301Config: []pfconfig.LBSSX1301Config{
					{
						Radios: []pfconfig.LBSRFConfig{
							{
								Enable:      true,
								Frequency:   867500000,
								AntennaGain: 3,
							},
							{
								Enable:      true,
								Frequency:   868500000,
								AntennaGain: 3,
							},
						},
						Channels: []shared.IFConfig{
							{Enable: true, Radio: 0, IFValue: 600000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 800000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 1000000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: -400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: -200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &shared.IFConfig{Enable: true, Radio: 0, IFValue: 800000, Bandwidth: 250000, SpreadFactor: 7, Datarate: 0},
						FSKChannel:          &shared.IFConfig{Enable: true, Radio: 0, IFValue: 1300000, Bandwidth: 0, SpreadFactor: 0, Datarate: 50000},
					},
				},
			},
			ExpectedStatusMessage: ttnpb.GatewayStatus{
				Versions: map[string]string{
					"station":  "test-station",
					"firmware": "1.0.0",
					"package":  "test-package",
					"platform": "test-model - Firmware 1.0.0 - Protocol 2",
				},
				Advanced: &pbtypes.Struct{
					Fields: map[string]*pbtypes.Value{
						"model": {
							Kind: &pbtypes.Value_StringValue{StringValue: "test-model"},
						},
						"features": {
							Kind: &pbtypes.Value_StringValue{StringValue: "prod gps"},
						},
					},
				},
			},
		},
		{
			Name: "VersionDebug",
			VersionQuery: lbslns.Version{
				Station:  "test-station-rc1",
				Firmware: "1.0.0",
				Package:  "test-package",
				Model:    "test-model",
				Protocol: 2,
				Features: "rmtsh gps",
			},
			ExpectedRouterConfig: pfconfig.RouterConfig{
				Region:         "EU863",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{863000000, 870000000},
				DataRates: [16][3]int{
					{12, 125, 0},
					{11, 125, 0},
					{10, 125, 0},
					{9, 125, 0},
					{8, 125, 0},
					{7, 125, 0},
					{7, 250, 0},
					{0, 0, 0},
				},
				NoCCA:       true,
				NoDwellTime: true,
				NoDutyCycle: true,
				SX1301Config: []pfconfig.LBSSX1301Config{
					{
						Radios: []pfconfig.LBSRFConfig{
							{
								Enable:      true,
								Frequency:   867500000,
								AntennaGain: 3,
							},
							{
								Enable:      true,
								Frequency:   868500000,
								AntennaGain: 3,
							},
						},
						Channels: []shared.IFConfig{
							{Enable: true, Radio: 0, IFValue: 600000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 800000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 1000000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: -400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: -200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: true, Radio: 0, IFValue: 400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &shared.IFConfig{Enable: true, Radio: 0, IFValue: 800000, Bandwidth: 250000, SpreadFactor: 7, Datarate: 0},
						FSKChannel:          &shared.IFConfig{Enable: true, Radio: 0, IFValue: 1300000, Bandwidth: 0, SpreadFactor: 0, Datarate: 50000},
					},
				},
			},
			ExpectedStatusMessage: ttnpb.GatewayStatus{
				Versions: map[string]string{
					"station":  "test-station-rc1",
					"firmware": "1.0.0",
					"package":  "test-package",
					"platform": "test-model - Firmware 1.0.0 - Protocol 2",
				},
				Advanced: &pbtypes.Struct{
					Fields: map[string]*pbtypes.Value{
						"model": {
							Kind: &pbtypes.Value_StringValue{StringValue: "test-model"},
						},
						"features": {
							Kind: &pbtypes.Value_StringValue{StringValue: "rmtsh gps"},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			reqVersion, err := json.Marshal(tc.VersionQuery)
			if err != nil {
				panic(err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, reqVersion); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}

			var readErr error
			var wg sync.WaitGroup
			resCh := make(chan []byte)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, data, err := conn.ReadMessage()
				if err != nil {
					readErr = err
					return
				}
				resCh <- data
			}()
			select {
			case res := <-resCh:
				var response pfconfig.RouterConfig
				if err := json.Unmarshal(res, &response); err != nil {
					t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
				}
				response.MuxTime = tc.ExpectedRouterConfig.(pfconfig.RouterConfig).MuxTime
				a.So(response, should.Resemble, tc.ExpectedRouterConfig)
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
			wg.Wait()
			if readErr != nil {
				t.Fatalf("Failed to read message: %v", err)
			}
			select {
			case stat := <-gsConn.Status():
				if a.So(stat.Time, should.NotBeNil) {
					a.So(time.Since(*ttnpb.StdTime(stat.Time)), should.BeLessThan, timeout)
					stat.Time = nil
				}
				a.So(stat, should.Resemble, &tc.ExpectedStatusMessage)
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
		})
	}
}

func TestTraffic(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayToken)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	gs := mock.NewServer(c)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go func() error {
		return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			web.ServeHTTP(w, r)
		}))
	}()
	servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

	wsConn, _, err := websocket.DefaultDialer.Dial(servAddr+testTrafficEndPoint, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer wsConn.Close()

	var gsConn *io.Connection
	select {
	case gsConn = <-gs.Connections():
	case <-time.After(timeout):
		t.Fatal("Connection timeout")
	}

	for _, tc := range []struct {
		Name                    string
		InputBSUpstream         interface{}
		InputNetworkDownstream  *ttnpb.DownlinkMessage
		InputDownlinkPath       *ttnpb.DownlinkPath
		ExpectedBSDownstream    interface{}
		ExpectedNetworkUpstream interface{}
	}{
		{
			Name: "JoinRequest",
			InputBSUpstream: lbslns.JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
				DevNonce: 18000,
				MIC:      12345678,
				RadioMetaData: lbslns.RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: lbslns.UpInfo{
						RxTime: 1548059982,
						XTime:  12666373963464220,
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ExpectedNetworkUpstream: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					Mic:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEui:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
						DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						DevNonce: [2]byte{0x46, 0x50},
					}},
				},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIds: &ttnpb.GatewayIdentifiers{
						GatewayId: "eui-0101010101010101",
						Eui:       &types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
					},
					Time:        ttnpb.ProtoTimePtr(time.Unix(1548059982, 0)),
					Timestamp:   (uint32)(12666373963464220 & 0xFFFFFFFF),
					Rssi:        89,
					ChannelRssi: 89,
					Snr:         9.25,
				}},
				Settings: &ttnpb.TxSettings{
					Frequency:  868300000,
					CodingRate: "4/5",
					Time:       ttnpb.ProtoTimePtr(time.Unix(1548059982, 0)),
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
		{
			Name: "UplinkFrame",
			InputBSUpstream: lbslns.UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      0x00,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "5fcc",
				MIC:        12345678,
				RadioMetaData: lbslns.RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: lbslns.UpInfo{
						RxTime: 1548059982,
						XTime:  12666373963464220,
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ExpectedNetworkUpstream: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					Mic:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
						FPort:      0,
						FrmPayload: []byte{0x5F, 0xCC},
						FHdr: &ttnpb.FHDR{
							DevAddr: [4]byte{0x11, 0x22, 0x33, 0x44},
							FCtrl: &ttnpb.FCtrl{
								Ack:    true,
								ClassB: true,
							},
							FCnt:  25,
							FOpts: []byte{0xFD},
						},
					}},
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "eui-0101010101010101",
							Eui:       &types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
						},
						Time:        ttnpb.ProtoTimePtr(time.Unix(1548059982, 0)),
						Timestamp:   (uint32)(12666373963464220 & 0xFFFFFFFF),
						Rssi:        89,
						ChannelRssi: 89,
						Snr:         9.25,
					},
				},
				Settings: &ttnpb.TxSettings{
					Frequency:  868300000,
					Time:       ttnpb.ProtoTimePtr(time.Unix(1548059982, 0)),
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					CodingRate: "4/5",
					DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
		{
			Name: "TimeSyncRequest",
			InputBSUpstream: lbslns.TimeSyncRequest{
				TxTime: 123.456,
			},
			ExpectedBSDownstream: lbslns.TimeSyncResponse{
				TxTime: 123.456,
			},
		},
		{
			Name: "Downlink",
			InputNetworkDownstream: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
					DevEui:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "testapp",
					},
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.CLASS_A,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
						Rx1DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx1Frequency:    868100000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
				CorrelationIds: []string{"correlation1", "correlation2"},
			},

			InputDownlinkPath: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
						1553759666,
						1553759666000,
						time.Unix(0, 1553759666*1000),
						nil,
					),
				},
			},
			ExpectedBSDownstream: lbslns.DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				Diid:        1,
				RxDelay:     1,
				Rx1Freq:     868100000,
				Rx1DR:       5,
				XTime:       12666375505739186,
				Priority:    25,
				MuxTime:     1554300787.123456,
			},
		},
		{
			Name: "FollowUpTxAck",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: ttnpb.TxAcknowledgment{
				DownlinkMessage: &ttnpb.DownlinkMessage{
					RawPayload: []byte("Ymxhamthc25kJ3M=="),
					EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
						DeviceId: "testdevice",
						DevEui:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: "testapp",
						},
					},
					Settings: &ttnpb.DownlinkMessage_Scheduled{
						Scheduled: &ttnpb.TxSettings{
							// Will only test that `Scheduled` field is set, not individual values.
						},
					},
					CorrelationIds: []string{"correlation1", "correlation2"},
				},
				Result:         ttnpb.TxAcknowledgment_SUCCESS,
				CorrelationIds: []string{"correlation1", "correlation2"},
			},
		},
		{
			Name: "RepeatedTxAck",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: ttnpb.TxAcknowledgment{
				DownlinkMessage: &ttnpb.DownlinkMessage{
					RawPayload: []byte("Ymxhamthc25kJ3M=="),
					EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
						DeviceId: "testdevice",
						DevEui:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: "testapp",
						},
					},
					Settings: &ttnpb.DownlinkMessage_Scheduled{
						Scheduled: &ttnpb.TxSettings{
							// Will only test that `Scheduled` field is set, not individual values.
						},
					},
					CorrelationIds: []string{"correlation1", "correlation2"},
				},
				Result:         ttnpb.TxAcknowledgment_SUCCESS,
				CorrelationIds: []string{"correlation1", "correlation2"},
			},
		},
		{
			Name: "RandomTxAck",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  2,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: ttnpb.TxAcknowledgment{},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			if tc.InputBSUpstream != nil {
				switch v := tc.InputBSUpstream.(type) {
				case lbslns.TxConfirmation:
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case ack := <-gsConn.TxAck():
						expected := tc.ExpectedNetworkUpstream.(ttnpb.TxAcknowledgment)
						if expected.DownlinkMessage.GetScheduled() != nil {
							if !a.So(ack.DownlinkMessage.GetScheduled(), should.NotBeNil) {
								t.Fatalf("Invalid downlink message settings: %v", ack.DownlinkMessage.Settings)
							}
							ack.DownlinkMessage.Settings = expected.DownlinkMessage.Settings
						}
						if !a.So(*ack, should.Resemble, expected) {
							t.Fatalf("Invalid TxAck: %v", ack)
						}
					case <-time.After(timeout):
						if tc.ExpectedNetworkUpstream != nil {
							t.Fatalf("Read message timeout")
						}
					}

				case lbslns.UplinkDataFrame, lbslns.JoinRequest:
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case up := <-gsConn.Up():
						a.So(time.Since(*ttnpb.StdTime(up.Message.ReceivedAt)), should.BeLessThan, timeout)
						up.Message.ReceivedAt = nil
						var payload ttnpb.Message
						a.So(lorawan.UnmarshalMessage(up.Message.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Message.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.Message.RawPayload)
						}
						up.Message.RawPayload = nil
						up.Message.RxMetadata[0].UplinkToken = nil
						expectedUp := tc.ExpectedNetworkUpstream.(ttnpb.UplinkMessage)
						a.So(up.Message, should.Resemble, &expectedUp)
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}

				case lbslns.TimeSyncRequest:
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}

					var readErr error
					var wg sync.WaitGroup
					resCh := make(chan []byte, 1)
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, data, err := wsConn.ReadMessage()
						if err != nil {
							readErr = err
							return
						}
						resCh <- data
					}()
					select {
					case res := <-resCh:
						expected := tc.ExpectedBSDownstream.(lbslns.TimeSyncResponse)
						var msg lbslns.TimeSyncResponse
						if err := json.Unmarshal(res, &msg); err != nil {
							t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
						}

						now := time.Now().UTC()
						a.So(msg.TxTime, should.Equal, expected.TxTime)
						a.So(gpstime.Parse(time.Duration(msg.GPSTime)*time.Microsecond), should.HappenBetween, now.Add(-time.Second), now.Add(time.Second))
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}
					wg.Wait()
					if readErr != nil {
						t.Fatalf("Failed to read message: %v", readErr)
					}
				}
			}

			if tc.InputNetworkDownstream != nil {
				if _, _, _, err := gsConn.ScheduleDown(tc.InputDownlinkPath, tc.InputNetworkDownstream); err != nil {
					t.Fatalf("Failed to send downlink: %v", err)
				}

				var readErr error
				var wg sync.WaitGroup
				resCh := make(chan []byte, 1)
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						readErr = err
						return
					}
					resCh <- data
				}()
				select {
				case res := <-resCh:
					switch tc.ExpectedBSDownstream.(type) {
					case lbslns.DownlinkMessage:
						var msg lbslns.DownlinkMessage
						if err := json.Unmarshal(res, &msg); err != nil {
							t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
						}
						msg.MuxTime = tc.ExpectedBSDownstream.(lbslns.DownlinkMessage).MuxTime
						if !a.So(msg, should.Resemble, tc.ExpectedBSDownstream.(lbslns.DownlinkMessage)) {
							t.Fatalf("Incorrect Downlink received: %s", string(res))
						}
					}
				case <-time.After(timeout):
					t.Fatalf("Read message timeout")
				}
				wg.Wait()
				if readErr != nil {
					t.Fatalf("Failed to read message: %v", readErr)
				}
			}
		})
	}
}

type testTime struct {
	Mux, Rx *time.Time
}

func (t testTime) getRefTime(drift time.Duration) float64 {
	time.Sleep(1 << 3 * test.Delay)
	now := time.Now()
	offset := now.Sub(*t.Rx)
	refTime := t.Mux.Add(offset)
	refTime = refTime.Add(-drift)
	return float64(refTime.UnixNano()) / float64(time.Second)
}

func TestRTT(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayToken)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	gs := mock.NewServer(c)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go func() error {
		return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			web.ServeHTTP(w, r)
		}))
	}()
	servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

	wsConn, _, err := websocket.DefaultDialer.Dial(servAddr+testTrafficEndPoint, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer wsConn.Close()

	var gsConn *io.Connection
	select {
	case gsConn = <-gs.Connections():
	case <-time.After(timeout):
		t.Fatal("Connection timeout")
	}

	testTime := testTime{}

	getTimeFromFloat64 := func(timeInFloat float64) *time.Time {
		sec, nsec := math.Modf(timeInFloat)
		retTime := time.Unix(int64(sec), int64(nsec*1e9))
		return &retTime
	}

	for _, tc := range []struct {
		Name                   string
		InputBSUpstream        interface{}
		InputNetworkDownstream *ttnpb.DownlinkMessage
		InputDownlinkPath      *ttnpb.DownlinkPath
		GatewayClockDrift      time.Duration
		ExpectedRTTStatsCount  int
	}{
		{
			Name: "JoinRequest",
			InputBSUpstream: lbslns.JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
				DevNonce: 18000,
				MIC:      12345678,
				RadioMetaData: lbslns.RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: lbslns.UpInfo{
						RxTime: 1548059982,
						XTime:  12666373963464220,
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ExpectedRTTStatsCount: 1,
		},
		{
			Name: "Downlink",
			InputNetworkDownstream: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
					DevEui:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "testapp",
					},
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.CLASS_A,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
						Rx1DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx1Frequency:    868100000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
				CorrelationIds: []string{"correlation1", "correlation2"},
			},

			InputDownlinkPath: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
						1553759666,
						1553759666000,
						time.Unix(0, 1553759666*1000),
						nil,
					),
				},
			},
		},
		{
			Name: "FollowUpTxAck",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
		},
		{
			Name: "RepeatedTxAck",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
		},
		{
			Name: "UplinkFrame",
			InputBSUpstream: lbslns.UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      0x00,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "5fcc",
				MIC:        12345678,
				RadioMetaData: lbslns.RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: lbslns.UpInfo{
						RxTime: 1548059982,
						XTime:  12666373963464220,
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ExpectedRTTStatsCount: 1,
		},
		{
			Name: "TxAckWithSmallClockDrift",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedRTTStatsCount: 1,
		},
		{
			Name: "TxAckWithClockDriftAboveThreshold",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedRTTStatsCount: 1,
			GatewayClockDrift:     (1 << 5 * test.Delay),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			if tc.InputBSUpstream != nil {
				switch v := tc.InputBSUpstream.(type) {
				case lbslns.TxConfirmation:
					// TxAck does not contain a RefTime.
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case ack := <-gsConn.TxAck():
						if ack.Result != ttnpb.TxAcknowledgment_SUCCESS {
							t.Fatalf("Tx acknowledgment failed")
						}
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}

				case lbslns.UplinkDataFrame:
					if testTime.Mux != nil {
						v.RefTime = testTime.getRefTime(tc.GatewayClockDrift)
					}
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case up := <-gsConn.Up():
						var payload ttnpb.Message
						a.So(lorawan.UnmarshalMessage(up.Message.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Message.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.Message.RawPayload)
						}
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}

				case lbslns.JoinRequest:
					if testTime.Mux != nil {
						v.RefTime = testTime.getRefTime(tc.GatewayClockDrift)
					}
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case up := <-gsConn.Up():
						var payload ttnpb.Message
						a.So(lorawan.UnmarshalMessage(up.Message.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Message.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.Message.RawPayload)
						}
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}
				}

				if testTime.Mux != nil {
					// Wait for stats to get updated
					time.Sleep(1 << 2 * test.Delay)

					// Atleast one downlink is needed for the first muxtime.
					min, max, median, _, count := gsConn.RTTStats(90, time.Now())
					if !a.So(count, should.Equal, tc.ExpectedRTTStatsCount) {
						t.Fatalf("Incorrect Stats entries recorded: %d", count)
					}
					if count > 0 {
						if !a.So(min, should.BeGreaterThan, 0) {
							t.Fatalf("Incorrect min: %s", min)
						}
						if tc.ExpectedRTTStatsCount > 1 {
							if !a.So(max, should.BeGreaterThan, min) {
								t.Fatalf("Incorrect max: %s", max)
							}
							if !a.So(median, should.BeBetween, min, max) {
								t.Fatalf("Incorrect median: %s", median)
							}
						}
					}
				}
			}

			if tc.InputNetworkDownstream != nil {
				if _, _, _, err := gsConn.ScheduleDown(tc.InputDownlinkPath, tc.InputNetworkDownstream); err != nil {
					t.Fatalf("Failed to send downlink: %v", err)
				}

				var readErr error
				var wg sync.WaitGroup
				resCh := make(chan []byte)
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						readErr = err
						return
					}
					resCh <- data
				}()
				select {
				case res := <-resCh:
					var msg lbslns.DownlinkMessage
					if err := json.Unmarshal(res, &msg); err != nil {
						t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
					}
					testTime.Mux = getTimeFromFloat64(msg.MuxTime)
					// Simulate downstream delay
					time.Sleep(1 << 2 * test.Delay)
					now := time.Now()
					testTime.Rx = &now
				case <-time.After(timeout):
					t.Fatalf("Read message timeout")
				}
				wg.Wait()
				if readErr != nil {
					t.Fatalf("Failed to read message: %v", err)
				}
			}
		})
	}
}

func TestPingPong(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayToken)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	gs := mock.NewServer(c)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go func() error {
		return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			web.ServeHTTP(w, r)
		}))
	}()
	servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())

	conn, _, err := websocket.DefaultDialer.Dial(servAddr+testTrafficEndPoint, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	pingCh := make(chan []byte)
	pongCh := make(chan []byte)

	// Read server ping
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			//  The ping/pong handlers are called only after ws.ReadMessage() receives a ping/pong message. The data read here is irrelevant.
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	conn.SetPingHandler(func(data string) error {
		pingCh <- []byte{}
		return nil
	})

	conn.SetPongHandler(func(data string) error {
		pongCh <- []byte{}
		return nil
	})

	select {
	case <-pingCh:
		t.Log("Received server ping")
		break
	case <-time.After(timeout):
		t.Fatalf("Server ping timeout")
	}

	// Client Ping, Server Pong
	if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		t.Fatalf("Failed to ping server: %v", err)
	}
	select {
	case <-pongCh:
		t.Log("Received server pong")
		break
	case <-time.After(timeout):
		t.Fatalf("Server pong timeout")
	}
	conn.Close() // The test below start a new connection per test. So this can be closed.

	// Test disconnection via ping pong
	for _, tc := range []struct {
		Name         string
		DisablePongs bool
		NoOfPongs    int
	}{
		{
			Name: "Regular ping-pong",
		},
		{
			Name:         "Disable pong",
			DisablePongs: true,
		},
		{
			Name:      "Stop responding after one pong",
			NoOfPongs: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), Config{
				WSPingInterval:       (1 << 5) * test.Delay,
				AllowUnauthenticated: true,
				UseTrafficTLSAddress: false,
				MissedPongThreshold:  2,
			})
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			lis, err := net.Listen("tcp", serverAddress)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			defer lis.Close()
			go func() error {
				return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					web.ServeHTTP(w, r)
				}))
			}()
			servAddr := fmt.Sprintf("ws://%s", lis.Addr().String())
			conn, _, err := websocket.DefaultDialer.Dial(servAddr+testTrafficEndPoint, nil)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Connection failed: %v", err)
			}
			defer conn.Close()

			handler := NewPingPongHandler(conn, tc.DisablePongs, tc.NoOfPongs)
			conn.SetPingHandler(handler.HandlePing)

			errCh := make(chan error)

			// Trigger server downstream.
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						//  The ping/pong handlers are called only after ws.ReadMessage() receives a ping/pong message. The data read here is irrelevant.
						_, _, err := conn.ReadMessage()
						if err != nil {
							errCh <- err
							return
						}
					}
				}
			}()

			// Wait for connection to setup
			time.After(1 << 8 * test.Delay)

			select {
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-time.After(timeout):
				if tc.DisablePongs || tc.NoOfPongs == 0 {
					// The timeout here is valid.
					// If tc.DisablePongs is true, client and server do ping pong forever.
					// If tc.NoOfPongs == 0, the server sends pings forever without checking pong.
					break
				}
				t.Fatal("Test time out")
			case err := <-errCh:
				if !tc.DisablePongs && tc.NoOfPongs == 1 {
					if websocket.IsUnexpectedCloseError(err) {
						// This is the error for WebSocket disconnection.
						break
					}
				}
				t.Fatalf("Unexpected error :%v", err)
			case err := <-handler.ErrCh():
				t.Fatal(err)
				break
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	t.Run("Accept", func(t *testing.T) {
		maxRate := uint(3)
		conf := config.RateLimiting{
			Profiles: []config.RateLimitingProfile{{
				Name:         "accept connections",
				MaxPerMin:    maxRate,
				MaxBurst:     maxRate,
				Associations: []string{"gs:accept:ws"},
			}},
		}
		withServer(t, defaultConfig, conf, func(t *testing.T, _ *mock.IdentityServer, serverAddress string) {
			a := assertions.New(t)
			for i := uint(0); i < maxRate; i++ {
				conn, _, err := websocket.DefaultDialer.Dial(serverAddress+testTrafficEndPoint, nil)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Connection failed: %v", err)
				}
				conn.Close()
			}

			for i := 0; i < 3; i++ {
				_, resp, err := websocket.DefaultDialer.Dial(serverAddress+testTrafficEndPoint, nil)
				a.So(err, should.NotBeNil)
				if !a.So(errors.IsResourceExhausted(errors.FromHTTPStatusCode(resp.StatusCode)), should.BeTrue) {
					t.FailNow()
				}

				a.So(resp.Header.Get("x-rate-limit-limit"), should.NotBeEmpty)
				a.So(resp.Header.Get("x-rate-limit-available"), should.NotBeEmpty)
				a.So(resp.Header.Get("x-rate-limit-reset"), should.NotBeEmpty)
				a.So(resp.Header.Get("x-rate-limit-retry"), should.NotBeEmpty)
			}
		})
	})
}
