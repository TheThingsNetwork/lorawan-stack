// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package crypto

import (
	"crypto/aes"
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/jacobsa/crypto/cmac"
)

// EncryptJoinAccept uses AES Decrypt to encrypt a JoinAccept message
// - The payload contains JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType) | MIC
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1, the NwkKey is used in reply to a JoinRequest
// - In LoRaWAN 1.1, the JSEncKey is used in reply to a RejoinRequest (type 0,1,2)
func EncryptJoinAccept(key types.AES128Key, payload []byte) (encrypted []byte, err error) {
	if len(payload) != 16 && len(payload) != 32 {
		return nil, errors.New("pkg/crypto: join-accept payload must be 16 or 32 bytes")
	}
	cipher, _ := aes.NewCipher(key[:])
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
	cipher, _ := aes.NewCipher(key[:])
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
		return mic, errors.New("pkg/crypto: join-request payload must be 19 bytes")
	}
	hash, _ := cmac.New(key[:])
	_, err = hash.Write(payload)
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
	hash, _ := cmac.New(key[:])
	_, err = hash.Write(payload)
	copy(mic[:], hash.Sum([]byte{}))
	return
}

// ComputeLegacyJoinAcceptMIC computes the Message Integrity Code for a JoinAccept message
// - The payload contains MHDR | JoinNonce/AppNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType)
// - In LoRaWAN 1.0, the AppKey is used
// - In LoRaWAN 1.1 with OptNeg=0, the NwkKey is used
func ComputeLegacyJoinAcceptMIC(key types.AES128Key, payload []byte) (mic [4]byte, err error) {
	if len(payload) != 13 && len(payload) != 29 {
		return mic, errors.New("pkg/crypto: join-accept payload must be 13 or 29 bytes")
	}
	hash, _ := cmac.New(key[:])
	_, err = hash.Write(payload)
	copy(mic[:], hash.Sum([]byte{}))
	return
}

// ComputeJoinAcceptMIC computes the Message Integrity Code for a JoinAccept message
// - The payload contains MHDR | JoinNonce | NetID | DevAddr | DLSettings | RxDelay | (CFList | CFListType)
// - the joinReqType is 0xFF in reply to a JoinRequest or the rejoin type in reply to a RejoinRequest
func ComputeJoinAcceptMIC(jsIntKey types.AES128Key, joinReqType byte, joinEUI types.EUI64, dn types.DevNonce, payload []byte) (mic [4]byte, err error) {
	if len(payload) != 13 && len(payload) != 29 {
		return mic, errors.New("pkg/crypto: join-accept payload must be 13 or 29 bytes")
	}
	hash, _ := cmac.New(jsIntKey[:])
	hash.Write(append(append([]byte{joinReqType}, joinEUI[:]...), dn[:]...))
	_, err = hash.Write(payload)
	copy(mic[:], hash.Sum([]byte{}))
	return
}
