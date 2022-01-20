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

package joinserver

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockInteropHandler struct {
	HandleJoinFunc     func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetHomeNetworkFunc func(context.Context, types.EUI64, types.EUI64) (*EndDeviceHomeNetwork, error)
	GetAppSKeyFunc     func(context.Context, *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error)
}

func (h mockInteropHandler) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, authorizer Authorizer) (*ttnpb.JoinResponse, error) {
	if h.HandleJoinFunc == nil {
		panic("HandleJoin should not be called")
	}
	return h.HandleJoinFunc(ctx, req)
}

func (h mockInteropHandler) GetHomeNetwork(ctx context.Context, joinEUI, devEUI types.EUI64, authorizer Authorizer) (*EndDeviceHomeNetwork, error) {
	if h.GetHomeNetworkFunc == nil {
		panic("GetHomeNetwork should not be called")
	}
	return h.GetHomeNetworkFunc(ctx, joinEUI, devEUI)
}

func (h mockInteropHandler) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest, authorizer Authorizer) (*ttnpb.AppSKeyResponse, error) {
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
			Name: "Normal/TS001-1.0.3",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_0_3),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_0_3,
				NetId:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return &ttnpb.JoinResponse{
					RawPayload: []byte{0x1, 0x2, 0x3, 0x4},
					Lifetime:   ttnpb.ProtoDurationPtr(1 * time.Hour),
					SessionKeys: &ttnpb.SessionKeys{
						SessionKeyId: []byte{0x1, 0x2, 0x3, 0x4},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KekLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
					},
				}, nil
			},
			ExpectedJoinAns: &interop.JoinAns{
				JsNsMessageHeader: interop.JsNsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinAns,
					},
					SenderID:   interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					ReceiverID: interop.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4},
				Result: interop.Result{
					ResultCode: interop.ResultSuccess,
				},
				Lifetime: 3600,
				NwkSKey: &interop.KeyEnvelope{
					KekLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				AppSKey: &interop.KeyEnvelope{
					KekLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				SessionKeyID: []byte{0x1, 0x2, 0x3, 0x4},
			},
		},
		{
			Name: "Normal/TS001-1.1",
			JoinReq: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_1,
				NetId:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return &ttnpb.JoinResponse{
					RawPayload: []byte{0x1, 0x2, 0x3, 0x4},
					Lifetime:   ttnpb.ProtoDurationPtr(1 * time.Hour),
					SessionKeys: &ttnpb.SessionKeys{
						SessionKeyId: []byte{0x1, 0x2, 0x3, 0x4},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							KekLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						NwkSEncKey: &ttnpb.KeyEnvelope{
							KekLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KekLabel:     "test",
							EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
						},
					},
				}, nil
			},
			ExpectedJoinAns: &interop.JoinAns{
				JsNsMessageHeader: interop.JsNsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinAns,
					},
					SenderID:   interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					ReceiverID: interop.NetID{0x0, 0x0, 0x13},
				},
				PHYPayload: []byte{0x1, 0x2, 0x3, 0x4},
				Result: interop.Result{
					ResultCode: interop.ResultSuccess,
				},
				Lifetime: 3600,
				FNwkSIntKey: &interop.KeyEnvelope{
					KekLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				SNwkSIntKey: &interop.KeyEnvelope{
					KekLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				NwkSEncKey: &interop.KeyEnvelope{
					KekLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				AppSKey: &interop.KeyEnvelope{
					KekLabel:     "test",
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
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_1,
				NetId:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return nil, errDecodePayload.New()
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
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_1,
				NetId:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return nil, errMICMismatch.New()
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
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
				CFList:     interop.Buffer{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0},
			},
			ExpectedJoinRequest: &ttnpb.JoinRequest{
				RawPayload:         []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:            types.DevAddr{0x1, 0x2, 0x3, 0x4},
				SelectedMacVersion: ttnpb.MACVersion_MAC_V1_1,
				NetId:              types.NetID{0x0, 0x0, 0x13},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			HandleJoinFunc: func() (*ttnpb.JoinResponse, error) {
				return nil, errReuseDevNonce.New()
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
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_1),
				DLSettings: interop.Buffer{0xef},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
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
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				PHYPayload: interop.Buffer{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x21, 0x22, 0x23},
				DevAddr:    interop.DevAddr{0x1, 0x2, 0x3, 0x4},
				DevEUI:     interop.EUI64{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_1),
				DLSettings: interop.Buffer{0xef, 0xff},
				RxDelay:    ttnpb.RxDelay_RX_DELAY_5,
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
		Name               string
		HomeNSReq          *interop.HomeNSReq
		ExpectedJoinEUI    types.EUI64
		ExpectedDevEUI     types.EUI64
		ErrorAssertion     func(*testing.T, error) bool
		GetHomeNetworkFunc func() (*EndDeviceHomeNetwork, error)
		ExpectedHomeNSAns  *interop.TTIHomeNSAns
	}{
		{
			Name: "Normal/TS002-1.0",
			HomeNSReq: &interop.HomeNSReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				DevEUI: interop.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			ExpectedJoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			GetHomeNetworkFunc: func() (*EndDeviceHomeNetwork, error) {
				return &EndDeviceHomeNetwork{
					NetID:                &types.NetID{0x42, 0xff, 0xff},
					NSID:                 &types.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0},
					TenantID:             "foo-tenant",
					NetworkServerAddress: "thethings.example.com",
				}, nil
			},
			ExpectedHomeNSAns: &interop.TTIHomeNSAns{
				HomeNSAns: interop.HomeNSAns{
					JsNsMessageHeader: interop.JsNsMessageHeader{
						MessageHeader: interop.MessageHeader{
							ProtocolVersion: interop.ProtocolV1_0,
							MessageType:     interop.MessageTypeJoinAns,
						},
						SenderID:   interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						ReceiverID: interop.NetID{0x0, 0x0, 0x13},
					},
					Result: interop.Result{
						ResultCode: interop.ResultSuccess,
					},
					HNetID: interop.NetID{0x42, 0xff, 0xff},
					// NOTE: HNSID is not returned as the field is not supported in LoRaWAN Backend Interfaces 1.0.
				},
				HTenantID:  "foo-tenant",
				HNSAddress: "thethings.example.com",
			},
		},
		{
			Name: "Normal/TS002-1.1",
			HomeNSReq: &interop.HomeNSReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_1,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					SenderNSID: &interop.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0xff},
					ReceiverID: interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				DevEUI: interop.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			ExpectedJoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			GetHomeNetworkFunc: func() (*EndDeviceHomeNetwork, error) {
				return &EndDeviceHomeNetwork{
					NetID: &types.NetID{0x42, 0xff, 0xff},
					NSID:  &types.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0},
				}, nil
			},
			ExpectedHomeNSAns: &interop.TTIHomeNSAns{
				HomeNSAns: interop.HomeNSAns{
					JsNsMessageHeader: interop.JsNsMessageHeader{
						MessageHeader: interop.MessageHeader{
							ProtocolVersion: interop.ProtocolV1_1,
							MessageType:     interop.MessageTypeJoinAns,
						},
						SenderID:     interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						ReceiverID:   interop.NetID{0x0, 0x0, 0x13},
						ReceiverNSID: &interop.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0xff},
					},
					Result: interop.Result{
						ResultCode: interop.ResultSuccess,
					},
					HNetID: interop.NetID{0x42, 0xff, 0xff},
					HNSID:  &interop.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
			},
		},
		{
			Name: "NoNetID",
			HomeNSReq: &interop.HomeNSReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x13},
					ReceiverID: interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				DevEUI: interop.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			ExpectedJoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ExpectedDevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			GetHomeNetworkFunc: func() (*EndDeviceHomeNetwork, error) {
				return &EndDeviceHomeNetwork{}, nil
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
					GetHomeNetworkFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64) (*EndDeviceHomeNetwork, error) {
						if !a.So(joinEUI, should.Resemble, tc.ExpectedJoinEUI) || !a.So(devEUI, should.Resemble, tc.ExpectedDevEUI) {
							t.FailNow()
						}
						return tc.GetHomeNetworkFunc()
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

func TestInteropAppSKeyRequest(t *testing.T) {
	errNotFound := errors.DefineNotFound("not_found", "not found")

	for _, tc := range []struct {
		Name                      string
		AppSKeyReq                *interop.AppSKeyReq
		ExpectedSessionKeyRequest *ttnpb.SessionKeyRequest
		ErrorAssertion            func(*testing.T, error) bool
		GetAppSKeyFunc            func() (*ttnpb.AppSKeyResponse, error)
		ExpectedAppSKeyAns        *interop.AppSKeyAns
	}{
		{
			Name: "Normal",
			AppSKeyReq: &interop.AppSKeyReq{
				AsJsMessageHeader: interop.AsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   "test.local",
					ReceiverID: interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				DevEUI:       interop.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: interop.Buffer{0x1, 0x2, 0x3, 0x4},
			},
			ExpectedSessionKeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyId: []byte{0x1, 0x2, 0x3, 0x4},
			},
			GetAppSKeyFunc: func() (*ttnpb.AppSKeyResponse, error) {
				return &ttnpb.AppSKeyResponse{
					AppSKey: &ttnpb.KeyEnvelope{
						KekLabel:     "test",
						EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
					},
				}, nil
			},
			ExpectedAppSKeyAns: &interop.AppSKeyAns{
				JsAsMessageHeader: interop.JsAsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinAns,
					},
					SenderID:   interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ReceiverID: "test.local",
				},
				Result: interop.Result{
					ResultCode: interop.ResultSuccess,
				},
				DevEUI: interop.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				AppSKey: interop.KeyEnvelope{
					KekLabel:     "test",
					EncryptedKey: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				},
				SessionKeyID: interop.Buffer{0x1, 0x2, 0x3, 0x4},
			},
		},
		{
			Name: "UnknownDevEUI",
			AppSKeyReq: &interop.AppSKeyReq{
				AsJsMessageHeader: interop.AsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						ProtocolVersion: interop.ProtocolV1_0,
						MessageType:     interop.MessageTypeJoinReq,
					},
					SenderID:   "test.local",
					ReceiverID: interop.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				DevEUI:       interop.EUI64{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42},
				SessionKeyID: interop.Buffer{0x1, 0x2, 0x3, 0x4},
			},
			ExpectedSessionKeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42},
				SessionKeyId: []byte{0x1, 0x2, 0x3, 0x4},
			},
			GetAppSKeyFunc: func() (*ttnpb.AppSKeyResponse, error) {
				return nil, errRegistryOperation.WithCause(errNotFound)
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.HaveSameErrorDefinitionAs, interop.ErrUnknownDevEUI)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := log.NewContext(test.Context(), test.GetLogger(t))
			a := assertions.New(t)

			srv := interopServer{
				JS: &mockInteropHandler{
					GetAppSKeyFunc: func(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
						if !a.So(req, should.Resemble, tc.ExpectedSessionKeyRequest) {
							t.FailNow()
						}
						return tc.GetAppSKeyFunc()
					},
				},
			}

			ans, err := srv.AppSKeyRequest(ctx, tc.AppSKeyReq)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			} else if tc.ErrorAssertion != nil {
				t.Fatal("Expected error")
			}

			a.So(ans, should.Resemble, tc.ExpectedAppSKeyAns)
		})
	}
}
