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

func TestAES128(t *testing.T) {
	a := assertions.New(t)

	aes := AES128Key{
		0x12, 0x34, 0xAE, 0x00, 0x3A, 0xB7, 0x38, 0x01,
		0x52, 0x31, 0x0B, 0x53, 0x3A, 0xB7, 0x38, 0x01,
	}

	a.So(aes.IsZero(), should.BeFalse)

	// JSON
	{
		jsonBytes, err := json.Marshal(aes)
		a.So(err, should.BeNil)
		jsonContent := string(jsonBytes)
		a.So(jsonContent, should.ContainSubstring, "1234AE003AB7380152310B533AB73801")
		var unmarshaledKey AES128Key
		a.So(unmarshaledKey.UnmarshalJSON(jsonBytes), should.BeNil)
		a.So(aes.Equal(unmarshaledKey), should.BeTrue)
	}

	// Text
	{
		textBytes, err := aes.MarshalText()
		a.So(err, should.BeNil)
		textContent := string(textBytes)
		a.So(textContent, should.Equal, "1234AE003AB7380152310B533AB73801")
		var unmarshaledKey AES128Key
		a.So(unmarshaledKey.UnmarshalText(textBytes), should.BeNil)
		a.So(aes.Equal(unmarshaledKey), should.BeTrue)
	}
}
