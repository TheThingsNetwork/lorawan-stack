// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package lbslns_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/iotest"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestFrontend(t *testing.T) {
	t.Parallel()
	wsPingInterval := (1 << 3) * test.Delay
	iotest.Frontend(t, iotest.FrontendConfig{
		SupportsStatus:         false,
		DetectsInvalidMessages: true,
		DetectsDisconnect:      true,
		AuthenticatesWithEUI:   true,
		IsAuthenticated:        true,
		DeduplicatesUplinks:    false,
		CustomConfig: func(config *gatewayserver.Config) {
			config.BasicStation = gatewayserver.BasicStationConfig{
				Listen: ":1887",
				Config: semtechws.Config{
					WSPingInterval:       wsPingInterval,
					MissedPongThreshold:  2,
					AllowUnauthenticated: true,
				},
			}
		},
		Link: func(
			ctx context.Context,
			t *testing.T,
			gs *gatewayserver.GatewayServer,
			ids *ttnpb.GatewayIdentifiers,
			key string,
			upCh <-chan *ttnpb.GatewayUp,
			downCh chan<- *ttnpb.GatewayDown,
		) error {
			if ids.Eui == nil {
				t.SkipNow()
			}
			wsConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://0.0.0.0:1887/traffic/eui-%08x", ids.Eui), nil)
			if err != nil {
				return err
			}
			defer wsConn.Close()
			ctx, cancel := errorcontext.New(ctx)
			// Write upstream.
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case msg := <-upCh:
						for _, uplink := range msg.UplinkMessages {
							var payload ttnpb.Message
							if err := lorawan.UnmarshalMessage(uplink.RawPayload, &payload); err != nil {
								// Ignore invalid uplinks
								continue
							}
							var bsUpstream []byte
							switch payload.MHdr.MType {
							case ttnpb.MType_JOIN_REQUEST:
								var jreq lbslns.JoinRequest
								err := jreq.FromUplinkMessage(uplink, band.EU_863_870)
								if err != nil {
									cancel(err)
									return
								}
								bsUpstream, err = jreq.MarshalJSON()
								if err != nil {
									cancel(err)
									return
								}
							case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP:
								var updf lbslns.UplinkDataFrame
								err := updf.FromUplinkMessage(uplink, band.EU_863_870)
								if err != nil {
									cancel(err)
									return
								}
								bsUpstream, err = updf.MarshalJSON()
								if err != nil {
									cancel(err)
									return
								}
							}
							if err := wsConn.WriteMessage(websocket.TextMessage, bsUpstream); err != nil {
								cancel(err)
								return
							}
						}
						if msg.TxAcknowledgment != nil {
							txConf := lbslns.TxConfirmation{
								Diid:  0,
								XTime: time.Now().Unix(),
							}
							bsUpstream, err := txConf.MarshalJSON()
							if err != nil {
								cancel(err)
								return
							}
							if err := wsConn.WriteMessage(websocket.TextMessage, bsUpstream); err != nil {
								cancel(err)
								return
							}
						}
					}
				}
			}()
			// Read downstream.
			go func() {
				for {
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						cancel(err)
						return
					}
					var msg lbslns.DownlinkMessage
					if err := json.Unmarshal(data, &msg); err != nil {
						cancel(err)
						return
					}
					dlmesg, err := msg.ToDownlinkMessage(band.EU_863_870)
					if err != nil {
						cancel(err)
						return
					}
					downCh <- &ttnpb.GatewayDown{
						DownlinkMessage: dlmesg,
					}
				}
			}()
			<-ctx.Done()
			return ctx.Err()
		},
	})
}
