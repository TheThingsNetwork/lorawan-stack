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
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
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

	testTrafficEndPoint = "ws://localhost:8100/api/v3/gs/io/basicstation/traffic/eui-0101010101010101"

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

	conn, _, err := websocket.DefaultDialer.Dial(testTrafficEndPoint, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

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

func TestUpstream(t *testing.T) {
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

	wsConn, _, err := websocket.DefaultDialer.Dial(testTrafficEndPoint, nil)
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
		Name          string
		TimeStamp     int64
		Message       interface{}
		UplinkMessage ttnpb.UplinkMessage
	}{
		{
			"JoinRequest",
			12666373963464220,
			messages.JoinRequest{
				MHdr:     0,
				DevEUI:   messages.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  messages.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
				DevNonce: 18000,
				MIC:      12345678,
				RadioMetaData: messages.RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: messages.UpInfo{
						RxTime: 1548059982,
						XTime:  12666373963464220,
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
						DevEUI:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						DevNonce: [2]byte{0x50, 0x46},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{&ttnpb.RxMetadata{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-0101010101010101"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
				}},
				Settings: ttnpb.TxSettings{
					CodingRate:    "4/5",
					Frequency:     868300000,
					DataRateIndex: 1,
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			req, err := json.Marshal(tc.Message)
			if err != nil {
				panic(err)
			}
			if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}
			select {
			case up := <-gsConn.Up():
				a.So(time.Since(up.ReceivedAt), should.BeLessThan, timeout)
				up.ReceivedAt = time.Time{}
				var payload ttnpb.Message
				a.So(lorawan.UnmarshalMessage(up.RawPayload, &payload), should.BeNil)
				if !a.So(&payload, should.Resemble, up.Payload) {
					t.Fatalf("Invalid RawPayload: %v", up.RawPayload)
				}
				up.RawPayload = nil
				up.RxMetadata[0].UplinkToken = nil //TODO: Figure out how to test this
				a.So(up, should.Resemble, &tc.UplinkMessage)
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
		})
	}

}
