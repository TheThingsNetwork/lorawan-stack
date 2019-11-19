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

	"go.thethings.network/lorawan-stack/pkg/types"
)

// deriveSKey derives a session key
func deriveSKey(key types.AES128Key, t byte, jn types.JoinNonce, joinEUI types.EUI64, dn types.DevNonce) (derived types.AES128Key) {
	buf := make([]byte, 16)
	buf[0] = t
	copy(buf[1:4], reverse(jn[:]))
	copy(buf[4:12], reverse(joinEUI[:]))
	copy(buf[12:14], reverse(dn[:]))
	block, _ := aes.NewCipher(key[:])
	block.Encrypt(derived[:], buf)
	return
}

// deriveLegacySKey derives a session key
func deriveLegacySKey(key types.AES128Key, t byte, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) (derived types.AES128Key) {
	buf := make([]byte, 16)
	buf[0] = t
	copy(buf[1:4], reverse(jn[:]))
	copy(buf[4:7], reverse(nid[:]))
	copy(buf[7:9], reverse(dn[:]))
	block, _ := aes.NewCipher(key[:])
	block.Encrypt(derived[:], buf)
	return
}

// DeriveFNwkSIntKey derives the LoRaWAN 1.1 Forwarding Network Session Integrity Key
func DeriveFNwkSIntKey(nwkKey types.AES128Key, jn types.JoinNonce, joinEUI types.EUI64, dn types.DevNonce) types.AES128Key {
	return deriveSKey(nwkKey, 0x01, jn, joinEUI, dn)
}

// DeriveSNwkSIntKey derives the LoRaWAN 1.1 Serving Network Session Integrity Key
func DeriveSNwkSIntKey(nwkKey types.AES128Key, jn types.JoinNonce, joinEUI types.EUI64, dn types.DevNonce) types.AES128Key {
	return deriveSKey(nwkKey, 0x03, jn, joinEUI, dn)
}

// DeriveNwkSEncKey derives the LoRaWAN 1.1 Network Session Encryption Key
func DeriveNwkSEncKey(nwkKey types.AES128Key, jn types.JoinNonce, joinEUI types.EUI64, dn types.DevNonce) types.AES128Key {
	return deriveSKey(nwkKey, 0x04, jn, joinEUI, dn)
}

// DeriveAppSKey derives the LoRaWAN Application Session Key
// - If a LoRaWAN 1.1 device joins a LoRaWAN 1.1 network, the AppKey is used as "key"
func DeriveAppSKey(key types.AES128Key, jn types.JoinNonce, joinEUI types.EUI64, dn types.DevNonce) types.AES128Key {
	return deriveSKey(key, 0x02, jn, joinEUI, dn)
}

// DeriveLegacyAppSKey derives the LoRaWAN Application Session Key
// - If a LoRaWAN 1.0 device joins a LoRaWAN 1.0/1.1 network, the AppKey is used as "key"
// - If a LoRaWAN 1.1 device joins a LoRaWAN 1.0 network, the NwkKey is used as "key"
func DeriveLegacyAppSKey(key types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveLegacySKey(key, 0x02, jn, nid, dn)
}

// DeriveLegacyNwkSKey derives the LoRaWAN 1.0 Network Session Key. AppNonce is entered as JoinNonce.
// - If a LoRaWAN 1.0 device joins a LoRaWAN 1.0/1.1 network, the AppKey is used as "key"
// - If a LoRaWAN 1.1 device joins a LoRaWAN 1.0 network, the NwkKey is used as "key"
func DeriveLegacyNwkSKey(appKey types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveLegacySKey(appKey, 0x01, jn, nid, dn)
}

// deriveKey derives a device key
func deriveDeviceKey(key types.AES128Key, t byte, devEUI types.EUI64) (derived types.AES128Key) {
	buf := make([]byte, 16)
	buf[0] = t
	copy(buf[1:9], reverse(devEUI[:]))
	block, _ := aes.NewCipher(key[:])
	block.Encrypt(derived[:], buf)
	return
}

// DeriveJSIntKey derives the Join Server Integrity Key
func DeriveJSIntKey(key types.AES128Key, devEUI types.EUI64) types.AES128Key {
	return deriveDeviceKey(key, 0x06, devEUI)
}

// DeriveJSEncKey derives the Join Server Encryption Key
func DeriveJSEncKey(key types.AES128Key, devEUI types.EUI64) types.AES128Key {
	return deriveDeviceKey(key, 0x05, devEUI)
}
