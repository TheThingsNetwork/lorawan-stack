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
				"SenderNSID": "010203",
				"SenderToken": "01",
				"MACVersion": "1.0.2",
				"PHYPayload": "010203040506",
				"DevEUI": "0102030405060708",
				"DevAddr": "01020304",
				"DLSettings": "FF",
				"RxDelay": 5,
				"CFList": "010203040506"
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
						SenderID:   NetID{0x1, 0x2, 0x3},
						ReceiverID: EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						SenderNSID: NetID{0x1, 0x2, 0x3},
					},
					MACVersion: MACVersion(ttnpb.MAC_V1_0_2),
					PHYPayload: Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
					DevEUI:     EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					DevAddr:    DevAddr{0x1, 0x2, 0x3, 0x4},
					DLSettings: Buffer{0xff},
					RxDelay:    ttnpb.RX_DELAY_5,
					CFList:     Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
				})
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusOK)
			},
		},
		{
			Name: "ValidJoinAns",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "JoinAns",
				"SenderID": "0102030405060708",
				"ReceiverID": "010203",
				"ReceiverNSID": "010203",
				"SenderToken": "01",
				"PHYPayload": "010203040506",
				"Result": "Success",
				"Lifetime": 3600,
				"NwkSKey": {
					"KEKLabel": "test",
					"AESKey": "000102030405060708090A0B0C0D0E0F"
				},
				"AppSKey": {
					"KEKLabel": "test",
					"AESKey": "000102030405060708090A0B0C0D0E0F"
				},
				"SessionKeyID": "01020304"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "JoinAns")
			},
			RequestMessageAssertion: func(t *testing.T, msg interface{}) bool {
				a := assertions.New(t)
				return a.So(msg, should.Resemble, &JoinAns{
					JsNsMessageHeader: JsNsMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							MessageType:     MessageTypeJoinAns,
							SenderToken:     []byte{0x1},
							TransactionID:   0,
						},
						SenderID:     EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						ReceiverID:   NetID{0x1, 0x2, 0x3},
						ReceiverNSID: NetID{0x1, 0x2, 0x3},
					},
					PHYPayload: Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
					Result:     ResultSuccess,
					Lifetime:   3600,
					NwkSKey: (*KeyEnvelope)(&ttnpb.KeyEnvelope{
						KEKLabel:     "test",
						EncryptedKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
					}),
					AppSKey: (*KeyEnvelope)(&ttnpb.KeyEnvelope{
						KEKLabel:     "test",
						EncryptedKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
					}),
					SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
				})
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusOK)
			},
		},
		{
			Name: "ValidAppSKeyReq",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "AppSKeyReq",
				"SenderID": "01020304",
				"ReceiverID": "0102030405060708",
				"SenderToken": "01",
				"DevEUI": "0102030405060708",
				"SessionKeyID": "01020304"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "AppSKeyReq")
			},
			RequestMessageAssertion: func(t *testing.T, msg interface{}) bool {
				a := assertions.New(t)
				return a.So(msg, should.Resemble, &AppSKeyReq{
					AsJsMessageHeader: AsJsMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							MessageType:     MessageTypeAppSKeyReq,
							SenderToken:     []byte{0x1},
							TransactionID:   0,
						},
						SenderID:   Buffer{0x1, 0x2, 0x3, 0x4},
						ReceiverID: EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					},
					DevEUI:       EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
				})
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusOK)
			},
		},
		{
			Name: "ValidAppSKeyAns",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "AppSKeyAns",
				"SenderID": "0102030405060708",
				"ReceiverID": "01020304",
				"SenderToken": "01",
				"Result": "Success",
				"DevEUI": "0102030405060708",
				"AppSKey": {
					"KEKLabel": "test",
					"AESKey": "000102030405060708090A0B0C0D0E0F"
				},
				"SessionKeyID": "01020304"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "AppSKeyAns")
			},
			RequestMessageAssertion: func(t *testing.T, msg interface{}) bool {
				a := assertions.New(t)
				return a.So(msg, should.Resemble, &AppSKeyAns{
					JsAsMessageHeader: JsAsMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							MessageType:     MessageTypeAppSKeyAns,
							SenderToken:     []byte{0x1},
							TransactionID:   0,
						},
						SenderID:   EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						ReceiverID: Buffer{0x1, 0x2, 0x3, 0x4},
					},
					Result: ResultSuccess,
					DevEUI: EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					AppSKey: KeyEnvelope{
						KEKLabel:     "test",
						EncryptedKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
					},
					SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
				})
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusOK)
			},
		},
		{
			Name: "ValidHomeNSReq",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "HomeNSReq",
				"SenderID": "010203",
				"ReceiverID": "0102030405060708",
				"SenderNSID": "010203",
				"SenderToken": "01",
				"DevEUI": "0102030405060708"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "HomeNSReq")
			},
			RequestMessageAssertion: func(t *testing.T, msg interface{}) bool {
				a := assertions.New(t)
				return a.So(msg, should.Resemble, &HomeNSReq{
					NsJsMessageHeader: NsJsMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							MessageType:     MessageTypeHomeNSReq,
							SenderToken:     []byte{0x1},
							TransactionID:   0,
						},
						SenderID:   NetID{0x1, 0x2, 0x3},
						ReceiverID: EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						SenderNSID: NetID{0x1, 0x2, 0x3},
					},
					DevEUI: EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				})
			},
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusOK)
			},
		},
		{
			Name: "ValidHomeNSAns",
			Request: []byte(`{
				"ProtocolVersion": "1.0",
				"MessageType": "HomeNSAns",
				"SenderID": "0102030405060708",
				"ReceiverID": "010203",
				"ReceiverNSID": "010203",
				"SenderToken": "01",
				"Result": "Success",
				"HNSID": "42FFFF",
				"HNetID": "42FFFF"
			}`),
			RequestHeaderAssertion: func(t *testing.T, header RawMessageHeader) bool {
				a := assertions.New(t)
				return a.So(header.MessageType, should.Equal, "HomeNSAns")
			},
			RequestMessageAssertion: func(t *testing.T, msg interface{}) bool {
				a := assertions.New(t)
				return a.So(msg, should.Resemble, &HomeNSAns{
					JsNsMessageHeader: JsNsMessageHeader{
						MessageHeader: MessageHeader{
							ProtocolVersion: "1.0",
							MessageType:     MessageTypeHomeNSAns,
							SenderToken:     []byte{0x1},
							TransactionID:   0,
						},
						SenderID:     EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						ReceiverID:   NetID{0x1, 0x2, 0x3},
						ReceiverNSID: NetID{0x1, 0x2, 0x3},
					},
					Result: ResultSuccess,
					HNSID:  NetID{0x42, 0xff, 0xff},
					HNetID: NetID{0x42, 0xff, 0xff},
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
