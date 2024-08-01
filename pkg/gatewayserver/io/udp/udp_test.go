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

package udp_test

import (
	"context"
	"encoding/base64"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/iotest"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestConnection(t *testing.T) {
	var (
		timeout    = (1 << 4) * test.Delay
		testConfig = Config{
			PacketHandlers:      2,
			PacketBuffer:        10,
			DownlinkPathExpires: 8 * timeout,
			ConnectionExpires:   20 * timeout,
			ScheduleLateTime:    0,
		}
	)

	a, ctx := test.New(t)
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, _, closeIS := mockis.New(ctx)
	defer closeIS()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()

	gs := mock.NewServer(c, is)
	addr, _ := net.ResolveUDPAddr("udp", ":0")
	lis, err := net.ListenUDP("udp", addr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	go Serve(ctx, gs, lis, testConfig)

	connections := &sync.Map{}
	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}

	conn, err := net.Dial("udp", lis.LocalAddr().String())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	for i, tc := range []struct {
		Name              string
		PacketType        encoding.PacketType
		Wait              time.Duration
		Connects          bool
		LosesDownlinkPath bool
		Disconnects       bool
	}{
		{
			Name:              "NewConnectionOnPush",
			PacketType:        encoding.PushData,
			Wait:              0,
			Connects:          true,
			LosesDownlinkPath: true,
			Disconnects:       false,
		},
		{
			Name:              "ExistingConnectionOnPull",
			PacketType:        encoding.PullData,
			Wait:              0,
			Connects:          false,
			LosesDownlinkPath: false,
			Disconnects:       false,
		},
		{
			Name:              "LoseDownlinkPath",
			PacketType:        encoding.PullData,
			Wait:              testConfig.DownlinkPathExpires * 150 / 100,
			Connects:          false,
			LosesDownlinkPath: true,
			Disconnects:       false,
		},
		{
			Name:              "RecoverDownlinkPathWithoutReconnect",
			PacketType:        encoding.PullData,
			Wait:              0,
			Connects:          false,
			LosesDownlinkPath: false,
			Disconnects:       false,
		},
		{
			Name:              "LoseConnection",
			PacketType:        encoding.PullData,
			Wait:              testConfig.ConnectionExpires * 150 / 100,
			Connects:          false,
			LosesDownlinkPath: true,
			Disconnects:       true,
		},
		{
			Name:              "Reconnect",
			PacketType:        encoding.PullData,
			Wait:              0,
			Connects:          true,
			LosesDownlinkPath: false,
			Disconnects:       false,
		},
	} {
		tcok := t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			// Send packet.
			var packet encoding.Packet
			var ackType encoding.PacketType
			if tc.PacketType == encoding.PushData {
				packet = generatePushData(eui, false, 0)
				ackType = encoding.PushAck
			} else {
				packet = generatePullData(eui)
				ackType = encoding.PullAck
			}
			packet.Token[1] = byte(i)
			buf, err := packet.MarshalBinary()
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			_, err = conn.Write(buf)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			expectAck(t, conn, true, ackType, packet.Token)

			// Optionally wait to lose downlink claim or connection expiry.
			time.Sleep(tc.Wait)

			// Assert disconnects.
			if tc.Disconnects {
				_, connected := connections.Load(eui)
				a.So(connected, should.BeFalse)
				return
			}

			// Asserts new or existing connection.
			conn := expectConnection(t, gs, connections, eui, tc.Connects)

			// Assert claim, give some time.
			<-time.After(timeout)
			hasClaim := gs.HasDownlinkClaim(ctx, conn.Gateway().GetIds())
			if tc.LosesDownlinkPath {
				a.So(hasClaim, should.BeFalse)
			} else {
				a.So(hasClaim, should.BeTrue)
			}
		})
		if !tcok {
			t.FailNow()
		}
	}

	cancelCtx()
}

func TestFrontend(t *testing.T) {
	t.Parallel()
	iotest.Frontend(t, iotest.FrontendConfig{
		DropsCRCFailure:      false,
		DropsInvalidLoRaWAN:  false,
		SupportsStatus:       true,
		DetectsDisconnect:    false,
		AuthenticatesWithEUI: true,
		IsAuthenticated:      false,
		DeduplicatesUplinks:  true,
		CustomRxMetadataAssertion: func(t *testing.T, actual, expected *ttnpb.RxMetadata) {
			t.Helper()
			a := assertions.New(t)
			a.So(actual.UplinkToken, should.NotBeEmpty)
			actual.UplinkToken = nil
			actual.ReceivedAt = nil
			expected.SignalRssi = nil
			a.So(actual, should.Resemble, expected)
		},
		CustomGatewayServerConfig: func(config *gatewayserver.Config) {
			config.UDP = gatewayserver.UDPConfig{
				Config: udp.Config{
					PacketHandlers:      2,
					PacketBuffer:        10,
					DownlinkPathExpires: (1 << 7) * test.Delay,
					ConnectionExpires:   (1 << 8) * test.Delay,
					ScheduleLateTime:    0,
					AddrChangeBlock:     (1 << 7) * test.Delay,
				},
				Listeners: map[string]string{
					":1700": test.EUFrequencyPlanID,
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
			// Engage downstream by sending PULL_DATA.
			go func() {
				var token byte
				initial := make(chan struct{}, 1)
				initial <- struct{}{}
				ticker := time.NewTicker((1 << 6) * test.Delay)
				for {
					select {
					case <-ctx.Done():
						ticker.Stop()
						return
					case <-ticker.C:
					case <-initial:
					}
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
			time.Sleep((1 << 9) * test.Delay) // Ensure that connection expires.
			return ctx.Err()
		},
	})
}

// TestRawData tests the raw data input and output of the UDP frontend.
// This includes garbage data, connection state, and the timing of downlink scheduling.
// This test is complementary to the generic TestFrontend.
func TestRawData(t *testing.T) {
	var (
		registeredGatewayID = ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
		timeout             = (1 << 4) * test.Delay
		testConfig          = Config{
			PacketHandlers:      2,
			PacketBuffer:        10,
			DownlinkPathExpires: 8 * timeout,
			ConnectionExpires:   20 * timeout,
			ScheduleLateTime:    0,
		}
	)

	a, ctx := test.New(t)
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, _, closeIS := mockis.New(ctx)
	defer closeIS()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()

	gs := mock.NewServer(c, is)
	addr, _ := net.ResolveUDPAddr("udp", ":0")
	lis, err := net.ListenUDP("udp", addr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	go Serve(ctx, gs, lis, testConfig)

	connections := &sync.Map{}
	eui1 := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	eui2 := types.EUI64{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
	eui3 := types.EUI64{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03}

	t.Run("Upstream", func(t *testing.T) {
		udpConn, err := net.Dial("udp", lis.LocalAddr().String())
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		for i, tc := range []struct {
			Name          string
			EUI           types.EUI64
			Raw           []byte
			Packet        encoding.Packet // Raw takes priority over Packet
			AckOK         bool
			ExpectConnect bool
		}{
			{
				Name:          "EOF",
				EUI:           eui1,
				Raw:           []byte{0x01, 0x02},
				AckOK:         false,
				ExpectConnect: false,
			},
			{
				Name:          "EOF",
				EUI:           eui1,
				Raw:           []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01},
				AckOK:         false,
				ExpectConnect: false,
			},
			{
				Name:          "InvalidPacketType",
				EUI:           eui1,
				Raw:           []byte{0x01, 0x00, 0x00, 0x01},
				AckOK:         false,
				ExpectConnect: false,
			},
			{
				Name:          "ValidNewConnection",
				EUI:           eui1,
				Packet:        generatePushData(eui1, true, 100*time.Microsecond),
				AckOK:         true,
				ExpectConnect: true,
			},
			{
				Name:          "ValidExistingConnection",
				EUI:           eui1,
				Packet:        generatePushData(eui1, false, 200*time.Microsecond, 300*time.Microsecond),
				AckOK:         true,
				ExpectConnect: false,
			},
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				var buf []byte
				if tc.Raw != nil {
					buf = tc.Raw
				} else {
					var err error
					buf, err = tc.Packet.MarshalBinary()
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}

				// Put unique token, write and expect acknowledgment.
				token := [2]byte{0x00, byte(i)}
				if len(buf) >= 4 {
					copy(buf[1:], token[:])
				}
				_, err := udpConn.Write(buf)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				expectAck(t, udpConn, tc.AckOK, encoding.PushAck, token)
				if !tc.AckOK {
					t.SkipNow()
				}

				// Expect a new connection or an existing.
				conn := expectConnection(t, gs, connections, tc.EUI, tc.ExpectConnect)

				// Expect upstream data.
				for _, p := range tc.Packet.Data.RxPacket {
					select {
					case up := <-conn.Up():
						data, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(p.Data, "="))
						a.So(err, should.BeNil)
						a.So(up.Message.RawPayload, should.Resemble, data)
					case <-time.After(timeout):
						t.Fatal("Receive expected uplink timeout")
					}
				}
				if tc.Packet.Data.Stat != nil {
					select {
					case <-conn.Status():
					case <-time.After(timeout):
						t.Fatal("Receive expected status timeout")
					}
				}
			})
			if !tcok {
				t.FailNow()
			}
		}
	})

	t.Run("Downstream", func(t *testing.T) {
		udpConn, err := net.Dial("udp", lis.LocalAddr().String())
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		for i, tc := range []struct {
			Name                 string
			EUI                  types.EUI64
			Packet               encoding.Packet
			AckOK                bool
			ExpectConnect        bool
			SyncClock            time.Duration
			Path                 *ttnpb.DownlinkPath
			Message              *ttnpb.DownlinkMessage
			ScheduleDownlinkLate bool
			ScheduledLate        bool
			SendTxAck            bool
		}{
			{
				Name:          "ValidExistingConnection",
				EUI:           eui1,
				Packet:        generatePullData(eui1),
				AckOK:         true,
				ExpectConnect: false,
			},
			{
				Name:          "ValidNewConnection",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: true,
			},
			{
				Name:          "TxScheduledLate",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     1 * time.Second, // Rx1 delay
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
							uint32(300*test.Delay/time.Microsecond),
							scheduling.ConcentratorTime(300*test.Delay),
							time.Unix(0, int64(300*test.Delay)),
							nil,
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
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
				},
				ScheduleDownlinkLate: false,
				ScheduledLate:        true, // Because Tx acknowledgment is not received.
				SendTxAck:            false,
			},
			{
				Name:          "TxPreferLateOK",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     1*time.Second + 300*test.Delay, // Rx1 delay + start time
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
							uint32(600*test.Delay/time.Microsecond),
							scheduling.ConcentratorTime(600*test.Delay),
							time.Unix(0, int64(600*test.Delay)),
							nil,
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x03},
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
				},
				ScheduleDownlinkLate: true,
				ScheduledLate:        true, // Because Tx acknowledgment is not received.
				SendTxAck:            true, // From now on, immediate scheduling takes priority over scheduling late preference.
			},
			{
				Name:          "TxPreferLateOverruled",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     15*time.Second + 1*time.Second, // Start time + Rx1 delay.
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
							uint32((15*time.Second+300*test.Delay)/time.Microsecond),
							scheduling.ConcentratorTime(15*time.Second+300*test.Delay),
							time.Unix(0, int64((15*time.Second+300*test.Delay))),
							nil,
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x04},
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
				},
				ScheduleDownlinkLate: true,
				ScheduledLate:        true, // Should be scheduled late as it's forced by ScheduleDownlinkLate.
				SendTxAck:            true,
			},
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)
				buf, err := tc.Packet.MarshalBinary()
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				// Put unique token, write and expect acknowledgment.
				token := [2]byte{0x00, byte(i)}
				copy(buf[1:], token[:])
				_, err = udpConn.Write(buf)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				expectAck(t, udpConn, tc.AckOK, encoding.PullAck, token)
				if !tc.AckOK {
					t.SkipNow()
				}

				// Expect a new connection or an existing.
				conn := expectConnection(t, gs, connections, tc.EUI, tc.ExpectConnect)

				if tc.Message == nil {
					t.SkipNow()
				}

				// Sync the clock at the given time.
				var clockSynced time.Time
				packet := generatePushData(eui2, false, tc.SyncClock)
				buf, err = packet.MarshalBinary()
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				token = [2]byte{0x01, byte(i)}
				copy(buf[1:], token[:])
				_, err = udpConn.Write(buf)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				clockSynced = time.Now()
				expectAck(t, udpConn, true, encoding.PushAck, token)
				time.Sleep(timeout) // Ensure that clock gets actually synced.

				// Send the downlink message, optionally buffer first.
				conn.Gateway().ScheduleDownlinkLate = tc.ScheduleDownlinkLate
				_, _, _, err = conn.ScheduleDown(tc.Path, tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				// Set expected time for the pull response.
				expectedTime := clockSynced
				if tc.ScheduledLate {
					expectedTime = expectedTime.Add(-tc.SyncClock)
					expectedTime = expectedTime.Add(time.Duration(tc.Message.GetScheduled().Timestamp) * time.Microsecond)
					expectedTime = expectedTime.Add(-testConfig.ScheduleLateTime)
				}

				// Read the response, taking care of expected time.
				var respBuf [65507]byte
				udpConn.SetReadDeadline(expectedTime.Add(timeout))
				n, err := udpConn.Read(respBuf[:])
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				actualTime := time.Now()
				var response encoding.Packet
				if err = response.UnmarshalBinary(respBuf[:n]); !a.So(err, should.BeNil) {
					t.FailNow()
				}

				// Assert packet type, content and time of arrival.
				a.So(response.PacketType, should.Equal, encoding.PullResp)
				expected, err := encoding.FromDownlinkMessage(tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(response.Data.TxPacket, should.Resemble, expected)
				a.So(actualTime, should.HappenOnOrBetween, expectedTime, expectedTime.Add(timeout))

				// Send TxAck.
				if tc.SendTxAck {
					packet := generateTxAck(eui2, encoding.TxErrNone)
					buf, err = packet.MarshalBinary()
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					token := [2]byte{0x02, byte(i)}
					copy(buf[1:], token[:])
					_, err = udpConn.Write(buf)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}
			})
			if !tcok {
				t.FailNow()
			}
		}
	})

	t.Run("TxAcknowledgment", func(t *testing.T) {
		a := assertions.New(t)

		udpConn, err := net.Dial("udp", lis.LocalAddr().String())
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		packet := generateTxAck(eui3, encoding.TxErrNone)
		buf, err := packet.MarshalBinary()
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to marshal Tx acknowledgment: %v", err)
		}
		token := [2]byte{0x00, 0xae}
		copy(buf[1:], token[:])
		_, err = udpConn.Write(buf)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to write Tx acknowledgment: %v", err)
		}

		conn := expectConnection(t, gs, connections, eui3, true)
		select {
		case <-conn.TxAck():
		case <-time.After(timeout):
			t.Fatal("Receive expected TxAck timeout")
		}
	})

	cancelCtx()
}
