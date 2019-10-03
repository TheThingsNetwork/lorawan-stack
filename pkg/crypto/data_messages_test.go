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
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestUplinkDownlinkEncryption(t *testing.T) {
	a := assertions.New(t)

	key := types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	addr := types.DevAddr{1, 2, 3, 4}

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

	key := types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	addr := types.DevAddr{1, 2, 3, 4}
	payloadWithoutMIC := []byte{
		0x40,                   // Unconfirmed Uplink
		0x04, 0x03, 0x02, 0x01, // DevAddr 01020304
		0x00,       // Empty FCtrl
		0x01, 0x00, // FCnt 1
		0x01,                   // FPort 1
		0x01, 0x02, 0x03, 0x04, // Data
	}

	mic, err := ComputeLegacyUplinkMIC(key, addr, 1, payloadWithoutMIC)
	a.So(err, should.BeNil)
	a.So(mic, should.Equal, [4]byte{0x3B, 0x07, 0x31, 0x82})

	mic, err = ComputeUplinkMIC(key, key, 0, 0, 0, addr, 1, payloadWithoutMIC)
	a.So(err, should.BeNil)
	a.So(mic, should.Equal, [4]byte{0x3B, 0x07, 0x3B, 0x07})

	mic, err = ComputeLegacyDownlinkMIC(key, addr, 1, payloadWithoutMIC)
	a.So(err, should.BeNil)
	a.So(mic, should.Equal, [4]byte{0xA5, 0x60, 0x9F, 0xA9})

	mic, err = ComputeDownlinkMIC(key, addr, 0, 1, payloadWithoutMIC)
	a.So(err, should.BeNil)
	a.So(mic, should.Equal, [4]byte{0xA5, 0x60, 0x9F, 0xA9})
}
