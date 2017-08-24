// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDevAddr(t *testing.T) {
	a := assertions.New(t)

	devAddr := DevAddr{0x26, 0x12, 0x34, 0x56}
	a.So(devAddr.NwkID(), should.Equal, 0x13)

	prefix := DevAddrPrefix{DevAddr{0x26}, 7}
	a.So(prefix.Matches(devAddr), should.BeTrue)

	addr := DevAddr{1, 2, 3, 4}
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{0, 0, 0, 0}, 0}), should.BeTrue)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 2, 3, 0}, 24}), should.BeTrue)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{2, 2, 3, 4}, 31}), should.BeFalse)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 1, 3, 4}, 31}), should.BeFalse)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 1, 1, 1}, 15}), should.BeFalse)
}

func ExampleDevAddr_MarshalText() {
	devAddr := DevAddr{0x26, 0x01, 0x26, 0xB4}
	text, err := devAddr.MarshalText()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(text))
	// Output: 260126B4
}

func ExampleDevAddr_UnmarshalText() {
	var devAddr DevAddr
	err := devAddr.UnmarshalText([]byte("2601A3C2"))
	if err != nil {
		panic(err)
	}

	devAddr2 := DevAddr{0x26, 0x01, 0xa3, 0xc2}
	fmt.Println(devAddr == devAddr2)
	// Output: true
}

func ExampleDevAddr_Mask() {
	devAddr := DevAddr{0x26, 0x01, 0x26, 0xB4}
	devAddrMasked := devAddr.Mask(16)
	devAddr2 := DevAddr{0x26, 0x01, 0x00, 0x00}

	fmt.Println(devAddrMasked == devAddr2)
	// Output: true
}

func ExampleDevAddr_NwkID() {
	devAddr := DevAddr{0x26, 0x01, 0x26, 0xB4}
	fmt.Printf("%#x", devAddr.NwkID())
	// Output: 0x13
}

func ExampleDevAddrPrefix_Matches() {
	devAddr := DevAddr{0x26, 0x00, 0x26, 0xB4}
	devAddr2 := DevAddr{0x26, 0x2a, 0x26, 0x8e}
	devAddrPrefix := DevAddrPrefix{
		DevAddr: DevAddr{0x26, 0x00, 0x00, 0x00},
		Length:  16,
	}
	fmt.Println(devAddrPrefix.Matches(devAddr))
	fmt.Println(devAddrPrefix.Matches(devAddr2))
	// Output:
	// true
	// false
}
