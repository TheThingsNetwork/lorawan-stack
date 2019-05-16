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
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	encoding "go.thethings.network/lorawan-stack/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestMemoryFirewall(t *testing.T) {
	ctx := test.Context()

	block := 10 * test.Delay
	v := NewMemoryFirewall(ctx, block)

	eui1 := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	eui2 := types.EUI64{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
	upAddr1 := net.UDPAddr{
		IP:   []byte{0x01, 0x01, 0x01, 0x01},
		Port: 1,
	}
	downAddr1 := net.UDPAddr{
		IP:   []byte{0x01, 0x01, 0x01, 0x01},
		Port: 2,
	}
	upAddr2 := net.UDPAddr{
		IP:   []byte{0x02, 0x02, 0x02, 0x02},
		Port: 1,
	}
	downAddr2 := net.UDPAddr{
		IP:   []byte{0x02, 0x02, 0x02, 0x02},
		Port: 2,
	}
	addr3 := net.UDPAddr{
		IP:   []byte{0x03, 0x03, 0x03, 0x03},
		Port: 3,
	}

	for i, tc := range []struct {
		Packet    encoding.Packet
		OK        bool
		WaitAfter time.Duration
	}{
		{
			Packet: encoding.Packet{
				GatewayAddr: &downAddr1,
				PacketType:  encoding.PullData,
			},
			OK: false, // no EUI
		},
		{
			Packet: encoding.Packet{
				GatewayEUI: &eui1,
				PacketType: encoding.PullData,
			},
			OK: false, // no address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui1,
				GatewayAddr: &downAddr1,
				PacketType:  encoding.PullData,
			},
			OK: true, // downstream 1
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui1,
				GatewayAddr: &downAddr1,
				PacketType:  encoding.PullData,
			},
			OK: true, // second time downstream 1 with same address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui1,
				GatewayAddr: &upAddr1,
				PacketType:  encoding.PushData,
			},
			OK: true, // upstream 1 with same address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &downAddr2,
				PacketType:  encoding.PullData,
			},
			OK: true, // first time downlink from 2
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &upAddr2,
				PacketType:  encoding.PushData,
			},
			OK: true, // first time uplink from 2
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &addr3,
				PacketType:  encoding.PullData,
			},
			WaitAfter: block * 2,
			OK:        false, // block change of downlink address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &addr3,
				PacketType:  encoding.PullData,
			},
			OK: true, // permit change of downlink address after block time
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			actual := v.Filter(tc.Packet)
			if !a.So(actual, should.Equal, tc.OK) {
				t.FailNow()
			}

			time.Sleep(tc.WaitAfter)
		})
	}
}
