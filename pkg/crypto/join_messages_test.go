// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package crypto

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestJoinAcceptEncryption(t *testing.T) {
	a := assertions.New(t)

	key := types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	bin := []byte{
		0x03, 0x02, 0x01, // AppNonce
		0x03, 0x02, 0x01, // NetID
		0x04, 0x03, 0x02, 0x01, // DevAddr
		0x00,                   // DLSettings
		0x01,                   // RxDelay
		0x32, 0xF5, 0x4A, 0xB3, // MIC
	}

	var enc, dec []byte
	var err error

	_, err = EncryptJoinAccept(key, nil)
	a.So(err, should.NotBeNil)
	_, err = DecryptJoinAccept(key, nil)
	a.So(err, should.NotBeNil)

	enc, err = EncryptJoinAccept(key, bin)
	a.So(err, should.BeNil)
	a.So(enc, should.Resemble, []byte{0xC9, 0xFB, 0xB2, 0x59, 0xE1, 0x16, 0x49, 0x09, 0x6A, 0x56, 0x8A, 0x9E, 0x3B, 0x71, 0x17, 0xC3})

	dec, err = DecryptJoinAccept(key, enc)
	a.So(err, should.BeNil)
	a.So(dec, should.Resemble, bin)
}

func TestJoinRequestMIC(t *testing.T) {
	a := assertions.New(t)

	key := types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	bin := []byte{
		0x00,                                           // JoinRequest
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, // JoinEUI
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, // DevEUI
		0x02, 0x01, // DevNonce
	}

	var mic [4]byte
	var err error

	_, err = ComputeJoinRequestMIC(key, nil)
	a.So(err, should.NotBeNil)

	mic, err = ComputeJoinRequestMIC(key, bin)
	a.So(err, should.BeNil)
	a.So(mic, should.Equal, [4]byte{0xE6, 0xE1, 0x0C, 0x55})
}

func TestRejoinRequestMIC(t *testing.T) {
	t.Skip("TODO: Test ComputeRejoinRequestMIC")
}

func TestJoinAcceptMIC(t *testing.T) {
	a := assertions.New(t)

	key := types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	bin := []byte{
		0x20,             // JoinAccept
		0x03, 0x02, 0x01, // AppNonce
		0x03, 0x02, 0x01, // NetID
		0x04, 0x03, 0x02, 0x01, // DevAddr
		0x00, // DLSettings
		0x01, // RxDelay
	}

	var mic [4]byte
	var err error

	_, err = ComputeLegacyJoinAcceptMIC(key, nil)
	a.So(err, should.NotBeNil)

	mic, err = ComputeLegacyJoinAcceptMIC(key, bin)
	a.So(err, should.BeNil)
	a.So(mic, should.Equal, [4]byte{0x32, 0xF5, 0x4A, 0xB3})

	t.Skip("TODO: Test LoRaWAN 1.1 ComputeJoinAcceptMIC")
}
