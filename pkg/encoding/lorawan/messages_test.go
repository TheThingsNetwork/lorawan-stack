// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"math"
	"testing"

	"github.com/smartystreets/assertions"
	_ "go.thethings.network/lorawan-stack/pkg/crypto" // Needed to make the populators work.
	. "go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var baseBytes = [...]byte{'t', 'e', 's', 't'}

func TestFCtrl(t *testing.T) {
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

			dst := append([]byte{}, baseBytes[:]...)
			b, err := AppendFCtrl(dst, tc.FCtrl, tc.IsUplink, tc.FOptsLen)
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, append(baseBytes[:], tc.Bytes...))
			}
			a.So(dst, should.Resemble, baseBytes[:])

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

func TestAppendFHDR(t *testing.T) {
	fCtrl := ttnpb.FCtrl{
		ADR: true,
		Ack: true,
	}
	for _, tc := range []struct {
		Bytes    []byte
		FHDR     ttnpb.FHDR
		IsUplink bool
	}{
		{
			Bytes: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			Bytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			IsUplink: true,
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0xfe, 0xff},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16 - 1,
				FCtrl:   fCtrl,
			},
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0xfe, 0xff},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16 - 1,
				FCtrl:   fCtrl,
			},
			IsUplink: true,
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0xff, 0xff},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16,
				FCtrl:   fCtrl,
			},
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0xff, 0xff},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16,
				FCtrl:   fCtrl,
			},
			IsUplink: true,
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0x00, 0x00},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16 + 1,
				FCtrl:   fCtrl,
			},
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0x00, 0x00},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16 + 1,
				FCtrl:   fCtrl,
			},
			IsUplink: true,
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0x01, 0x00},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16 + 2,
				FCtrl:   fCtrl,
			},
		},
		{
			Bytes: []byte{0xff, 0xff, 0xff, 0x42, 0b1_0_1_0_0000, 0x01, 0x00},
			FHDR: ttnpb.FHDR{
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				FCnt:    math.MaxUint16 + 2,
				FCtrl:   fCtrl,
			},
			IsUplink: true,
		},
	} {
		dirStr := "downlink"
		if tc.IsUplink {
			dirStr = "uplink"
		}
		t.Run(fmt.Sprintf("%s/DevAddr:%v,FCnt:%v,FOpts:(%s)", dirStr, tc.FHDR.DevAddr, tc.FHDR.FCnt, tc.FHDR.FOpts), func(t *testing.T) {
			a := assertions.New(t)

			dst := append([]byte{}, baseBytes[:]...)
			b, err := AppendFHDR(dst, tc.FHDR, tc.IsUplink)
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, append(baseBytes[:], tc.Bytes...))
			}
			a.So(dst, should.Resemble, baseBytes[:])
		})
	}
}
