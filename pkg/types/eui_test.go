// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestEUI64(t *testing.T) {
	a := assertions.New(t)

	eui := EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42}

	prefix := EUI64Prefix{EUI64{0x26}, 7}
	a.So(prefix.Matches(eui), should.BeTrue)

	// Prefix list
	{
		addr := EUI64{1, 2, 3, 4}
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{0, 0, 0, 0}, 0}), should.BeTrue)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{1, 2, 3, 0}, 24}), should.BeTrue)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{2, 2, 3, 4}, 31}), should.BeFalse)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{1, 1, 3, 4}, 31}), should.BeFalse)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{1, 1, 1, 1}, 15}), should.BeFalse)
	}

	// JSON unmarshalling
	{
		jsonBytes, err := json.Marshal(eui)
		if !a.So(err, should.BeNil) {
			panic(err)
		}
		jsonContent := string(jsonBytes)
		a.So(jsonContent, should.ContainSubstring, "2612345642424242")

		jsonBytes, err = json.Marshal(prefix)
		if !a.So(err, should.BeNil) {
			panic(err)
		}
		jsonContent = string(jsonBytes)
		a.So(jsonContent, should.ContainSubstring, "2600000000000000/7")
	}
}
