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

package types_test

import (
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNetID(t *testing.T) {
	for _, tc := range []struct {
		NetID   NetID
		Type    byte
		ID      []byte
		IDBits  int
		Numeric uint32
	}{
		{
			NetID{0x00, 0x00, 0x2f},
			0,
			[]byte{0x2f},
			6,
			0x00002f,
		},
		{
			NetID{0x20, 0x00, 0x2f},
			1,
			[]byte{0x2f},
			6,
			0x20002f,
		},
		{
			NetID{0x40, 0x00, 0xef},
			2,
			[]byte{0x0, 0xef},
			9,
			0x4000ef,
		},
		{
			NetID{0x7f, 0xff, 0x42},
			3,
			[]byte{0x1f, 0xff, 0x42},
			21,
			0x7fff42,
		},
		{
			NetID{0x9f, 0xff, 0x42},
			4,
			[]byte{0x1f, 0xff, 0x42},
			21,
			0x9fff42,
		},
		{
			NetID{0xbf, 0xff, 0x42},
			5,
			[]byte{0x1f, 0xff, 0x42},
			21,
			0xbfff42,
		},
		{
			NetID{0xdf, 0xff, 0x42},
			6,
			[]byte{0x1f, 0xff, 0x42},
			21,
			0xdfff42,
		},
		{
			NetID{0xff, 0xff, 0x42},
			7,
			[]byte{0x1f, 0xff, 0x42},
			21,
			0xffff42,
		},
	} {
		t.Run(string(tc.Type+'0'), func(t *testing.T) {
			a := assertions.New(t)

			netID, err := NewNetID(tc.Type, tc.ID)
			a.So(err, should.BeNil)
			if !a.So(netID, should.Equal, tc.NetID) {
				return
			}

			a.So(netID.Type(), should.Equal, tc.Type)
			a.So(netID.ID(), should.Resemble, tc.ID)
			a.So(netID.IDBits(), should.Equal, tc.IDBits)
			a.So(netID.MarshalNumber(), should.Equal, tc.Numeric)

			var newNetID NetID
			a.So(newNetID.UnmarshalNumber(tc.Numeric), should.BeNil)
			a.So(newNetID, should.Equal, netID)
		})
	}
}
