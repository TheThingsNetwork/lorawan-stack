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

package basicstation_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstation"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstation/messages"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredGatewayUID = "0101010101010101"
	registeredGatewayID  = ttnpb.GatewayIdentifiers{GatewayID: "eui-0101010101010101"}
	registeredGateway    = ttnpb.Gateway{GatewayIdentifiers: registeredGatewayID, FrequencyPlanID: "EU_863_870"}
	registeredGatewayKey = "test-key"

	discoveryEndPoint      = "ws://localhost:8100/api/v3/gs/io/basicstation/discover"
	connectionRootEndPoint = "ws://localhost:8100/api/v3/gs/io/basicstation/traffic/"

	timeout = 10 * test.Delay
)

func TestAuthentication(t *testing.T) {
	// TODO: Test authentication. We're gonna provision authentication tokens, which may be API keys.
	// https://github.com/TheThingsIndustries/lorawan-stack/issues/1413
}

func TestDiscover(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx = newContextWithRightsFetcher(ctx)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: ":8100",
			},
		},
	})
	gs := mock.NewServer()
	srv := New(ctx, gs)
	c.RegisterWeb(srv)
	if err := c.Start(); err != nil {
		panic(err)
	}
	defer c.Close()

	// Test Endpoints
	for i, tc := range []struct {
		URL                string
		ExpectedError      error
		ExpectedStatusCode int
	}{
		{
			"ws://localhost:8100/router-info", // This would ideally be proxied to "ws://server-address:port/api/v3/gs/io/basicstation/discover" but is invalid for unit tests.
			websocket.ErrBadHandshake,
			http.StatusNotFound,
		},
		{
			discoveryEndPoint + "/router-58a0:cbff:fe80:f8",
			websocket.ErrBadHandshake,
			http.StatusNotFound,
		},
		{
			discoveryEndPoint + "/eui-0101010101010101",
			websocket.ErrBadHandshake,
			http.StatusNotFound,
		},
	} {
		t.Run(fmt.Sprintf("InvalidDiscoveryEndPoint/%d", i), func(t *testing.T) {
			_, res, err := websocket.DefaultDialer.Dial(tc.URL, nil)
			if res.StatusCode != tc.ExpectedStatusCode {
				t.Fatalf("Unexpected response received: %v", res.Status)
			}
			if !a.So(err, should.Equal, tc.ExpectedError) {
				t.Fatalf("Connection failed: %v", err)
			}
		})
	}

	// Test Queries
	for i, tc := range []struct {
		Query interface{}
	}{
		{
			messages.DiscoverQuery{},
		},
		{
			struct{}{},
		},
		{
			struct {
				EUI string `json:"route"`
			}{EUI: `"01-02-03-04-05-06-07-08"`},
		},
		{
			struct {
				EUI string `json:"router"`
			}{EUI: `"01-02-03-04-05-06-07-08-09"`},
		},
		{
			struct {
				EUI string `json:"router"`
			}{EUI: `"01:02:03:04:05:06:07:08:09"`},
		},
		{
			struct {
				EUI string `json:"router"`
			}{EUI: `"01:02:03:04:05:06:07-08"`},
		},
	} {
		t.Run(fmt.Sprintf("InvalidQuery/%d", i), func(t *testing.T) {
			conn, _, err := websocket.DefaultDialer.Dial(discoveryEndPoint, nil)
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

			resCh := make(chan []byte)
			go func() {
				_, data, err := conn.ReadMessage()
				if err != nil {
					t.Fatalf("Failed to read message: %v", err)
				}
				resCh <- data
			}()
			select {
			case res := <-resCh:
				var response messages.DiscoverResponse
				if err := json.Unmarshal(res, &response); err != nil {
					t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
				}
				a.So(response, should.Resemble, messages.DiscoverResponse{
					Error: "Invalid request",
				})
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
		})
	}

	for i, tc := range []struct {
		EndPointEUI string
		EUI         types.EUI64
		Query       interface{}
	}{
		{
			"1111111111111111",
			types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			messages.DiscoverQuery{EUI: messages.EUI{Prefix: "router", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}}},
		},
	} {
		t.Run(fmt.Sprintf("ValidQuery/%d", i), func(t *testing.T) {
			conn, _, err := websocket.DefaultDialer.Dial(discoveryEndPoint, nil)
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

			resCh := make(chan []byte)
			go func() {
				_, data, err := conn.ReadMessage()
				if err != nil {
					t.Fatalf("Failed to read message: %v", err)
				}
				resCh <- data
			}()
			select {
			case res := <-resCh:
				var response messages.DiscoverResponse
				if err := json.Unmarshal(res, &response); err != nil {
					t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
				}
				a.So(response, should.Resemble, messages.DiscoverResponse{
					EUI: messages.EUI{Prefix: "router", EUI64: tc.EUI},
					Muxs: messages.EUI{
						Prefix: "muxs",
					},
					URI: connectionRootEndPoint + "eui-" + tc.EndPointEUI,
				})
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
		})
	}
}

func TestVersion(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx = newContextWithRightsFetcher(ctx)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: ":8100",
			},
		},
	})
	gs := mock.NewServer()
	srv := New(ctx, gs)
	c.RegisterWeb(srv)
	if err := c.Start(); err != nil {
		panic(err)
	}
	defer c.Close()
	gs.RegisterGateway(ctx, registeredGatewayID, &registeredGateway)

	testTrafficEndPoint := "ws://localhost:8100/api/v3/gs/io/basicstation/traffic/eui-0101010101010101"

	for i, tc := range []struct {
		Name                 string
		VersionQuery         interface{}
		ExpectedRouterConfig interface{}
	}{
		{
			"VersionProd",
			messages.Version{
				Station:  "test-station",
				Firmware: "1.0.0",
				Package:  "test-package",
				Model:    "test-model",
				Protocol: 2,
				Features: []string{"prod", "gps"},
			},
			messages.RouterConfig{
				Region:         "EU863",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{863000000, 870000000},
				DataRates: [16][3]int{
					[3]int{12, 125, 0},
					[3]int{11, 125, 0},
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{7, 250, 0},
					[3]int{0, 0, 0},
				},
			},
		},
		{
			"VersionDebug",
			messages.Version{
				Station:  "test-station",
				Firmware: "1.0.0",
				Package:  "test-package",
				Model:    "test-model",
				Protocol: 2,
				Features: []string{"rmtsh", "gps"},
			},
			messages.RouterConfig{
				Region:         "EU863",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{863000000, 870000000},
				DataRates: [16][3]int{
					[3]int{12, 125, 0},
					[3]int{11, 125, 0},
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{7, 250, 0},
					[3]int{0, 0, 0},
				},
				NoCCA:       true,
				NoDwellTime: true,
				NoDutyCycle: true,
			},
		},
	} {
		t.Run(fmt.Sprintf("VersionMessage/%d", i), func(t *testing.T) {
			conn, _, err := websocket.DefaultDialer.Dial(testTrafficEndPoint, nil)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Connection failed: %v", err)
			}
			defer conn.Close()
			req, err := json.Marshal(messages.DiscoverQuery{EUI: messages.EUI{Prefix: "router", EUI64: types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}}})
			if err != nil {
				panic(err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, req); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}

			reqVersion, err := json.Marshal(tc.VersionQuery)
			if err != nil {
				panic(err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, reqVersion); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}

			resCh := make(chan []byte)
			go func() {
				_, data, err := conn.ReadMessage()
				if err != nil {
					t.Fatalf("Failed to read message: %v", err)
				}
				resCh <- data
			}()
			select {
			case res := <-resCh:
				var response messages.RouterConfig
				if err := json.Unmarshal(res, &response); err != nil {
					t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
				}
				a.So(response, should.Resemble, tc.ExpectedRouterConfig)
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
		})
	}

	//TODO: test hardcoded URL, connect without discover.
}

func TestTraffic(t *testing.T) {
	// TODO: Test traffic, see gRPC frontend.
}
