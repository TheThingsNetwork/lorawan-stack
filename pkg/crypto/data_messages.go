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

	"github.com/jacobsa/crypto/cmac"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func encrypt(key types.AES128Key, dir uint8, addr types.DevAddr, fCnt uint32, payload []byte) (encrypted []byte, err error) {
	k := len(payload) / aes.BlockSize
	if len(payload)%aes.BlockSize != 0 {
		k++
	}
	encrypted = make([]byte, 0, k*16)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err) // types.AES128Key
	}
	var a [aes.BlockSize]byte
	a[0] = 0x01
	a[5] = dir
	copy(a[6:10], reverse(addr[:]))
	binary.LittleEndian.PutUint32(a[10:14], fCnt)
	var s [aes.BlockSize]byte
	var b [aes.BlockSize]byte
	for i := 0; i < k; i++ {
		copy(b[:], payload[i*aes.BlockSize:])
		a[15] = uint8(i + 1)
		cipher.Encrypt(s[:], a[:])
		for j := 0; j < aes.BlockSize; j++ {
			b[j] = b[j] ^ s[j]
		}
		encrypted = append(encrypted, b[:]...)
	}
	return encrypted[:len(payload)], nil
}

// EncryptUplink encrypts an uplink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func EncryptUplink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([]byte, error) {
	return encrypt(key, 0, addr, fCnt, payload)
}

// DecryptUplink decrypts an uplink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func DecryptUplink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([]byte, error) {
	return encrypt(key, 0, addr, fCnt, payload)
}

// EncryptDownlink encrypts a downlink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func EncryptDownlink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([]byte, error) {
	return encrypt(key, 1, addr, fCnt, payload)
}

// DecryptDownlink decrypts a downlink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func DecryptDownlink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([]byte, error) {
	return encrypt(key, 1, addr, fCnt, payload)
}

func computeMIC(key types.AES128Key, dir uint8, confFCnt uint16, addr types.DevAddr, fCnt uint32, payload []byte) (mic [4]byte, err error) {
	hash, _ := cmac.New(key[:])
	var b0 [aes.BlockSize]byte
	b0[0] = 0x49
	binary.LittleEndian.PutUint16(b0[1:3], confFCnt)
	b0[5] = dir
	copy(b0[6:10], reverse(addr[:]))
	binary.LittleEndian.PutUint32(b0[10:14], fCnt)
	b0[15] = uint8(len(payload))
	_, err = hash.Write(b0[:])
	if err != nil {
		return
	}
	_, err = hash.Write(payload)
	if err != nil {
		return
	}
	copy(mic[:], hash.Sum([]byte{}))
	return
}

// ComputeLegacyUplinkMIC computes the Uplink Message Integrity Code.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - The NwkSKey is used
func ComputeLegacyUplinkMIC(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([4]byte, error) {
	return computeMIC(key, 0, 0, addr, fCnt, payload)
}

// ComputeUplinkMIC computes the Uplink Message Integrity Code.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - If this uplink has the ACK bit set, confFCnt must be set to the FCnt of the last downlink.
func ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey types.AES128Key, confFCnt uint32, txDRIdx uint8, txChIdx uint8, addr types.DevAddr, fCnt uint32, payload []byte) (mic [4]byte, err error) {
	m0, err := computeMIC(fNwkSIntKey, 0, 0, addr, fCnt, payload)
	if err != nil {
		return mic, err
	}
	copy(mic[2:], m0[:])
	hash, _ := cmac.New(sNwkSIntKey[:])
	var b0 [aes.BlockSize]byte
	b0[0] = 0x49
	binary.LittleEndian.PutUint16(b0[1:3], uint16(confFCnt))
	b0[3] = txDRIdx
	b0[4] = txChIdx
	b0[5] = 0
	copy(b0[6:10], reverse(addr[:]))
	binary.LittleEndian.PutUint32(b0[10:14], fCnt)
	b0[15] = uint8(len(payload))
	_, err = hash.Write(b0[:])
	if err != nil {
		return
	}
	_, err = hash.Write(payload)
	if err != nil {
		return
	}
	copy(mic[:2], hash.Sum([]byte{}))
	return
}

// ComputeLegacyDownlinkMIC computes the Downlink Message Integrity Code.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - The NwkSKey is used
func ComputeLegacyDownlinkMIC(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([4]byte, error) {
	return computeMIC(key, 1, 0, addr, fCnt, payload)
}

// ComputeDownlinkMIC computes the Downlink Message Integrity Code.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - If this downlink has the ACK bit set, confFCnt must be set to the FCnt of the last uplink
// - The SNwkSIntKey is used
func ComputeDownlinkMIC(key types.AES128Key, addr types.DevAddr, confFCnt uint32, fCnt uint32, payload []byte) ([4]byte, error) {
	return computeMIC(key, 1, uint16(confFCnt), addr, fCnt, payload)
}
