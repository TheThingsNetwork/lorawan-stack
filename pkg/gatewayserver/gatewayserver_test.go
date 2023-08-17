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

package gatewayserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"github.com/smarty/assertions"
	ttnpbv2 "go.thethings.network/lorawan-stack-legacy/v2/pkg/ttnpb"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/lbslns"
	gsredis "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/mock"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	registeredGatewayID  = "eui-aaee000000000000"
	registeredGatewayKey = "secret"
	registeredGatewayEUI = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	unregisteredGatewayEUI = types.EUI64{0xBB, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	timeout        = (1 << 5) * test.Delay
	wsPingInterval = (1 << 3) * test.Delay
)

func TestGatewayServer(t *testing.T) {
	// The component's gRPC address is bound for each iteration and cannot be used in parallel.
	for _, rtc := range []struct { //nolint:paralleltest
		Name                   string
		Setup                  func(context.Context, string, string, gatewayserver.GatewayConnectionStatsRegistry) (*component.Component, *gatewayserver.GatewayServer, error)
		PostSetup              func(context.Context, *component.Component, *mockis.MockDefinition)
		SkipProtocols          func(string) bool
		SupportsLocationUpdate bool
	}{
		{
			Name: "IdentityServer",
			Setup: func(ctx context.Context, isAddr string, nsAddr string, statsRegistry gatewayserver.GatewayConnectionStatsRegistry) (*component.Component, *gatewayserver.GatewayServer, error) {
				c := componenttest.NewComponent(t, &component.Config{
					ServiceBase: config.ServiceBase{
						GRPC: config.GRPC{
							Listen:                      ":9187",
							AllowInsecureForCredentials: true,
						},
						Cluster: cluster.Config{
							IdentityServer: isAddr,
							NetworkServer:  nsAddr,
						},
						FrequencyPlans: config.FrequencyPlansConfig{
							ConfigSource: "static",
							Static:       test.StaticFrequencyPlans,
						},
					},
				})
				gsConfig := &gatewayserver.Config{
					RequireRegisteredGateways:         false,
					UpdateGatewayLocationDebounceTime: 0,
					UpdateConnectionStatsInterval:     (1 << 5) * test.Delay,
					ConnectionStatsTTL:                (1 << 6) * test.Delay,
					ConnectionStatsDisconnectTTL:      (1 << 7) * test.Delay,
					Stats:                             statsRegistry,
					FetchGatewayInterval:              time.Minute,
					FetchGatewayJitter:                1,
					MQTT: config.MQTT{
						Listen: ":1882",
					},
					UDP: gatewayserver.UDPConfig{
						Config: udp.Config{
							PacketHandlers:      2,
							PacketBuffer:        10,
							DownlinkPathExpires: 1 * time.Second,
							ConnectionExpires:   2 * time.Second,
							ScheduleLateTime:    0,
							AddrChangeBlock:     2 * time.Second,
						},
						Listeners: map[string]string{
							":1700": test.EUFrequencyPlanID,
						},
					},
					BasicStation: gatewayserver.BasicStationConfig{
						Listen: ":1887",
						Config: ws.Config{
							WSPingInterval:       wsPingInterval,
							MissedPongThreshold:  2,
							AllowUnauthenticated: true,
						},
					},
				}

				er := gatewayserver.NewIS(c)
				gs, err := gatewayserver.New(c, gsConfig,
					gatewayserver.WithRegistry(er),
				)
				if err != nil {
					return nil, nil, err
				}
				return c, gs, nil
			},
			PostSetup: func(ctx context.Context, c *component.Component, is *mockis.MockDefinition) {
				mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
				mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

				ids := &ttnpb.GatewayIdentifiers{
					GatewayId: registeredGatewayID,
					Eui:       registeredGatewayEUI.Bytes(),
				}
				gtw := mockis.DefaultGateway(ids, true, true)
				is.GatewayRegistry().Add(ctx, ids, registeredGatewayKey, gtw, testRights...)
			},
			SkipProtocols: func(string) bool {
				return false
			},
			SupportsLocationUpdate: true,
		},
	} {
		t.Run(rtc.Name, func(t *testing.T) {
			a, ctx := test.New(t)

			is, isAddr, closeIS := mockis.New(ctx)
			defer closeIS()
			ns, nsAddr := mock.StartNS(ctx)

			var statsRegistry gatewayserver.GatewayConnectionStatsRegistry
			if os.Getenv("TEST_REDIS") == "1" {
				statsRedisClient, statsFlush := test.NewRedis(ctx, "gatewayserver_test")
				defer statsFlush()
				defer statsRedisClient.Close()
				registry := &gsredis.GatewayConnectionStatsRegistry{
					Redis:   statsRedisClient,
					LockTTL: test.Delay << 10,
				}
				if err := registry.Init(ctx); !a.So(err, should.BeNil) {
					t.FailNow()
				}
				statsRegistry = registry
			}

			c, gs, err := rtc.Setup(ctx, isAddr, nsAddr, statsRegistry)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to setup server :%v", err)
			}
			defer c.Close()
			roles := gs.Roles()
			a.So(len(roles), should.Equal, 1)
			a.So(roles[0], should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER)

			config, err := gs.GetConfig(ctx)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			componenttest.StartComponent(t, c)

			rtc.PostSetup(ctx, c, is)

			time.Sleep(timeout) // Wait for setup to be completed.

			for _, ptc := range []struct {
				Protocol               string
				SupportsStatus         bool
				DetectsInvalidMessages bool
				DetectsDisconnect      bool
				TimeoutOnInvalidAuth   bool
				HasAuth                bool
				DeduplicatesUplinks    bool
				ValidAuth              func(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string) bool
				Link                   func(ctx context.Context, t *testing.T, ids *ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error
			}{
				{
					Protocol:            "grpc",
					SupportsStatus:      true,
					HasAuth:             true,
					DetectsDisconnect:   true,
					DeduplicatesUplinks: true,

					ValidAuth: func(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string) bool {
						return ids.GatewayId == registeredGatewayID && key == registeredGatewayKey
					},
					Link: func(ctx context.Context, t *testing.T, ids *ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
						conn, err := grpc.Dial(":9187", append(rpcclient.DefaultDialOptions(ctx), grpc.WithInsecure(), grpc.WithBlock())...)
						if err != nil {
							return err
						}
						defer conn.Close()
						md := rpcmetadata.MD{
							ID:            ids.GatewayId,
							AuthType:      "Bearer",
							AuthValue:     key,
							AllowInsecure: true,
						}
						client := ttnpb.NewGtwGsClient(conn)
						_, err = client.GetConcentratorConfig(ctx, ttnpb.Empty, grpc.PerRPCCredentials(md))
						if err != nil {
							return err
						}
						link, err := client.LinkGateway(ctx, grpc.PerRPCCredentials(md))
						if err != nil {
							return err
						}
						ctx, cancel := errorcontext.New(ctx)
						// Write upstream.
						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								case msg := <-upCh:
									if err := link.Send(msg); err != nil {
										cancel(err)
										return
									}
								}
							}
						}()
						// Read downstream.
						go func() {
							for {
								msg, err := link.Recv()
								if err != nil {
									cancel(err)
									return
								}
								downCh <- msg
							}
						}()
						<-ctx.Done()
						return ctx.Err()
					},
				},
				{
					Protocol:             "mqtt",
					SupportsStatus:       true,
					HasAuth:              true,
					DetectsDisconnect:    true,
					TimeoutOnInvalidAuth: true, // The MQTT client keeps reconnecting on invalid auth.
					ValidAuth: func(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string) bool {
						return ids.GatewayId == registeredGatewayID && key == registeredGatewayKey
					},
					Link: func(ctx context.Context, t *testing.T, ids *ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
						if ids.GatewayId == "" {
							t.SkipNow()
						}
						ctx, cancel := errorcontext.New(ctx)
						clientOpts := mqtt.NewClientOptions()
						clientOpts.AddBroker("tcp://0.0.0.0:1882")
						clientOpts.SetUsername(unique.ID(ctx, ids))
						clientOpts.SetPassword(key)
						clientOpts.SetAutoReconnect(false)
						clientOpts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
							cancel(err)
						})
						client := mqtt.NewClient(clientOpts)
						if token := client.Connect(); !token.WaitTimeout(timeout) {
							return context.DeadlineExceeded
						} else if err := token.Error(); err != nil {
							return err
						}
						defer client.Disconnect(uint(timeout / time.Millisecond))
						// Write upstream.
						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								case up := <-upCh:
									for _, msg := range up.UplinkMessages {
										buf, err := proto.Marshal(msg)
										if err != nil {
											cancel(err)
											return
										}
										if token := client.Publish(fmt.Sprintf("v3/%v/up", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
											cancel(token.Error())
											return
										}
									}
									if up.GatewayStatus != nil {
										buf, err := proto.Marshal(up.GatewayStatus)
										if err != nil {
											cancel(err)
											return
										}
										if token := client.Publish(fmt.Sprintf("v3/%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
											cancel(token.Error())
											return
										}
									}
									if up.TxAcknowledgment != nil {
										buf, err := proto.Marshal(up.TxAcknowledgment)
										if err != nil {
											cancel(err)
											return
										}
										if token := client.Publish(fmt.Sprintf("v3/%v/down/ack", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
											cancel(token.Error())
											return
										}
									}
								}
							}
						}()
						// Read downstream.
						token := client.Subscribe(fmt.Sprintf("v3/%v/down", unique.ID(ctx, ids)), 1, func(_ mqtt.Client, raw mqtt.Message) {
							var msg ttnpb.GatewayDown
							if err := proto.Unmarshal(raw.Payload(), &msg); err != nil {
								cancel(err)
								return
							}
							downCh <- &msg
						})
						if token.Wait() && token.Error() != nil {
							return token.Error()
						}
						<-ctx.Done()
						return ctx.Err()
					},
				},
				{
					Protocol:            "udp",
					SupportsStatus:      true,
					DeduplicatesUplinks: true,
					ValidAuth: func(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string) bool {
						return ids.Eui != nil
					},
					Link: func(ctx context.Context, t *testing.T, ids *ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
						if ids.Eui == nil {
							t.SkipNow()
						}
						upConn, err := net.Dial("udp", ":1700")
						if err != nil {
							return err
						}
						downConn, err := net.Dial("udp", ":1700")
						if err != nil {
							return err
						}
						ctx, cancel := errorcontext.New(ctx)
						// Write upstream.
						go func() {
							var token byte
							var readBuf [65507]byte
							for {
								select {
								case <-ctx.Done():
									return
								case up := <-upCh:
									token++
									packet := encoding.Packet{
										GatewayEUI:      types.MustEUI64(ids.Eui),
										ProtocolVersion: encoding.Version1,
										Token:           [2]byte{0x00, token},
										PacketType:      encoding.PushData,
										Data:            &encoding.Data{},
									}
									packet.Data.RxPacket, packet.Data.Stat, packet.Data.TxPacketAck = encoding.FromGatewayUp(up)
									if packet.Data.TxPacketAck != nil {
										packet.PacketType = encoding.TxAck
									}
									writeBuf, err := packet.MarshalBinary()
									if err != nil {
										cancel(err)
										return
									}
									switch packet.PacketType {
									case encoding.PushData:
										if _, err := upConn.Write(writeBuf); err != nil {
											cancel(err)
											return
										}
										if _, err := upConn.Read(readBuf[:]); err != nil {
											cancel(err)
											return
										}
									case encoding.TxAck:
										if _, err := downConn.Write(writeBuf); err != nil {
											cancel(err)
											return
										}
									}
								}
							}
						}()
						// Engage downstream by sending PULL_DATA every 10ms.
						go func() {
							var token byte
							ticker := time.NewTicker(10 * time.Millisecond)
							for {
								select {
								case <-ctx.Done():
									ticker.Stop()
									return
								case <-ticker.C:
									token++
									pull := encoding.Packet{
										GatewayEUI:      types.MustEUI64(ids.Eui),
										ProtocolVersion: encoding.Version1,
										Token:           [2]byte{0x01, token},
										PacketType:      encoding.PullData,
									}
									buf, err := pull.MarshalBinary()
									if err != nil {
										cancel(err)
										return
									}
									if _, err := downConn.Write(buf); err != nil {
										cancel(err)
										return
									}
								}
							}
						}()
						// Read downstream; PULL_RESP and PULL_ACK.
						go func() {
							var buf [65507]byte
							for {
								n, err := downConn.Read(buf[:])
								if err != nil {
									cancel(err)
									return
								}
								packetBuf := make([]byte, n)
								copy(packetBuf, buf[:])
								var packet encoding.Packet
								if err := packet.UnmarshalBinary(packetBuf); err != nil {
									cancel(err)
									return
								}
								switch packet.PacketType {
								case encoding.PullResp:
									msg, err := encoding.ToDownlinkMessage(packet.Data.TxPacket)
									if err != nil {
										cancel(err)
										return
									}
									downCh <- &ttnpb.GatewayDown{
										DownlinkMessage: msg,
									}
								}
							}
						}()
						<-ctx.Done()
						time.Sleep(config.UDP.ConnectionExpires * 150 / 100) // Ensure that connection expires.
						return ctx.Err()
					},
				},
				{
					Protocol:               "basicstation",
					SupportsStatus:         false,
					DetectsDisconnect:      true,
					DetectsInvalidMessages: true,
					HasAuth:                true,
					ValidAuth: func(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string) bool {
						return ids.Eui != nil
					},
					Link: func(ctx context.Context, t *testing.T, ids *ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
						if ids.Eui == nil {
							t.SkipNow()
						}
						wsConn, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:1887/traffic/"+registeredGatewayID, nil)
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
				},
			} {
				if rtc.SkipProtocols(ptc.Protocol) {
					continue
				}
				t.Run(fmt.Sprintf("Authenticate/%v", ptc.Protocol), func(t *testing.T) {
					for _, ctc := range []struct {
						Name string
						ID   *ttnpb.GatewayIdentifiers
						Key  string
					}{
						{
							Name: "ValidIDAndKey",
							ID:   &ttnpb.GatewayIdentifiers{GatewayId: registeredGatewayID},
							Key:  registeredGatewayKey,
						},
						{
							Name: "InvalidKey",
							ID:   &ttnpb.GatewayIdentifiers{GatewayId: registeredGatewayID},
							Key:  "invalid-key",
						},
						{
							Name: "InvalidIDAndKey",
							ID:   &ttnpb.GatewayIdentifiers{GatewayId: "invalid-gateway"},
							Key:  "invalid-key",
						},
						{
							Name: "RegisteredEUI",
							ID:   &ttnpb.GatewayIdentifiers{Eui: registeredGatewayEUI.Bytes()},
						},
						{
							Name: "UnregisteredEUI",
							ID:   &ttnpb.GatewayIdentifiers{Eui: unregisteredGatewayEUI.Bytes()},
						},
					} {
						t.Run(ctc.Name, func(t *testing.T) {
							ctx, cancel := context.WithCancel(ctx)
							upCh := make(chan *ttnpb.GatewayUp)
							downCh := make(chan *ttnpb.GatewayDown)

							upEvents := map[string]events.Channel{}
							for _, event := range []string{"gs.gateway.connect"} {
								upEvents[event] = make(events.Channel, 5)
							}
							defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{
								PublishFunc: func(evs ...events.Event) {
									for _, ev := range evs {
										ev := ev

										switch name := ev.Name(); name {
										case "gs.gateway.connect":
											go func() {
												upEvents[name] <- ev
											}()
										default:
											t.Logf("%s event published", name)
										}
									}
								},
							})()

							connectedWithInvalidAuth := make(chan struct{}, 1)
							expectedProperLink := make(chan struct{}, 1)
							go func() {
								select {
								case <-upEvents["gs.gateway.connect"]:
									if !ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
										connectedWithInvalidAuth <- struct{}{}
									}
								case <-time.After(timeout):
									if ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
										expectedProperLink <- struct{}{}
									}
								}
								time.Sleep(test.Delay)
								cancel()
							}()
							err := ptc.Link(ctx, t, ctc.ID, ctc.Key, upCh, downCh)
							if !errors.IsCanceled(err) && ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
								t.Fatalf("Expect canceled context but have %v", err)
							}
							select {
							case <-connectedWithInvalidAuth:
								t.Fatal("Expected link error due to invalid auth")
							case <-expectedProperLink:
								t.Fatal("Expected proper link")
							default:
							}
						})
					}
				})

				// Wait for gateway disconnection to be processed.
				time.Sleep(timeout)

				t.Run(fmt.Sprintf("DetectDisconnect/%v", ptc.Protocol), func(t *testing.T) {
					if !ptc.DetectsDisconnect {
						t.SkipNow()
					}

					id := &ttnpb.GatewayIdentifiers{
						GatewayId: registeredGatewayID,
						Eui:       registeredGatewayEUI.Bytes(),
					}

					ctx1, fail1 := errorcontext.New(ctx)
					defer fail1(context.Canceled)
					go func() {
						upCh := make(chan *ttnpb.GatewayUp)
						downCh := make(chan *ttnpb.GatewayDown)
						err := ptc.Link(ctx1, t, id, registeredGatewayKey, upCh, downCh)
						fail1(err)
					}()
					select {
					case <-ctx1.Done():
						t.Fatalf("Expected no link error on first connection but have %v", ctx1.Err())
					case <-time.After(timeout):
					}

					ctx2, cancel2 := context.WithDeadline(ctx, time.Now().Add(4*timeout))
					upCh := make(chan *ttnpb.GatewayUp)
					downCh := make(chan *ttnpb.GatewayDown)
					err := ptc.Link(ctx2, t, id, registeredGatewayKey, upCh, downCh)
					cancel2()
					if !errors.IsDeadlineExceeded(err) {
						t.Fatalf("Expected deadline exceeded on second connection but have %v", err)
					}
					select {
					case <-ctx1.Done():
						t.Logf("First connection failed when second connected with %v", ctx1.Err())
					case <-time.After(4 * timeout):
						t.Fatalf("Expected link failure on first connection when second connected")
					}
				})

				// Wait for gateway disconnection to be processed.
				time.Sleep(2 * config.ConnectionStatsDisconnectTTL)

				t.Run(fmt.Sprintf("Traffic/%v", ptc.Protocol), func(t *testing.T) {
					a := assertions.New(t)

					ctx, cancel := context.WithCancel(ctx)
					upCh := make(chan *ttnpb.GatewayUp)
					downCh := make(chan *ttnpb.GatewayDown)
					ids := &ttnpb.GatewayIdentifiers{
						GatewayId: registeredGatewayID,
						Eui:       registeredGatewayEUI.Bytes(),
					}
					// Setup a stats client with independent context to query whether the gateway is connected and statistics on
					// upstream and downstream.
					statsConn, err := grpc.Dial(":9187", append(rpcclient.DefaultDialOptions(test.Context()), grpc.WithInsecure(), grpc.WithBlock())...)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					defer statsConn.Close()
					statsCtx := metadata.AppendToOutgoingContext(test.Context(),
						"id", ids.GatewayId,
						"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
					)
					statsClient := ttnpb.NewGsClient(statsConn)

					// The gateway should not be connected before testing traffic.
					t.Run("NotConnected", func(t *testing.T) {
						_, err := statsClient.GetGatewayConnectionStats(statsCtx, ids)
						if !a.So(errors.IsNotFound(err), should.BeTrue) {
							t.Fatal("Expected gateway not to be connected yet, but it is")
						}
					})

					if ptc.SupportsStatus && ptc.HasAuth && rtc.SupportsLocationUpdate {
						t.Run("UpdateLocation", func(t *testing.T) {
							for _, tc := range []struct {
								Name           string
								UpdateLocation bool
								Up             *ttnpb.GatewayUp
								ExpectLocation *ttnpb.Location
							}{
								{
									Name:           "NoUpdate",
									UpdateLocation: false,
									Up: &ttnpb.GatewayUp{
										GatewayStatus: &ttnpb.GatewayStatus{
											Time: timestamppb.New(time.Unix(424242, 0)),
											AntennaLocations: []*ttnpb.Location{
												{
													Source:    ttnpb.LocationSource_SOURCE_GPS,
													Altitude:  10,
													Latitude:  12,
													Longitude: 14,
												},
											},
										},
									},
									ExpectLocation: &ttnpb.Location{
										Source: ttnpb.LocationSource_SOURCE_GPS,
									},
								},
								{
									Name:           "NoLocation",
									UpdateLocation: true,
									Up: &ttnpb.GatewayUp{
										GatewayStatus: &ttnpb.GatewayStatus{
											Time: timestamppb.New(time.Unix(424242, 0)),
										},
									},
									ExpectLocation: &ttnpb.Location{
										Source: ttnpb.LocationSource_SOURCE_GPS,
									},
								},
								{
									Name:           "Update",
									UpdateLocation: true,
									Up: &ttnpb.GatewayUp{
										GatewayStatus: &ttnpb.GatewayStatus{
											Time: timestamppb.New(time.Unix(42424242, 0)),
											AntennaLocations: []*ttnpb.Location{
												{
													Source:    ttnpb.LocationSource_SOURCE_GPS,
													Altitude:  10,
													Latitude:  12,
													Longitude: 14,
												},
											},
										},
									},
									ExpectLocation: &ttnpb.Location{
										Source:    ttnpb.LocationSource_SOURCE_GPS,
										Altitude:  10,
										Latitude:  12,
										Longitude: 14,
									},
								},
							} {
								t.Run(tc.Name, func(t *testing.T) {
									a := assertions.New(t)

									gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
										GatewayIds: ids,
									})
									a.So(err, should.BeNil)

									gtw.Antennas[0].Location = &ttnpb.Location{
										Source: ttnpb.LocationSource_SOURCE_GPS,
									}
									gtw.UpdateLocationFromStatus = tc.UpdateLocation
									gtw, err = is.GatewayRegistry().Update(ctx, &ttnpb.UpdateGatewayRequest{
										Gateway:   gtw,
										FieldMask: ttnpb.FieldMask("antennas", "update_location_from_status"),
									})
									a.So(err, should.BeNil)
									a.So(gtw.UpdateLocationFromStatus, should.Equal, tc.UpdateLocation)

									ctx, cancel := context.WithCancel(ctx)
									upCh := make(chan *ttnpb.GatewayUp)
									downCh := make(chan *ttnpb.GatewayDown)

									wg := &sync.WaitGroup{}
									wg.Add(1)
									var linkErr error
									go func() {
										defer wg.Done()
										linkErr = ptc.Link(ctx, t, ids, registeredGatewayKey, upCh, downCh)
									}()

									select {
									case upCh <- tc.Up:
									case <-time.After(timeout):
										t.Fatalf("Failed to send message to upstream channel")
									}

									time.Sleep(timeout)
									gtw, err = is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
										GatewayIds: ids,
									})
									a.So(err, should.BeNil)
									a.So(gtw.Antennas[0].Location, should.Resemble, tc.ExpectLocation)

									cancel()
									wg.Wait()
									if !errors.IsCanceled(linkErr) {
										t.Fatalf("Expected context canceled, but have %v", linkErr)
									}
								})
							}
						})
					}

					if rtc.SupportsLocationUpdate {
						t.Run("LocationMetadata", func(t *testing.T) {
							location := &ttnpb.Location{
								Source:    ttnpb.LocationSource_SOURCE_GPS,
								Altitude:  10,
								Latitude:  12,
								Longitude: 14,
							}
							up := &ttnpb.GatewayUp{
								UplinkMessages: []*ttnpb.UplinkMessage{
									{
										Settings: &ttnpb.TxSettings{
											DataRate: &ttnpb.DataRate{
												Modulation: &ttnpb.DataRate_Lora{
													Lora: &ttnpb.LoRaDataRate{
														SpreadingFactor: 7,
														Bandwidth:       250000,
														CodingRate:      band.Cr4_5,
													},
												},
											},
											Frequency: 867900000,
											Timestamp: 100,
										},
										RxMetadata: []*ttnpb.RxMetadata{
											{
												AntennaIndex: 0,
												GatewayIds:   ids,
												Timestamp:    100,
												Rssi:         -69,
												ChannelRssi:  -69,
												Snr:          11,
											},
										},
									},
								},
							}
							for _, locationPublic := range []bool{false, true} {
								t.Run(fmt.Sprintf("LocationPublic=%v", locationPublic), func(t *testing.T) {
									a := assertions.New(t)
									mockGtw := mockis.DefaultGateway(ids, false, false)
									mockGtw.LocationPublic = locationPublic
									is.GatewayRegistry().Add(ctx, ids, registeredGatewayKey, mockGtw, testRights...)

									gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
										GatewayIds: ids,
									})
									a.So(err, should.BeNil)
									a.So(gtw.LocationPublic, should.Equal, locationPublic)
									gtw.LocationPublic = locationPublic
									gtw.Antennas[0].Location = location
									gtw, err = is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
										GatewayIds: ids,
									})
									a.So(err, should.BeNil)
									a.So(gtw.LocationPublic, should.Equal, locationPublic)

									ctx, cancel := context.WithCancel(ctx)
									upCh := make(chan *ttnpb.GatewayUp)
									downCh := make(chan *ttnpb.GatewayDown)

									wg := &sync.WaitGroup{}
									wg.Add(1)
									var linkErr error
									go func() {
										defer wg.Done()
										linkErr = ptc.Link(ctx, t, ids, registeredGatewayKey, upCh, downCh)
									}()

									for _, locationInRxMetadata := range []bool{false, true} {
										t.Run(fmt.Sprintf("RxMetadata=%v", locationInRxMetadata), func(t *testing.T) {
											if !locationInRxMetadata && locationPublic {
												// Disabled, because this is inconsistent amongst frontends
												// - gRPC and MQTT: location is in RxMetadata
												// - UDP and BasicStation: location is not in RxMetadata
												t.SkipNow()
											}
											a := assertions.New(t)

											if locationInRxMetadata {
												up.UplinkMessages[0].RxMetadata[0].Location = location
											} else {
												up.UplinkMessages[0].RxMetadata[0].Location = nil
											}
											up.UplinkMessages[0].RawPayload = randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6)

											select {
											case upCh <- up:
											case <-time.After(timeout):
												t.Fatalf("Failed to send message to upstream channel")
											}

											select {
											case msg := <-ns.Up():
												if a.So(len(msg.RxMetadata), should.Equal, 1) {
													if locationPublic {
														a.So(msg.RxMetadata[0].Location, should.Resemble, location)
													} else {
														a.So(msg.RxMetadata[0].Location, should.BeNil)
													}
												}
											case <-time.After(2 * timeout):
												t.Fatalf("Failed to get message")
											}
										})
									}

									cancel()
									wg.Wait()
									if !errors.IsCanceled(linkErr) {
										t.Fatalf("Expected context canceled, but have %v", linkErr)
									}
								})
							}
						})
					}
					// Wait for gateway disconnection to be processed.
					time.Sleep(timeout)

					wg := &sync.WaitGroup{}
					wg.Add(1)
					var linkErr error
					go func() {
						defer wg.Done()
						linkErr = ptc.Link(ctx, t, ids, registeredGatewayKey, upCh, downCh)
					}()

					// Expected location for RxMetadata
					var location *ttnpb.Location
					if rtc.SupportsLocationUpdate {
						gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
							GatewayIds: ids,
						})
						a.So(err, should.BeNil)
						location = gtw.Antennas[0].Location
					}

					duplicatePayload := randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6)

					t.Run("Upstream", func(t *testing.T) {
						uplinkCount := 0
						for _, tc := range []struct {
							Name                         string
							Up                           *ttnpb.GatewayUp
							Received                     []uint32 // Timestamps of uplink messages in Up that are received.
							Dropped                      []uint32 // Timestamps of uplink messages in Up that are dropped.
							PublicLocation               bool     // If gateway location is public, it should be in RxMetadata
							UplinkCount                  int      // Number of expected uplinks
							RepeatUpEvent                bool     // Expect event for repeated uplinks
							SkipIfDetectsInvalidMessages bool     // Skip this test if the frontend detects invalid messages
						}{
							{
								Name: "GatewayStatus",
								Up: &ttnpb.GatewayUp{
									GatewayStatus: &ttnpb.GatewayStatus{
										Time: timestamppb.New(time.Unix(424242, 0)),
									},
								},
							},
							{
								Name: "TxAck",
								Up: &ttnpb.GatewayUp{
									TxAcknowledgment: &ttnpb.TxAcknowledgment{
										Result: ttnpb.TxAcknowledgment_SUCCESS,
									},
								},
							},
							{
								Name: "CRCFailure",
								Up: &ttnpb.GatewayUp{
									UplinkMessages: []*ttnpb.UplinkMessage{
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 7,
															Bandwidth:       250000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 867900000,
												Timestamp: 100,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   100,
													Rssi:        -69,
													ChannelRssi: -69,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 2, 2),
											CrcStatus:  wrapperspb.Bool(false),
										},
									},
								},
								Received:                     []uint32{100},
								Dropped:                      []uint32{100},
								SkipIfDetectsInvalidMessages: true,
							},
							{
								Name: "OneValidLoRa",
								Up: &ttnpb.GatewayUp{
									UplinkMessages: []*ttnpb.UplinkMessage{
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 7,
															Bandwidth:       250000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 867900000,
												Timestamp: 200,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   200,
													Rssi:        -69,
													ChannelRssi: -69,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
										},
									},
								},
								Received: []uint32{200},
							},
							{
								Name: "OneValidLoRaAndTwoRepeated",
								Up: &ttnpb.GatewayUp{
									UplinkMessages: []*ttnpb.UplinkMessage{
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 7,
															Bandwidth:       250000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 867900000,
												Timestamp: 301,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   301,
													Rssi:        -42,
													ChannelRssi: -42,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: duplicatePayload,
										},
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 7,
															Bandwidth:       250000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 867900000,
												Timestamp: 300,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   300,
													Rssi:        -69,
													ChannelRssi: -69,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: duplicatePayload,
										},
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 7,
															Bandwidth:       250000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 867900000,
												Timestamp: 300,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   300,
													Rssi:        -69,
													ChannelRssi: -69,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: duplicatePayload,
										},
									},
								},
								Received:      []uint32{301},
								UplinkCount:   1,
								RepeatUpEvent: true,
							},
							{
								Name: "OneValidFSK",
								Up: &ttnpb.GatewayUp{
									UplinkMessages: []*ttnpb.UplinkMessage{
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Fsk{
														Fsk: &ttnpb.FSKDataRate{
															BitRate: 50000,
														},
													},
												},
												Frequency: 867900000,
												Timestamp: 400,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   400,
													Rssi:        -69,
													ChannelRssi: -69,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
										},
									},
								},
								Received: []uint32{400},
							},
							{
								Name: "OneGarbageWithStatus",
								Up: &ttnpb.GatewayUp{
									UplinkMessages: []*ttnpb.UplinkMessage{
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 9,
															Bandwidth:       125000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 868500000,
												Timestamp: 500,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   500,
													Rssi:        -112,
													ChannelRssi: -112,
													Snr:         2,
													Location:    location,
												},
											},
											RawPayload: []byte{0xff, 0x02, 0x03}, // Garbage; doesn't get forwarded.
										},
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 7,
															Bandwidth:       125000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 868100000,
												Timestamp: 501,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   501,
													Rssi:        -69,
													ChannelRssi: -69,
													Snr:         11,
													Location:    location,
												},
											},
											RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
										},
										{
											Settings: &ttnpb.TxSettings{
												DataRate: &ttnpb.DataRate{
													Modulation: &ttnpb.DataRate_Lora{
														Lora: &ttnpb.LoRaDataRate{
															SpreadingFactor: 12,
															Bandwidth:       125000,
															CodingRate:      band.Cr4_5,
														},
													},
												},
												Frequency: 867700000,
												Timestamp: 502,
											},
											RxMetadata: []*ttnpb.RxMetadata{
												{
													GatewayIds:  ids,
													Timestamp:   502,
													Rssi:        -36,
													ChannelRssi: -36,
													Snr:         5,
													Location:    location,
												},
											},
											RawPayload: randomJoinRequestPayload(
												types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
												types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
											),
										},
									},
									GatewayStatus: &ttnpb.GatewayStatus{
										Time: timestamppb.New(time.Unix(4242424, 0)),
									},
								},
								Received: []uint32{501, 502},
							},
						} {
							t.Run(tc.Name, func(t *testing.T) {
								a := assertions.New(t)

								if tc.SkipIfDetectsInvalidMessages && ptc.DetectsInvalidMessages {
									t.Skip("Skipping test case because gateway detects invalid messages")
								}

								upEvents := map[string]events.Channel{}
								for _, event := range []string{"gs.up.receive", "gs.down.tx.success", "gs.down.tx.fail", "gs.status.receive", "gs.io.up.repeat"} {
									upEvents[event] = make(events.Channel, 5)
								}
								defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{
									PublishFunc: func(evs ...events.Event) {
										for _, ev := range evs {
											ev := ev
											switch name := ev.Name(); name {
											case "gs.up.receive", "gs.down.tx.success", "gs.down.tx.fail", "gs.status.receive", "gs.io.up.repeat":
												go func() {
													upEvents[name] <- ev
												}()
											default:
												t.Logf("%s event published", name)
											}
										}
									},
								})()

								select {
								case upCh <- tc.Up:
								case <-time.After(timeout):
									t.Fatalf("Failed to send message to upstream channel")
								}
								if tc.UplinkCount > 0 {
									uplinkCount += tc.UplinkCount
								} else if ptc.DetectsInvalidMessages {
									uplinkCount += len(tc.Received)
								} else {
									uplinkCount += len(tc.Up.UplinkMessages)
								}

								if tc.RepeatUpEvent && !ptc.DeduplicatesUplinks {
									select {
									case evt := <-upEvents["gs.io.up.repeat"]:
										a.So(evt.Name(), should.Equal, "gs.io.up.repeat")
									case <-time.After(timeout):
										t.Fatal("Expected repeat uplink event timeout")
									}
								}

								received := make(map[uint32]struct{})
								forwarded := make(map[uint32]struct{})
								for _, t := range tc.Received {
									received[t] = struct{}{}
									forwarded[t] = struct{}{}
								}
								for _, t := range tc.Dropped {
									delete(forwarded, t)
								}
								for len(received) > 0 {
									if len(forwarded) > 0 {
										select {
										case msg := <-ns.Up():
											var expected *ttnpb.UplinkMessage
											for _, up := range tc.Up.UplinkMessages {
												if ts := up.Settings.Timestamp; ts == msg.Settings.Timestamp {
													if _, ok := forwarded[ts]; !ok {
														t.Fatalf("Not expecting message %v", msg)
													}
													expected = up
													delete(forwarded, ts)
													break
												}
											}
											if expected == nil {
												t.Fatalf("Received unexpected message with timestamp %d", msg.Settings.Timestamp)
											}
											a.So(time.Since(*ttnpb.StdTime(msg.ReceivedAt)), should.BeLessThan, timeout)
											a.So(msg.Settings, should.Resemble, expected.Settings)
											a.So(len(msg.RxMetadata), should.Equal, len(expected.RxMetadata))
											for i, md := range msg.RxMetadata {
												a.So(md.UplinkToken, should.NotBeEmpty)
												md.UplinkToken = nil
												md.ReceivedAt = nil
												a.So(md, should.Resemble, expected.RxMetadata[i])
											}
											a.So(msg.RawPayload, should.Resemble, expected.RawPayload)
										case <-time.After(timeout):
											t.Fatal("Expected uplink timeout")
										}
									}
									select {
									case evt := <-upEvents["gs.up.receive"]:
										a.So(evt.Name(), should.Equal, "gs.up.receive")
										msg := evt.Data().(*ttnpb.GatewayUplinkMessage)
										delete(received, msg.Message.Settings.Timestamp)
									case <-time.After(timeout):
										t.Fatal("Expected uplink event timeout")
									}
								}
								if expected := tc.Up.TxAcknowledgment; expected != nil {
									select {
									case <-upEvents["gs.down.tx.success"]:
									case evt := <-upEvents["gs.down.tx.fail"]:
										received, ok := evt.Data().(ttnpb.TxAcknowledgment_Result)
										if !ok {
											t.Fatal("No acknowledgment attached to the downlink emission fail event")
										}
										a.So(received, should.Resemble, expected.Result)
									case <-time.After(timeout):
										t.Fatal("Expected Tx acknowledgment event timeout")
									}
									select {
									case ack := <-ns.TxAck():
										if txAck := ack.GetTxAck(); a.So(txAck, should.NotBeNil) {
											a.So(txAck.Result, should.Resemble, expected.Result)
											a.So(txAck.DownlinkMessage, should.Resemble, expected.DownlinkMessage)
										}
									case <-time.After(timeout):
										t.Fatal("Expected Tx acknowledgment event timeout")
									}
								}
								if tc.Up.GatewayStatus != nil && ptc.SupportsStatus {
									select {
									case <-upEvents["gs.status.receive"]:
									case <-time.After(timeout):
										t.Fatal("Expected gateway status event timeout")
									}
								}

								time.Sleep(2 * timeout)

								conn, ok := gs.GetConnection(ctx, ids)
								a.So(ok, should.BeTrue)

								stats, paths := conn.Stats()
								a.So(stats, should.NotBeNil)
								a.So(paths, should.NotBeEmpty)

								stats, err := statsClient.GetGatewayConnectionStats(statsCtx, ids)
								if !a.So(err, should.BeNil) {
									t.FailNow()
								}
								a.So(stats.UplinkCount, should.Equal, uplinkCount)

								if tc.Up.GatewayStatus != nil && ptc.SupportsStatus {
									if !a.So(stats.LastStatus, should.NotBeNil) {
										t.FailNow()
									}
									a.So(stats.LastStatus.Time, should.Resemble, tc.Up.GatewayStatus.Time)
								}
							})
						}
					})

					t.Run("Downstream", func(t *testing.T) {
						ctx := clusterauth.NewContext(test.Context(), nil)
						downlinkCount := 0
						for _, tc := range []struct {
							Name                     string
							Message                  *ttnpb.DownlinkMessage
							ErrorAssertion           func(error) bool
							RxWindowDetailsAssertion []func(error) bool
						}{
							{
								Name: "InvalidSettingsType",
								Message: &ttnpb.DownlinkMessage{
									RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
									Settings: &ttnpb.DownlinkMessage_Scheduled{
										Scheduled: &ttnpb.TxSettings{
											DataRate: &ttnpb.DataRate{
												Modulation: &ttnpb.DataRate_Lora{
													Lora: &ttnpb.LoRaDataRate{
														SpreadingFactor: 12,
														Bandwidth:       125000,
														CodingRate:      band.Cr4_5,
													},
												},
											},
											Frequency: 869525000,
											Downlink: &ttnpb.TxSettings_Downlink{
												TxPower: 10,
											},
											Timestamp: 100,
										},
									},
								},
								ErrorAssertion: errors.IsInvalidArgument, // Network Server may send Tx request only.
							},
							{
								Name: "NotConnected",
								Message: &ttnpb.DownlinkMessage{
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class: ttnpb.Class_CLASS_C,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIds: &ttnpb.GatewayIdentifiers{
																GatewayId: "not-connected",
															},
														},
													},
												},
											},
											FrequencyPlanId: test.EUFrequencyPlanID,
										},
									},
								},
								ErrorAssertion: errors.IsAborted, // The gateway is not connected.
							},
							{
								Name: "ValidClassA",
								Message: &ttnpb.DownlinkMessage{
									RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class: ttnpb.Class_CLASS_A,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: io.MustUplinkToken(
															&ttnpb.GatewayAntennaIdentifiers{
																GatewayIds: &ttnpb.GatewayIdentifiers{
																	GatewayId: registeredGatewayID,
																},
															},
															10000000,
															10000000000,
															time.Unix(0, 10000000*1000),
															nil,
														),
													},
												},
											},
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
								},
							},
							{
								Name: "ValidClassAWithoutFrequencyPlanInTxRequest",
								Message: &ttnpb.DownlinkMessage{
									RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class: ttnpb.Class_CLASS_A,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: io.MustUplinkToken(
															&ttnpb.GatewayAntennaIdentifiers{
																GatewayIds: &ttnpb.GatewayIdentifiers{
																	GatewayId: registeredGatewayID,
																},
															},
															20000000,
															20000000000,
															time.Unix(0, 20000000*1000),
															nil,
														),
													},
												},
											},
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
											Rx1Frequency: 868100000,
										},
									},
								},
							},
							{
								Name: "ConflictClassA",
								Message: &ttnpb.DownlinkMessage{
									RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class: ttnpb.Class_CLASS_A,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: io.MustUplinkToken(
															&ttnpb.GatewayAntennaIdentifiers{
																GatewayIds: &ttnpb.GatewayIdentifiers{
																	GatewayId: registeredGatewayID,
																},
															},
															10000000,
															10000000000,
															time.Unix(0, 10000000*1000),
															nil,
														),
													},
												},
											},
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
								},
								ErrorAssertion: errors.IsAborted,
								RxWindowDetailsAssertion: []func(error) bool{
									errors.IsAlreadyExists,      // Rx1 conflicts with previous.
									errors.IsFailedPrecondition, // Rx2 not provided.
								},
							},
							{
								Name: "ValidClassC",
								Message: &ttnpb.DownlinkMessage{
									RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class: ttnpb.Class_CLASS_C,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIds: &ttnpb.GatewayIdentifiers{
																GatewayId: registeredGatewayID,
															},
														},
													},
												},
											},
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
								},
							},
							{
								Name: "ValidClassCWithoutFrequencyPlanInTxRequest",
								Message: &ttnpb.DownlinkMessage{
									RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class: ttnpb.Class_CLASS_C,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIds: &ttnpb.GatewayIdentifiers{
																GatewayId: registeredGatewayID,
															},
														},
													},
												},
											},
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
											Rx1Frequency: 868100000,
										},
									},
								},
							},
						} {
							t.Run(tc.Name, func(t *testing.T) {
								a := assertions.New(t)

								_, err := gs.ScheduleDownlink(ctx, tc.Message)
								if err != nil {
									if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
										t.Fatalf("Unexpected error: %v", err)
									}
									if tc.RxWindowDetailsAssertion != nil {
										a.So(err, should.HaveSameErrorDefinitionAs, gatewayserver.ErrSchedule)
										if !a.So(errors.Details(err), should.HaveLength, 1) {
											t.FailNow()
										}
										details := errors.Details(err)[0].(*ttnpb.ScheduleDownlinkErrorDetails)
										if !a.So(details, should.NotBeNil) || !a.So(details.PathErrors, should.HaveLength, 1) {
											t.FailNow()
										}
										errSchedulePathCause := errors.Cause(ttnpb.ErrorDetailsFromProto(details.PathErrors[0]))
										a.So(errors.IsAborted(errSchedulePathCause), should.BeTrue)
										for i, assert := range tc.RxWindowDetailsAssertion {
											if !a.So(errors.Details(errSchedulePathCause), should.HaveLength, 1) {
												t.FailNow()
											}
											errSchedulePathCauseDetails := errors.Details(errSchedulePathCause)[0].(*ttnpb.ScheduleDownlinkErrorDetails)
											if !a.So(errSchedulePathCauseDetails, should.NotBeNil) {
												t.FailNow()
											}
											if i >= len(errSchedulePathCauseDetails.PathErrors) {
												t.Fatalf("Expected error in Rx window %d", i+1)
											}
											errRxWindow := ttnpb.ErrorDetailsFromProto(errSchedulePathCauseDetails.PathErrors[i])
											if !a.So(assert(errRxWindow), should.BeTrue) {
												t.Fatalf("Unexpected Rx window %d error: %v", i+1, errRxWindow)
											}
										}
									}
									return
								} else if tc.ErrorAssertion != nil {
									t.Fatalf("Expected error")
								}
								downlinkCount++

								select {
								case msg := <-downCh:
									settings := msg.DownlinkMessage.GetScheduled()
									a.So(settings, should.NotBeNil)
								case <-time.After(timeout):
									t.Fatal("Expected downlink timeout")
								}

								time.Sleep(2 * timeout)

								conn, ok := gs.GetConnection(ctx, ids)
								a.So(ok, should.BeTrue)

								stats, paths := conn.Stats()
								a.So(stats, should.NotBeNil)
								a.So(paths, should.NotBeEmpty)

								stats, err = statsClient.GetGatewayConnectionStats(statsCtx, ids)
								if !a.So(err, should.BeNil) {
									t.FailNow()
								}
								a.So(stats.DownlinkCount, should.Equal, downlinkCount)
							})
						}
					})

					cancel()
					wg.Wait()
					if !errors.IsCanceled(linkErr) {
						t.Fatalf("Expected context canceled, but have %v", linkErr)
					}

					// Wait for disconnection to be processed.
					time.Sleep(2 * config.ConnectionStatsDisconnectTTL)

					// After canceling the context and awaiting the link, the connection should be gone.
					t.Run("Disconnected", func(t *testing.T) {
						_, err := statsClient.GetGatewayConnectionStats(statsCtx, ids)
						if !a.So(errors.IsNotFound(err), should.BeTrue) {
							t.Fatalf("Expected gateway to be disconnected, but it's not")
						}
					})
				})
			}

			t.Run("Shutdown", func(t *testing.T) {
				if statsRegistry == nil {
					t.Skip("Stats registry disabled")
				}

				ids := &ttnpb.GatewayIdentifiers{
					GatewayId: registeredGatewayID,
					Eui:       registeredGatewayEUI.Bytes(),
				}

				conn, err := grpc.Dial(":9187", append(rpcclient.DefaultDialOptions(ctx), grpc.WithInsecure(), grpc.WithBlock())...)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				defer conn.Close()
				md := rpcmetadata.MD{
					ID:            ids.GatewayId,
					AuthType:      "Bearer",
					AuthValue:     registeredGatewayKey,
					AllowInsecure: true,
				}
				_, err = ttnpb.NewGtwGsClient(conn).LinkGateway(ctx, grpc.PerRPCCredentials(md))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				gs.Close()
				time.Sleep(2 * config.ConnectionStatsTTL)

				_, err = statsRegistry.Get(ctx, ids)
				a.So(errors.IsNotFound(err), should.BeTrue)
			})
		})
	}
}

func TestUpdateVersionInfo(t *testing.T) { //nolint:paralleltest
	a, ctx := test.New(t)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.GatewayRegistry().SetRegisteredGateway(&ttnpb.GatewayIdentifiers{
		GatewayId: registeredGatewayID,
		Eui:       registeredGatewayEUI.Bytes(),
	})

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9187",
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
	defer c.Close()

	gatewayFetchInterval := test.Delay

	gsConfig := &gatewayserver.Config{
		FetchGatewayInterval:   gatewayFetchInterval,
		FetchGatewayJitter:     1,
		UpdateVersionInfoDelay: test.Delay,
		MQTTV2: config.MQTT{
			Listen: ":1881",
		},
	}

	er := gatewayserver.NewIS(c)
	gs, err := gatewayserver.New(c, gsConfig,
		gatewayserver.WithRegistry(er),
	)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Failed to setup server :%v", err)
	}
	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER)
	a.So(err, should.BeNil)

	componenttest.StartComponent(t, c)
	time.Sleep(timeout) // Wait for component to start.

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	gtwIDs := &ttnpb.GatewayIdentifiers{
		GatewayId: registeredGatewayID,
		Eui:       registeredGatewayEUI.Bytes(),
	}

	mockGtw := mockis.DefaultGateway(gtwIDs, true, true)
	is.GatewayRegistry().Add(ctx, gtwIDs, registeredGatewayKey, mockGtw, testRights...)
	time.Sleep(timeout) // Wait for setup to be completed.

	linkFn := func(ctx context.Context, t *testing.T, ids *ttnpb.GatewayIdentifiers, key string, statCh <-chan *ttnpbv2.StatusMessage) error {
		ctx, cancel := errorcontext.New(ctx)
		clientOpts := mqtt.NewClientOptions()
		clientOpts.AddBroker("tcp://0.0.0.0:1881")
		clientOpts.SetUsername(unique.ID(ctx, ids))
		clientOpts.SetPassword(key)
		clientOpts.SetAutoReconnect(false)
		clientOpts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			cancel(err)
		})
		client := mqtt.NewClient(clientOpts)
		if token := client.Connect(); !token.WaitTimeout(timeout) {
			return context.DeadlineExceeded
		} else if err := token.Error(); err != nil {
			return err
		}
		defer client.Disconnect(uint(timeout / time.Millisecond))
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case stat := <-statCh:
					buf, err := proto.Marshal(stat)
					if err != nil {
						cancel(err)
						return
					}
					if token := client.Publish(fmt.Sprintf("%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
						cancel(token.Error())
						return
					}

				}
			}
		}()
		<-ctx.Done()
		return ctx.Err()
	}

	statCh := make(chan *ttnpbv2.StatusMessage)
	ids := &ttnpb.GatewayIdentifiers{
		GatewayId: registeredGatewayID,
		Eui:       registeredGatewayEUI.Bytes(),
	}
	go func() {
		linkFn(ctx, t, ids, registeredGatewayKey, statCh)
	}()

	for _, tc := range []struct {
		Name               string
		Stat               *ttnpbv2.StatusMessage
		ExpectedAttributes map[string]string
	}{
		{
			Name: "FirstStat",
			Stat: &ttnpbv2.StatusMessage{
				Platform: "The Things Gateway v1 - BL r9-12345678 (2006-01-02T15:04:05Z) - Firmware v1.2.3-12345678 (2006-01-02T15:04:05Z)",
			},
			ExpectedAttributes: map[string]string{
				"model":    "The Things Kickstarter Gateway v1",
				"firmware": "v1.2.3-12345678",
			},
		},
		{
			Name: "SubsequentStatNoUpdate",
			Stat: &ttnpbv2.StatusMessage{
				Platform: "The Things Gateway v1 - BL r9-12345678 (2006-01-02T15:04:05Z) - Firmware v2.0.0-00000000 (2006-01-02T15:04:05Z)",
			},
			ExpectedAttributes: map[string]string{
				"model":    "The Things Kickstarter Gateway v1",
				"firmware": "v1.2.3-12345678",
			},
		},
	} {
		t.Run(fmt.Sprintf("UpdateVersionInfo/%s", tc.Name), func(t *testing.T) {
			select {
			case statCh <- tc.Stat:
			case <-time.After(timeout):
				t.Fatalf("Failed to send status message to upstream channel")
			}
			time.Sleep(timeout)
			gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: ids,
			})
			a.So(err, should.BeNil)
			a.So(gtw.Attributes, should.Resemble, tc.ExpectedAttributes)
		})
	}

	// Test Disconnection on delete.
	// Setup a stats client with independent context to query whether the gateway is connected and statistics on
	// upstream and downstream.
	statsConn, err := grpc.Dial(":9187", append(rpcclient.DefaultDialOptions(test.Context()), grpc.WithInsecure(), grpc.WithBlock())...)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer statsConn.Close()
	statsCtx := metadata.AppendToOutgoingContext(test.Context(),
		"id", ids.GatewayId,
		"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
	)
	statsClient := ttnpb.NewGsClient(statsConn)

	stat, err := statsClient.GetGatewayConnectionStats(statsCtx, gtwIDs)
	a.So(err, should.BeNil)
	a.So(stat, should.NotBeNil)

	// Delete and wait for fetch interval.
	is.GatewayRegistry().Delete(ctx, gtwIDs)
	time.Sleep(gatewayFetchInterval << 7)

	stat, err = statsClient.GetGatewayConnectionStats(statsCtx, gtwIDs)
	a.So(errors.IsNotFound(err), should.BeTrue)
	a.So(stat, should.BeNil)

	gs.Close()
	time.Sleep(timeout)
}

func TestBatchGetStatus(t *testing.T) {
	a, ctx := test.New(t)
	t.Parallel()

	for _, tc := range []struct { //nolint:paralleltest
		Name      string
		WithRedis bool
	}{
		{
			Name:      "Redis",
			WithRedis: true,
		},
		{
			Name: "NilRegistry",
		},
	} {
		t.Run(fmt.Sprintf("BatchGetStatus/%s", tc.Name), func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			is, isAddr, closeIS := mockis.New(ctx)
			defer closeIS()

			c := componenttest.NewComponent(t, &component.Config{
				ServiceBase: config.ServiceBase{
					GRPC: config.GRPC{
						Listen:                      ":9187",
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
			defer c.Close()

			gatewayFetchInterval := test.Delay

			gsConfig := &gatewayserver.Config{
				FetchGatewayInterval:   gatewayFetchInterval,
				FetchGatewayJitter:     1,
				UpdateVersionInfoDelay: test.Delay,
				MQTTV2: config.MQTT{
					Listen: ":1881",
				},
			}

			if tc.WithRedis && os.Getenv("TEST_REDIS") == "1" {
				statsRedisClient, statsFlush := test.NewRedis(ctx, "gatewayserver_test")
				defer statsFlush()
				defer statsRedisClient.Close()
				registry := &gsredis.GatewayConnectionStatsRegistry{
					Redis:   statsRedisClient,
					LockTTL: timeout,
				}
				if err := registry.Init(ctx); err != nil {
					t.Fatalf("Failed to setup stats registry :%v", err)
				}
				gsConfig.Stats = registry
			}

			er := gatewayserver.NewIS(c)
			gs, err := gatewayserver.New(c, gsConfig,
				gatewayserver.WithRegistry(er),
			)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to setup server :%v", err)
			}
			roles := gs.Roles()
			a.So(len(roles), should.Equal, 1)
			a.So(roles[0], should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER)
			a.So(err, should.BeNil)

			componenttest.StartComponent(t, c)
			time.Sleep(timeout) // Wait for component to start.

			mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

			linkFn := func(ctx context.Context, t *testing.T,
				ids *ttnpb.GatewayIdentifiers, key string, statCh <-chan *ttnpbv2.StatusMessage,
			) error {
				t.Helper()
				ctx, cancel := errorcontext.New(ctx)
				clientOpts := mqtt.NewClientOptions()
				clientOpts.AddBroker("tcp://0.0.0.0:1881")
				clientOpts.SetUsername(unique.ID(ctx, ids))
				clientOpts.SetPassword(key)
				clientOpts.SetAutoReconnect(false)
				clientOpts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
					cancel(err)
				})
				client := mqtt.NewClient(clientOpts)
				if token := client.Connect(); !token.WaitTimeout(timeout) {
					return context.DeadlineExceeded
				} else if err := token.Error(); err != nil {
					return err
				}
				defer client.Disconnect(uint(timeout / time.Millisecond))
				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case stat := <-statCh:
							buf, err := proto.Marshal(stat)
							if err != nil {
								cancel(err)
								return
							}
							if token := client.Publish(
								fmt.Sprintf("%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
								cancel(token.Error())
								return
							}
						}
					}
				}()
				<-ctx.Done()
				return ctx.Err()
			}

			gtwIDs1 := &ttnpb.GatewayIdentifiers{
				GatewayId: registeredGatewayID,
				Eui:       registeredGatewayEUI.Bytes(),
			}

			gtwIDs2 := &ttnpb.GatewayIdentifiers{
				GatewayId: "eui-aaee000000000001",
			}

			statsConn, err := grpc.Dial(":9187",
				append(rpcclient.DefaultDialOptions(test.Context()), grpc.WithInsecure(), grpc.WithBlock())...)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			defer statsConn.Close()
			statsClient := ttnpb.NewGsClient(statsConn)
			statsCtx := metadata.AppendToOutgoingContext(test.Context(),
				"id", gtwIDs1.GatewayId,
				"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
			)

			request := &ttnpb.BatchGetGatewayConnectionStatsRequest{
				GatewayIds: []*ttnpb.GatewayIdentifiers{
					gtwIDs1,
					gtwIDs2,
				},
			}

			// Get Stats before creation.
			res, err := statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.NotBeNil)
			a.So(res, should.BeNil)

			mockGtw1 := mockis.DefaultGateway(gtwIDs1, true, true)
			is.GatewayRegistry().Add(ctx, gtwIDs1, registeredGatewayKey, mockGtw1, testRights...)
			time.Sleep(timeout) // Wait for setup to be completed.

			mockGtw2 := mockis.DefaultGateway(gtwIDs2, true, true)
			is.GatewayRegistry().Add(ctx, gtwIDs2, registeredGatewayKey, mockGtw2, testRights...)
			time.Sleep(timeout) // Wait for setup to be completed.

			// Get Stats before connection.
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
			a.So(len(res.Entries), should.Equal, 0)

			// Connect first gateway.
			statCh1 := make(chan *ttnpbv2.StatusMessage)
			go func() {
				_ = linkFn(ctx, t, gtwIDs1, registeredGatewayKey, statCh1)
			}()
			time.Sleep(timeout) // Wait for connection to be completed.

			// Get Stats
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
			a.So(len(res.Entries), should.Equal, 1)
			a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)

			// Connect second gateway.
			ctxWithCancel, gtwConnCancel := context.WithCancel(ctx)
			statCh2 := make(chan *ttnpbv2.StatusMessage)
			go func() {
				_ = linkFn(ctxWithCancel, t, gtwIDs2, registeredGatewayKey, statCh2)
			}()
			time.Sleep(timeout) // Wait for connection to be completed.

			// Get Stats
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
			a.So(len(res.Entries), should.Equal, 2)
			a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)
			a.So(res.Entries[gtwIDs2.GatewayId], should.NotBeNil)

			// Disconnect second gateway.
			gtwConnCancel()
			time.Sleep(timeout) // Wait for connection to be closed.

			cfg, err := gs.GetConfig(ctx)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to get Gateway Server configuration :%v", err)
			}
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)

			if cfg.Stats != nil {
				// Only stats entries in the registry are persisted until Redis TTL after disconnection.
				// These entries will have the `disconnected_at` field set.
				a.So(len(res.Entries), should.Equal, 2)
				a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)
				a.So(res.Entries[gtwIDs2.GatewayId], should.NotBeNil)
			} else {
				a.So(len(res.Entries), should.Equal, 1)
				a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)
			}
			// Close the Gateway Server.
			gs.Close()
			time.Sleep(timeout)
		})
	}
}
