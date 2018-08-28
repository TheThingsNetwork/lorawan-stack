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

package gatewayserver_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp/encoding"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	registeredGatewayID  = "eui-aaee000000000000"
	registeredGatewayKey = "secret"
	registeredGatewayEUI = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	unregisteredGatewayID  = "eui-bbff000000000000"
	unregisteredGatewayEUI = types.EUI64{0xBB, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	timeout = 10 * time.Millisecond
)

func TestGatewayServer(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	is, isAddr := startMockIS(ctx)
	ns, nsAddr := startMockNS(ctx)

	band, err := band.GetByID("EU_863_870")
	if err != nil {
		panic(err)
	}
	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen: ":9184",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
		},
	})
	c.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
	config := gatewayserver.Config{
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
	mustHavePeer(ctx, c, ttnpb.PeerInfo_IDENTITY_SERVER)
	is.add(ctx, ttnpb.GatewayIdentifiers{
		GatewayID: registeredGatewayID,
		EUI:       &registeredGatewayEUI,
	}, registeredGatewayKey)

	for _, ptc := range []struct {
		Protocol  string
		ValidAuth func(ctx context.Context, id ttnpb.GatewayIdentifiers, key string) bool
		Link      func(ctx context.Context, t *testing.T, id ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error
	}{
		{
			Protocol: "grpc",
			ValidAuth: func(ctx context.Context, id ttnpb.GatewayIdentifiers, key string) bool {
				return id.GatewayID == registeredGatewayID && key == registeredGatewayKey
			},
			Link: func(ctx context.Context, t *testing.T, id ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
				a := assertions.New(t)
				conn, err := grpc.Dial(":9184", grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					return err
				}
				defer conn.Close()
				ctx = metadata.AppendToOutgoingContext(ctx,
					"id", id.GatewayID,
					"authorization", fmt.Sprintf("Key %v", key),
				)
				client := ttnpb.NewGtwGsClient(conn)
				fp, err := client.GetFrequencyPlan(ctx, &ttnpb.GetFrequencyPlanRequest{FrequencyPlanID: test.EUFrequencyPlanID})
				if err != nil {
					return err
				}
				a.So(fp.BandID, should.Equal, "EU_863_870")
				link, err := client.LinkGateway(ctx)
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
			ValidAuth: func(ctx context.Context, id ttnpb.GatewayIdentifiers, key string) bool {
				return id.GatewayID == registeredGatewayID && key == registeredGatewayKey
			},
			Link: func(ctx context.Context, t *testing.T, id ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
				if id.GatewayID == "" {
					t.SkipNow()
				}
				clientOpts := mqtt.NewClientOptions()
				clientOpts.AddBroker("tcp://0.0.0.0:1882")
				clientOpts.SetUsername(unique.ID(ctx, id))
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
								if token := client.Publish(fmt.Sprintf("v3/%v/up", id.GatewayID), 1, false, buf); token.Wait() && token.Error() != nil {
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
								if token := client.Publish(fmt.Sprintf("v3/%v/status", id.GatewayID), 1, false, buf); token.Wait() && token.Error() != nil {
									errCh <- token.Error()
									return
								}
							}
						}
					}
				}()
				// Read downstream.
				token := client.Subscribe(fmt.Sprintf("v3/%v/down", id.GatewayID), 1, func(_ mqtt.Client, raw mqtt.Message) {
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
			ValidAuth: func(ctx context.Context, id ttnpb.GatewayIdentifiers, key string) bool {
				return id.EUI != nil
			},
			Link: func(ctx context.Context, t *testing.T, id ttnpb.GatewayIdentifiers, key string, upCh <-chan *ttnpb.GatewayUp, downCh chan<- *ttnpb.GatewayDown) error {
				if id.EUI == nil {
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
								GatewayEUI:      id.EUI,
								ProtocolVersion: encoding.Version1,
								Token:           [2]byte{0x00, token},
								PacketType:      encoding.PushData,
								Data:            &encoding.Data{},
							}
							for _, msg := range up.UplinkMessages {
								var modulation string
								var dataRate types.DataRate
								switch msg.Settings.Modulation {
								case ttnpb.Modulation_LORA:
									modulation = "LORA"
									dataRate.LoRa = fmt.Sprintf("SF%dBW%d", msg.Settings.SpreadingFactor, msg.Settings.Bandwidth/1000)
								case ttnpb.Modulation_FSK:
									modulation = "FSK"
									dataRate.FSK = msg.Settings.BitRate
								}
								packet.Data.RxPacket = append(packet.Data.RxPacket, &encoding.RxPacket{
									Freq: float64(msg.Settings.Frequency) / 1000000,
									Chan: uint8(msg.Settings.ChannelIndex),
									Modu: modulation,
									DatR: encoding.DataRate{DataRate: dataRate},
									CodR: msg.Settings.CodingRate,
									Size: uint16(len(msg.RawPayload)),
									Data: base64.StdEncoding.EncodeToString(msg.RawPayload),
									Tmst: uint32(msg.RxMetadata[0].Timestamp / 1000),
									RSSI: int16(msg.RxMetadata[0].RSSI),
									LSNR: float64(msg.RxMetadata[0].SNR),
								})
							}
							if up.GatewayStatus != nil {
								packet.Data.Stat = &encoding.Stat{
									Time: encoding.ExpandedTime(up.GatewayStatus.Time),
								}
							}
							writeBuf, err := packet.MarshalBinary()
							if err != nil {
								errCh <- err
								return
							}
							if _, err := upConn.Write(writeBuf); err != nil {
								errCh <- err
								return
							}
							if _, err := upConn.Read(readBuf[:]); err != nil {
								errCh <- err
								return
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
								GatewayEUI:      id.EUI,
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
							msg := ttnpb.DownlinkMessage{
								Settings: ttnpb.TxSettings{
									CodingRate:            packet.Data.TxPacket.CodR,
									Frequency:             uint64(packet.Data.TxPacket.Freq * 1000000),
									PolarizationInversion: packet.Data.TxPacket.IPol,
									TxPower:               int32(packet.Data.TxPacket.Powe),
								},
								TxMetadata: ttnpb.TxMetadata{
									Timestamp: uint64(packet.Data.TxPacket.Tmst * 1000),
								},
							}
							msg.RawPayload, _ = base64.RawStdEncoding.DecodeString(strings.TrimRight(packet.Data.TxPacket.Data, "="))
							switch packet.Data.TxPacket.Modu {
							case "LORA":
								msg.Settings.Modulation = ttnpb.Modulation_LORA
								msg.TxMetadata.EnableCRC = !packet.Data.TxPacket.NCRC
								for i := len(band.DataRates) - 1; i >= 0; i-- {
									dr := band.DataRates[i].Rate
									if dr == packet.Data.TxPacket.DatR.DataRate {
										if sf, _ := dr.SpreadingFactor(); err == nil {
											msg.Settings.SpreadingFactor = uint32(sf)
										}
										if bw, _ := dr.Bandwidth(); err == nil {
											msg.Settings.Bandwidth = bw
										}
										break
									}
								}
							case "FSK":
								msg.Settings.Modulation = ttnpb.Modulation_FSK
								msg.Settings.BitRate = packet.Data.TxPacket.DatR.FSK
							}
							downCh <- &ttnpb.GatewayDown{
								DownlinkMessage: &msg,
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
			id := ttnpb.GatewayIdentifiers{
				GatewayID: registeredGatewayID,
				EUI:       &registeredGatewayEUI,
			}

			// Setup a stats client with independent context to query whether the gateway is connected and statistics on
			// upstream and downstream.
			statsConn, err := grpc.Dial(":9184", grpc.WithInsecure(), grpc.WithBlock())
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			defer statsConn.Close()
			statsCtx := metadata.AppendToOutgoingContext(test.Context(),
				"id", id.GatewayID,
				"authorization", fmt.Sprintf("Key %v", registeredGatewayKey),
			)
			statsClient := ttnpb.NewGsClient(statsConn)

			// The gateway should not be connected before testing traffic.
			t.Run("NotConnected", func(t *testing.T) {
				_, err := statsClient.GetGatewayConnectionStats(statsCtx, &id)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatal("Expected gateway not to be connected yet, but it is")
				}
			})

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := ptc.Link(ctx, t, id, registeredGatewayKey, upCh, downCh)
				if !errors.IsCanceled(err) {
					t.Fatalf("Expected context canceled, but have %v", err)
				}
			}()

			t.Run("Upstream", func(t *testing.T) {
				uplinkCount := 0
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
						Name: "OneValidLoRa",
						Up: &ttnpb.GatewayUp{
							UplinkMessages: []*ttnpb.UplinkMessage{
								{
									Settings: ttnpb.TxSettings{
										Modulation:      ttnpb.Modulation_LORA,
										SpreadingFactor: 7,
										Bandwidth:       250000,
										CodingRate:      "4/5",
										Frequency:       867900000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: id,
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
										Modulation: ttnpb.Modulation_FSK,
										BitRate:    50000,
										Frequency:  867900000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: id,
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
										Modulation:      ttnpb.Modulation_LORA,
										SpreadingFactor: 9,
										Bandwidth:       125000,
										CodingRate:      "4/5",
										Frequency:       868500000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: id,
											Timestamp:          1234560000,
											RSSI:               -112,
											SNR:                2,
										},
									},
									RawPayload: []byte{0xff, 0x02, 0x03}, // Garbage; doesn't get forwarded.
								},
								{
									Settings: ttnpb.TxSettings{
										Modulation:      ttnpb.Modulation_LORA,
										SpreadingFactor: 7,
										Bandwidth:       125000,
										CodingRate:      "4/5",
										Frequency:       868100000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: id,
											Timestamp:          4242000,
											RSSI:               -69,
											SNR:                11,
										},
									},
									RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
								},
								{
									Settings: ttnpb.TxSettings{
										Modulation:      ttnpb.Modulation_LORA,
										SpreadingFactor: 12,
										Bandwidth:       125000,
										CodingRate:      "4/5",
										Frequency:       867700000,
									},
									RxMetadata: []*ttnpb.RxMetadata{
										{
											GatewayIdentifiers: id,
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

						upCh <- tc.Up
						uplinkCount += len(tc.Up.UplinkMessages)

						for _, msgIdx := range tc.Forwards {
							select {
							case msg := <-ns.upCh:
								expected := tc.Up.UplinkMessages[msgIdx]
								a.So(msg.Settings, should.Resemble, expected.Settings)
								a.So(msg.RxMetadata, should.Resemble, expected.RxMetadata)
								a.So(msg.RawPayload, should.Resemble, expected.RawPayload)
							case <-time.After(timeout):
								t.Fatal("Expected uplink timeout")
							}
						}

						// Wait for gateway status to be processed; no event available here.
						time.Sleep(50 * time.Millisecond)

						stats, err := statsClient.GetGatewayConnectionStats(statsCtx, &id)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						a.So(stats.UplinkCount, should.Equal, uplinkCount)
						if tc.Up.GatewayStatus != nil {
							a.So(stats.LastStatus.Time, should.Equal, tc.Up.GatewayStatus.Time)
						}
					})
				}
			})

			t.Run("Downstream", func(t *testing.T) {
				ctx := cluster.NewContext(test.Context(), nil)
				downlinkCount := 0
				for _, tc := range []struct {
					Name       string
					Message    *ttnpb.DownlinkMessage
					ValidError func(error) bool
				}{
					{
						Name: "NotConnected",
						Message: &ttnpb.DownlinkMessage{
							TxMetadata: ttnpb.TxMetadata{
								GatewayIdentifiers: ttnpb.GatewayIdentifiers{
									GatewayID: "not-connected",
								},
							},
						},
						ValidError: errors.IsNotFound,
					},
					{
						Name: "Valid",
						Message: &ttnpb.DownlinkMessage{
							RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
							Settings: ttnpb.TxSettings{
								Modulation:      ttnpb.Modulation_LORA,
								SpreadingFactor: 12,
								Bandwidth:       125000,
								CodingRate:      "4/5",
								Frequency:       869525000,
							},
							TxMetadata: ttnpb.TxMetadata{
								GatewayIdentifiers: id,
								Timestamp:          0,
							},
						},
					},
					{
						Name: "Conflict",
						Message: &ttnpb.DownlinkMessage{
							RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
							Settings: ttnpb.TxSettings{
								Modulation:      ttnpb.Modulation_LORA,
								SpreadingFactor: 10,
								Bandwidth:       125000,
								CodingRate:      "4/5",
								Frequency:       869525000,
							},
							TxMetadata: ttnpb.TxMetadata{
								GatewayIdentifiers: id,
								Timestamp:          0,
							},
						},
						ValidError: errors.IsResourceExhausted,
					},
				} {
					t.Run(tc.Name, func(t *testing.T) {
						a := assertions.New(t)

						_, err := gs.ScheduleDownlink(ctx, tc.Message)
						if tc.ValidError != nil {
							if !a.So(tc.ValidError(err), should.BeTrue) {
								t.Fatalf("Expected a different error, but have %v", err)
							}
							return
						} else if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						downlinkCount++

						select {
						case msg := <-downCh:
							msg.DownlinkMessage.TxMetadata.GatewayIdentifiers = id
							a.So(msg.DownlinkMessage.Settings, should.Resemble, tc.Message.Settings)
							a.So(msg.DownlinkMessage.TxMetadata, should.Resemble, tc.Message.TxMetadata)
							a.So(msg.DownlinkMessage.RawPayload, should.Resemble, tc.Message.RawPayload)
						case <-time.After(timeout):
							t.Fatal("Expected downlink timeout")
						}

						stats, err := statsClient.GetGatewayConnectionStats(statsCtx, &id)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						a.So(stats.DownlinkCount, should.Equal, downlinkCount)
					})
				}
			})

			cancel()
			wg.Wait()

			// After canceling the context and awaiting the link, the connection should be gone.
			t.Run("Disconnected", func(t *testing.T) {
				_, err := statsClient.GetGatewayConnectionStats(statsCtx, &id)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Expected gateway to be disconnected, but it's not")
				}
			})
		})
	}
}
