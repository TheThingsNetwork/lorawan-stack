// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package interop

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestExtensions(t *testing.T) { //nolint:paralleltest
	a := assertions.New(t)

	actual := TTIHomeNSAns{
		HomeNSAns: HomeNSAns{
			JsNsMessageHeader: JsNsMessageHeader{
				MessageHeader: MessageHeader{
					ProtocolVersion: ProtocolV1_1,
					TransactionID:   42,
					MessageType:     MessageTypeHomeNSAns,
				},
				SenderID:     EUI64{0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				ReceiverID:   NetID{0x42, 0x0, 0x0},
				ReceiverNSID: &EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			Result: Result{
				ResultCode: ResultSuccess,
			},
			HNSID:  &EUI64{0x42, 0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0},
			HNetID: NetID{0x42, 0x42, 0x0},
		},
		TTIVSExtension: TTIVSExtension{
			HTenantID:  "foo-tenant",
			HNSAddress: "thethings.example.com",
		},
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "\t")
	err := enc.Encode(actual)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	data := buf.Bytes()
	a.So(string(data), should.Equal, `{
	"ProtocolVersion": "1.1",
	"TransactionID": 42,
	"MessageType": "HomeNSAns",
	"SenderID": "4242000000000000",
	"ReceiverID": "420000",
	"ReceiverNSID": "4242420000000000",
	"Result": {
		"ResultCode": "Success"
	},
	"HNSID": "4242424200000000",
	"HNetID": "424200",
	"VSExtension": {
		"VendorID": "EC656E",
		"Object": {
			"TTSV3": {
				"HTenantID": "foo-tenant",
				"HNSAddress": "thethings.example.com"
			}
		}
	}
}
`)

	var expected TTIHomeNSAns
	err = json.Unmarshal(data, &expected)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	a.So(actual, should.Resemble, expected)
}
