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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func isNoError(err error) bool { return err == nil }

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
		Packet     encoding.Packet
		ErrorCheck func(error) bool
		WaitAfter  time.Duration
	}{
		{
			Packet: encoding.Packet{
				GatewayAddr: &downAddr1,
				PacketType:  encoding.PullData,
			},
			ErrorCheck: errors.IsInvalidArgument, // no EUI
		},
		{
			Packet: encoding.Packet{
				GatewayEUI: &eui1,
				PacketType: encoding.PullData,
			},
			ErrorCheck: errors.IsInvalidArgument, // no address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui1,
				GatewayAddr: &downAddr1,
				PacketType:  encoding.PullData,
			},
			ErrorCheck: isNoError, // downstream 1
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui1,
				GatewayAddr: &downAddr1,
				PacketType:  encoding.PullData,
			},
			ErrorCheck: isNoError, // second time downstream 1 with same address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui1,
				GatewayAddr: &upAddr1,
				PacketType:  encoding.PushData,
			},
			ErrorCheck: isNoError, // upstream 1 with same address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &downAddr2,
				PacketType:  encoding.PullData,
			},
			ErrorCheck: isNoError, // first time downlink from 2
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &upAddr2,
				PacketType:  encoding.PushData,
			},
			ErrorCheck: isNoError, // first time uplink from 2
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &addr3,
				PacketType:  encoding.PullData,
			},
			WaitAfter:  block * 2,
			ErrorCheck: errors.IsFailedPrecondition, // block change of downlink address
		},
		{
			Packet: encoding.Packet{
				GatewayEUI:  &eui2,
				GatewayAddr: &addr3,
				PacketType:  encoding.PullData,
			},
			ErrorCheck: isNoError, // permit change of downlink address after block time
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			err := v.Filter(tc.Packet)
			if !a.So(tc.ErrorCheck(err), should.BeTrue) {
				t.FailNow()
			}

			time.Sleep(tc.WaitAfter)
		})
	}
}
