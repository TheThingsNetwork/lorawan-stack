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

package crypto_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestComputePingOffset(t *testing.T) {
	for _, tc := range []struct {
		BeaconTime uint32
		DevAddr    types.DevAddr
		PingPeriod uint16

		ExpectedPingOffset uint16
		ErrorAssertion     func(t *testing.T, err error) bool
	}{
		{
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			PingPeriod: 31,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			PingPeriod: 4097,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			PingPeriod:         32,
			ExpectedPingOffset: 6,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			BeaconTime:         0xff42,
			DevAddr:            types.DevAddr{0x00, 0x42, 0x00, 0xff},
			PingPeriod:         4096,
			ExpectedPingOffset: 3994,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(fmt.Sprintf("beacon_time(%d)/dev_addr(%s)/ping_period(%d)", tc.BeaconTime, tc.DevAddr, tc.PingPeriod), func(t *testing.T) {
			a := assertions.New(t)
			p, err := ComputePingOffset(tc.BeaconTime, tc.DevAddr, tc.PingPeriod)
			a.So(p, should.Equal, tc.ExpectedPingOffset)
			a.So(tc.ErrorAssertion(t, err), should.BeTrue)
		})
	}
}
