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
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestRateLimitingFirewall(t *testing.T) {
	ctx := test.Context()

	eui1 := &types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	eui2 := &types.EUI64{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}

	t.Run("FloodCheck", func(t *testing.T) {
		for _, tc := range []struct {
			rateLimiting   bool
			errorAssertion func(error) bool
		}{
			{
				rateLimiting:   true,
				errorAssertion: errors.IsResourceExhausted,
			},
			{
				rateLimiting:   false,
				errorAssertion: func(err error) bool { return err == nil },
			},
		} {
			t.Run(fmt.Sprintf("RateLimiting=%v", tc.rateLimiting), func(t *testing.T) {
				a := assertions.New(t)
				f := NewMemoryFirewall(ctx, time.Hour)
				if tc.rateLimiting {
					f = NewRateLimitingFirewall(f, 3, time.Hour)
				}

				for i := 0; i < 4; i++ {
					err := f.Filter(encoding.Packet{
						GatewayEUI: eui1,
						GatewayAddr: &net.UDPAddr{
							IP:   []byte{0x03, 0x03, 0x03, 0x03},
							Port: 3,
						},
						PacketType: encoding.PullData,
					})

					if i < 3 {
						a.So(err, should.BeNil)
					} else {
						a.So(tc.errorAssertion(err), should.BeTrue)
					}
				}

				// Ensure filtering is not affected by port
				err := f.Filter(encoding.Packet{
					GatewayEUI: eui1,
					GatewayAddr: &net.UDPAddr{
						IP:   []byte{0x03, 0x03, 0x03, 0x03},
						Port: 4,
					},
					PacketType: encoding.PullData,
				})
				a.So(tc.errorAssertion(err), should.BeTrue)

				// Ensure other gateways are not affected
				a.So(f.Filter(encoding.Packet{
					GatewayEUI: eui2,
					GatewayAddr: &net.UDPAddr{
						IP:   []byte{0x03, 0x03, 0x03, 0x04},
						Port: 4,
					},
					PacketType: encoding.PullData,
				}), should.BeNil)
			})
		}
	})

	t.Run("Recover", func(t *testing.T) {
		a := assertions.New(t)
		duration := (1 << 6) * test.Delay
		f := NewRateLimitingFirewall(NewMemoryFirewall(ctx, time.Hour), 3, duration)

		packet := encoding.Packet{
			GatewayEUI: eui1,
			GatewayAddr: &net.UDPAddr{
				IP:   []byte{0x03, 0x03, 0x03, 0x03},
				Port: 4,
			},
			PacketType: encoding.PullData,
		}

		a.So(f.Filter(packet), should.BeNil)
		a.So(f.Filter(packet), should.BeNil)
		a.So(f.Filter(packet), should.BeNil)
		a.So(errors.IsResourceExhausted(f.Filter(packet)), should.BeTrue)
		a.So(errors.IsResourceExhausted(f.Filter(packet)), should.BeTrue)

		time.Sleep(2 * duration)
		a.So(f.Filter(packet), should.BeNil)
	})
}
