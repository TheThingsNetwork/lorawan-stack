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
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	_ "go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/datarate"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var testRights = []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO, ttnpb.Right_RIGHT_GATEWAY_LINK}

func generatePushData(eui types.EUI64, status bool, timestamps ...time.Duration) encoding.Packet {
	packet := encoding.Packet{
		GatewayEUI:      &eui,
		ProtocolVersion: encoding.Version1,
		Token:           [2]byte{0x00, 0x00},
		PacketType:      encoding.PushData,
		Data: &encoding.Data{
			RxPacket: make([]*encoding.RxPacket, len(timestamps)),
		},
	}
	if status {
		packet.Data.Stat = &encoding.Stat{
			Time: encoding.ExpandedTime(time.Now()),
		}
	}
	for i, t := range timestamps {
		rawPayload := randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, i)
		abs := encoding.CompactTime(time.Unix(0, 0).Add(t))
		packet.Data.RxPacket[i] = &encoding.RxPacket{
			Freq: 868.1,
			Chan: 2,
			Modu: "LORA",
			CodR: band.Cr4_5,
			DatR: datarate.DR{
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							SpreadingFactor: 7,
							Bandwidth:       125000,
							CodingRate:      band.Cr4_5,
						},
					},
				},
			},
			Size: uint16(len(rawPayload)),
			Data: base64.StdEncoding.EncodeToString(rawPayload),
			Tmst: uint32(t / time.Microsecond),
			Time: &abs,
		}
	}
	return packet
}

func generatePullData(eui types.EUI64) encoding.Packet {
	return encoding.Packet{
		GatewayEUI:      &eui,
		ProtocolVersion: encoding.Version1,
		Token:           [2]byte{0x00, 0x00},
		PacketType:      encoding.PullData,
	}
}

func generateTxAck(eui types.EUI64, err encoding.TxError) encoding.Packet {
	return encoding.Packet{
		GatewayEUI:      &eui,
		ProtocolVersion: encoding.Version1,
		Token:           [2]byte{0x00, 0x00},
		PacketType:      encoding.TxAck,
		Data: &encoding.Data{
			TxPacketAck: &encoding.TxPacketAck{
				Error: err,
			},
		},
	}
}

func expectAck(t *testing.T, conn net.Conn, expect bool, packetType encoding.PacketType, token [2]byte) {
	var (
		timeout = (1 << 4) * test.Delay
		buf     [65507]byte
	)
	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.Read(buf[:])
	if err != nil {
		if !expect {
			return
		}
		t.Fatal("Failed to read acknowledgment")
	}
	var ack encoding.Packet
	if err := ack.UnmarshalBinary(buf[0:n]); err != nil {
		t.Fatal("Failed to unmarshal acknowledgment")
	}
	if ack.PacketType != packetType {
		t.Fatalf("Packet type %v is not %v", ack.PacketType, packetType)
	}
	if !bytes.Equal(ack.Token[:], token[:]) {
		t.Fatal("Received acknowledgment with unexpected token")
	}
	if !expect {
		t.Fatal("Should not have received acknowledgment for this token")
	}
}

func expectConnection(t *testing.T, server mock.Server, connections *sync.Map, eui types.EUI64, expectNew bool) *io.Connection {
	t.Helper()
	var (
		timeout = (1 << 4) * test.Delay
		a       = assertions.New(t)
		conn    *io.Connection
	)
	select {
	case conn = <-server.Connections():
		if !a.So(expectNew, should.BeTrue) {
			t.Fatal("Should not have a new connection")
		}
		actual := types.MustEUI64(conn.Gateway().GetIds().GetEui())
		if !actual.Equal(eui) {
			t.Fatalf("New connection for unexpected EUI %v", actual)
		}
		if _, loaded := connections.LoadOrStore(eui, conn); loaded {
			t.Fatalf("Gateway %v already connected", eui)
		}
		go func() {
			<-conn.Context().Done()
			connections.Delete(eui)
		}()
	case <-time.After(timeout):
		if !a.So(expectNew, should.BeFalse) {
			t.Fatal("New connection timeout")
		} else if v, loaded := connections.Load(eui); !loaded {
			t.Fatal("Expected existing connection")
		} else {
			conn = v.(*io.Connection)
		}
	}
	return conn
}

func randomUpDataPayload(devAddr types.DevAddr, fPort uint32, size int) []byte {
	var fNwkSIntKey, sNwkSIntKey, appSKey types.AES128Key
	rand.Read(fNwkSIntKey[:])
	rand.Read(sNwkSIntKey[:])
	rand.Read(appSKey[:])

	pld := &ttnpb.MACPayload{
		FHdr: &ttnpb.FHDR{
			DevAddr: devAddr.Bytes(),
			FCnt:    42,
		},
		FPort:      fPort,
		FrmPayload: random.Bytes(size),
	}
	buf, err := crypto.EncryptUplink(appSKey, devAddr, pld.FHdr.FCnt, pld.FrmPayload)
	if err != nil {
		panic(err)
	}
	pld.FrmPayload = buf

	msg := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			Payload: &ttnpb.Message_MacPayload{
				MacPayload: pld,
			},
		},
	}
	buf, err = lorawan.MarshalMessage(msg.Payload)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, 0, 5, 0, devAddr, pld.FHdr.FCnt, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}
