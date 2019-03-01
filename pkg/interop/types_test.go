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

package interop_test

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type testMessage struct {
	MACVersion interop.MACVersion
	Buffer     interop.Buffer `json:",omitempty"`
}

func TestMarshalTypes(t *testing.T) {
	a := assertions.New(t)

	{
		msg := &testMessage{
			MACVersion: interop.MACVersion(ttnpb.MAC_V1_0_2),
			Buffer:     interop.Buffer([]byte{0x1, 0x2, 0x3}),
		}
		data, err := json.Marshal(msg)
		if a.So(err, should.BeNil) {
			a.So(string(data), should.Equal, `{"MACVersion":"1.0.2","Buffer":"010203"}`)
		}
	}

	{
		msg := &testMessage{
			MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
		}
		data, err := json.Marshal(msg)
		if a.So(err, should.BeNil) {
			a.So(string(data), should.Equal, `{"MACVersion":"1.1"}`)
		}
	}
}

func TestUnmarshalTypes(t *testing.T) {
	a := assertions.New(t)

	{
		data := []byte(`{"MACVersion":"1.0.2","Buffer":"010203"}`)
		var msg testMessage
		err := json.Unmarshal(data, &msg)
		if a.So(err, should.BeNil) {
			a.So(msg, should.Resemble, testMessage{
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_0_2),
				Buffer:     interop.Buffer([]byte{0x1, 0x2, 0x3}),
			})
		}
	}

	{
		data := []byte(`{"MACVersion":"1.1","Buffer":"0x010203"}`)
		var msg testMessage
		err := json.Unmarshal(data, &msg)
		if a.So(err, should.BeNil) {
			a.So(msg, should.Resemble, testMessage{
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				Buffer:     interop.Buffer([]byte{0x1, 0x2, 0x3}),
			})
		}
	}

	{
		data := []byte(`{"MACVersion":"1.0"}`)
		var msg testMessage
		err := json.Unmarshal(data, &msg)
		if a.So(err, should.BeNil) {
			a.So(msg, should.Resemble, testMessage{
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_0),
			})
		}
	}
}
