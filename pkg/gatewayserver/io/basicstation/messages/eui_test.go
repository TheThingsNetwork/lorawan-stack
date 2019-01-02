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

package messages_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstation/messages"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestMarshalEUI(t *testing.T) {
	a := assertions.New(t)

	eui := messages.EUI(types.EUI64{0xaa, 0xbb, 0x00, 0x01, 0x02, 0x03, 0x42, 0xff})
	data, err := json.Marshal(eui)
	a.So(err, should.BeNil)
	a.So(string(data), should.Equal, `"aabb:1:203:42ff"`)
}

func TestUnmarshalEUI(t *testing.T) {
	for i, tc := range []struct {
		Input  string
		Output types.EUI64
		OK     bool
	}{
		{
			Input:  `"aa-bb-cc-01-02-03-42-ff"`,
			Output: types.EUI64{0xaa, 0xbb, 0xcc, 0x01, 0x02, 0x03, 0x42, 0xff},
			OK:     true,
		},
		{
			Input: `"aa-bb-cc-01-02-03"`,
			OK:    false, // Too short.
		},
		{
			Input: `aa-bb-cc-01-02-03-42-ff`,
			OK:    false, // Not a string.
		},
		{
			Input: `"aa-bb-cc-01-02-03-42-xx"`,
			OK:    false, // Invalid hex.
		},
		{
			Input:  `"aabb:cc01:0203:42ff"`,
			Output: types.EUI64{0xaa, 0xbb, 0xcc, 0x01, 0x02, 0x03, 0x42, 0xff},
			OK:     true,
		},
		{
			Input:  `"aabb:01:203:42ff"`,
			Output: types.EUI64{0xaa, 0xbb, 0x00, 0x01, 0x02, 0x03, 0x42, 0xff},
			OK:     true,
		},
		{
			Input:  `"aabb:01::"`,
			Output: types.EUI64{0xaa, 0xbb, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			OK:     true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			var eui messages.EUI
			err := json.Unmarshal([]byte(tc.Input), &eui)
			if tc.OK {
				a.So(err, should.BeNil)
				a.So(types.EUI64(eui), should.Resemble, tc.Output)
			} else {
				a.So(err, should.NotBeNil)
			}
		})
	}
}
