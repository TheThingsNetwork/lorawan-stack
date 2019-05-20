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

package interop

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestParseMessage(t *testing.T) {
	for _, tc := range []struct {
		Name                    string
		Request                 []byte
		RequestHeaderAssertion  func(*testing.T, RawMessageHeader) bool
		RequestMessageAssertion func(*testing.T, interface{}) bool
		ResponseAssertion       func(*testing.T, int, []byte) bool
	}{
		{
			Name:    "Empty",
			Request: nil,
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusBadRequest)
			},
		},
		{
			Name:    "InvalidJSON",
			Request: []byte("invalid"),
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusBadRequest)
			},
		},
		{
			Name:    "EmptyJSON",
			Request: []byte(`{}`),
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusBadRequest)
			},
		},
		{
			Name:    "InvalidMessageType",
			Request: []byte(`{"ProtocolVersion":"1.1","MessageType":"Invalid"}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "Invalid")
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusBadRequest) &&
					a.So(data, should.BeEmpty)
			},
		},
		{
			Name: "InvalidProtocolVersion",
			Request: []byte(`{
				"ProtocolVersion": "2.0",
				"MessageType": "JoinReq",
				"SenderID": "01",
				"SenderToken": "01"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "JoinReq")
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				if !a.So(statusCode, should.Equal, http.StatusBadRequest) {
					t.FailNow()
				}
				var errorMsg ErrorMessage
				if err := json.Unmarshal(data, &errorMsg); err != nil {
					t.Fatalf("Unmarshal error message failed: %v", err)
				}
				return a.So(errorMsg, should.Resemble, ErrorMessage{
					RawMessageHeader: RawMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "2.0",
							TransactionID:   0,
							MessageType:     MessageTypeJoinAns,
							ReceiverToken:   []byte{0x1},
						},
						ReceiverID: "01",
					},
					Result: ResultInvalidProtocolVersion,
				})
			},
		},
		{
			Name: "InvalidJoinReq",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "JoinReq",
				"SenderID": "01",
				"SenderToken": "01",
				"MACVersion": "1.0.2",
				"PHYPayload": "010203040506"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "JoinReq")
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				if !a.So(statusCode, should.Equal, http.StatusBadRequest) {
					t.FailNow()
				}
				var errorMsg ErrorMessage
				if err := json.Unmarshal(data, &errorMsg); err != nil {
					t.Fatalf("Unmarshal error message failed: %v", err)
				}
				return a.So(errorMsg, should.Resemble, ErrorMessage{
					RawMessageHeader: RawMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							TransactionID:   0,
							MessageType:     MessageTypeJoinAns,
							ReceiverToken:   []byte{0x1},
						},
						ReceiverID: "01",
					},
					Result: ResultMalformedMessage,
				})
			},
		},
		{
			Name: "ValidJoinReq",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "JoinReq",
				"SenderID": "010203",
				"ReceiverID": "0102030405060708",
				"SenderToken": "01",
				"MACVersion": "1.0.2",
				"PHYPayload": "010203040506",
				"DevEUI": "0102030405060708",
				"DevAddr": "01020304",
				"DLSettings": "010203040506",
				"RxDelay": 5,
				"CFList": "010203040506",
				"CFListType": 1
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "JoinReq")
			},
			RequestMessageAssertion: func(t *testing.T, msg interface{}) bool {
				a := assertions.New(t)
				return a.So(msg, should.Resemble, &JoinReq{
					NsJsMessageHeader: NsJsMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							MessageType:     MessageTypeJoinReq,
							SenderToken:     []byte{0x1},
							TransactionID:   0,
						},
						SenderID:   types.NetID{0x1, 0x2, 0x3},
						ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					},
					MACVersion: MACVersion(ttnpb.MAC_V1_0_2),
					PHYPayload: Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
					DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
					DLSettings: Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
					RxDelay:    ttnpb.RX_DELAY_5,
					CFList:     Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
					CFListType: ttnpb.CFListType_CHANNEL_MASKS,
				})
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusOK)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			server := echo.New()
			server.HTTPErrorHandler = ErrorHandler
			server.POST("/", func(c echo.Context) error {
				header := c.Get(headerKey).(*RawMessageHeader)
				if tc.RequestHeaderAssertion != nil && !tc.RequestHeaderAssertion(t, *header) {
					t.Fatal("Header assertion failed")
				}
				msg := c.Get(messageKey)
				if tc.RequestMessageAssertion != nil && !tc.RequestMessageAssertion(t, msg) {
					t.Fatal("Request assertion failed")
				}
				c.NoContent(http.StatusOK)
				return nil
			}, parseMessage())

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(tc.Request))
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}
			if !tc.ResponseAssertion(t, res.StatusCode, data) {
				t.Fatal("Response assertion failed")
			}
		})
	}
}
