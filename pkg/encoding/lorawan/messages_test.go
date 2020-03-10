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

package lorawan_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	_ "go.thethings.network/lorawan-stack/pkg/crypto" // Needed to make the populators work.
	. "go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestFCtrl(t *testing.T) {
	base := [...]byte{'t', 'e', 's', 't'}
	for _, tc := range []struct {
		Bytes    []byte
		FCtrl    ttnpb.FCtrl
		FOptsLen uint8
		IsUplink bool
	}{
		{
			Bytes: []byte{0},
		},
		{
			Bytes:    []byte{0},
			IsUplink: true,
		},
		{
			Bytes: []byte{0b1_0_0_0_0010},
			FCtrl: ttnpb.FCtrl{
				ADR: true,
			},
			FOptsLen: 2,
		},
		{
			Bytes: []byte{0b1_0_0_0_0010},
			FCtrl: ttnpb.FCtrl{
				ADR: true,
			},
			FOptsLen: 2,
			IsUplink: true,
		},
		{
			Bytes: []byte{0b1_0_1_1_0100},
			FCtrl: ttnpb.FCtrl{
				ADR:      true,
				Ack:      true,
				FPending: true,
			},
			FOptsLen: 4,
		},
		{
			Bytes: []byte{0b1_1_1_1_0100},
			FCtrl: ttnpb.FCtrl{
				ADR:       true,
				ADRAckReq: true,
				Ack:       true,
				ClassB:    true,
			},
			FOptsLen: 4,
			IsUplink: true,
		},
	} {
		var name string
		if tc.IsUplink {
			name = fmt.Sprintf("uplink/ADR:%v,ADRACKReq:%v,ACK:%v,ClassB:%v,FOptsLen:%d", tc.FCtrl.ADR, tc.FCtrl.ADRAckReq, tc.FCtrl.Ack, tc.FCtrl.ClassB, tc.FOptsLen)
		} else {
			name = fmt.Sprintf("downlink/ADR:%v,ACK:%v,FPending:%v,FOptsLen:%d", tc.FCtrl.ADR, tc.FCtrl.Ack, tc.FCtrl.FPending, tc.FOptsLen)
		}
		t.Run(name, func(t *testing.T) {
			a := assertions.New(t)

			dst := append([]byte{}, base[:]...)
			b, err := AppendFCtrl(dst, tc.FCtrl, tc.IsUplink, tc.FOptsLen)
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, append(base[:], tc.Bytes...))
			}
			a.So(dst, should.Resemble, base[:])

			var fCtrl ttnpb.FCtrl
			b = append([]byte{}, tc.Bytes...)
			err = UnmarshalFCtrl(b, &fCtrl, tc.IsUplink)
			if a.So(err, should.BeNil) {
				a.So(fCtrl, should.Resemble, tc.FCtrl)
			}
			a.So(b, should.Resemble, tc.Bytes)
		})
	}
}
