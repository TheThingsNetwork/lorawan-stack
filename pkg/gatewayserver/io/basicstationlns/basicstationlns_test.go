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

package basicstationlns_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstationlns"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstationlns/messages"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredGatewayUID   = "eui-0101010101010101"
	registeredGatewayID    = ttnpb.GatewayIdentifiers{GatewayID: registeredGatewayUID}
	registeredGateway      = ttnpb.Gateway{GatewayIdentifiers: registeredGatewayID, FrequencyPlanID: "EU_863_870"}
	registeredGatewayToken = "secrettoken"

	discoveryEndPoint      = "ws://localhost:8100/api/v3/gs/io/basicstation/discover"
	connectionRootEndPoint = "ws://localhost:8100/api/v3/gs/io/basicstation/traffic/"

	testTrafficEndPoint = "ws://localhost:8100/api/v3/gs/io/basicstation/traffic/eui-0101010101010101"

	timeout = 10 * test.Delay
)

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestClientTokenAuth(t *testing.T) {
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
	test.Must(nil, c.Start())
	defer c.Close()

	gs.RegisterGateway(ctx, registeredGatewayID, &registeredGateway)

	for _, tc := range []struct {
		Name           string
		GatewayID      string
		AuthToken      string
		ErrorAssertion func(err error) bool
	}{
		{
			Name:           "RegisteredGatewayAndValidKey",
			GatewayID:      registeredGatewayID.GatewayID,
			AuthToken:      registeredGatewayToken,
			ErrorAssertion: nil,
		},
		{
			Name:      "RegisteredGatewayAndInValidKey",
			GatewayID: registeredGatewayID.GatewayID,
			AuthToken: "invalidToken",
			ErrorAssertion: func(err error) bool {
				return err == websocket.ErrBadHandshake
			},
		},
		{
			Name:           "RegisteredGatewayAndNoKey",
			GatewayID:      registeredGatewayID.GatewayID,
			ErrorAssertion: nil,
		},
		{
			Name:      "UnregisteredGateway",
			GatewayID: "eui-1122334455667788",
			AuthToken: registeredGatewayToken,
			ErrorAssertion: func(err error) bool {
				return err == websocket.ErrBadHandshake
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			a := assertions.New(t)
			h := http.Header{}
			h.Set("Authorization", tc.AuthToken)
			conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8100/api/v3/gs/io/basicstation/traffic/"+tc.GatewayID, h)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			}
			if conn != nil {
				conn.Close()
			}
		})
	}

}

func TestClientSideTLS(t *testing.T) {
	// TODO: https://github.com/TheThingsNetwork/lorawan-stack/issues/558
}
func TestDiscover(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

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
	test.Must(nil, c.Start())
	defer c.Close()

	// Invalid Endpoints
	for i, tc := range []struct {
		URL string
	}{
		{
			URL: "ws://localhost:8100/router-info",
		},
		{
			URL: discoveryEndPoint + "/router-58a0:cbff:fe80:f8",
		},
		{
			URL: discoveryEndPoint + "/eui-0101010101010101",
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
	for i, tc := range []struct {
		Query    interface{}
		Response messages.DiscoverResponse
	}{
		{
			Query:    messages.DiscoverQuery{},
			Response: messages.DiscoverResponse{Error: "Invalid request"},
		},
		{
			Query:    struct{}{},
			Response: messages.DiscoverResponse{Error: "Invalid request"},
		},
		{
			Query: struct {
				EUI string `json:"route"`
			}{EUI: `"01-02-03-04-05-06-07-08"`},
			Response: messages.DiscoverResponse{Error: "Invalid request"},
		},
	} {
		t.Run(fmt.Sprintf("InvalidQuery/%d", i), func(t *testing.T) {
			a := assertions.New(t)
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
					close(resCh)
					if err == websocket.ErrBadHandshake {
						return
					}
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
				a.So(response, should.Resemble, tc.Response)
			case <-time.After(timeout):
				t.Fatal("Read message timeout")
			}
		})
	}

	for i, tc := range []struct {
		Query interface{}
	}{
		{
			Query: struct {
				EUI string `json:"router"`
			}{EUI: `"01-02-03-04-05-06-07-08-09"`},
		},
		{
			Query: struct {
				EUI string `json:"router"`
			}{EUI: `"01:02:03:04:05:06:07:08:09"`},
		},
		{
			Query: struct {
				EUI string `json:"router"`
			}{EUI: `"01:02:03:04:05:06:07-08"`},
		},
	} {
		t.Run(fmt.Sprintf("InvalidQuery/%d", i), func(t *testing.T) {
			a := assertions.New(t)
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

			go func() {
				_, _, err := conn.ReadMessage()
				if err == nil {
					t.Fatalf("Expected connection closure with error but received none")
				}
			}()
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
			Query: messages.DiscoverQuery{
				EUI: basicstation.EUI{
					Prefix: "router",
					EUI64:  types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("ValidQuery/%d", i), func(t *testing.T) {
			a := assertions.New(t)
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
					close(resCh)
					if err == websocket.ErrBadHandshake {
						return
					} else {
						t.Fatalf("Failed to read message: %v", err)
					}
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
					EUI: basicstation.EUI{Prefix: "router", EUI64: tc.EUI},
					Muxs: basicstation.EUI{
						Prefix: "muxs",
					},
					URI: connectionRootEndPoint + "eui-" + tc.EndPointEUI,
				})
			case <-time.After(timeout):
				t.Fatalf("Read message timeout")
			}
			conn.Close()
		})
	}
}

func TestVersion(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

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
	test.Must(nil, c.Start())
	defer c.Close()
	gs.RegisterGateway(ctx, registeredGatewayID, &registeredGateway)

	conn, _, err := websocket.DefaultDialer.Dial(testTrafficEndPoint, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	for _, tc := range []struct {
		Name                 string
		VersionQuery         interface{}
		ExpectedRouterConfig interface{}
	}{
		// These tests will fail without unmarshalling of Sx1301_conf https://github.com/TheThingsNetwork/lorawan-stack/issues/808
		// {
		// 	Name: "VersionProd",
		// 	VersionQuery: messages.Version{
		// 		Station:  "test-station",
		// 		Firmware: "1.0.0",
		// 		Package:  "test-package",
		// 		Model:    "test-model",
		// 		Protocol: 2,
		// 		Features: "prod gps",
		// 	},
		// 	ExpectedRouterConfig: messages.RouterConfig{
		// 		Region:         "EU863",
		// 		HardwareSpec:   "sx1301/1",
		// 		FrequencyRange: []int{863000000, 870000000},
		// 		DataRates: [16][3]int{
		// 			{12, 125, 0},
		// 			{11, 125, 0},
		// 			{10, 125, 0},
		// 			{9, 125, 0},
		// 			{8, 125, 0},
		// 			{7, 125, 0},
		// 			{7, 250, 0},
		// 			{0, 0, 0},
		// 		},
		// 	},
		// },
		// {
		// 	Name: "VersionDebug",
		// 	VersionQuery: messages.Version{
		// 		Station:  "test-station",
		// 		Firmware: "1.0.0",
		// 		Package:  "test-package",
		// 		Model:    "test-model",
		// 		Protocol: 2,
		// 		Features: "rmtsh gps",
		// 	},
		// 	ExpectedRouterConfig: messages.RouterConfig{
		// 		Region:         "EU863",
		// 		HardwareSpec:   "sx1301/1",
		// 		FrequencyRange: []int{863000000, 870000000},
		// 		DataRates: [16][3]int{
		// 			{12, 125, 0},
		// 			{11, 125, 0},
		// 			{10, 125, 0},
		// 			{9, 125, 0},
		// 			{8, 125, 0},
		// 			{7, 125, 0},
		// 			{7, 250, 0},
		// 			{0, 0, 0},
		// 		},
		// 		NoCCA:       true,
		// 		NoDwellTime: true,
		// 		NoDutyCycle: true,
		// 	},
		// },
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
}
func TestTraffic(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

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
	test.Must(nil, c.Start())
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
		Name                    string
		InputBSUpstream         interface{}
		InputNetworkDownstream  *ttnpb.DownlinkMessage
		InputDownlinkPath       *ttnpb.DownlinkPath
		ExpectedBSDownstream    interface{}
		ExpectedNetworkUpstream interface{}
	}{
		{
			Name: "JoinRequest",
			InputBSUpstream: messages.JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			ExpectedNetworkUpstream: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
						DevEUI:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						DevNonce: [2]byte{0x46, 0x50},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-0101010101010101",
						EUI: &types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
					},
					Time:      &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp: (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:      89,
					SNR:       9.25,
				}},
				Settings: ttnpb.TxSettings{
					Frequency:  868300000,
					CodingRate: "4/5",
					Time:       &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
		{
			Name: "UplinkFrame",
			InputBSUpstream: messages.UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      0x00,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "5fcc",
				MIC:        12345678,
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
			ExpectedNetworkUpstream: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
						FPort:      0,
						FRMPayload: []byte{0x5F, 0xCC},
						FHDR: ttnpb.FHDR{
							DevAddr: [4]byte{0x11, 0x22, 0x33, 0x44},
							FCtrl: ttnpb.FCtrl{
								Ack:    true,
								ClassB: true,
							},
							FCnt:  25,
							FOpts: []byte{0xFD},
						},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-0101010101010101",
						EUI: &types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
					},
					Time:      &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp: (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:      89,
					SNR:       9.25},
				},
				Settings: ttnpb.TxSettings{
					Frequency:  868300000,
					Time:       &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					CodingRate: "4/5",
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
		{
			Name: "Downlink",
			InputNetworkDownstream: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIDs: &ttnpb.EndDeviceIdentifiers{
					DeviceID: "testdevice",
					DevEUI:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_A,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay:         ttnpb.RX_DELAY_1,
						Rx1DataRateIndex: 5,
						Rx1Frequency:     868100000,
					},
				},
				CorrelationIDs: []string{"correlation1", "correlation2"},
			},

			InputDownlinkPath: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID}, 1553759666),
				},
			},
			ExpectedBSDownstream: messages.DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-00",
				DeviceClass: 0,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				Diid:        1,
				RxDelay:     1,
				Rx1Freq:     868100000,
				Rx1DR:       5,
				XTime:       1553759666,
				Priority:    25,
				MuxTime:     1554300787.123456,
			},
		},
		{
			Name: "FollowUpTxAck",
			InputBSUpstream: messages.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: ttnpb.TxAcknowledgment{
				CorrelationIDs: []string{"correlation1", "correlation2"},
				Result:         ttnpb.TxAcknowledgment_SUCCESS,
			},
		},
		{
			Name: "RepeatedTxAck",
			InputBSUpstream: messages.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: ttnpb.TxAcknowledgment{
				CorrelationIDs: []string{"correlation1", "correlation2"},
				Result:         ttnpb.TxAcknowledgment_SUCCESS,
			},
		},
		{
			Name: "RandomTxAck",
			InputBSUpstream: messages.TxConfirmation{
				Diid:  2,
				XTime: 1548059982,
			},
			ExpectedNetworkUpstream: nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			if tc.InputBSUpstream != nil {
				switch v := tc.InputBSUpstream.(type) {
				case messages.TxConfirmation:
					req, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					if err := wsConn.WriteMessage(websocket.TextMessage, req); err != nil {
						t.Fatalf("Failed to write message: %v", err)
					}
					select {
					case ack := <-gsConn.TxAck():
						if !a.So(*ack, should.Resemble, tc.ExpectedNetworkUpstream) {
							t.Fatalf("Invalid TxAck: %v", ack)
						}
					case <-time.After(timeout):
						if tc.ExpectedNetworkUpstream != nil {
							t.Fatalf("Read message timeout")
						}
					}

				case messages.UplinkDataFrame, messages.JoinRequest:
					req, err := json.Marshal(v)
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
						up.RxMetadata[0].UplinkToken = nil
						expectedUp := tc.ExpectedNetworkUpstream.(ttnpb.UplinkMessage)
						a.So(up, should.Resemble, &expectedUp)
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}
				}
			}

			if tc.InputNetworkDownstream != nil {
				if _, err := gsConn.SendDown(tc.InputDownlinkPath, tc.InputNetworkDownstream); err != nil {
					t.Fatalf("Failed to send downlink: %v", err)
				}

				resCh := make(chan []byte)
				go func() {
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						t.Fatalf("Failed to read message: %v", err)
					}
					resCh <- data
				}()
				select {
				case res := <-resCh:
					switch tc.ExpectedBSDownstream.(type) {
					case messages.DownlinkMessage:
						var msg messages.DownlinkMessage
						if err := json.Unmarshal(res, &msg); err != nil {
							t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
						}
						msg.XTime = tc.ExpectedBSDownstream.(messages.DownlinkMessage).XTime
						msg.MuxTime = tc.ExpectedBSDownstream.(messages.DownlinkMessage).MuxTime
						if !a.So(msg, should.Resemble, tc.ExpectedBSDownstream.(messages.DownlinkMessage)) {
							t.Fatalf("Incorrect Downlink received: %s", string(res))
						}
					}
				case <-time.After(timeout):
					t.Fatalf("Read message timeout")
				}
			}
		})
	}

}

func TestRTT(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

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
	test.Must(nil, c.Start())
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

	var MuxTime, RxTime float64
	for _, tc := range []struct {
		Name                   string
		InputBSUpstream        interface{}
		InputNetworkDownstream *ttnpb.DownlinkMessage
		InputDownlinkPath      *ttnpb.DownlinkPath
		WaitTime               time.Duration
		ExpectedRTTStatsCount  int
	}{
		{
			Name: "JoinRequest",
			InputBSUpstream: messages.JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			ExpectedRTTStatsCount: 0,
		},
		{
			Name: "Downlink",
			InputNetworkDownstream: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIDs: &ttnpb.EndDeviceIdentifiers{
					DeviceID: "testdevice",
					DevEUI:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_A,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx1Delay:         ttnpb.RX_DELAY_1,
						Rx1DataRateIndex: 5,
						Rx1Frequency:     868100000,
					},
				},
				CorrelationIDs: []string{"correlation1", "correlation2"},
			},

			InputDownlinkPath: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID}, 1553759666),
				},
			},
		},
		{
			Name: "FollowUpTxAck",
			InputBSUpstream: messages.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedRTTStatsCount: 1,
			WaitTime:              1 << 4 * test.Delay,
		},
		{
			Name: "RepeatedTxAck",
			InputBSUpstream: messages.TxConfirmation{
				Diid:  1,
				XTime: 1548059982,
			},
			ExpectedRTTStatsCount: 2,
			WaitTime:              0,
		},
		{
			Name: "UplinkFrame",
			InputBSUpstream: messages.UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      0x00,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "5fcc",
				MIC:        12345678,
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
			ExpectedRTTStatsCount: 3,
			WaitTime:              1 << 3 * test.Delay,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			if tc.InputBSUpstream != nil {
				switch v := tc.InputBSUpstream.(type) {
				case messages.TxConfirmation:
					if MuxTime != 0 {
						time.Sleep(tc.WaitTime)
						now := float64(time.Now().UnixNano()) / float64(time.Second)
						v.RefTime = now - RxTime + MuxTime
					}
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
							t.Fatalf("Tx Acknowledgement failed")
						}
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}

				case messages.UplinkDataFrame:
					if MuxTime != 0 {
						time.Sleep(tc.WaitTime)
						now := float64(time.Now().UnixNano()) / float64(time.Second)
						v.RefTime = now - RxTime + MuxTime
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
						a.So(lorawan.UnmarshalMessage(up.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.RawPayload)
						}
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}

				case messages.JoinRequest:
					if MuxTime != 0 {
						time.Sleep(tc.WaitTime)
						now := float64(time.Now().Unix()) + float64(time.Now().Nanosecond())/(1e9)
						v.RefTime = now - RxTime + MuxTime
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
						a.So(lorawan.UnmarshalMessage(up.RawPayload, &payload), should.BeNil)
						if !a.So(&payload, should.Resemble, up.Payload) {
							t.Fatalf("Invalid RawPayload: %v", up.RawPayload)
						}
					case <-time.After(timeout):
						t.Fatalf("Read message timeout")
					}
				}

				if MuxTime > 0 {
					// Atleast one downlink is needed for the first muxtime
					min, max, median, count := gsConn.RTTStats()
					if !a.So(count, should.Equal, tc.ExpectedRTTStatsCount) {
						t.Fatalf("Incorrect Stats entries recorded: %d", count)
					}
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

			if tc.InputNetworkDownstream != nil {
				if _, err := gsConn.SendDown(tc.InputDownlinkPath, tc.InputNetworkDownstream); err != nil {
					t.Fatalf("Failed to send downlink: %v", err)
				}

				resCh := make(chan []byte)
				go func() {
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						t.Fatalf("Failed to read message: %v", err)
					}
					resCh <- data
				}()
				select {
				case res := <-resCh:
					var msg messages.DownlinkMessage
					if err := json.Unmarshal(res, &msg); err != nil {
						t.Fatalf("Failed to unmarshal response `%s`: %v", string(res), err)
					}
					MuxTime = msg.MuxTime
					RxTime = float64(time.Now().Unix()) + float64(time.Now().Nanosecond())/(1e9)
				case <-time.After(timeout):
					t.Fatalf("Read message timeout")
				}
			}
		})
	}

}
