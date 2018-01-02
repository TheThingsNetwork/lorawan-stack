// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

func TestSortDevAddrPrefixList(t *testing.T) {
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

func TestDevAddrPrefix_UnmarshalText(t *testing.T) {
	a := assertions.New(t)

	var prefix DevAddrPrefix

	err := prefix.UnmarshalText([]byte("26000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix.DevAddr, should.Equal, DevAddr{0x26, 0x00, 0x00, 0x00})
	a.So(prefix.Length, should.Equal, 7)

	err = prefix.UnmarshalText([]byte("27000000/0"))
	a.So(err, should.BeNil)
	a.So(prefix.DevAddr, should.Equal, DevAddr{0x27, 0x00, 0x00, 0x00})
	a.So(prefix.Length, should.Equal, 0)

	err = prefix.UnmarshalText([]byte("27000000/32"))
	a.So(err, should.BeNil)
	a.So(prefix.DevAddr, should.Equal, DevAddr{0x27, 0x00, 0x00, 0x00})
	a.So(prefix.Length, should.Equal, 32)

	err = prefix.UnmarshalText([]byte("01000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix.DevAddr, should.Equal, DevAddr{0x01, 0x00, 0x00, 0x00})
	a.So(prefix.Length, should.Equal, 7)
}

func TestDevAddrPrefix_NbItems(t *testing.T) {
	a := assertions.New(t)

	var prefix DevAddrPrefix

	err := prefix.UnmarshalText([]byte("26000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix.NbItems(), should.Equal, 33554432)

	err = prefix.UnmarshalText([]byte("27000000/0"))
	a.So(err, should.BeNil)
	a.So(prefix.NbItems(), should.Equal, uint64(4294967296))

	err = prefix.UnmarshalText([]byte("27000000/32"))
	a.So(err, should.BeNil)
	a.So(prefix.NbItems(), should.Equal, 1)

	err = prefix.UnmarshalText([]byte("01000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix.NbItems(), should.Equal, 33554432)
}

func TestDevAddr_number(t *testing.T) {
	a := assertions.New(t)

	var addr1 DevAddr
	err := addr1.UnmarshalText([]byte("26000000"))
	a.So(err, should.BeNil)
	a.So(addr1.MarshalNumber(), should.Equal, 637534208)

	var addr2 DevAddr
	err = addr2.UnmarshalText([]byte("27000000"))
	a.So(err, should.BeNil)
	a.So(addr2.MarshalNumber(), should.Equal, 654311424)

	var addr3 DevAddr
	addr3.UnmarshalNumber(654311424)
	a.So(addr3, should.Equal, addr2)

	var addr4 DevAddr
	addr4.UnmarshalNumber(637534208)
	a.So(addr4, should.Equal, addr1)
}

func TestDevAddrPrefix_numbers(t *testing.T) {
	a := assertions.New(t)

	var prefix1 DevAddrPrefix
	err := prefix1.UnmarshalText([]byte("26000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix1.firstNumericDevAddrCovered(), should.Equal, 637534208)
	a.So(prefix1.FirstDevAddrCovered(), should.Equal, DevAddr{0x26, 0x00, 0x00, 0x00})
	a.So(prefix1.LastDevAddrCovered(), should.Equal, DevAddr{0x27, 0xff, 0xff, 0xff})
	a.So(prefix1.lastNumericDevAddrCovered(), should.Equal, 671088639)
	a.So(prefix1.NbItems(), should.Equal, 33554432)

	var prefix2 DevAddrPrefix
	err = prefix2.UnmarshalText([]byte("27000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix2.firstNumericDevAddrCovered(), should.Equal, 637534208)
	a.So(prefix2.FirstDevAddrCovered(), should.Equal, DevAddr{0x26, 0x00, 0x00, 0x00})

	var prefix3 DevAddrPrefix
	err = prefix3.UnmarshalText([]byte("27000000/8"))
	a.So(err, should.BeNil)
	a.So(prefix3.firstNumericDevAddrCovered(), should.Equal, 654311424)
	a.So(prefix3.FirstDevAddrCovered(), should.Equal, DevAddr{0x27, 0x00, 0x00, 0x00})

	var prefix4 DevAddrPrefix
	err = prefix4.UnmarshalText([]byte("27000000/0"))
	a.So(err, should.BeNil)
	a.So(prefix4.firstNumericDevAddrCovered(), should.Equal, 0)
	a.So(prefix4.FirstDevAddrCovered(), should.Equal, DevAddr{0x00, 0x00, 0x00, 0x00})
	a.So(prefix4.LastDevAddrCovered(), should.Equal, DevAddr{0xff, 0xff, 0xff, 0xff})
	a.So(prefix4.lastNumericDevAddrCovered(), should.Equal, uint64(4294967295))
	a.So(prefix4.NbItems(), should.Equal, uint64(4294967296))

	var prefix5 DevAddrPrefix
	err = prefix5.UnmarshalText([]byte("01000000/7"))
	a.So(err, should.BeNil)
	a.So(prefix5.firstNumericDevAddrCovered(), should.Equal, 0)
	a.So(prefix5.FirstDevAddrCovered(), should.Equal, DevAddr{0x00, 0x00, 0x00, 0x00})
	a.So(prefix5.LastDevAddrCovered(), should.Equal, DevAddr{0x01, 0xff, 0xff, 0xff})
	a.So(prefix5.lastNumericDevAddrCovered(), should.Equal, 33554431)
	a.So(prefix5.NbItems(), should.Equal, 33554432)

	a.So(prefix1.firstNumericDevAddrCovered(), should.Equal, prefix2.firstNumericDevAddrCovered())
}
