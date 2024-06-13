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

package semtechws_test

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

	"github.com/gorilla/websocket"
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws/id6"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws/lbslns"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	pfconfig "go.thethings.network/lorawan-stack/v3/pkg/pfconfig/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	serverAddress          = "127.0.0.1:0"
	registeredGatewayUID   = "eui-0101010101010101"
	registeredGatewayID    = &ttnpb.GatewayIdentifiers{GatewayId: registeredGatewayUID}
	registeredGatewayToken = "secrettoken"

	discoveryEndPoint      = "/router-info"
	connectionRootEndPoint = "/traffic/"

	testTrafficEndPoint = "/traffic/eui-0101010101010101"

	timeout             = (1 << 7) * test.Delay
	trafficTestWaitTime = (1 << 7) * test.Delay
	defaultConfig       = Config{
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

	is, isAddr, cancelIS := mockis.New(ctx)
	defer cancelIS()

	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayToken, testGtw, testRights...)
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
	gs := mock.NewServer(c, is)

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
		go http.Serve(lis, web) // nolint:errcheck,gosec
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
					return errors.Is(err, websocket.ErrBadHandshake)
				},
			},
			{
				Name:      "RegisteredGatewayAndNoKey",
				GatewayID: registeredGatewayID.GatewayId,
				ErrorAssertion: func(err error) bool {
					if ttc.AllowUnauthenticated && err == nil {
						return true
					}
					return errors.Is(err, websocket.ErrBadHandshake)
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
					return errors.Is(err, websocket.ErrBadHandshake)
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

	is, isAddr, cancelIS := mockis.New(ctx)
	defer cancelIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayToken, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go http.Serve(lis, web) // nolint:errcheck,gosec
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
		Query    any
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
					if errors.Is(err, websocket.ErrBadHandshake) {
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
		Query any
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
		Query       any
	}{
		{
			EndPointEUI: "1111111111111111",
			EUI:         types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			Query: lbslns.DiscoverQuery{
				EUI: id6.EUI{
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
					if errors.Is(err, websocket.ErrBadHandshake) {
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
					EUI: id6.EUI{Prefix: "router", EUI64: tc.EUI},
					Muxs: id6.EUI{
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

	is, _, cancelIS := mockis.New(ctx)
	defer cancelIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayToken, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)

	gs.RegisterGateway(ctx, registeredGatewayID, &ttnpb.Gateway{
		Ids:             registeredGatewayID,
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
	go http.Serve(lis, web) // nolint:errcheck,gosec
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
		VersionQuery          any
		ExpectedRouterConfig  any
		ExpectedStatusMessage *ttnpb.GatewayStatus
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
				Beacon: &pfconfig.BeaconingConfig{
					DR:     ttnpb.DataRateIndex_DATA_RATE_3,
					Layout: [3]int{2, 8, 17},
					Freqs:  []uint64{869525000},
				},
			},
			ExpectedStatusMessage: &ttnpb.GatewayStatus{
				Versions: map[string]string{
					"station":  "test-station",
					"firmware": "1.0.0",
					"package":  "test-package",
					"platform": "test-model - Firmware 1.0.0 - Protocol 2",
				},
				Advanced: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"model": {
							Kind: &structpb.Value_StringValue{StringValue: "test-model"},
						},
						"features": {
							Kind: &structpb.Value_StringValue{StringValue: "prod gps"},
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
				Beacon: &pfconfig.BeaconingConfig{
					DR:     ttnpb.DataRateIndex_DATA_RATE_3,
					Layout: [3]int{2, 8, 17},
					Freqs:  []uint64{869525000},
				},
			},
			ExpectedStatusMessage: &ttnpb.GatewayStatus{
				Versions: map[string]string{
					"station":  "test-station-rc1",
					"firmware": "1.0.0",
					"package":  "test-package",
					"platform": "test-model - Firmware 1.0.0 - Protocol 2",
				},
				Advanced: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"model": {
							Kind: &structpb.Value_StringValue{StringValue: "test-model"},
						},
						"features": {
							Kind: &structpb.Value_StringValue{StringValue: "rmtsh gps"},
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
				a.So(stat, should.Resemble, tc.ExpectedStatusMessage)
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

	is, isAddr, cancelIS := mockis.New(ctx)
	defer cancelIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayToken, testGtw, testRights...)
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
	gs := mock.NewServer(c, is)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go http.Serve(lis, web) // nolint:errcheck,gosec
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

	now := time.Now().UTC()
	clock := mockClock{}
	clock.Start(ctx, now)

	for _, tc := range []struct {
		Name                    string
		InputBSUpstream         any
		InputNetworkDownstream  *ttnpb.DownlinkMessage
		ExpectedBSDownstream    any
		ExpectedNetworkUpstream proto.Message
	}{
		{
			Name: "JoinRequest",
			InputBSUpstream: lbslns.JoinRequest{
				MHdr:     0,
				DevEUI:   id6.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  id6.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
				DevNonce: 18000,
				MIC:      12345678,
				RadioMetaData: lbslns.RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: lbslns.UpInfo{
						RSSI: 89,
						SNR:  9.25,
					},
				},
			},
			ExpectedNetworkUpstream: &ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					Mic:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEui:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}.Bytes(),
						DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
						DevNonce: []byte{0x46, 0x50},
					}},
				},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIds: &ttnpb.GatewayIdentifiers{
						GatewayId: "eui-0101010101010101",
						Eui:       types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}.Bytes(),
					},
					Rssi:        89,
					ChannelRssi: 89,
					Snr:         9.25,
				}},
				Settings: &ttnpb.TxSettings{
					Frequency: 868300000,
					DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
						CodingRate:      band.Cr4_5,
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
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ExpectedNetworkUpstream: &ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					Mic:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
						FPort:      0,
						FrmPayload: []byte{0x5F, 0xCC},
						FHdr: &ttnpb.FHDR{
							DevAddr: []byte{0x11, 0x22, 0x33, 0x44},
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
							Eui:       types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}.Bytes(),
						},
						Time:        timestamppb.New(time.Unix(1548059982, 0)),
						Timestamp:   (uint32)(12666373963464220 & 0xFFFFFFFF),
						Rssi:        89,
						ChannelRssi: 89,
						Snr:         9.25,
					},
				},
				Settings: &ttnpb.TxSettings{
					Frequency: 868300000,
					Time:      timestamppb.New(time.Unix(1548059982, 0)),
					Timestamp: (uint32)(12666373963464220 & 0xFFFFFFFF),
					DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
						CodingRate:      band.Cr4_5,
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
					DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "testapp",
					},
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_A,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
						Rx1DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
								},
							},
						},
						Rx1Frequency:    868100000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
				CorrelationIds: []string{"correlation1", "correlation2"},
			},
			ExpectedBSDownstream: lbslns.DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				Diid:        1,
				Priority:    25,
				MuxTime:     1554300787.123456,
				TimestampDownlinkMessage: &lbslns.TimestampDownlinkMessage{
					RxDelay: 1,
					Rx1Freq: 868100000,
					Rx1DR:   5,
				},
			},
		},
		{
			Name: "FollowUpTxAck",
			InputBSUpstream: lbslns.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: &ttnpb.TxAcknowledgment{
				DownlinkMessage: &ttnpb.DownlinkMessage{
					RawPayload: []byte("Ymxhamthc25kJ3M=="),
					EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
						DeviceId: "testdevice",
						DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
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
			ExpectedNetworkUpstream: &ttnpb.TxAcknowledgment{
				DownlinkMessage: &ttnpb.DownlinkMessage{
					RawPayload: []byte("Ymxhamthc25kJ3M=="),
					EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
						DeviceId: "testdevice",
						DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
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
			ExpectedNetworkUpstream: &ttnpb.TxAcknowledgment{},
		},
		{
			Name: "AbsoluteTimeDownlink",
			InputNetworkDownstream: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
					DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "testapp",
					},
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_C,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
						Rx1DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
								},
							},
						},
						Rx1Frequency:    868100000,
						FrequencyPlanId: test.EUFrequencyPlanID,
						AbsoluteTime:    timestamppb.New(now.Add(30 * time.Second)),
					},
				},
				CorrelationIds: []string{"correlation1", "correlation2"},
			},
			ExpectedBSDownstream: lbslns.DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 1,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				Diid:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
				AbsoluteTimeDownlinkMessage: &lbslns.AbsoluteTimeDownlinkMessage{
					Freq:    868100000,
					DR:      5,
					GPSTime: TimeToGPSTime(now.Add(30 * time.Second)),
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			if tc.InputBSUpstream != nil {
				timestamp := clock.GetTimestamp()
				upXTime := clock.GetXTimeForTimestamp(timestamp)
				switch v := tc.InputBSUpstream.(type) {
				case lbslns.TxConfirmation:
					v.XTime = upXTime
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case ack := <-gsConn.TxAck():
						expected := tc.ExpectedNetworkUpstream.(*ttnpb.TxAcknowledgment)
						if expected.DownlinkMessage.GetScheduled() != nil {
							if !a.So(ack.DownlinkMessage.GetScheduled(), should.NotBeNil) {
								t.Fatalf("Invalid downlink message settings: %v", ack.DownlinkMessage.Settings)
							}
							ack.DownlinkMessage.Settings = expected.DownlinkMessage.Settings
						}
						if !a.So(ack, should.Resemble, expected) {
							t.Fatalf("Invalid TxAck: %v", ack)
						}
					case <-time.After(timeout):
						if tc.ExpectedNetworkUpstream != nil {
							t.Fatalf("Read message timeout")
						}
					}

				case lbslns.UplinkDataFrame:
					now := time.Unix(time.Now().UTC().Unix(), 0)
					v.UpInfo.XTime = upXTime
					v.UpInfo.RxTime = float64(now.Unix())
					v.UpInfo.GPSTime = TimeToGPSTime(now)
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
						var payload ttnpb.Message
						a.So(lorawan.UnmarshalMessage(up.Message.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Message.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.Message.RawPayload)
						}

						expectedUp := ttnpb.Clone(tc.ExpectedNetworkUpstream).(*ttnpb.UplinkMessage)
						expectedUp.ReceivedAt = up.Message.ReceivedAt
						expectedUp.RawPayload = up.Message.RawPayload

						// Set the correct xtime and timestamps for the assertion.
						for i, md := range expectedUp.RxMetadata {
							md.UplinkToken = up.Message.RxMetadata[i].UplinkToken
							md.Timestamp = timestamp
							md.Time = timestamppb.New(now)
							md.GpsTime = timestamppb.New(now)
							md.ReceivedAt = expectedUp.ReceivedAt
						}
						expectedUp.Settings.Timestamp = timestamp
						expectedUp.Settings.Time = ttnpb.ProtoTime(&now)
						a.So(up.Message, should.Resemble, expectedUp)
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}
				case lbslns.JoinRequest:
					now := time.Unix(time.Now().UTC().Unix(), 0)
					v.UpInfo.XTime = upXTime
					v.UpInfo.RxTime = float64(now.Unix())
					v.UpInfo.GPSTime = TimeToGPSTime(now)
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
						var payload ttnpb.Message
						a.So(lorawan.UnmarshalMessage(up.Message.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Message.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.Message.RawPayload)
						}

						expectedUp := ttnpb.Clone(tc.ExpectedNetworkUpstream).(*ttnpb.UplinkMessage)
						expectedUp.ReceivedAt = up.Message.ReceivedAt
						expectedUp.RawPayload = up.Message.RawPayload

						// Set the correct xtime and timestamps for the assertion.
						for i, md := range expectedUp.RxMetadata {
							md.UplinkToken = up.Message.RxMetadata[i].UplinkToken
							md.Timestamp = timestamp
							md.Time = timestamppb.New(now)
							md.GpsTime = timestamppb.New(now)
							md.ReceivedAt = expectedUp.ReceivedAt
						}
						expectedUp.Settings.Timestamp = timestamp
						expectedUp.Settings.Time = ttnpb.ProtoTime(&now)
						a.So(up.Message, should.Resemble, expectedUp)
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
						a.So(
							TimeFromGPSTime(msg.GPSTime),
							should.HappenBetween,
							now.Add(-time.Second),
							now.Add(time.Second),
						)
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
				var (
					downlinkPath *ttnpb.DownlinkPath
					down         = tc.InputNetworkDownstream
					now          = time.Unix(time.Now().UTC().Unix(), 0)
					timeStamp    = clock.GetTimestamp()
					dlClass      = down.GetRequest().Class
				)
				if dlClass == ttnpb.Class_CLASS_A {
					downlinkPath = &ttnpb.DownlinkPath{
						Path: &ttnpb.DownlinkPath_UplinkToken{
							UplinkToken: io.MustUplinkToken(
								&ttnpb.GatewayAntennaIdentifiers{GatewayIds: registeredGatewayID},
								timeStamp,
								0,
								now,
								nil,
							),
						},
					}
				} else {
					downlinkPath = &ttnpb.DownlinkPath{
						Path: &ttnpb.DownlinkPath_Fixed{
							Fixed: &ttnpb.GatewayAntennaIdentifiers{
								GatewayIds: registeredGatewayID,
							},
						},
					}
				}

				if _, _, _, err := gsConn.ScheduleDown(downlinkPath, down); err != nil {
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
						expected := tc.ExpectedBSDownstream.(lbslns.DownlinkMessage)
						msg.MuxTime = expected.MuxTime
						if dlClass == ttnpb.Class_CLASS_A {
							expected.XTime = clock.GetXTimeForTimestamp(timeStamp)
						}
						if !a.So(msg, should.Resemble, expected) {
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
			time.Sleep(trafficTestWaitTime)
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

	is, isAddr, cancelIS := mockis.New(ctx)
	defer cancelIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayToken, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)

	web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), defaultConfig)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()
	go http.Serve(lis, web) // nolint:errcheck,gosec
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
		InputBSUpstream        any
		InputNetworkDownstream *ttnpb.DownlinkMessage
		InputDownlinkPath      *ttnpb.DownlinkPath
		GatewayClockDrift      time.Duration
		ExpectedRTTStatsCount  int
	}{
		{
			Name: "JoinRequest",
			InputBSUpstream: lbslns.JoinRequest{
				MHdr:     0,
				DevEUI:   id6.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  id6.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
					DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "testapp",
					},
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_A,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
						Rx1DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
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
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: registeredGatewayID},
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

	is, isAddr, cancelIS := mockis.New(ctx)
	defer cancelIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayToken, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)

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
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			web, err := New(ctx, gs, lbslns.NewFormatter(maxValidRoundTripDelay), Config{
				WSPingInterval:       (1 << 4) * test.Delay,
				MissedPongThreshold:  2,
				AllowUnauthenticated: true,
				UseTrafficTLSAddress: false,
			})
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			lis, err := net.Listen("tcp", serverAddress)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			defer lis.Close()
			go http.Serve(lis, web) // nolint:errcheck,gosec
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
						// The ping/pong handlers are called only after ReadMessage() receives a ping/pong message.
						// The data read here is irrelevant.
						_, _, err := conn.ReadMessage()
						if err != nil {
							errCh <- err
							return
						}
					}
				}
			}()

			// Wait for connection to setup
			time.After(timeout)

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
				Associations: []string{"gs:accept:semtechws/lbslns"},
			}},
		}
		withServer(t, defaultConfig, conf, func(t *testing.T, _ *mockis.MockDefinition, serverAddress string) {
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
