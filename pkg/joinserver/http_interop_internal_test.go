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

package joinserver

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type mockInteropHandler struct {
	HandleJoinFunc   func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetHomeNetIDFunc func(context.Context, types.EUI64, types.EUI64) (*types.NetID, error)
	GetAppSKeyFunc   func(context.Context, *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error)
}

func (h mockInteropHandler) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if h.HandleJoinFunc == nil {
		panic("HandleJoin should not be called")
	}
	return h.HandleJoinFunc(ctx, req)
}

func (h mockInteropHandler) GetHomeNetID(ctx context.Context, joinEUI, devEUI types.EUI64) (*types.NetID, error) {
	if h.GetHomeNetIDFunc == nil {
		panic("GetHomeNetID should not be called")
	}
	return h.GetHomeNetIDFunc(ctx, joinEUI, devEUI)
}

func (h mockInteropHandler) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	if h.GetAppSKeyFunc == nil {
		panic("GetAppSKey should not be called")
	}
	return h.GetAppSKeyFunc(ctx, req)
}

func TestInteropJoinRequest(t *testing.T) {
	for _, tc := range []struct {
		Name                string
		JoinReq             *interop.JoinReq
		ExpectedJoinRequest *ttnpb.JoinRequest
		ErrorAssertion      func(*testing.T, error) bool
		HandleJoinFunc      func() (*ttnpb.JoinResponse, error)
		ExpectedJoinAns     *interop.JoinAns
	}{
		{
			Name: "Normal/1.0.3",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_0_3),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMACVersion: ttnpb.MAC_V1_0_3,
				NetID:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: ttnpb.RX_DELAY_5,
				CFList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return &ttnpb.JoinResponse{
					RawPayload: []byte{0x1, 0x2, 0x3, 0x4},
					Lifetime:   1 * time.Hour,
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
					},
				}, nil
			},
			ExpectedJoinAns: &interop.JoinAns{
				JsNsMessageHeader: interop.JsNsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinAns,
					},
					SenderID:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					ReceiverID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: []byte{0x1, 0x2, 0x3, 0x4},
				Result:     interop.ResultSuccess,
				Lifetime:   3600,
				NwkSKey: &interop.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				AppSKey: &interop.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
			},
		},
		{
			Name: "Normal/1.1",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMACVersion: ttnpb.MAC_V1_1,
				NetID:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: ttnpb.RX_DELAY_5,
				CFList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return &ttnpb.JoinResponse{
					RawPayload: []byte{0x1, 0x2, 0x3, 0x4},
					Lifetime:   1 * time.Hour,
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						NwkSEncKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
					},
				}, nil
			},
			ExpectedJoinAns: &interop.JoinAns{
				JsNsMessageHeader: interop.JsNsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinAns,
					},
					SenderID:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					ReceiverID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: []byte{0x1, 0x2, 0x3, 0x4},
				Result:     interop.ResultSuccess,
				Lifetime:   3600,
				FNwkSIntKey: &interop.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				SNwkSIntKey: &interop.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				NwkSEncKey: &interop.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				AppSKey: &interop.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
			},
		},
		{
			Name: "Error/Decode",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMACVersion: ttnpb.MAC_V1_1,
				NetID:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: ttnpb.RX_DELAY_5,
				CFList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return nil, errDecodePayload
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrMalformedMessage)
			},
		},
		{
			Name: "Error/MIC",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMACVersion: ttnpb.MAC_V1_1,
				NetID:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: ttnpb.RX_DELAY_5,
				CFList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return nil, errMICMismatch
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrMIC)
			},
		},
		{
			Name: "Error/Join",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMACVersion: ttnpb.MAC_V1_1,
				NetID:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: ttnpb.RX_DELAY_5,
				CFList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return nil, errReuseDevNonce
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrJoinReq)
			},
		},
		{
			Name: "InvalidCFList",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0x42},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrMalformedMessage)
			},
		},
		{
			Name: "InvalidDLSettings",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				DevAddr:    types.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     types.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MAC_V1_1),
				DLSettings: interop.Buffer{0xef, 0xff},
				RxDelay:    ttnpb.RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrMalformedMessage)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := log.NewContext(test.Context(), test.GetLogger(t))
			a := assertions.New(t)

			srv := interopServer{
				JS: &mockInteropHandler{
					HandleJoinFunc: func(ctx context.Context, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
						if !a.So(req, should.Resemble, tc.ExpectedJoinRequest) {
							t.FailNow()
						}
						return tc.HandleJoinFunc()
					},
				},
			}

			ans, err := srv.JoinRequest(ctx, tc.JoinReq)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			} else if tc.ErrorAssertion != nil {
				t.Fatal("Expected error")
			}

			a.So(ans, should.Resemble, tc.ExpectedJoinAns)
		})
	}
}

func TestInteropHomeNSRequest(t *testing.T) {
	for _, tc := range []struct {
		Name              string
		HomeNSReq         *interop.HomeNSReq
		ExpectedJoinEUI   types.EUI64
		ExpectedDevEUI    types.EUI64
		ErrorAssertion    func(*testing.T, error) bool
		GetNetIDFunc      func() (*types.NetID, error)
		ExpectedHomeNSAns *interop.HomeNSAns
	}{
		{
			Name: "Normal",
			HomeNSReq: &interop.HomeNSReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				DevEUI: types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			ExpectedJoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			GetNetIDFunc: func() (*types.NetID, error) {
				return &types.NetID{0x42, 0xff, 0xff}, nil
			},
			ExpectedHomeNSAns: &interop.HomeNSAns{
				JsNsMessageHeader: interop.JsNsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinAns,
					},
					SenderID:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ReceiverID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverNSID: types.NetID{0x0, 0x0, 0x13},
				},
				HNSID:  types.NetID{0x42, 0xff, 0xff},
				HNetID: types.NetID{0x42, 0xff, 0xff},
			},
		},
		{
			Name: "NoNetID",
			HomeNSReq: &interop.HomeNSReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: "1.0",
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   types.NetID{0x0, 0x0, 0x13},
					ReceiverID: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					SenderNSID: types.NetID{0x0, 0x0, 0x13},
				},
				DevEUI: types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			ExpectedJoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			GetNetIDFunc: func() (*types.NetID, error) {
				return nil, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrActivation)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := log.NewContext(test.Context(), test.GetLogger(t))
			a := assertions.New(t)

			srv := interopServer{
				JS: &mockInteropHandler{
					GetHomeNetIDFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64) (*types.NetID, error) {
						if !a.So(joinEUI, should.Resemble, tc.ExpectedJoinEUI) || !a.So(devEUI, should.Resemble, tc.ExpectedDevEUI) {
							t.FailNow()
						}
						return tc.GetNetIDFunc()
					},
				},
			}

			ans, err := srv.HomeNSRequest(ctx, tc.HomeNSReq)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			} else if tc.ErrorAssertion != nil {
				t.Fatal("Expected error")
			}

			a.So(ans, should.Resemble, tc.ExpectedHomeNSAns)
		})
	}
}
