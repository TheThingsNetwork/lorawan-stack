// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package crypto

import (
	"crypto/aes"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// deriveSKey derives a session key
func deriveSKey(key types.AES128Key, t byte, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) (derived types.AES128Key) {
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
func DeriveFNwkSIntKey(nwkKey types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveSKey(nwkKey, 0x01, jn, nid, dn)
}

// DeriveSNwkSIntKey derives the LoRaWAN 1.1 Forwarding Network Session Integrity Key
func DeriveSNwkSIntKey(nwkKey types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveSKey(nwkKey, 0x03, jn, nid, dn)
}

// DeriveNwkSEncKey derives the LoRaWAN 1.1 Forwarding Network Session Integrity Key
func DeriveNwkSEncKey(nwkKey types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveSKey(nwkKey, 0x04, jn, nid, dn)
}

// DeriveAppSKey derives the LoRaWAN Application Session Key
// - If a LoRaWAN 1.0 device joins a LoRaWAN 1.0/1.1 network, the AppKey is used as "key"
// - If a LoRaWAN 1.1 device joins a LoRaWAN 1.1 network, the AppKey is used as "key"
// - If a LoRaWAN 1.1 device joins a LoRaWAN 1.0 network, the NwkKey is used as "key"
func DeriveAppSKey(key types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveSKey(key, 0x02, jn, nid, dn)
}

// DeriveNwkSKey derives the LoRaWAN 1.0 Network Session Key
// - Here the AppNonce is entered as JoinNonce
func DeriveNwkSKey(appKey types.AES128Key, jn types.JoinNonce, nid types.NetID, dn types.DevNonce) types.AES128Key {
	return deriveSKey(appKey, 0x01, jn, nid, dn)
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
