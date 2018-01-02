// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestKEK(t *testing.T) {
	a := assertions.New(t)

	var table = []struct {
		plaintext  string
		kek        string
		ciphertext string
	}{
		{"00112233445566778899AABBCCDDEEFF", "000102030405060708090A0B0C0D0E0F", "1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5"},
		{"00112233445566778899AABBCCDDEEFF", "000102030405060708090A0B0C0D0E0F1011121314151617", "96778B25AE6CA435F92B5B97C050AED2468AB8A17AD84E5D"},
		{"00112233445566778899AABBCCDDEEFF", "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F", "64E8C3F9CE0F5BA263E9777905818A2A93C8191E7D6E8AE7"},
		{"00112233445566778899AABBCCDDEEFF0001020304050607", "000102030405060708090A0B0C0D0E0F1011121314151617", "031D33264E15D33268F24EC260743EDCE1C6C7DDEE725A936BA814915C6762D2"},
		{"00112233445566778899AABBCCDDEEFF0001020304050607", "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F", "A8F9BC1612C68B3FF6E6F4FBE30E71E4769C8B80A32CB8958CD5D17D6B254DA1"},
		{"00112233445566778899AABBCCDDEEFF000102030405060708090A0B0C0D0E0F", "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F", "28C9F404C4B810F4CBCCB35CFB87F8263F5786E2D80ED326CBC7F0E71A99F43BFB988B9B7A02DD21"},
	}

	for _, tt := range table {
		plaintext, _ := hex.DecodeString(tt.plaintext)
		kek, _ := hex.DecodeString(tt.kek)
		ciphertext, _ := hex.DecodeString(tt.ciphertext)

		wrapped, err := WrapKey(plaintext, kek)
		a.So(err, should.BeNil)
		a.So(wrapped, should.Resemble, ciphertext)

		unwrapped, err := UnwrapKey(ciphertext, kek)
		a.So(err, should.BeNil)
		a.So(unwrapped, should.Resemble, plaintext)
	}

	var err error

	kek, _ := hex.DecodeString("101112131415161718191A1B1C1D1E1F")
	noBlock, _ := hex.DecodeString("10111213141516")
	tooShort, _ := hex.DecodeString("1011121314151617")

	_, err = WrapKey(tooShort, kek)
	a.So(err, should.NotBeNil)

	_, err = UnwrapKey(tooShort, kek)
	a.So(err, should.NotBeNil)

	_, err = WrapKey(noBlock, kek)
	a.So(err, should.NotBeNil)

	_, err = UnwrapKey(noBlock, kek)
	a.So(err, should.NotBeNil)

	data, _ := hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")

	_, err = WrapKey(data, tooShort)
	a.So(err, should.NotBeNil)

	_, err = UnwrapKey(data, tooShort)
	a.So(err, should.NotBeNil)

	_, err = UnwrapKey(data, kek)
	a.So(err, should.NotBeNil)

}
