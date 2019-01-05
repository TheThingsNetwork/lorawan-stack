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

package types

import "encoding/binary"

type randy interface {
	Int63() int64
}

func randyUint64(r randy) uint64 {
	return uint64(r.Int63())>>31 | uint64(r.Int63())<<32
}

func randy8(r randy) (b [8]byte) {
	binary.LittleEndian.PutUint64(b[:], randyUint64(r))
	return
}

// NewPopulatedDevNonce returns a random DevNonce.
func NewPopulatedDevNonce(r randy) *DevNonce {
	var dn DevNonce
	r8 := randy8(r)
	copy(dn[:], r8[:])
	return &dn
}

// NewPopulatedJoinNonce returns a random JoinNonce.
func NewPopulatedJoinNonce(r randy) *JoinNonce {
	var jn JoinNonce
	r8 := randy8(r)
	copy(jn[:], r8[:])
	return &jn
}

// NewPopulatedNetID returns a random NetID.
func NewPopulatedNetID(r randy) *NetID {
	var id NetID
	r8 := randy8(r)
	copy(id[:], r8[:])
	return &id
}

// NewPopulatedDevAddr returns a random DevAddr.
func NewPopulatedDevAddr(r randy) *DevAddr {
	var addr DevAddr
	r8 := randy8(r)
	copy(addr[:], r8[:])
	return &addr
}

// NewPopulatedDevAddrPrefix returns a random DevAddrPrefix.
func NewPopulatedDevAddrPrefix(r randy) *DevAddrPrefix {
	var prefix DevAddrPrefix
	prefix.DevAddr = *NewPopulatedDevAddr(r)
	prefix.Length = uint8(r.Int63())
	return &prefix
}

// NewPopulatedEUI64 returns a random EUI64.
func NewPopulatedEUI64(r randy) *EUI64 {
	var eui EUI64
	r8 := randy8(r)
	copy(eui[:], r8[:])
	return &eui
}

// NewPopulatedAES128Key returns a random AES128Key.
func NewPopulatedAES128Key(r randy) *AES128Key {
	var key AES128Key
	r8a := randy8(r)
	r8b := randy8(r)
	copy(key[:8], r8a[:])
	copy(key[8:], r8b[:])
	return &key
}
