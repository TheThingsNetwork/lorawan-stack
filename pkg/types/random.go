// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"encoding/binary"
)

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

// NewPopulatedDevNonce returns a random DevNonce
func NewPopulatedDevNonce(r randy) (dn DevNonce) {
	r8 := randy8(r)
	copy(dn[:], r8[:])
	return
}

// NewPopulatedJoinNonce returns a random JoinNonce
func NewPopulatedJoinNonce(r randy) (jn JoinNonce) {
	r8 := randy8(r)
	copy(jn[:], r8[:])
	return
}

// NewPopulatedNetID returns a random NetID
func NewPopulatedNetID(r randy) (id NetID) {
	r8 := randy8(r)
	copy(id[:], r8[:])
	return
}

// NewPopulatedDevAddr returns a random DevAddr
func NewPopulatedDevAddr(r randy) (addr DevAddr) {
	r8 := randy8(r)
	copy(addr[:], r8[:])
	return
}

// NewPopulatedDevAddrPrefix returns a random DevAddrPrefix
func NewPopulatedDevAddrPrefix(r randy) (prefix DevAddrPrefix) {
	prefix.DevAddr = NewPopulatedDevAddr(r)
	prefix.Length = uint8(r.Int63())
	return
}

// NewPopulatedEUI64 returns a random EUI64
func NewPopulatedEUI64(r randy) (eui EUI64) {
	r8 := randy8(r)
	copy(eui[:], r8[:])
	return
}

// NewPopulatedAES128Key returns a random AES128Key
func NewPopulatedAES128Key(r randy) (key AES128Key) {
	r8a := randy8(r)
	r8b := randy8(r)
	copy(key[:8], r8a[:])
	copy(key[8:], r8b[:])
	return
}
