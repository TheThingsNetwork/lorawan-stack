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

package types_test

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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

	// Number unmarshalling
	{
		number := eui.MarshalNumber()
		a.So(number, should.Equal, uint64(2743312668105523778))
		unmarshaledEui := EUI64{}
		unmarshaledEui.UnmarshalNumber(number)
		a.So(unmarshaledEui, should.Equal, eui)
	}
}
