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
	"crypto/aes"
	"fmt"

	"github.com/jacobsa/crypto/cmac"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// EncryptJoinAccept uses AES Decrypt to encrypt a JoinAccept message
// - The payload contains JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType) | MIC
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey is used in reply to a JoinRequest
// - In LoRaWAN 1.1, the JSEncKey is used in reply to a RejoinRequest (type 0,1,2)
func EncryptJoinAccept(key types.AES128Key, payload []byte) (encrypted []byte, err error) {
	if len(payload) != 16 && len(payload) != 32 {
		return nil, errors.Errorf("pkg/crypto: join-accept payload must be 16 or 32 bytes, got %d", len(payload))
	}
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	encrypted = make([]byte, len(payload))
	for i := 0; i < len(encrypted); i += 16 {
		cipher.Decrypt(encrypted[i:i+16], payload[i:i+16])
	}
	return
}

// DecryptJoinAccept uses AES Encrypt to decrypt a JoinAccept message
// - The returned payload contains JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType) | MIC
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey or JSEncKey is used
func DecryptJoinAccept(key types.AES128Key, encrypted []byte) (payload []byte, err error) {
	if len(encrypted) != 16 && len(encrypted) != 32 {
		return nil, errors.New("pkg/crypto: encrypted join-accept payload must be 16 or 32 bytes")
	}
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	payload = make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i += 16 {
		cipher.Encrypt(payload[i:i+16], encrypted[i:i+16])
	}
	return
}

// ComputeJoinRequestMIC computes the Message Integrity Code for a JoinRequest message
// - The payload contains MHDR | JoinEUI/AppEUI | DevEUI | DevNonce
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey is used
func ComputeJoinRequestMIC(key types.AES128Key, payload []byte) (mic [4]byte, err error) {
	if len(payload) != 19 {
		return mic, errors.Errorf("pkg/crypto: expected join-request payload length to equal 19, got %d", len(payload))
	}
	hash, err := cmac.New(key[:])
	if err != nil {
		return mic, err
	}
	_, err = hash.Write(payload)
	if err != nil {
		return mic, err
	}
	copy(mic[:], hash.Sum([]byte{}))
	return
}

// ComputeRejoinRequestMIC computes the Message Integrity Code for a RejoinRequest message
// - For a type 0 or 2 RejoinRequest, the payload contains MHDR | RejoinType | NetID | DevEUI | RJcount0
// - For a type 0 or 2 RejoinRequest, the SNwkSIntKey is used
// - For a type 1 RejoinRequest, the payload contains MHDR | RejoinType | JoinEUI | DevEUI | RJcount1
// - For a type 1 RejoinRequest, the JSIntKey is used
func ComputeRejoinRequestMIC(key types.AES128Key, payload []byte) (mic [4]byte, err error) {
	if len(payload) < 2 {
		return mic, errors.New("pkg/crypto: insufficient rejoin-request payload bytes")
	}
	rejoinType := payload[1]
	switch rejoinType {
	case 0, 2:
		if len(payload) != 15 {
			return mic, fmt.Errorf("pkg/crypto: rejoin-request type %d payload must be 15 bytes", rejoinType)
		}
	case 1:
		if len(payload) != 20 {
			return mic, errors.New("pkg/crypto: rejoin-request type 1 payload must be 20 bytes")
		}
	}
	hash, err := cmac.New(key[:])
	if err != nil {
		return mic, err
	}
	_, err = hash.Write(payload)
	if err != nil {
		return mic, err
	}
	copy(mic[:], hash.Sum([]byte{}))
	return
}

// ComputeLegacyJoinAcceptMIC computes the Message Integrity Code for a JoinAccept message
// - The payload contains MHDR | JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType)
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1 with OptNeg=0, the NwkKey is used
func ComputeLegacyJoinAcceptMIC(key types.AES128Key, payload []byte) (mic [4]byte, err error) {
	if n := len(payload); n != 13 && n != 29 {
		return mic, errors.Errorf("pkg/crypto: join-accept payload must be 13 or 29 bytes, got %d", n)
	}
	hash, err := cmac.New(key[:])
	if err != nil {
		return mic, err
	}
	_, err = hash.Write(payload)
	if err != nil {
		return mic, err
	}
	copy(mic[:], hash.Sum([]byte{}))
	return
}

// ComputeJoinAcceptMIC computes the Message Integrity Code for a JoinAccept message
// - The payload contains MHDR | JoinNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType)
// - the joinReqType is 0xFF in reply to a JoinRequest or the rejoin type in reply to a RejoinRequest
func ComputeJoinAcceptMIC(jsIntKey types.AES128Key, joinReqType byte, joinEUI types.EUI64, dn types.DevNonce, payload []byte) (mic [4]byte, err error) {
	if n := len(payload); n != 13 && n != 29 {
		return mic, errors.Errorf("pkg/crypto: join-accept payload must be 13 or 29 bytes, got %d", n)
	}
	hash, err := cmac.New(jsIntKey[:])
	if err != nil {
		return mic, err
	}
	_, err = hash.Write(append(append([]byte{joinReqType}, joinEUI[:]...), dn[:]...))
	if err != nil {
		return mic, err
	}
	_, err = hash.Write(payload)
	if err != nil {
		return mic, err
	}
	copy(mic[:], hash.Sum([]byte{}))
	return
}
