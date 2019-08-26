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

package crypto

import (
	"crypto/aes"
	"encoding/binary"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

var iv = [8]byte{0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6}

func concat(a, b [8]byte) []byte {
	c := make([]byte, aes.BlockSize)
	copy(c[:8], a[:])
	copy(c[8:], b[:])
	return c
}

func xor(b [8]byte, t uint64) (c [8]byte) {
	val := binary.BigEndian.Uint64(b[:]) ^ t
	binary.BigEndian.PutUint64(c[:], val)
	return
}

func msb(b [16]byte) (c [8]byte) { copy(c[:], b[:8]); return }
func lsb(b [16]byte) (c [8]byte) { copy(c[:], b[8:]); return }

var errInvalidKeyLength = errInvalidSize("key_length", "key length", "16, 24 or 32")

// WrapKey implements the RFC 3394 Wrap algorithm
func WrapKey(plaintext, kek []byte) ([]byte, error) {
	length := len(plaintext)
	if length%8 != 0 {
		return nil, errInvalidKeyLength.WithAttributes("size", length)
	}

	var n = length / 8
	if n < 2 {
		return nil, errInvalidKeyLength.WithAttributes("size", length)
	}

	cipher, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}

	// Set A to initial value
	var a = iv

	// Fill R blocks
	var r = make([][8]byte, n)
	for i := 0; i < n; i++ {
		copy(r[i][:], plaintext[i*8:(i+1)*8])
	}

	// Run the algorithm
	for j := 0; j <= 5; j++ {
		for i := 1; i <= n; i++ {
			var b [aes.BlockSize]byte
			cipher.Encrypt(b[:], concat(a, r[i-1]))
			a = xor(msb(b), uint64((n*j)+i))
			r[i-1] = lsb(b)
		}
	}

	// Build the result
	ciphertext := make([]byte, 0, 8*(n+1))
	ciphertext = append(ciphertext, a[:]...)
	for i := 0; i < n; i++ {
		ciphertext = append(ciphertext, r[i][:]...)
	}

	return ciphertext, nil
}

var errCorruptKey = errors.DefineCorruption("corrupt_key", "corrupt key data")

// UnwrapKey implements the RFC 3394 Unwrap algorithm
func UnwrapKey(ciphertext, kek []byte) ([]byte, error) {
	length := len(ciphertext)
	if length%8 != 0 {
		return nil, errInvalidKeyLength.WithAttributes("size", length)
	}

	var n = (length / 8) - 1
	if n < 2 {
		return nil, errInvalidKeyLength.WithAttributes("size", length)
	}

	cipher, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}

	// Set A to C[0]
	var a [8]byte
	copy(a[:], ciphertext[:8])

	// Fill R blocks
	var r = make([][8]byte, n)
	for i := 0; i < n; i++ {
		copy(r[i][:], ciphertext[(i+1)*8:(i+2)*8])
	}

	// Run the algorithm
	for j := 5; j >= 0; j-- {
		for i := n; i >= 1; i-- {
			var b [aes.BlockSize]byte
			cipher.Decrypt(b[:], concat(xor(a, uint64(n*j+i)), r[i-1]))
			a = msb(b)
			r[i-1] = lsb(b)
		}
	}

	// Check for corruption
	if a != iv {
		return nil, errCorruptKey
	}

	// Build the result
	plaintext := make([]byte, 0, 8*n)
	for i := 0; i < n; i++ {
		plaintext = append(plaintext, r[i][:]...)
	}

	return plaintext, nil
}
