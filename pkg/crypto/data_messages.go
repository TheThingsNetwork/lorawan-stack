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
	"fmt"
	"math"

	"github.com/jacobsa/crypto/cmac"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// IV Header values for downlink MIC A computation.
const (
	IvHeaderRx1  = 0x4A
	IvHeaderRx2  = 0x4B
	IvHeaderRxj1 = 0x50
	IvHeaderRxj2 = 0x51
)

type encryptionOptions struct {
	frameTypeConstant [4]byte
}

// EncryptionOption encapsulates custom LoRaWAN encryption options.
type EncryptionOption func(*encryptionOptions)

// WithFrameTypeConstant specifies which bytes are to be used for
// positions [1:5] during encryption. These 4 bytes uniquely identify
// the frame type in order to avoid collisions between different frames
// which may have the same counter values. See the
// `FOpts Encryption, Usage of FCntDwn Errata on the LoRaWAN L2 1.1 Specification`
// erratum for more details on why these collisions may be an issue
// and which values are used by different frame types.
// In the absence of this option, the constant [0x00, 0x00, 0x00, 0x00]
// is used.
func WithFrameTypeConstant(identifier [4]byte) EncryptionOption {
	return func(encOpts *encryptionOptions) {
		copy(encOpts.frameTypeConstant[:], identifier[:])
	}
}

func encryptMessage(key types.AES128Key, dir uint8, addr types.DevAddr, fCnt uint32, payload []byte, opts ...EncryptionOption) ([]byte, error) {
	encOpts := &encryptionOptions{}
	for _, opt := range opts {
		opt(encOpts)
	}
	k := len(payload) / aes.BlockSize
	if len(payload)%aes.BlockSize != 0 {
		k++
	}
	if k > math.MaxUint8 {
		panic(fmt.Sprintf("k value of %d overflows byte", k))
	}
	encrypted := make([]byte, 0, k*16)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err) // types.AES128Key
	}
	var a [aes.BlockSize]byte
	a[0] = 0x01
	copy(a[1:5], encOpts.frameTypeConstant[:])
	a[5] = dir
	copy(a[6:10], reverse(addr[:]))
	binary.LittleEndian.PutUint32(a[10:14], fCnt)
	var s [aes.BlockSize]byte
	var b [aes.BlockSize]byte
	for i := uint8(0); i < uint8(k); i++ {
		copy(b[:], payload[i*aes.BlockSize:])
		a[15] = i + 1
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
func EncryptUplink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte, opts ...EncryptionOption) ([]byte, error) {
	return encryptMessage(key, 0, addr, fCnt, payload, opts...)
}

// DecryptUplink decrypts an uplink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func DecryptUplink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte, opts ...EncryptionOption) ([]byte, error) {
	return encryptMessage(key, 0, addr, fCnt, payload, opts...)
}

// EncryptDownlink encrypts a downlink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func EncryptDownlink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte, opts ...EncryptionOption) ([]byte, error) {
	return encryptMessage(key, 1, addr, fCnt, payload, opts...)
}

// DecryptDownlink decrypts a downlink payload
// - The payload contains the FRMPayload bytes
// - For FPort>0, the AppSKey is used
// - For FPort=0, the NwkSEncKey/NwkSKey is used
func DecryptDownlink(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte, opts ...EncryptionOption) ([]byte, error) {
	return encryptMessage(key, 1, addr, fCnt, payload, opts...)
}

func computeMIC(key types.AES128Key, dir uint8, confFCnt uint16, addr types.DevAddr, fCnt uint32, payload []byte) ([4]byte, error) {
	hash, _ := cmac.New(key[:])
	var b0 [aes.BlockSize]byte
	b0[0] = 0x49
	binary.LittleEndian.PutUint16(b0[1:3], confFCnt)
	b0[5] = dir
	copy(b0[6:10], reverse(addr[:]))
	binary.LittleEndian.PutUint32(b0[10:14], fCnt)
	b0[15] = uint8(len(payload))
	_, err := hash.Write(b0[:])
	if err != nil {
		return [4]byte{}, err
	}
	_, err = hash.Write(payload)
	if err != nil {
		return [4]byte{}, err
	}
	var mic [4]byte
	copy(mic[:], hash.Sum([]byte{}))
	return mic, nil
}

// ComputeLegacyUplinkMIC computes the Uplink Message Integrity Code.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - The NwkSKey is used
func ComputeLegacyUplinkMIC(key types.AES128Key, addr types.DevAddr, fCnt uint32, payload []byte) ([4]byte, error) {
	return computeMIC(key, 0, 0, addr, fCnt, payload)
}

// ComputeUplinkMICFromLegacy computes the Uplink Message Integrity Code from legacy MIC.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - If this uplink has the ACK bit set, confFCnt must be set to the FCnt of the last downlink.
func ComputeUplinkMICFromLegacy(cmacF [4]byte, sNwkSIntKey types.AES128Key, confFCnt uint32, txDRIdx uint8, txChIdx uint8, addr types.DevAddr, fCnt uint32, payload []byte) ([4]byte, error) {
	sHash, _ := cmac.New(sNwkSIntKey[:])
	var b1 [aes.BlockSize]byte
	b1[0] = 0x49
	binary.LittleEndian.PutUint16(b1[1:3], uint16(confFCnt))
	b1[3] = txDRIdx
	b1[4] = txChIdx
	copy(b1[6:10], reverse(addr[:]))
	binary.LittleEndian.PutUint32(b1[10:14], fCnt)
	b1[15] = uint8(len(payload))
	_, err := sHash.Write(b1[:])
	if err != nil {
		return [4]byte{}, err
	}
	_, err = sHash.Write(payload)
	if err != nil {
		return [4]byte{}, err
	}
	var mic [4]byte
	copy(mic[:2], sHash.Sum([]byte{}))
	copy(mic[2:], cmacF[:])
	return mic, nil
}

// ComputeUplinkMIC computes the Uplink Message Integrity Code.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - If this uplink has the ACK bit set, confFCnt must be set to the FCnt of the last downlink.
func ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey types.AES128Key, confFCnt uint32, txDRIdx uint8, txChIdx uint8, addr types.DevAddr, fCnt uint32, payload []byte) ([4]byte, error) {
	cmacF, err := computeMIC(fNwkSIntKey, 0, 0, addr, fCnt, payload)
	if err != nil {
		return [4]byte{}, err
	}
	return ComputeUplinkMICFromLegacy(cmacF, sNwkSIntKey, confFCnt, txDRIdx, txChIdx, addr, fCnt, payload)
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

// ComputeDownlinkMICA computes the Downlink Message Integrity Code A Scheme.
// - The payload contains MHDR | FHDR | FPort | FRMPayload
// - The SNwkSIntKey is used.
func ComputeDownlinkMICA(
	key types.AES128Key,
	ivHeader uint8,
	txDrUp uint8,
	txChUp uint8,
	cnt uint16,
	addr types.DevAddr,
	fCnt uint32,
	payload []byte,
) ([4]byte, error) {
	hash, _ := cmac.New(key[:])
	var aa [aes.BlockSize]byte
	aa[0] = ivHeader
	aa[3] = txDrUp
	aa[4] = txChUp
	binary.LittleEndian.PutUint16(aa[5:7], cnt)
	copy(aa[7:11], reverse(addr[:]))
	binary.LittleEndian.PutUint32(aa[11:15], fCnt)
	aa[15] = uint8(len(payload))
	_, err := hash.Write(aa[:])
	if err != nil {
		return [4]byte{}, err
	}
	_, err = hash.Write(payload)
	if err != nil {
		return [4]byte{}, err
	}
	var mic [4]byte
	copy(mic[:], hash.Sum([]byte{}))
	return mic, nil
}
