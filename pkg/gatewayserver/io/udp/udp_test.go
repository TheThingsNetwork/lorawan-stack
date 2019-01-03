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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mock"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	encoding "go.thethings.network/lorawan-stack/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredGatewayID = ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}

	timeout = (1 << 4) * test.Delay

	testConfig = Config{
		PacketHandlers:      2,
		PacketBuffer:        10,
		DownlinkPathExpires: 5 * timeout,
		ConnectionExpires:   12 * timeout,
		ScheduleLateTime:    0,
	}
)

func TestConnection(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)

	gs := mock.NewServer()
	addr, _ := net.ResolveUDPAddr("udp", ":0")
	lis, err := net.ListenUDP("udp", addr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	Start(ctx, gs, lis, testConfig)

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
			hasClaim := gs.HasDownlinkClaim(ctx, conn.Gateway().GatewayIdentifiers)
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

func TestTraffic(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)

	gs := mock.NewServer()
	addr, _ := net.ResolveUDPAddr("udp", ":0")
	lis, err := net.ListenUDP("udp", addr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	Start(ctx, gs, lis, testConfig)

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

				// Put unique token, write and expect acknowledgement.
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
						a.So(up.RawPayload, should.Resemble, data)
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
			Name               string
			EUI                types.EUI64
			Packet             encoding.Packet
			AckOK              bool
			ExpectConnect      bool
			SyncClock          *time.Duration
			Path               *ttnpb.DownlinkPath
			Message            *ttnpb.DownlinkMessage
			PreferScheduleLate bool
			ScheduledLate      bool
			SendTxAck          bool
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
				Name:          "TxImmediate",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID},
							uint32(5*time.Second/time.Microsecond),
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							Rx1Delay:         ttnpb.RX_DELAY_1,
							Rx1DataRateIndex: 5,
							Rx1Frequency:     868100000,
						},
					},
				},
				PreferScheduleLate: false,
				ScheduledLate:      false, // Should come immediately as late scheduling is not preferred.
				SendTxAck:          false,
			},
			{
				Name:          "TxPreferLateNoClock",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID},
							uint32(10*time.Second/time.Microsecond),
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x02},
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							Rx1Delay:         ttnpb.RX_DELAY_1,
							Rx1DataRateIndex: 5,
							Rx1Frequency:     868100000,
						},
					},
				},
				PreferScheduleLate: true,
				ScheduledLate:      false, // Should come immediately as there is no clock.
				SendTxAck:          false,
			},
			{
				Name:          "TxPreferLateOK",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     durationPtr(1 * time.Second), // Rx1 delay
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID},
							uint32(150*test.Delay/time.Microsecond),
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x03},
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							Rx1Delay:         ttnpb.RX_DELAY_1,
							Rx1DataRateIndex: 5,
							Rx1Frequency:     868100000,
						},
					},
				},
				PreferScheduleLate: true,
				ScheduledLate:      true, // Should be scheduled late.
				SendTxAck:          true, // From now on, immediate scheduling takes priority over scheduling late preference.
			},
			{
				Name:          "TxPreferLateOverruled",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     durationPtr(0),
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID},
							uint32(15*time.Second/time.Microsecond),
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x04},
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							Rx1Delay:         ttnpb.RX_DELAY_1,
							Rx1DataRateIndex: 5,
							Rx1Frequency:     868100000,
						},
					},
				},
				PreferScheduleLate: true,
				ScheduledLate:      false, // Should be scheduled immediately as it's overruled (JIT queue enabled by TxAck).
				SendTxAck:          true,
			},
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)
				buf, err := tc.Packet.MarshalBinary()
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				// Put unique token, write and expect acknowledgement.
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

				// Sync the clock at 0, i.e. approximate time.Now().
				var clockSynced time.Time
				if tc.SyncClock != nil {
					packet := generatePushData(eui2, false, *tc.SyncClock)
					buf, err = packet.MarshalBinary()
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					token := [2]byte{0x01, byte(i)}
					copy(buf[1:], token[:])
					_, err = udpConn.Write(buf)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					clockSynced = time.Now()
					expectAck(t, udpConn, true, encoding.PushAck, token)
					time.Sleep(timeout) // Ensure that clock gets actually synced.
				}

				// Send the downlink message, optionally buffer first.
				conn.Gateway().ScheduleDownlinkLate = tc.PreferScheduleLate
				_, err = conn.SendDown(tc.Path, tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				// Set expected time for the pull response.
				expectedTime := time.Now()
				if tc.ScheduledLate {
					if tc.SyncClock != nil {
						expectedTime = expectedTime.Add(-*tc.SyncClock)
						expectedTime = expectedTime.Add(-time.Since(clockSynced))
					}
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
			t.Fatalf("Failed to marshal Tx acknowledgement: %v", err)
		}
		token := [2]byte{0x00, 0xae}
		copy(buf[1:], token[:])
		_, err = udpConn.Write(buf)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to write Tx acknowledgement: %v", err)
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
