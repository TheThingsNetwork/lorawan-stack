// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNetID(t *testing.T) {
	for _, tc := range []struct {
		NetID  NetID
		Type   byte
		ID     []byte
		IDBits int
	}{
		{
			NetID{0x00, 0x00, 0x2f},
			0,
			[]byte{0x2f},
			6,
		},
		{
			NetID{0x20, 0x00, 0x2f},
			1,
			[]byte{0x2f},
			6,
		},
		{
			NetID{0x40, 0x00, 0xef},
			2,
			[]byte{0x0, 0xef},
			9,
		},
		{
			NetID{0x7f, 0xff, 0x42},
			3,
			[]byte{0x1f, 0xff, 0x42},
			21,
		},
		{
			NetID{0x9f, 0xff, 0x42},
			4,
			[]byte{0x1f, 0xff, 0x42},
			21,
		},
		{
			NetID{0xbf, 0xff, 0x42},
			5,
			[]byte{0x1f, 0xff, 0x42},
			21,
		},
		{
			NetID{0xdf, 0xff, 0x42},
			6,
			[]byte{0x1f, 0xff, 0x42},
			21,
		},
		{
			NetID{0xff, 0xff, 0x42},
			7,
			[]byte{0x1f, 0xff, 0x42},
			21,
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
		})
	}
}
