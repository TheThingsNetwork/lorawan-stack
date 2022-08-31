// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/TheThingsIndustries/protoc-gen-go-flags/flagsplugin"
	"github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/customflags"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errInvalidDevAddr = errors.DefineInvalidArgument("invalid_dev_addr", "invalid DevAddr")

// DevAddr is a 32-bit LoRaWAN device address.
type DevAddr [4]byte

// MinDevAddr is the lowest value possible for a DevAddr.
var MinDevAddr = DevAddr{0x00, 0x00, 0x00, 0x00}

// MaxDevAddr is the highest value possible for a DevAddr.
var MaxDevAddr = DevAddr{0xFF, 0xFF, 0xFF, 0xFF}

// IsZero returns true iff the type is zero.
func (addr DevAddr) IsZero() bool { return addr == [4]byte{} }

func (addr DevAddr) String() string   { return strings.ToUpper(hex.EncodeToString(addr[:])) }
func (addr DevAddr) GoString() string { return addr.String() }
func (addr DevAddr) Bytes() []byte {
	b := make([]byte, 4)
	copy(b, addr[:])
	return b
}

// GetDevAddr gets a typed DevAddr from the bytes.
// It returns nil, nil if b is nil.
// It returns an error if unmarshaling fails.
func GetDevAddr(b []byte) (*DevAddr, error) {
	if b == nil {
		return nil, nil
	}
	var t DevAddr
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return &t, nil
}

// MustDevAddr returns a typed DevAddr from the bytes.
// It returns nil if the bytes are empty.
// It panics if unmarshaling results in an error.
func MustDevAddr(b []byte) *DevAddr {
	t, err := GetDevAddr(b)
	if err != nil {
		panic(err)
	}
	return t
}

// OrZero returns the DevAddr value, or a zero value if the DevAddr was nil.
func (addr *DevAddr) OrZero() DevAddr {
	if addr != nil {
		return *addr
	}
	return DevAddr{}
}

// Equal returns true iff addresses are equal.
func (addr DevAddr) Equal(other DevAddr) bool { return addr == other }

// Size implements the Sizer interface.
func (addr DevAddr) Size() int { return 4 }

// Marshal implements the proto.Marshaler interface.
func (addr DevAddr) Marshal() ([]byte, error) { return addr.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (addr DevAddr) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, addr[:]) }

// Unmarshal implements the proto.Unmarshaler interface.
func (addr *DevAddr) Unmarshal(data []byte) error { return addr.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (addr DevAddr) MarshalJSON() ([]byte, error) { return marshalJSONHexBytes(addr[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (addr *DevAddr) UnmarshalJSON(data []byte) error {
	*addr = [4]byte{}
	if err := unmarshalJSONHexBytes(addr[:], data); err != nil {
		return errInvalidDevAddr.WithCause(err)
	}
	return nil
}

// MarshalProtoJSON implements the jsonplugin.Marshaler interface.
func (addr *DevAddr) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if addr == nil {
		s.WriteNil()
		return
	}
	s.WriteString(fmt.Sprintf("%X", addr[:]))
}

// UnmarshalProtoJSON implements the jsonplugin.Unmarshaler interface.
func (addr *DevAddr) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	*addr = [4]byte{}
	b, err := hex.DecodeString(s.ReadString())
	if err != nil {
		s.SetError(err)
		return
	}
	if len(b) != 4 {
		s.SetError(errInvalidDevAddr.WithCause(errInvalidLength.WithAttributes("want", 4, "got", len(b))))
		return
	}
	copy(addr[:], b)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (addr DevAddr) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(addr[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (addr *DevAddr) UnmarshalBinary(data []byte) error {
	*addr = [4]byte{}
	if err := unmarshalBinaryBytes(addr[:], data); err != nil {
		return errInvalidDevAddr.WithCause(err)
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (addr DevAddr) MarshalText() ([]byte, error) { return marshalTextBytes(addr[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (addr *DevAddr) UnmarshalText(data []byte) error {
	*addr = [4]byte{}
	if err := unmarshalTextBytes(addr[:], data); err != nil {
		return errInvalidDevAddr.WithCause(err)
	}
	return nil
}

// MarshalNumber returns the DevAddr in a decimal form.
func (addr DevAddr) MarshalNumber() uint32 {
	return binary.BigEndian.Uint32(addr[:])
}

// UnmarshalNumber retrieves a DevAddr from a decimal form.
func (addr *DevAddr) UnmarshalNumber(n uint32) {
	*addr = [4]byte{}
	binary.BigEndian.PutUint32(addr[:], n)
}

// NetIDType returns the NetID type of the DevAddr.
func (addr DevAddr) NetIDType() (byte, bool) {
	const prefix = 0b11111110
	for i := byte(0); i <= 7; i++ {
		prefixLength := i + 1
		typePrefix := byte(prefix << (7 - i))
		if addr[0]>>(8-prefixLength) == typePrefix>>(8-prefixLength) {
			return i, true
		}
	}
	return 0, false
}

// NwkAddr returns NwkAddr of the DevAddr.
func (addr DevAddr) NwkAddr() ([]byte, bool) {
	netIDType, ok := addr.NetIDType()
	if !ok {
		return nil, false
	}
	switch netIDType {
	case 0:
		return []byte{addr[0] & 0x01, addr[1], addr[2], addr[3]}, true
	case 1:
		return []byte{addr[1], addr[2], addr[3]}, true
	case 2:
		return []byte{addr[1] & 0x0f, addr[2], addr[3]}, true
	case 3:
		return []byte{addr[1] & 0x01, addr[2], addr[3]}, true
	case 4:
		return []byte{addr[2] & 0x7f, addr[3]}, true
	case 5:
		return []byte{addr[2] & 0x1f, addr[3]}, true
	case 6:
		return []byte{addr[2] & 0x03, addr[3]}, true
	case 7:
		return []byte{addr[3] & 0x7f}, true
	}
	panic("unreachable")
}

// NetID returns NetID of the DevAddr.
func (addr DevAddr) NetID() (NetID, bool) {
	netIDType, ok := addr.NetIDType()
	if !ok {
		return NetID{}, false
	}
	switch netIDType {
	case 0:
		return NetID{0b000_00000, 0x0, (addr[0] & 0x7f) >> 1}, true
	case 1:
		return NetID{0b001_00000, 0x0, addr[0] & 0x3f}, true
	case 2:
		return NetID{0b010_00000, (addr[0] & 0x1f) >> 4, (addr[0] << 4) | (addr[1] >> 4)}, true
	case 3:
		return NetID{0b011_00000, (addr[0] >> 1) & 0x07, (addr[0] << 7) | (addr[1] >> 1)}, true
	case 4:
		return NetID{0b100_00000, ((addr[0] & 0x07) << 1) | (addr[1] >> 7), (addr[1] << 1) | (addr[2] >> 7)}, true
	case 5:
		return NetID{0b101_00000, ((addr[0] & 0x03) << 3) | (addr[1] >> 5), (addr[1] << 3) | (addr[2] >> 5)}, true
	case 6:
		return NetID{0b110_00000, ((addr[0] & 0x01) << 6) | (addr[1] >> 2), (addr[1] << 6) | (addr[2] >> 2)}, true
	case 7:
		return NetID{0b111_00000 | addr[1]>>7, (addr[1] << 1) | (addr[2] >> 7), (addr[2] << 1) | (addr[3] >> 7)}, true
	}
	panic("unreachable")
}

// NwkAddrBits returns the length of NwkAddr field of netID in bits.
func NwkAddrBits(netID NetID) uint {
	switch netID.Type() {
	case 0:
		return 25
	case 1:
		return 24
	case 2:
		return 20
	case 3:
		return 17
	case 4:
		return 15
	case 5:
		return 13
	case 6:
		return 10
	case 7:
		return 7
	}
	panic("unreachable")
}

// NwkAddrLength returns the length of NwkAddr field of netID in bytes.
func NwkAddrLength(netID NetID) int {
	return int((NwkAddrBits(netID) + 7) / 8)
}

var errNwkAddrLength = errors.DefineInternal("nwk_addr_bits", "too many bits set in NwkAddr")

// NewDevAddr returns new DevAddr.
func NewDevAddr(netID NetID, nwkAddr []byte) (addr DevAddr, err error) {
	if len(nwkAddr) < 4 {
		nwkAddr = append(make([]byte, 4-len(nwkAddr)), nwkAddr...)
	}
	if nwkAddr[0]&(0xfe<<((NwkAddrBits(netID)-1)%8)) > 0 {
		return DevAddr{}, errNwkAddrLength.New()
	}
	copy(addr[:], nwkAddr)

	nwkID := netID.ID()
	t := netID.Type()
	switch t {
	case 0:
		addr[0] |= nwkID[0] << 1
	case 1:
		addr[0] |= nwkID[0]
	case 2:
		addr[1] |= nwkID[1] << 4
		addr[0] |= nwkID[1] >> 4
		addr[0] |= nwkID[0] << 4
	case 3:
		addr[1] |= nwkID[2] << 1
		addr[0] |= nwkID[2] >> 7
		addr[0] |= nwkID[1] << 1
	case 4:
		addr[2] |= nwkID[2] << 7
		addr[1] |= nwkID[2] >> 1
		addr[1] |= nwkID[1] << 7
		addr[0] |= nwkID[1] >> 1
	case 5:
		addr[2] |= nwkID[2] << 5
		addr[1] |= nwkID[2] >> 3
		addr[1] |= nwkID[1] << 5
		addr[0] |= nwkID[1] >> 3
	case 6:
		addr[2] |= nwkID[2] << 2
		addr[1] |= nwkID[2] >> 6
		addr[1] |= nwkID[1] << 2
		addr[0] |= nwkID[1] >> 6
	case 7:
		addr[3] |= nwkID[2] << 7
		addr[2] |= nwkID[2] >> 1
		addr[2] |= nwkID[1] << 7
		addr[1] |= nwkID[1] >> 1
		addr[1] |= nwkID[0] << 7
	}
	addr[0] |= 0xfe << (7 - t)
	return addr, nil
}

// DevAddrPrefix is a DevAddr with a prefix length.
type DevAddrPrefix struct {
	DevAddr DevAddr
	Length  uint8
}

// IsZero returns true iff the type is zero.
func (prefix DevAddrPrefix) IsZero() bool { return prefix.Length == 0 }

func (prefix DevAddrPrefix) String() string {
	return fmt.Sprintf("%s/%d", prefix.DevAddr, prefix.Length)
}

func (prefix DevAddrPrefix) GoString() string { return prefix.String() }

func (prefix DevAddrPrefix) Bytes() []byte {
	return append(prefix.DevAddr.Bytes(), prefix.Length)
}

// GetDevAddrPrefix gets a typed DevAddrPrefix from the bytes.
// It returns nil, nil if b is nil.
// It returns an error if unmarshaling fails.
func GetDevAddrPrefix(b []byte) (*DevAddrPrefix, error) {
	if b == nil {
		return nil, nil
	}
	var t DevAddrPrefix
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return &t, nil
}

// MustDevAddrPrefix returns a typed DevAddrPrefix from the bytes.
// It returns nil if the bytes are empty.
// It panics if unmarshaling results in an error.
func MustDevAddrPrefix(b []byte) *DevAddrPrefix {
	t, err := GetDevAddrPrefix(b)
	if err != nil {
		panic(err)
	}
	return t
}

// OrZero returns the DevAddr prefix value, or a zero value if the DevAddr prefix was nil.
func (prefix *DevAddrPrefix) OrZero() DevAddrPrefix {
	if prefix != nil {
		return *prefix
	}
	return DevAddrPrefix{}
}

// Equal returns true iff prefixes are equal.
func (prefix DevAddrPrefix) Equal(other DevAddrPrefix) bool {
	return prefix.Length == other.Length && prefix.DevAddr.Equal(other.DevAddr)
}

// Size implements the Sizer interface.
func (prefix DevAddrPrefix) Size() int { return 5 }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (prefix DevAddrPrefix) MarshalTo(data []byte) (int, error) {
	return marshalBinaryBytesTo(data, append(prefix.DevAddr[:], prefix.Length))
}

// Marshal implements the proto.Marshaler interface.
func (prefix DevAddrPrefix) Marshal() ([]byte, error) { return prefix.MarshalBinary() }

// Unmarshal implements the proto.Unmarshaler interface.
func (prefix *DevAddrPrefix) Unmarshal(data []byte) error { return prefix.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (prefix DevAddrPrefix) MarshalJSON() ([]byte, error) {
	str := append([]byte(`"`+hex.EncodeToString(prefix.DevAddr[:])), '/')
	str = append(str, []byte(fmt.Sprintf("%d", prefix.Length))...)
	return append(str, '"'), nil
}

var errInvalidDevAddrPrefix = errors.DefineInvalidArgument("dev_addr_prefix", "invalid DevAddr prefix")

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prefix *DevAddrPrefix) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		*prefix = DevAddrPrefix{}
		return nil
	}
	if len(data) != 12 && len(data) != 13 {
		return errInvalidDevAddrPrefix.New()
	}
	if data[0] != '"' || data[len(data)-1] != '"' {
		return errInvalidJSON.WithAttributes("json", string(data))
	}
	if data[9] != '/' {
		return errInvalidDevAddrPrefix.New()
	}
	b := make([]byte, hex.DecodedLen(8))
	n, err := hex.Decode(b, data[1:9])
	if err != nil {
		return err
	}
	if n != 4 || copy(prefix.DevAddr[:], b) != 4 {
		return errInvalidDevAddrPrefix.New()
	}
	length, err := strconv.Atoi(string(data[10 : len(data)-1]))
	if err != nil {
		return err
	}
	prefix.Length = uint8(length)
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (prefix DevAddrPrefix) MarshalBinary() ([]byte, error) {
	return marshalBinaryBytes(append(prefix.DevAddr[:], prefix.Length))
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (prefix *DevAddrPrefix) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		*prefix = DevAddrPrefix{}
		return nil
	}
	if len(data) != 5 {
		return errInvalidDevAddrPrefix.New()
	}
	if err := prefix.DevAddr.UnmarshalBinary(data[:4]); err != nil {
		return err
	}
	prefix.Length = data[4]
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (prefix DevAddrPrefix) MarshalText() ([]byte, error) {
	b, err := prefix.DevAddr.MarshalText()
	if err != nil {
		return nil, err
	}
	return append(append(b, '/'), []byte(strconv.Itoa(int(prefix.Length)))...), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (prefix *DevAddrPrefix) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*prefix = DevAddrPrefix{}
		return nil
	}
	if len(data) != 10 && len(data) != 11 {
		return errInvalidDevAddrPrefix.New()
	}
	if data[8] != '/' {
		return errInvalidDevAddrPrefix.New()
	}
	if err := prefix.DevAddr.UnmarshalText(data[:8]); err != nil {
		return err
	}
	// transform length from number character range
	if len(data) == 10 {
		prefix.Length = data[9] - '0'
	} else {
		prefix.Length = (data[9]-'0')*10 + (data[10] - '0')
	}
	return nil
}

// UnmarshalConfigString implements the config.Configurable interface.
func (prefix *DevAddrPrefix) UnmarshalConfigString(s string) error {
	return prefix.UnmarshalText([]byte(s))
}

// ConfigString implements the config.Stringer interface.
func (prefix DevAddrPrefix) ConfigString() string {
	return prefix.String()
}

// MarshalDevAddrPrefixSlice marshals a slice of DevAddrPrefixes to JSON.
func MarshalDevAddrPrefixSlice(s *jsonplugin.MarshalState, bs [][]byte) {
	vs := make([]string, len(bs))
	for i, b := range bs {
		prefix := MustDevAddrPrefix(b)
		vs[i] = prefix.String()
	}
	s.WriteStringArray(vs)
}

// UnmarshalDevAddrPrefixSlice unmarshals a slice of DevAddrPrefixes from JSON.
func UnmarshalDevAddrPrefixSlice(s *jsonplugin.UnmarshalState) [][]byte {
	vs := s.ReadStringArray()
	if s.Err() != nil {
		return nil
	}
	bs := make([][]byte, len(vs))
	for i, v := range vs {
		var prefix DevAddrPrefix
		if err := prefix.UnmarshalText([]byte(v)); err != nil {
			s.SetError(err)
			return nil
		}
		bs[i] = prefix.Bytes()
	}
	return bs
}

// WithPrefix returns the DevAddr, but with the first length bits replaced by the Prefix.
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

// Mask returns a copy of the DevAddr with only the first "bits" bits.
func (addr DevAddr) Mask(bits uint8) (masked DevAddr) {
	return (DevAddr{}).WithPrefix(DevAddrPrefix{addr, bits})
}

// HasPrefix returns true iff the DevAddr has a prefix of given length.
func (addr DevAddr) HasPrefix(prefix DevAddrPrefix) bool { return prefix.Matches(addr) }

// Matches returns true iff the DevAddr matches the prefix.
func (prefix DevAddrPrefix) Matches(addr DevAddr) bool {
	return addr.Mask(prefix.Length) == prefix.DevAddr.Mask(prefix.Length)
}

// Copy stores a copy of addr in x and returns it.
func (addr DevAddr) Copy(x *DevAddr) *DevAddr {
	copy(x[:], addr[:])
	return x
}

// GetDevAddrFromFlag gets a DevAddr from a named flag in the flag set.
func GetDevAddrFromFlag(fs *pflag.FlagSet, name string) (value DevAddr, set bool, err error) {
	flag := fs.Lookup(name)
	var devAddr DevAddr
	if flag == nil {
		return devAddr, false, &flagsplugin.ErrFlagNotFound{FlagName: name}
	}
	if !flag.Changed {
		return devAddr, flag.Changed, nil
	}
	if err := devAddr.Unmarshal(flag.Value.(*customflags.ExactBytesValue).Value); err != nil {
		return devAddr, false, err
	}
	return devAddr, flag.Changed, nil
}

// GetDevAddrPrefixSliceFromFlag gets a DevAddrPrefix slice from a named flag in the flag set.
func GetDevAddrPrefixSliceFromFlag(fs *pflag.FlagSet, name string) (value [][]byte, set bool, err error) {
	flag := fs.Lookup(name)
	if flag == nil {
		return nil, false, &flagsplugin.ErrFlagNotFound{FlagName: name}
	}
	value = make([][]byte, len(flag.Value.(*flagsplugin.StringSliceValue).Values))
	for i, v := range flag.Value.(*flagsplugin.StringSliceValue).Values {
		var prefix DevAddrPrefix
		if err := prefix.UnmarshalText([]byte(v.Value)); err != nil {
			return nil, false, err
		}
		value[i] = prefix.Bytes()
	}
	return value, flag.Changed, nil
}
