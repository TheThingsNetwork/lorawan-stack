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

	"github.com/jacobsa/crypto/cmac"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var errInvalidJoinAcceptMessageSize = errInvalidSize("join_accept_message", "join-accept message", "16 or 32")

// EncryptJoinAccept uses AES Decrypt to encrypt a join-accept message
// - The payload contains JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType) | MIC
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey is used in reply to a JoinRequest
// - In LoRaWAN 1.1, the JSEncKey is used in reply to a RejoinRequest (type 0,1,2)
func EncryptJoinAccept(key types.AES128Key, payload []byte) ([]byte, error) {
	if len(payload) != 16 && len(payload) != 32 {
		return nil, errInvalidJoinAcceptMessageSize.WithAttributes("size", len(payload))
	}
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	encrypted := make([]byte, len(payload))
	for i := 0; i < len(encrypted); i += 16 {
		cipher.Decrypt(encrypted[i:i+16], payload[i:i+16])
	}
	return encrypted, nil
}

// DecryptJoinAccept uses AES Encrypt to decrypt a join-accept message
// - The returned payload contains JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType) | MIC
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey or JSEncKey is used
func DecryptJoinAccept(key types.AES128Key, encrypted []byte) ([]byte, error) {
	if len(encrypted) != 16 && len(encrypted) != 32 {
		return nil, errInvalidJoinAcceptMessageSize.WithAttributes("size", len(encrypted))
	}
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	payload := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i += 16 {
		cipher.Encrypt(payload[i:i+16], encrypted[i:i+16])
	}
	return payload, nil
}

var errInvalidJoinRequestPayloadSize = errInvalidSize("join_request_payload", "join-request payload", "19")

// ComputeJoinRequestMIC computes the Message Integrity Code for a join-request message
// - The payload contains MHDR | JoinEUI/AppEUI | DevEUI | DevNonce
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey is used
func ComputeJoinRequestMIC(key types.AES128Key, payload []byte) ([4]byte, error) {
	if len(payload) != 19 {
		return [4]byte{}, errInvalidJoinRequestPayloadSize.WithAttributes("size", len(payload))
	}
	hash, err := cmac.New(key[:])
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

var (
	errrInvalidRejoinRequestSize    = errInvalidSize("rejoin_request", "rejoin-request", "15 or 20")
	errrInvalidRejoinRequestType0_2 = errInvalidSize("rejoin_request_0_2", "rejoin-request type 0 or 2", "15")
	errrInvalidRejoinRequestType1   = errInvalidSize("rejoin_request_1", "rejoin-request type 1", "20")
)

// ComputeRejoinRequestMIC computes the Message Integrity Code for a RejoinRequest message
// - For a type 0 or 2 RejoinRequest, the payload contains MHDR | RejoinType | NetID | DevEUI | RJcount0
// - For a type 0 or 2 RejoinRequest, the SNwkSIntKey is used
// - For a type 1 RejoinRequest, the payload contains MHDR | RejoinType | JoinEUI | DevEUI | RJcount1
// - For a type 1 RejoinRequest, the JSIntKey is used
func ComputeRejoinRequestMIC(key types.AES128Key, payload []byte) ([4]byte, error) {
	if len(payload) != 15 && len(payload) != 20 {
		return [4]byte{}, errrInvalidRejoinRequestSize.WithAttributes("size", len(payload))
	}
	rejoinType := payload[1]
	switch rejoinType {
	case 0, 2:
		if len(payload) != 15 {
			return [4]byte{}, errrInvalidRejoinRequestType0_2.WithAttributes("size", len(payload))
		}
	case 1:
		if len(payload) != 20 {
			return [4]byte{}, errrInvalidRejoinRequestType1.WithAttributes("size", len(payload))
		}
	}
	hash, err := cmac.New(key[:])
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

var errInvalidJoinAcceptPayloadSize = errInvalidSize("join_accept_payload", "JoinAccept payload", "13 or 29")

// ComputeLegacyJoinAcceptMIC computes the Message Integrity Code for a join-accept message
// - The payload contains MHDR | JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType)
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1 with OptNeg=0, the NwkKey is used
func ComputeLegacyJoinAcceptMIC(key types.AES128Key, payload []byte) ([4]byte, error) {
	if n := len(payload); n != 13 && n != 29 {
		return [4]byte{}, errInvalidJoinAcceptPayloadSize.WithAttributes("size", len(payload))
	}
	hash, err := cmac.New(key[:])
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

// ComputeJoinAcceptMIC computes the Message Integrity Code for a join-accept message
// - The payload contains MHDR | JoinNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType)
// - the joinReqType is 0xFF in reply to a join-request or the rejoin type in reply to a RejoinRequest
func ComputeJoinAcceptMIC(jsIntKey types.AES128Key, joinReqType byte, joinEUI types.EUI64, dn types.DevNonce, payload []byte) ([4]byte, error) {
	if n := len(payload); n != 13 && n != 29 {
		return [4]byte{}, errInvalidJoinAcceptPayloadSize.WithAttributes("size", len(payload))
	}
	hash, err := cmac.New(jsIntKey[:])
	if err != nil {
		return [4]byte{}, err
	}
	_, err = hash.Write(append(append([]byte{joinReqType}, reverse(joinEUI[:])...), reverse(dn[:])...))
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
