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
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	encoding "go.thethings.network/lorawan-stack/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	registeredGatewayID  = "test-gateway"
	registeredGatewayKey = "secret"
	registeredGatewayEUI = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	unregisteredGatewayID  = "eui-bbff000000000000"
	unregisteredGatewayEUI = types.EUI64{0xBB, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	timeout = (1 << 5) * test.Delay
)

func TestGatewayServer(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	is, isAddr := startMockIS(ctx)
	ns, nsAddr := startMockNS(ctx)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9187",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
		},
	})
	c.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)
	config := &gatewayserver.Config{
		RequireRegisteredGateways: false,
		MQTT: gatewayserver.MQTTConfig{
			Listen: ":1882",
		},
		UDP: gatewayserver.UDPConfig{
			Config: udp.Config{
				PacketHandlers:      2,
				PacketBuffer:        10,
				DownlinkPathExpires: 100 * time.Millisecond,
				ConnectionExpires:   250 * time.Millisecond,
				ScheduleLateTime:    0,
				AddrChangeBlock:     250 * time.Millisecond,
			},
			Listeners: map[string]string{
				":1700": test.EUFrequencyPlanID,
			},
		},
	}
	gs, err := gatewayserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

	test.Must(nil, c.Start())
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.PeerInfo_NETWORK_SERVER)
	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)
	is.add(ctx, ttnpb.GatewayIdentifiers{
		GatewayID: registeredGatewayID,
		EUI:       &registeredGatewayEUI,
	}, registeredGatewayKey)

	for _, ptc := range []struct {
		Protocol  string
		ValidAuth func(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string) bool
		Link      func(ctx context.Context, t *testing.T, ids ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error
	}{
		{
			Protocol: "grpc",
			ValidAuth: func(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string) bool {
				return ids.GatewayID == registeredGatewayID && key == registeredGatewayKey
			},
			Link: func(ctx context.Context, t *testing.T, ids ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
				conn, err := grpc.Dial(":9187", append(rpcclient.DefaultDialOptions(ctx), grpc.WithInsecure(), grpc.WithBlock())...)
				if err != nil {
					return err
				}
				defer conn.Close()
				md := rpcmetadata.MD{
					ID:            ids.GatewayID,
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
				errCh := make(chan error, 2)
				// Write upstream.
				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case msg := <-upCh:
							if err := link.Send(msg); err != nil {
								errCh <- err
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
							errCh <- err
							return
						}
						downCh <- msg
					}
				}()
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		},
		{
			Protocol: "mqtt",
			ValidAuth: func(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string) bool {
				return ids.GatewayID == registeredGatewayID && key == registeredGatewayKey
			},
			Link: func(ctx context.Context, t *testing.T, ids ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
				if ids.GatewayID == "" {
					t.SkipNow()
				}
				clientOpts := mqtt.NewClientOptions()
				clientOpts.AddBroker("tcp://0.0.0.0:1882")
				clientOpts.SetUsername(unique.ID(ctx, ids))
				clientOpts.SetPassword(key)
				client := mqtt.NewClient(clientOpts)
				if token := client.Connect(); !token.WaitTimeout(timeout) {
					return errors.New("connect timeout")
				} else if token.Error() != nil {
					return token.Error()
				}
				defer client.Disconnect(uint(timeout / time.Millisecond))
				errCh := make(chan error, 2)
				// Write upstream.
				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case up := <-upCh:
							for _, msg := range up.UplinkMessages {
								buf, err := msg.Marshal()
								if err != nil {
									errCh <- err
									return
								}
								if token := client.Publish(fmt.Sprintf("v3/%v/up", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
									errCh <- token.Error()
									return
								}
							}
							if up.GatewayStatus != nil {
								buf, err := up.GatewayStatus.Marshal()
								if err != nil {
									errCh <- err
									return
								}
								if token := client.Publish(fmt.Sprintf("v3/%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
									errCh <- token.Error()
									return
								}
							}
							if up.TxAcknowledgment != nil {
								buf, err := up.TxAcknowledgment.Marshal()
								if err != nil {
									errCh <- err
									return
								}
								if token := client.Publish(fmt.Sprintf("v3/%v/down/ack", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
									errCh <- token.Error()
									return
								}
							}
						}
					}
				}()
				// Read downstream.
				token := client.Subscribe(fmt.Sprintf("v3/%v/down", unique.ID(ctx, ids)), 1, func(_ mqtt.Client, raw mqtt.Message) {
					var msg ttnpb.GatewayDown
					if err := msg.Unmarshal(raw.Payload()); err != nil {
						errCh <- err
						return
					}
					downCh <- &msg
				})
				if token.Wait() && token.Error() != nil {
					return token.Error()
				}
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		},
		{
			Protocol: "udp",
			ValidAuth: func(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string) bool {
				return ids.EUI != nil
			},
			Link: func(ctx context.Context, t *testing.T, ids ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
				if ids.EUI == nil {
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
				errCh := make(chan error, 3)
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
								GatewayEUI:      ids.EUI,
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
								errCh <- err
								return
							}
							switch packet.PacketType {
							case encoding.PushData:
								if _, err := upConn.Write(writeBuf); err != nil {
									errCh <- err
									return
								}
								if _, err := upConn.Read(readBuf[:]); err != nil {
									errCh <- err
									return
								}
							case encoding.TxAck:
								if _, err := downConn.Write(writeBuf); err != nil {
									errCh <- err
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
								GatewayEUI:      ids.EUI,
								ProtocolVersion: encoding.Version1,
								Token:           [2]byte{0x01, token},
								PacketType:      encoding.PullData,
							}
							buf, err := pull.MarshalBinary()
							if err != nil {
								errCh <- err
								return
							}
							if _, err := downConn.Write(buf); err != nil {
								errCh <- err
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
							errCh <- err
							return
						}
						packetBuf := make([]byte, n)
						copy(packetBuf, buf[:])
						var packet encoding.Packet
						if err := packet.UnmarshalBinary(packetBuf); err != nil {
							errCh <- err
							return
						}
						switch packet.PacketType {
						case encoding.PullResp:
							msg, err := encoding.ToDownlinkMessage(packet.Data.TxPacket)
							if err != nil {
								errCh <- err
								return
							}
							downCh <- &ttnpb.GatewayDown{
								DownlinkMessage: msg,
							}
						}
					}
				}()
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					time.Sleep(config.UDP.ConnectionExpires * 150 / 100) // Ensure that connection expires.
					return ctx.Err()
				}
			},
		},
	} {
		t.Run(fmt.Sprintf("Authenticate/%v", ptc.Protocol), func(t *testing.T) {
			for _, ctc := range []struct {
				Name string
				ID   ttnpb.GatewayIdentifiers
				Key  string
			}{
				{
					Name: "ValidIDAndKey",
					ID:   ttnpb.GatewayIdentifiers{GatewayID: registeredGatewayID},
					Key:  registeredGatewayKey,
				},
				{
					Name: "InvalidKey",
					ID:   ttnpb.GatewayIdentifiers{GatewayID: registeredGatewayID},
					Key:  "invalid-key",
				},
				{
					Name: "InvalidIDAndKey",
					ID:   ttnpb.GatewayIdentifiers{GatewayID: "invalid-gateway"},
					Key:  "invalid-key",
				},
				{
					Name: "RegisteredEUI",
					ID:   ttnpb.GatewayIdentifiers{EUI: &registeredGatewayEUI},
				},
				{
					Name: "UnregisteredEUI",
					ID:   ttnpb.GatewayIdentifiers{EUI: &unregisteredGatewayEUI},
				},
			} {
				t.Run(ctc.Name, func(t *testing.T) {
					ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
					upCh := make(chan *ttnpb.GatewayUp)
					downCh := make(chan *ttnpb.GatewayDown)
					err := ptc.Link(ctx, t, ctc.ID, ctc.Key, upCh, downCh)
					cancel()
					if errors.IsDeadlineExceeded(err) {
						if !ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
							t.Fatal("Expected link error due to invalid auth")
						}
					} else if ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
						t.Fatalf("Expected deadline exceeded with valid auth, but have %v", err)
					}
				})
			}
		})

		t.Run(fmt.Sprintf("Traffic/%v", ptc.Protocol), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancel := context.WithCancel(ctx)
			upCh := make(chan *ttnpb.GatewayUp)
			downCh := make(chan *ttnpb.GatewayDown)
			ids := ttnpb.GatewayIdentifiers{
				GatewayID: registeredGatewayID,
				EUI:       &registeredGatewayEUI,
			}

			// Setup a stats client with independent context to query whether the gateway is connected and statistics on
			// upstream and downstream.
			statsConn, err := grpc.Dial(":9187", append(rpcclient.DefaultDialOptions(test.Context()), grpc.WithInsecure(), grpc.WithBlock())...)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			defer statsConn.Close()
			statsCtx := metadata.AppendToOutgoingContext(test.Context(),
				"id", ids.GatewayID,
				"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
			)
			statsClient := ttnpb.NewGsClient(statsConn)

			// The gateway should not be connected before testing traffic.
			t.Run("NotConnected", func(t *testing.T) {
				_, err := statsClient.GetGatewayConnectionStats(statsCtx, &ids)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatal("Expected gateway not to be connected yet, but it is")
				}
			})

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := ptc.Link(ctx, t, ids, registeredGatewayKey, upCh, downCh)
				if !errors.IsCanceled(err) {
					t.Fatalf("Expected context canceled, but have %v", err)
				}
			}()

			t.Run("Upstream", func(t *testing.T) {
				uplinkCount := 0
				upEvents := map[string]events.Channel{}
				for _, event := range []string{"gs.up.receive", "gs.down.tx.success", "gs.down.tx.fail", "gs.status.receive"} {
					ch := make(events.Channel, 5)
					events.Subscribe(event, ch)
					defer events.Unsubscribe(event, ch)
					upEvents[event] = ch
				}
				for _, tc := range []struct {
					Name     string
					Up       *ttnpb.GatewayUp
					Forwards []int // Indices of uplink messages in Up that are being forwarded.
				}{
					{
						Name: "GatewayStatus",
						Up: &ttnpb.GatewayUp{
							GatewayStatus: &ttnpb.GatewayStatus{
								Time: time.Unix(424242, 0),
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
						Name: "OneValidLoRa",
						Up: &ttnpb.GatewayUp{
							UplinkMessages: []*ttnpb.UplinkMessage{
								{
									Settings: ttnpb.TxSettings{
										DataRate: ttnpb.DataRate{
											Modulation: &ttnpb.DataRate_LoRa{
												LoRa: &ttnpb.LoRaDataRate{
													SpreadingFactor: 7,
													Bandwidth:       250000,
												},
											},
										},
										CodingRate: "4/5",
										Frequency:  867900000,
										Timestamp:  4242000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: ids,
											Timestamp:          4242000,
											RSSI:               -69,
											SNR:                11,
										},
									},
									RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
								},
							},
						},
						Forwards: []int{0},
					},
					{
						Name: "OneValidFSK",
						Up: &ttnpb.GatewayUp{
							UplinkMessages: []*ttnpb.UplinkMessage{
								{
									Settings: ttnpb.TxSettings{
										DataRate: ttnpb.DataRate{
											Modulation: &ttnpb.DataRate_FSK{
												FSK: &ttnpb.FSKDataRate{
													BitRate: 50000,
												},
											},
										},
										Frequency: 867900000,
										Timestamp: 4242000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: ids,
											Timestamp:          4242000,
											RSSI:               -69,
											SNR:                11,
										},
									},
									RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
								},
							},
						},
						Forwards: []int{0},
					},
					{
						Name: "OneGarbageWithStatus",
						Up: &ttnpb.GatewayUp{
							UplinkMessages: []*ttnpb.UplinkMessage{
								{
									Settings: ttnpb.TxSettings{
										DataRate: ttnpb.DataRate{
											Modulation: &ttnpb.DataRate_LoRa{
												LoRa: &ttnpb.LoRaDataRate{
													SpreadingFactor: 9,
													Bandwidth:       125000,
												},
											},
										},
										CodingRate: "4/5",
										Frequency:  868500000,
										Timestamp:  1234560000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: ids,
											Timestamp:          1234560000,
											RSSI:               -112,
											SNR:                2,
										},
									},
									RawPayload: []byte{0xff, 0x02, 0x03}, // Garbage; doesn't get forwarded.
								},
								{
									Settings: ttnpb.TxSettings{
										DataRate: ttnpb.DataRate{
											Modulation: &ttnpb.DataRate_LoRa{
												LoRa: &ttnpb.LoRaDataRate{
													SpreadingFactor: 7,
													Bandwidth:       125000,
												},
											},
										},
										CodingRate: "4/5",
										Frequency:  868100000,
										Timestamp:  4242000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: ids,
											Timestamp:          4242000,
											RSSI:               -69,
											SNR:                11,
										},
									},
									RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
								},
								{
									Settings: ttnpb.TxSettings{
										DataRate: ttnpb.DataRate{
											Modulation: &ttnpb.DataRate_LoRa{
												LoRa: &ttnpb.LoRaDataRate{
													SpreadingFactor: 12,
													Bandwidth:       125000,
												},
											},
										},
										CodingRate: "4/5",
										Frequency:  867700000,
										Timestamp:  2424000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: ids,
											Timestamp:          2424000,
											RSSI:               -36,
											SNR:                5,
										},
									},
									RawPayload: randomJoinRequestPayload(
										types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
										types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
									),
								},
							},
							GatewayStatus: &ttnpb.GatewayStatus{
								Time: time.Unix(4242424, 0),
							},
						},
						Forwards: []int{1, 2},
					},
				} {
					t.Run(tc.Name, func(t *testing.T) {
						a := assertions.New(t)

						select {
						case upCh <- tc.Up:
						case <-time.After(timeout):
							t.Fatalf("Failed to send message to upstream channel")
						}
						uplinkCount += len(tc.Up.UplinkMessages)

						for _, msgIdx := range tc.Forwards {
							select {
							case msg := <-ns.upCh:
								expected := tc.Up.UplinkMessages[msgIdx]
								a.So(time.Since(msg.ReceivedAt), should.BeLessThan, timeout)
								a.So(msg.Settings, should.Resemble, expected.Settings)
								for _, md := range msg.RxMetadata {
									a.So(md.UplinkToken, should.NotBeEmpty)
									md.UplinkToken = nil
								}
								a.So(msg.RxMetadata, should.Resemble, expected.RxMetadata)
								a.So(msg.RawPayload, should.Resemble, expected.RawPayload)
							case <-time.After(timeout):
								t.Fatal("Expected uplink timeout")
							}
							select {
							case evt := <-upEvents["gs.up.receive"]:
								a.So(evt.Name(), should.Equal, "gs.up.receive")
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
						}
						if tc.Up.GatewayStatus != nil {
							select {
							case <-upEvents["gs.status.receive"]:
							case <-time.After(timeout):
								t.Fatal("Expected gateway status event timeout")
							}
						}

						// Wait for gateway status to be processed; no event available here.
						time.Sleep(timeout)

						stats, err := statsClient.GetGatewayConnectionStats(statsCtx, &ids)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						a.So(stats.UplinkCount, should.Equal, uplinkCount)
						if tc.Up.GatewayStatus != nil {
							if !a.So(stats.LastStatus, should.NotBeNil) {
								t.FailNow()
							}
							a.So(stats.LastStatus.Time, should.Equal, tc.Up.GatewayStatus.Time)
						}
					})
				}
			})

			t.Run("Downstream", func(t *testing.T) {
				ctx := cluster.NewContext(test.Context(), nil)
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
									DataRate: ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_LoRa{
											LoRa: &ttnpb.LoRaDataRate{
												SpreadingFactor: 12,
												Bandwidth:       125000,
											},
										},
									},
									CodingRate: "4/5",
									Frequency:  869525000,
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
									Class: ttnpb.CLASS_C,
									DownlinkPaths: []*ttnpb.DownlinkPath{
										{
											Path: &ttnpb.DownlinkPath_Fixed{
												Fixed: &ttnpb.GatewayAntennaIdentifiers{
													GatewayIdentifiers: ttnpb.GatewayIdentifiers{
														GatewayID: "not-connected",
													},
												},
											},
										},
									},
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
									Class: ttnpb.CLASS_A,
									DownlinkPaths: []*ttnpb.DownlinkPath{
										{
											Path: &ttnpb.DownlinkPath_UplinkToken{
												UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{
													GatewayIdentifiers: ttnpb.GatewayIdentifiers{
														GatewayID: registeredGatewayID,
													},
												}, 100),
											},
										},
									},
									Priority:         ttnpb.TxSchedulePriority_NORMAL,
									Rx1Delay:         ttnpb.RX_DELAY_1,
									Rx1DataRateIndex: 5,
									Rx1Frequency:     868100000,
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
									Class: ttnpb.CLASS_A,
									DownlinkPaths: []*ttnpb.DownlinkPath{
										{
											Path: &ttnpb.DownlinkPath_UplinkToken{
												UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{
													GatewayIdentifiers: ttnpb.GatewayIdentifiers{
														GatewayID: registeredGatewayID,
													},
												}, 100),
											},
										},
									},
									Priority:         ttnpb.TxSchedulePriority_NORMAL,
									Rx1Delay:         ttnpb.RX_DELAY_1,
									Rx1DataRateIndex: 5,
									Rx1Frequency:     868100000,
								},
							},
						},
						ErrorAssertion: errors.IsAborted,
						RxWindowDetailsAssertion: []func(error) bool{
							errors.IsResourceExhausted,  // Rx1 conflicts with previous.
							errors.IsFailedPrecondition, // Rx2 not provided.
						},
					},
					{
						Name: "ValidClassC",
						Message: &ttnpb.DownlinkMessage{
							RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
							Settings: &ttnpb.DownlinkMessage_Request{
								Request: &ttnpb.TxRequest{
									Class: ttnpb.CLASS_C,
									DownlinkPaths: []*ttnpb.DownlinkPath{
										{
											Path: &ttnpb.DownlinkPath_Fixed{
												Fixed: &ttnpb.GatewayAntennaIdentifiers{
													GatewayIdentifiers: ttnpb.GatewayIdentifiers{
														GatewayID: registeredGatewayID,
													},
												},
											},
										},
									},
									Priority:         ttnpb.TxSchedulePriority_NORMAL,
									Rx1Delay:         ttnpb.RX_DELAY_1,
									Rx1DataRateIndex: 5,
									Rx1Frequency:     868100000,
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
								a.So(errors.Details(err), should.HaveLength, 1)
								errSchedulePathCause := errors.Cause(errors.Details(err)[0].(error))
								a.So(errors.IsAborted(errSchedulePathCause), should.BeTrue)
								for i, assert := range tc.RxWindowDetailsAssertion {
									if i >= len(errors.Details(errSchedulePathCause)) {
										t.Fatalf("Expected error in Rx window %d", i+1)
									}
									errRxWindow := errors.Details(errSchedulePathCause)[i].(error)
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

						stats, err := statsClient.GetGatewayConnectionStats(statsCtx, &ids)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						a.So(stats.DownlinkCount, should.Equal, downlinkCount)
					})
				}
			})

			cancel()
			wg.Wait()

			// Wait for disconnection to be processed.
			time.Sleep(timeout)

			// After canceling the context and awaiting the link, the connection should be gone.
			t.Run("Disconnected", func(t *testing.T) {
				_, err := statsClient.GetGatewayConnectionStats(statsCtx, &ids)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Expected gateway to be disconnected, but it's not")
				}
			})
		})
	}
}
