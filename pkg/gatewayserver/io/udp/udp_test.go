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
	"go.thethings.network/lorawan-stack/pkg/band"
	iotesting "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/testing"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	encoding "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp/encoding"
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

	testConfig = Config{
		PacketHandlers:      2,
		PacketBuffer:        10,
		DownlinkPathExpires: 250 * time.Millisecond,
		ConnectionExpires:   1 * time.Second,
		ScheduleLateTime:    0,
	}

	timeout = 10 * time.Millisecond
)

func TestConnection(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)

	gs := iotesting.NewServer()
	addr, _ := net.ResolveUDPAddr("udp", ":0")
	lis, err := net.ListenUDP("udp", addr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	Start(ctx, gs, lis, testConfig)

	connections := &sync.Map{}
	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	band, err := band.GetByID("EU_863_870")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	conn, err := net.Dial("udp", lis.LocalAddr().String())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	for i, tc := range []struct {
		Name            string
		PacketType      encoding.PacketType
		Wait            time.Duration
		Connects        bool
		HasDownlinkPath bool
		Disconnects     bool
	}{
		{
			Name:            "NewConnectionOnPush",
			PacketType:      encoding.PushData,
			Wait:            0,
			Connects:        true,
			HasDownlinkPath: true,
			Disconnects:     false,
		},
		{
			Name:            "ExistingConnectionOnPull",
			PacketType:      encoding.PullData,
			Wait:            0,
			Connects:        false,
			HasDownlinkPath: false,
			Disconnects:     false,
		},
		{
			Name:            "LoseDownlinkPath",
			PacketType:      encoding.PullData,
			Wait:            testConfig.DownlinkPathExpires * 150 / 100,
			Connects:        false,
			HasDownlinkPath: true,
			Disconnects:     false,
		},
		{
			Name:            "RecoverDownlinkPathWithoutReconnect",
			PacketType:      encoding.PullData,
			Wait:            0,
			Connects:        false,
			HasDownlinkPath: false,
			Disconnects:     false,
		},
		{
			Name:            "LoseConnection",
			PacketType:      encoding.PullData,
			Wait:            testConfig.ConnectionExpires * 150 / 100,
			Connects:        false,
			HasDownlinkPath: true,
			Disconnects:     true,
		},
		{
			Name:            "Reconnect",
			PacketType:      encoding.PullData,
			Wait:            0,
			Connects:        true,
			HasDownlinkPath: false,
			Disconnects:     false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			// Send packet.
			var packet encoding.Packet
			var ackType encoding.PacketType
			if tc.PacketType == encoding.PushData {
				packet = generatePushData(eui, band, false, 0)
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
			hasClaim, err := gs.HasDownlinkClaim(ctx, conn.Gateway().GatewayIdentifiers)
			if a.So(err, should.BeNil) {
				if tc.HasDownlinkPath {
					a.So(hasClaim, should.BeFalse)
				} else {
					a.So(hasClaim, should.BeTrue)
				}
			}
		})
	}

	cancelCtx()
}

func TestTraffic(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)

	gs := iotesting.NewServer()
	addr, _ := net.ResolveUDPAddr("udp", ":0")
	lis, err := net.ListenUDP("udp", addr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	Start(ctx, gs, lis, testConfig)

	connections := &sync.Map{}
	eui1 := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	eui2 := types.EUI64{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
	band, err := band.GetByID("EU_863_870")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

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
				Packet:        generatePushData(eui1, band, true, 100*time.Microsecond),
				AckOK:         true,
				ExpectConnect: true,
			},
			{
				Name:          "ValidExistingConnection",
				EUI:           eui1,
				Packet:        generatePushData(eui1, band, false, 200*time.Microsecond, 300*time.Microsecond),
				AckOK:         true,
				ExpectConnect: false,
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
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
			SyncClock          bool
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
				Name:          "DownlinkMessageImmediate",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     false,
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
					Settings: ttnpb.TxSettings{
						Modulation:      ttnpb.Modulation_LORA,
						Bandwidth:       125000,
						SpreadingFactor: 7,
						CodingRate:      "4/5",
						Frequency:       869525000,
					},
					TxMetadata: ttnpb.TxMetadata{
						Timestamp: uint64(5 * time.Second),
					},
				},
				PreferScheduleLate: false,
				ScheduledLate:      false, // Should come immediately as late scheduling is not preferred.
				SendTxAck:          false,
			},
			{
				Name:          "DownlinkMessagePreferLateNoClock",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     false,
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x02},
					Settings: ttnpb.TxSettings{
						Modulation:      ttnpb.Modulation_LORA,
						Bandwidth:       125000,
						SpreadingFactor: 7,
						CodingRate:      "4/5",
						Frequency:       869525000,
					},
					TxMetadata: ttnpb.TxMetadata{
						Timestamp: uint64(10 * time.Second),
					},
				},
				PreferScheduleLate: true,
				ScheduledLate:      false, // Should come immediately as there is no clock.
				SendTxAck:          false,
			},
			{
				Name:          "DownlinkMessagePreferLateOK",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     true,
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x03},
					Settings: ttnpb.TxSettings{
						Modulation:      ttnpb.Modulation_LORA,
						Bandwidth:       125000,
						SpreadingFactor: 7,
						CodingRate:      "4/5",
						Frequency:       869525000,
					},
					TxMetadata: ttnpb.TxMetadata{
						Timestamp: uint64(200 * time.Millisecond),
					},
				},
				PreferScheduleLate: true,
				ScheduledLate:      true, // Should be scheduled late.
				SendTxAck:          true, // From now on, immediate scheduling takes priority over scheduling late preference.
			},
			{
				Name:          "DownlinkMessagePreferLateOverruled",
				EUI:           eui2,
				Packet:        generatePullData(eui2),
				AckOK:         true,
				ExpectConnect: false,
				SyncClock:     true,
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x04},
					Settings: ttnpb.TxSettings{
						Modulation:      ttnpb.Modulation_LORA,
						Bandwidth:       125000,
						SpreadingFactor: 7,
						CodingRate:      "4/5",
						Frequency:       869525000,
					},
					TxMetadata: ttnpb.TxMetadata{
						Timestamp: uint64(15 * time.Second),
					},
				},
				PreferScheduleLate: true,
				ScheduledLate:      false, // Should be scheduled immediately as it's overruled.
				SendTxAck:          true,
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
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
				if tc.SyncClock {
					packet := generatePushData(eui2, band, false, 0)
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
					expectAck(t, udpConn, true, encoding.PushAck, token)
				}

				// Set expected time for the pull response.
				expectedTime := time.Now()
				if tc.ScheduledLate {
					expectedTime = expectedTime.Add(time.Duration(tc.Message.TxMetadata.Timestamp))
					expectedTime = expectedTime.Add(-testConfig.ScheduleLateTime)
				}

				// Send the downlink message, optionally buffer first.
				conn.Gateway().ScheduleDownlinkLate = tc.PreferScheduleLate
				err = conn.SendDown(tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
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
				expected, err := encoding.TranslateDownstream(tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(*response.Data.TxPacket, should.Resemble, expected)
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
		}
	})

	cancelCtx()
}
