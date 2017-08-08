// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// DevAddr is a 32-bit LoRaWAN device address
type DevAddr [4]byte

// IsZero returns true iff the type is zero
func (addr DevAddr) IsZero() bool { return addr == [4]byte{} }

// String implements the Stringer interface
func (addr DevAddr) String() string { return strings.ToUpper(hex.EncodeToString(addr[:])) }

// GoString implements the GoStringer interface
func (addr DevAddr) GoString() string { return addr.String() }

// Size implements the Sizer interface
func (addr DevAddr) Size() int { return 4 }

// MarshalTo implements the proto.MarshalerTo interface
func (addr DevAddr) MarshalTo(data []byte) (int, error) { return copy(data, addr[:]), nil }

// Unmarshal implements the proto.Unmarshaler interface
func (addr *DevAddr) Unmarshal(data []byte) error {
	*addr = [4]byte{}
	if len(data) != 4 || copy(addr[:], data) != 4 {
		return ErrInvalidLength
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (addr DevAddr) MarshalBinary() ([]byte, error) { return addr[:], nil }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (addr *DevAddr) UnmarshalBinary(data []byte) error { return addr.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (addr DevAddr) MarshalText() ([]byte, error) { return []byte(addr.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (addr *DevAddr) UnmarshalText(data []byte) error {
	if len(data) != 8 {
		return ErrInvalidLength
	}
	_, err := hex.Decode(addr[:], data)
	if err != nil {
		return err
	}
	return nil
}

// ErrInvalidPrefix can be returned when unmarshaling an invalid slice into a prefix
var ErrInvalidPrefix = errors.New("invalid device address prefix")

// DevAddrPrefix is a DevAddr with a prefix length
type DevAddrPrefix struct {
	DevAddr DevAddr
	Length  uint8
}

// IsZero returns true iff the type is zero
func (prefix DevAddrPrefix) IsZero() bool { return prefix.Length == 0 }

// String implements the Stringer interface
func (prefix DevAddrPrefix) String() string {
	return fmt.Sprintf("%s/%d", prefix.DevAddr, prefix.Length)
}

// GoString implements the GoStringer interface
func (prefix DevAddrPrefix) GoString() string { return prefix.String() }

// Size implements the Sizer interface
func (prefix DevAddrPrefix) Size() int { return 5 }

// MarshalTo implements the proto.MarshalerTo interface
func (prefix DevAddrPrefix) MarshalTo(data []byte) (int, error) {
	return copy(data, append(prefix.DevAddr[:], prefix.Length)), nil
}

// Unmarshal implements the proto.Unmarshaler interface
func (prefix *DevAddrPrefix) Unmarshal(data []byte) error {
	if len(data) != 5 || copy(prefix.DevAddr[:], data[:4]) != 4 {
		return ErrInvalidLength
	}
	prefix.Length = data[4]
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (prefix DevAddrPrefix) MarshalBinary() ([]byte, error) {
	return append(prefix.DevAddr[:], prefix.Length), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (prefix *DevAddrPrefix) UnmarshalBinary(data []byte) error { return prefix.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (prefix DevAddrPrefix) MarshalText() ([]byte, error) { return []byte(prefix.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (prefix *DevAddrPrefix) UnmarshalText(data []byte) (err error) {
	slash := bytes.IndexByte(data, '/')
	if slash == -1 {
		return ErrInvalidPrefix
	}
	err = prefix.DevAddr.UnmarshalText(data[:slash])
	if err != nil {
		return err
	}
	length, err := strconv.ParseUint(string(data[slash+1:]), 10, 8)
	if err != nil {
		return err
	}
	prefix.Length = uint8(length)
	return nil
}

// NwkID of the DevAddr.
func (addr DevAddr) NwkID() byte {
	return addr[0] >> 1
}

// WithPrefix returns the DevAddr, but with the first length bits replaced by the Prefix
func (addr DevAddr) WithPrefix(prefix DevAddrPrefix) (prefixed DevAddr) {
	k := uint(prefix.Length)
	for i := 0; i < 4; i++ {
		if k >= 8 {
			prefixed[i] = prefix.DevAddr[i] & 0xff
			k -= 8
			continue
		}
		prefixed[i] = (prefix.DevAddr[i] & ^byte(0xff>>k)) | (addr[i] & byte(0xff>>k))
		k = 0
	}
	return
}

// Mask returns a copy of the DevAddr with only the first "bits" bits
func (addr DevAddr) Mask(bits uint8) (masked DevAddr) {
	return (DevAddr{}).WithPrefix(DevAddrPrefix{addr, bits})
}

// HasPrefix returns true iff the DevAddr has a prefix of given length
func (addr DevAddr) HasPrefix(prefix DevAddrPrefix) bool { return prefix.Matches(addr) }

// Matches returns true iff the DevAddr matches the prefix
func (prefix DevAddrPrefix) Matches(addr DevAddr) bool {
	return addr.Mask(prefix.Length) == prefix.DevAddr.Mask(prefix.Length)
}
