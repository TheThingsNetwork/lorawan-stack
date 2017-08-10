// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package crypto

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestUplinkDownlinkEncryption(t *testing.T) {
	a := assertions.New(t)

	var key = types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	var addr = types.DevAddr{1, 2, 3, 4}

	var res []byte

	res, _ = EncryptUplink(key, addr, 1, []byte{1, 2, 3, 4})
	a.So(res, should.Resemble, []byte{0xCF, 0xF3, 0x0B, 0x4E})
	res, _ = DecryptUplink(key, addr, 1, []byte{0xCF, 0xF3, 0x0B, 0x4E})
	a.So(res, should.Resemble, []byte{1, 2, 3, 4})

	res, _ = EncryptDownlink(key, addr, 1, []byte{1, 2, 3, 4})
	a.So(res, should.Resemble, []byte{0x4E, 0x75, 0xF4, 0x40})
	res, _ = DecryptDownlink(key, addr, 1, []byte{0x4E, 0x75, 0xF4, 0x40})
	a.So(res, should.Resemble, []byte{1, 2, 3, 4})
}

func TestUplinkDownlinkMIC(t *testing.T) {
	a := assertions.New(t)

	var key = types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	var addr = types.DevAddr{1, 2, 3, 4}
	payloadWithoutMIC := []byte{
		0x40,                   // Unconfirmed Uplink
		0x04, 0x03, 0x02, 0x01, // DevAddr 01020304
		0x00,       // Empty FCtrl
		0x01, 0x00, // FCnt 1
		0x01,                   // FPort 1
		0x01, 0x02, 0x03, 0x04, // Data
	}

	var mic [4]byte

	mic, _ = ComputeLegacyUplinkMIC(key, addr, 1, payloadWithoutMIC)
	a.So(mic, should.Equal, [4]byte{0x3B, 0x07, 0x31, 0x82})

	mic, _ = ComputeUplinkMIC(key, key, 0, 0, 0, addr, 1, payloadWithoutMIC)
	a.So(mic, should.Equal, [4]byte{0x3B, 0x07, 0x3B, 0x07})

	mic, _ = ComputeDownlinkMIC(key, addr, 1, payloadWithoutMIC)
	a.So(mic, should.Equal, [4]byte{0xA5, 0x60, 0x9F, 0xA9})
}
