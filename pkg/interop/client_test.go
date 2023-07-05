// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestGetAppSKey(t *testing.T) { //nolint:paralleltest
	for _, tc := range []struct { //nolint:paralleltest
		Name              string
		NewServer         func(*assertions.Assertion) *httptest.Server
		AsID              string
		Request           *ttnpb.SessionKeyRequest
		ResponseAssertion func(*assertions.Assertion, *ttnpb.AppSKeyResponse) bool
		ErrorAssertion    func(*assertions.Assertion, error) bool
	}{
		{
			Name: "Backend Interfaces 1.0/UnknownDevEUI",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyReq",
						"SenderID":        "test-as",
						"ReceiverID":      "70B3D57ED0000000",
						"DevEUI":          "0102030405060708",
						"SessionKeyID":    "016BFA7BAD4756346A674981E75CDBDC",
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyAns",
						"SenderID":        "70B3D57ED0000000",
						"ReceiverID":      "test-as",
						"Result": map[string]any{
							"ResultCode": "UnknownDevEUI",
						},
					}))
				}))
			},
			AsID: "test-as",
			Request: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}.Bytes(),
				DevEui:       types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
				SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.AppSKeyResponse) bool {
				return a.So(resp, should.BeNil)
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.HaveSameErrorDefinitionAs, ErrUnknownDevEUI)
			},
		},
		{
			Name: "Backend Interfaces 1.1/UnknownDevEUI",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyReq",
						"SenderID":        "test-as",
						"ReceiverID":      "EC656E0000000000",
						"DevEUI":          "0102030405060708",
						"SessionKeyID":    "016BFA7BAD4756346A674981E75CDBDC",
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyAns",
						"SenderID":        "EC656E0000000000",
						"ReceiverID":      "test-as",
						"Result": map[string]any{
							"ResultCode": "UnknownDevEUI",
						},
					}))
				}))
			},
			AsID: "test-as",
			Request: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00}.Bytes(),
				DevEui:       types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
				SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.AppSKeyResponse) bool {
				return a.So(resp, should.BeNil)
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.HaveSameErrorDefinitionAs, ErrUnknownDevEUI)
			},
		},
		{
			Name: "Backend Interfaces 1.0/Success",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyReq",
						"SenderID":        "test-as",
						"ReceiverID":      "70B3D57ED0000000",
						"DevEUI":          "0102030405060708",
						"SessionKeyID":    "016BFA7BAD4756346A674981E75CDBDC",
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyAns",
						"SenderID":        "70B3D57ED0000000",
						"ReceiverID":      "test-as",
						"Result": map[string]any{
							"ResultCode": "Success",
						},
						"Lifetime": 0,
						"AppSKey": map[string]any{
							"KEKLabel": "as:010042",
							"AESKey":   "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69",
						},
						"SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC",
					}))
				}))
			},
			AsID: "test-as",
			Request: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}.Bytes(),
				DevEui:       types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
				SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.AppSKeyResponse) bool {
				return a.So(resp, should.Resemble, &ttnpb.AppSKeyResponse{
					AppSKey: &ttnpb.KeyEnvelope{
						KekLabel:     "as:010042",
						EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69}, //nolint:lll
					},
				})
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.BeNil)
			},
		},
		{
			Name: "Backend Interfaces 1.1/Success",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyReq",
						"SenderID":        "test-as",
						"ReceiverID":      "EC656E0000000000",
						"DevEUI":          "0102030405060708",
						"SessionKeyID":    "016BFA7BAD4756346A674981E75CDBDC",
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "AppSKeyAns",
						"SenderID":        "EC656E0000000000",
						"ReceiverID":      "test-as",
						"Result": map[string]any{
							"ResultCode": "Success",
						},
						"Lifetime": 0,
						"AppSKey": map[string]any{
							"KEKLabel": "as:010042",
							"AESKey":   "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69",
						},
						"SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC",
					}))
				}))
			},
			AsID: "test-as",
			Request: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00}.Bytes(),
				DevEui:       types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
				SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.AppSKeyResponse) bool {
				return a.So(resp, should.Resemble, &ttnpb.AppSKeyResponse{
					AppSKey: &ttnpb.KeyEnvelope{
						KekLabel:     "as:010042",
						EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69}, //nolint:lll
					},
				})
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.BeNil)
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			ctx = log.NewContext(ctx, test.GetLogger(t))

			srv := tc.NewServer(a)
			defer srv.Close()

			c := componenttest.NewComponent(t, &component.Config{})
			componenttest.StartComponent(t, c)
			defer c.Close()

			cl, err := NewClient(ctx, config.InteropClient{
				ConfigSource: "directory",
				Directory:    "testdata/client",
			}, c, SelectorApplicationServer)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to create new client: %s", err)
			}

			res, err := cl.GetAppSKey(ctx, tc.AsID, tc.Request)
			if a.So(tc.ErrorAssertion(a, err), should.BeTrue) {
				a.So(tc.ResponseAssertion(a, res), should.BeTrue)
			} else if err != nil {
				t.Errorf("Received unexpected error: %v", errors.Stack(err))
			}
		})
	}
}

func TestHandleJoinRequest(t *testing.T) { //nolint:paralleltest
	for _, tc := range []struct { //nolint:paralleltest
		Name              string
		NewServer         func(*assertions.Assertion) *httptest.Server
		NetID             types.NetID
		NSID              *types.EUI64
		Request           *ttnpb.JoinRequest
		ResponseAssertion func(*assertions.Assertion, *ttnpb.JoinResponse) bool
		ErrorAssertion    func(*assertions.Assertion, error) bool
	}{
		{
			Name: "Backend Interfaces 1.0/MICFailed",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "JoinReq",
						"SenderID":        "42FFFF",
						"ReceiverID":      "70B3D57ED0000000",
						"MACVersion":      "1.0.3",
						"PHYPayload":      "00000000D07ED5B370080706050403020100003851F0B6",
						"DevEUI":          "0102030405060708",
						"DevAddr":         "01020304",
						"DLSettings":      "00",
						"RxDelay":         5.0,
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "JoinAns",
						"SenderID":        "70B3D57ED0000000",
						"ReceiverID":      "42FFFF",
						"Result": map[string]any{
							"ResultCode": "MICFailed",
						},
					}))
				}))
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				return a.So(resp, should.BeNil)
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.HaveSameErrorDefinitionAs, ErrMIC)
			},
		},
		{
			Name: "Backend Interfaces 1.1/MICFailed",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "JoinReq",
						"SenderID":        "42FFFF",
						"SenderNSID":      "70B3D57ED0000001",
						"ReceiverID":      "EC656E0000000000",
						"MACVersion":      "1.0.3",
						"PHYPayload":      "00000000D07ED5B370080706050403020100003851F0B6",
						"DevEUI":          "0102030405060708",
						"DevAddr":         "01020304",
						"DLSettings":      "00",
						"RxDelay":         5.0,
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "JoinAns",
						"SenderID":        "EC656E0000000000",
						"ReceiverID":      "42FFFF",
						"ReceiverNSID":    "70B3D57ED0000001",
						"Result": map[string]any{
							"ResultCode": "MICFailed",
						},
					}))
				}))
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				return a.So(resp, should.BeNil)
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.HaveSameErrorDefinitionAs, ErrMIC)
			},
		},
		{
			Name: "Backend Interfaces 1.0/Success",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0.0,
						"MessageType":     "JoinReq",
						"SenderID":        "42FFFF",
						"ReceiverID":      "70B3D57ED0000000",
						"MACVersion":      "1.0.3",
						"PHYPayload":      "00000000D07ED5B370080706050403020100003851F0B6",
						"DevEUI":          "0102030405060708",
						"DevAddr":         "01020304",
						"DLSettings":      "00",
						"RxDelay":         5.0,
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.0",
						"TransactionID":   0,
						"MessageType":     "JoinAns",
						"ReceiverToken":   "01",
						"SenderID":        "70B3D57ED0000000",
						"ReceiverID":      "42FFFF",
						"PHYPayload":      "204D675073BB4153B23653EFA82C1F3A49E19C2A8696C9A34BF492674779E4BEFA",
						"Result": map[string]any{
							"ResultCode": "Success",
						},
						"Lifetime": 0,
						"NwkSKey": map[string]any{
							"KEKLabel": "ns:42ffff",
							"AESKey":   "EB56FE6681999F25D548CFEDD4A6528B331BB5ADE1CAF17F",
						},
						"AppSKey": map[string]any{
							"KEKLabel": "as:010042",
							"AESKey":   "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69",
						},
						"SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC",
					}))
				}))
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			NSID:  types.MustEUI64([]byte{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0xFF}),
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				return a.So(resp, should.Resemble, &ttnpb.JoinResponse{
					RawPayload: []byte{0x20, 0x4d, 0x67, 0x50, 0x73, 0xbb, 0x41, 0x53, 0xb2, 0x36, 0x53, 0xef, 0xa8, 0x2c, 0x1f, 0x3a, 0x49, 0xe1, 0x9c, 0x2a, 0x86, 0x96, 0xc9, 0xa3, 0x4b, 0xf4, 0x92, 0x67, 0x47, 0x79, 0xe4, 0xbe, 0xfa}, //nolint:lll
					SessionKeys: &ttnpb.SessionKeys{
						SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "ns:42ffff",
							EncryptedKey: []byte{0xeb, 0x56, 0xfe, 0x66, 0x81, 0x99, 0x9f, 0x25, 0xd5, 0x48, 0xcf, 0xed, 0xd4, 0xa6, 0x52, 0x8b, 0x33, 0x1b, 0xb5, 0xad, 0xe1, 0xca, 0xf1, 0x7f}, //nolint:lll
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KekLabel:     "as:010042",
							EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69}, //nolint:lll
						},
					},
					Lifetime: durationpb.New(0),
				})
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.BeNil)
			},
		},
		{
			Name: "Backend Interfaces 1.1/Success/With Session Key ID",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "JoinReq",
						"SenderID":        "42FFFF",
						"SenderNSID":      "70B3D57ED0000001",
						"ReceiverID":      "EC656E0000000000",
						"MACVersion":      "1.0.3",
						"PHYPayload":      "00000000D07ED5B370080706050403020100003851F0B6",
						"DevEUI":          "0102030405060708",
						"DevAddr":         "01020304",
						"DLSettings":      "00",
						"RxDelay":         5.0,
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0,
						"MessageType":     "JoinAns",
						"ReceiverToken":   "01",
						"SenderID":        "EC656E0000000000",
						"ReceiverID":      "42FFFF",
						"ReceiverNSID":    "70B3D57ED0000001",
						"PHYPayload":      "204D675073BB4153B23653EFA82C1F3A49E19C2A8696C9A34BF492674779E4BEFA",
						"Result": map[string]any{
							"ResultCode": "Success",
						},
						"Lifetime": 0,
						"NwkSKey": map[string]any{
							"KEKLabel": "ns:42ffff",
							"AESKey":   "EB56FE6681999F25D548CFEDD4A6528B331BB5ADE1CAF17F",
						},
						"AppSKey": map[string]any{
							"KEKLabel": "as:010042",
							"AESKey":   "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69",
						},
						"SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC",
					}))
				}))
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			NSID:  types.MustEUI64([]byte{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0xFF}),
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				return a.So(resp, should.Resemble, &ttnpb.JoinResponse{
					RawPayload: []byte{0x20, 0x4d, 0x67, 0x50, 0x73, 0xbb, 0x41, 0x53, 0xb2, 0x36, 0x53, 0xef, 0xa8, 0x2c, 0x1f, 0x3a, 0x49, 0xe1, 0x9c, 0x2a, 0x86, 0x96, 0xc9, 0xa3, 0x4b, 0xf4, 0x92, 0x67, 0x47, 0x79, 0xe4, 0xbe, 0xfa}, //nolint:lll
					SessionKeys: &ttnpb.SessionKeys{
						SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "ns:42ffff",
							EncryptedKey: []byte{0xeb, 0x56, 0xfe, 0x66, 0x81, 0x99, 0x9f, 0x25, 0xd5, 0x48, 0xcf, 0xed, 0xd4, 0xa6, 0x52, 0x8b, 0x33, 0x1b, 0xb5, 0xad, 0xe1, 0xca, 0xf1, 0x7f}, //nolint:lll
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KekLabel:     "as:010042",
							EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69}, //nolint:lll
						},
					},
					Lifetime: durationpb.New(0),
				})
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.BeNil)
			},
		},
		{
			Name: "Backend Interfaces 1.1/Success/Without Session Key ID",
			NewServer: func(a *assertions.Assertion) *httptest.Server {
				return newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := a
					a.So(r.Method, should.Equal, http.MethodPost)
					b := test.Must(io.ReadAll(r.Body))
					var req map[string]any
					test.Must[any](nil, json.Unmarshal(b, &req))
					a.So(req, should.Resemble, map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0.0,
						"MessageType":     "JoinReq",
						"SenderID":        "42FFFF",
						"SenderNSID":      "70B3D57ED0000001",
						"ReceiverID":      "EC656E0000000000",
						"MACVersion":      "1.0.3",
						"PHYPayload":      "00000000D07ED5B370080706050403020100003851F0B6",
						"DevEUI":          "0102030405060708",
						"DevAddr":         "01020304",
						"DLSettings":      "00",
						"RxDelay":         5.0,
					})
					test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
						"ProtocolVersion": "1.1",
						"TransactionID":   0,
						"MessageType":     "JoinAns",
						"ReceiverToken":   "01",
						"SenderID":        "EC656E0000000000",
						"ReceiverID":      "42FFFF",
						"ReceiverNSID":    "70B3D57ED0000001",
						"PHYPayload":      "204D675073BB4153B23653EFA82C1F3A49E19C2A8696C9A34BF492674779E4BEFA",
						"Result": map[string]any{
							"ResultCode": "Success",
						},
						"Lifetime": 0,
						"NwkSKey": map[string]any{
							"KEKLabel": "ns:42ffff",
							"AESKey":   "EB56FE6681999F25D548CFEDD4A6528B331BB5ADE1CAF17F",
						},
						"AppSKey": map[string]any{
							"KEKLabel": "as:010042",
							"AESKey":   "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69",
						},
					}))
				}))
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				if !a.So(GeneratedSessionKeyID(resp.SessionKeys.SessionKeyId), should.BeTrue) {
					return false
				}
				return a.So(resp, should.Resemble, &ttnpb.JoinResponse{
					RawPayload: []byte{0x20, 0x4d, 0x67, 0x50, 0x73, 0xbb, 0x41, 0x53, 0xb2, 0x36, 0x53, 0xef, 0xa8, 0x2c, 0x1f, 0x3a, 0x49, 0xe1, 0x9c, 0x2a, 0x86, 0x96, 0xc9, 0xa3, 0x4b, 0xf4, 0x92, 0x67, 0x47, 0x79, 0xe4, 0xbe, 0xfa}, //nolint:lll
					SessionKeys: &ttnpb.SessionKeys{
						SessionKeyId: resp.SessionKeys.SessionKeyId,
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "ns:42ffff",
							EncryptedKey: []byte{0xeb, 0x56, 0xfe, 0x66, 0x81, 0x99, 0x9f, 0x25, 0xd5, 0x48, 0xcf, 0xed, 0xd4, 0xa6, 0x52, 0x8b, 0x33, 0x1b, 0xb5, 0xad, 0xe1, 0xca, 0xf1, 0x7f}, //nolint:lll
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KekLabel:     "as:010042",
							EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69}, //nolint:lll
						},
					},
					Lifetime: durationpb.New(0),
				})
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.BeNil)
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			ctx = log.NewContext(ctx, test.GetLogger(t))

			srv := tc.NewServer(a)
			defer srv.Close()

			c := componenttest.NewComponent(t, &component.Config{})
			componenttest.StartComponent(t, c)
			defer c.Close()

			cl, err := NewClient(ctx, config.InteropClient{
				ConfigSource: "directory",
				Directory:    "testdata/client",
			}, c, SelectorNetworkServer)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to create new client: %s", err)
			}

			res, err := cl.HandleJoinRequest(ctx, tc.NetID, tc.NSID, tc.Request)
			if a.So(tc.ErrorAssertion(a, err), should.BeTrue) {
				a.So(tc.ResponseAssertion(a, res), should.BeTrue)
			} else if err != nil {
				t.Errorf("Received unexpected error: %v", errors.Stack(err))
			}
		})
	}
}

func TestJoinServerRace(t *testing.T) { //nolint:paralleltest
	for _, tc := range []struct { //nolint:paralleltest
		Name              string
		NewServers        func(*assertions.Assertion) []*httptest.Server
		NetID             types.NetID
		Request           *ttnpb.JoinRequest
		ResponseAssertion func(*assertions.Assertion, *ttnpb.JoinResponse) bool
		ErrorAssertion    func(*assertions.Assertion, error) bool
	}{
		{
			Name: "Slowest server successful",
			NewServers: func(a *assertions.Assertion) []*httptest.Server {
				return []*httptest.Server{
					newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						time.Sleep(test.Delay << 4)
						test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
							"ProtocolVersion": "1.0",
							"TransactionID":   0,
							"MessageType":     "JoinAns",
							"ReceiverToken":   "01",
							"SenderID":        "70B3D57ED0000000",
							"ReceiverID":      "42FFFF",
							"PHYPayload":      "204D675073BB4153B23653EFA82C1F3A49E19C2A8696C9A34BF492674779E4BEFA",
							"Result": map[string]any{
								"ResultCode": "Success",
							},
							"Lifetime": 0,
							"NwkSKey": map[string]any{
								"KEKLabel": "ns:42ffff",
								"AESKey":   "EB56FE6681999F25D548CFEDD4A6528B331BB5ADE1CAF17F",
							},
							"AppSKey": map[string]any{
								"KEKLabel": "as:010042",
								"AESKey":   "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69",
							},
							"SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC",
						}))
					})),
					newTLSServer(9184, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
							"ProtocolVersion": "1.0",
							"TransactionID":   0.0,
							"MessageType":     "JoinAns",
							"SenderID":        "EC656E0000000001",
							"ReceiverID":      "42FFFF",
							"Result": map[string]any{
								"ResultCode": "MICFailed",
							},
						}))
					})),
				}
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x01}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				return a.So(resp, should.Resemble, &ttnpb.JoinResponse{
					RawPayload: []byte{0x20, 0x4d, 0x67, 0x50, 0x73, 0xbb, 0x41, 0x53, 0xb2, 0x36, 0x53, 0xef, 0xa8, 0x2c, 0x1f, 0x3a, 0x49, 0xe1, 0x9c, 0x2a, 0x86, 0x96, 0xc9, 0xa3, 0x4b, 0xf4, 0x92, 0x67, 0x47, 0x79, 0xe4, 0xbe, 0xfa}, //nolint:lll
					SessionKeys: &ttnpb.SessionKeys{
						SessionKeyId: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc}, //nolint:lll
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "ns:42ffff",
							EncryptedKey: []byte{0xeb, 0x56, 0xfe, 0x66, 0x81, 0x99, 0x9f, 0x25, 0xd5, 0x48, 0xcf, 0xed, 0xd4, 0xa6, 0x52, 0x8b, 0x33, 0x1b, 0xb5, 0xad, 0xe1, 0xca, 0xf1, 0x7f}, //nolint:lll
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KekLabel:     "as:010042",
							EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69}, //nolint:lll
						},
					},
					Lifetime: durationpb.New(0),
				})
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.BeNil)
			},
		},
		{
			Name: "Fastest server error",
			NewServers: func(a *assertions.Assertion) []*httptest.Server {
				return []*httptest.Server{
					newTLSServer(9183, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// This is the slowest response.
						time.Sleep(test.Delay << 4)
						test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
							"ProtocolVersion": "1.0",
							"TransactionID":   0.0,
							"MessageType":     "JoinAns",
							"SenderID":        "EC656E0000000001",
							"ReceiverID":      "42FFFF",
							"Result": map[string]any{
								"ResultCode": "JoinReqFailed",
							},
						}))
					})),
					newTLSServer(9184, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						test.Must[any](nil, json.NewEncoder(w).Encode(map[string]any{
							"ProtocolVersion": "1.0",
							"TransactionID":   0.0,
							"MessageType":     "JoinAns",
							"SenderID":        "EC656E0000000001",
							"ReceiverID":      "42FFFF",
							"Result": map[string]any{
								"ResultCode": "MICFailed",
							},
						}))
					})),
				}
			},
			NetID: types.NetID{0x42, 0xff, 0xff},
			Request: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04}.Bytes(),
				RxDelay:            ttnpb.RxDelay_RX_DELAY_5,
				DownlinkSettings:   &ttnpb.DLSettings{},
				Payload: &ttnpb.Message{
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui: types.EUI64{0xec, 0x65, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x01}.Bytes(),
							DevEui:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}.Bytes(),
						},
					},
				},
				RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6}, //nolint:lll
			},
			ResponseAssertion: func(a *assertions.Assertion, resp *ttnpb.JoinResponse) bool {
				return a.So(resp, should.BeNil)
			},
			ErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(err, should.HaveSameErrorDefinitionAs, ErrMIC)
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			ctx = log.NewContext(ctx, test.GetLogger(t))

			srvs := tc.NewServers(a)
			for _, srv := range srvs {
				defer srv.Close() //nolint:revive
			}

			c := componenttest.NewComponent(t, &component.Config{})
			componenttest.StartComponent(t, c)
			defer c.Close()

			cl, err := NewClient(ctx, config.InteropClient{
				ConfigSource: "directory",
				Directory:    "testdata/client",
			}, c, SelectorNetworkServer)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to create new client: %s", err)
			}

			res, err := cl.HandleJoinRequest(ctx, tc.NetID, nil, tc.Request)
			if a.So(tc.ErrorAssertion(a, err), should.BeTrue) {
				a.So(tc.ResponseAssertion(a, res), should.BeTrue)
			} else if err != nil {
				t.Errorf("Received unexpected error: %v", errors.Stack(err))
			}
		})
	}
}
