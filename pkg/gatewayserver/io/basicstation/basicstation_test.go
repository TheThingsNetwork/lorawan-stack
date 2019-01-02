// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	registeredGatewayUID = "test-gateway"
	registeredGatewayID  = ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}
	registeredGatewayKey = "test-key"

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

	url := "ws://localhost:8100/api/v3/gs/io/basicstation/discover"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	eui := messages.EUI{
		Prefix: "router",
		EUI64:  types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
	}

	query := messages.DiscoverQuery{
		EUI: eui,
	}
	req, err := json.Marshal(query)
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
			EUI: eui,
			Muxs: messages.EUI{
				Prefix: "muxs",
			},
			URI: "ws://localhost:8100/api/v3/gs/io/basicstation/traffic/eui-0101010101010101",
		})
	case <-time.After(timeout):
		t.Fatalf("Read message timeout")
	}
}

func TestTraffic(t *testing.T) {
	// TODO: Test traffic, see gRPC frontend.
}
