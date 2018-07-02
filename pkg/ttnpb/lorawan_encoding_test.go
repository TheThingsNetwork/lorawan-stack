// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	. "go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func lorawanEncodingTestName(v interface{}) string {
	switch v := v.(type) {
	case *Message:
		switch v.MType {
		case MType_UNCONFIRMED_UP:
			return "Uplink(Unconfirmed)"
		case MType_UNCONFIRMED_DOWN:
			return "Downlink(Unconfirmed)"
		case MType_CONFIRMED_UP:
			return "Uplink(Confirmed)"
		case MType_CONFIRMED_DOWN:
			return "Downlink(Confirmed)"
		case MType_JOIN_REQUEST:
			return "JoinRequest"
		case MType_JOIN_ACCEPT:
			return "JoinAccept(Encrypted)"
		case MType_REJOIN_REQUEST:
			return fmt.Sprintf("RejoinRequest%d", v.GetRejoinRequestPayload().RejoinType)
		}
	case *JoinAcceptPayload:
		if v.CFList == nil {
			return "JoinAcceptPayload(no CFList)"
		}
		return fmt.Sprintf("JoinAcceptPayload(CFListType %d)", v.CFList.Type)
	}
	panic("Unmatched type")
}

type message interface {
	lorawan.Marshaler
	lorawan.Appender
	lorawan.Unmarshaler
}

func TestLoRaWANEncodingRandomized(t *testing.T) {
	r := test.Randy

	for i, expected := range []message{
		NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), false),
		NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), true),
		NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), false),
		NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), true),
		NewPopulatedMessageJoinRequest(r),
		NewPopulatedMessageJoinAccept(r, false),
		NewPopulatedMessageRejoinRequest(r, 0),
		NewPopulatedMessageRejoinRequest(r, 1),
		NewPopulatedMessageRejoinRequest(r, 2),

		NewPopulatedJoinAcceptPayload(r, false),
	} {
		t.Run(fmt.Sprintf("%d/%s", i, lorawanEncodingTestName(expected)), func(t *testing.T) {
			a := assertions.New(t)

			b, err := expected.MarshalLoRaWAN()
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)

			ret, err := expected.AppendLoRaWAN(make([]byte, 0))
			a.So(err, should.BeNil)
			a.So(ret, should.Resemble, b)

			msg := reflect.New(reflect.Indirect(reflect.ValueOf(expected)).Type()).Interface().(lorawan.Unmarshaler)
			a.So(msg.UnmarshalLoRaWAN(b), should.BeNil)
			if !a.So(msg, should.Resemble, expected) {
				pretty.Ldiff(t, msg, expected)
			}
		})
	}
}

func TestLoRaWANEncodingRaw(t *testing.T) {
	for i, tc := range []struct {
		Message message
		Bytes   []byte
	}{
		{
			&Message{
				MHDR: MHDR{MType: MType_JOIN_REQUEST, Major: 0},
				Payload: &Message_JoinRequestPayload{&JoinRequestPayload{
					JoinEUI:  types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEUI:   types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevNonce: types.DevNonce{0x42, 0xff},
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x00,

				/* MACPayload */
				/** JoinEUI **/
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/** DevEUI **/
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/** DevNonce **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_JOIN_ACCEPT, Major: 0},
				Payload: &Message_JoinAcceptPayload{&JoinAcceptPayload{
					Encrypted: []byte{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				}},
			},
			[]byte{
				/* MHDR */
				0x20,
				/* Encrypted */
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_UNCONFIRMED_UP, Major: 0},
				Payload: &Message_MACPayload{&MACPayload{
					FHDR: FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: FCtrl{
							ADR:       true,
							ADRAckReq: false,
							Ack:       true,
							ClassB:    true,
							FPending:  false,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FRMPayload: []byte{0xfe, 0xff},
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x40,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0x42, 0xff, 0xff, 0xff,
				/*** FCtrl ***/
				0xb2,
				/*** FCnt ***/
				0x42, 0xff,
				/*** FOpts ***/
				0xfe, 0xff,

				/** FPort **/
				0x42,

				/** FRMPayload **/
				0xfe, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_UNCONFIRMED_DOWN, Major: 0},
				Payload: &Message_MACPayload{&MACPayload{
					FHDR: FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: FCtrl{
							ADR:       true,
							ADRAckReq: false,
							Ack:       true,
							ClassB:    false,
							FPending:  true,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FRMPayload: []byte{0xfe, 0xff},
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x60,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0x42, 0xff, 0xff, 0xff,
				/*** FCtrl ***/
				0xb2,
				/*** FCnt ***/
				0x42, 0xff,
				/*** FOpts ***/
				0xfe, 0xff,

				/** FPort **/
				0x42,

				/** FRMPayload **/
				0xfe, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_CONFIRMED_UP, Major: 0},
				Payload: &Message_MACPayload{&MACPayload{
					FHDR: FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: FCtrl{
							ADR:       true,
							ADRAckReq: false,
							Ack:       true,
							ClassB:    true,
							FPending:  false,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FRMPayload: []byte{0xfe, 0xff},
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x80,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0x42, 0xff, 0xff, 0xff,
				/*** FCtrl ***/
				0xb2,
				/*** FCnt ***/
				0x42, 0xff,
				/*** FOpts ***/
				0xfe, 0xff,

				/** FPort **/
				0x42,

				/** FRMPayload **/
				0xfe, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_CONFIRMED_DOWN, Major: 0},
				Payload: &Message_MACPayload{&MACPayload{
					FHDR: FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: FCtrl{
							ADR:       true,
							ADRAckReq: false,
							Ack:       true,
							ClassB:    false,
							FPending:  true,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FRMPayload: []byte{0xfe, 0xff},
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xa0,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0x42, 0xff, 0xff, 0xff,
				/*** FCtrl ***/
				0xb2,
				/*** FCnt ***/
				0x42, 0xff,
				/*** FOpts ***/
				0xfe, 0xff,

				/** FPort **/
				0x42,

				/** FRMPayload **/
				0xfe, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_REJOIN_REQUEST, Major: 0},
				Payload: &Message_RejoinRequestPayload{&RejoinRequestPayload{
					RejoinType: 0,
					NetID:      types.NetID{0x42, 0xff, 0xff},
					DevEUI:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					RejoinCnt:  0xff42,
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xc0,

				/* MACPayload */
				/** RejoinType **/
				0x00,
				/** NetID **/
				0x42, 0xff, 0xff,
				/** DevEUI **/
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/** RejoinCnt **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_REJOIN_REQUEST, Major: 0},
				Payload: &Message_RejoinRequestPayload{&RejoinRequestPayload{
					RejoinType: 1,
					JoinEUI:    types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEUI:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					RejoinCnt:  0xff42,
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xc0,

				/* MACPayload */
				/** RejoinType **/
				0x01,
				/** JoinEUI **/
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/** DevEUI **/
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/** RejoinCnt **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&Message{
				MHDR: MHDR{MType: MType_REJOIN_REQUEST, Major: 0},
				Payload: &Message_RejoinRequestPayload{&RejoinRequestPayload{
					RejoinType: 2,
					NetID:      types.NetID{0x42, 0xff, 0xff},
					DevEUI:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					RejoinCnt:  0xff42,
				}},
				MIC: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xc0,

				/* MACPayload */
				/** RejoinType **/
				0x02,
				/** NetID **/
				0x42, 0xff, 0xff,
				/** DevEUI **/
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/** RejoinCnt **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			&JoinAcceptPayload{
				JoinNonce: types.JoinNonce{0x42, 0xff, 0xff},
				NetID:     types.NetID{0x42, 0xff, 0xff},
				DevAddr:   types.DevAddr{0x42, 0xff, 0xff, 0xff},
				DLSettings: DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			[]byte{
				/* JoinNonce */
				0x42, 0xff, 0xff,
				/* NetID */
				0x42, 0xff, 0xff,
				/* DevAddr */
				0x42, 0xff, 0xff, 0xff,
				/* DLSettings */
				0xef,
				/* RxDelay */
				0x42,
			},
		},
		{
			&JoinAcceptPayload{
				JoinNonce: types.JoinNonce{0x42, 0xff, 0xff},
				NetID:     types.NetID{0x42, 0xff, 0xff},
				DevAddr:   types.DevAddr{0x42, 0xff, 0xff, 0xff},
				DLSettings: DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList: &CFList{
					Type: CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			[]byte{
				/* JoinNonce */
				0x42, 0xff, 0xff,
				/* NetID */
				0x42, 0xff, 0xff,
				/* DevAddr */
				0x42, 0xff, 0xff, 0xff,
				/* DLSettings */
				0xef,
				/* RxDelay */
				0x42,
				/* CFList */
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				/* CFListType */
				0x0,
			},
		},
		{
			&JoinAcceptPayload{
				JoinNonce: types.JoinNonce{0x42, 0xff, 0xff},
				NetID:     types.NetID{0x42, 0xff, 0xff},
				DevAddr:   types.DevAddr{0x42, 0xff, 0xff, 0xff},
				DLSettings: DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x6,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList: &CFList{
					Type: CFListType_CHANNEL_MASKS,
					ChMasks: []bool{
						false, true, false, false, false, false, true, false,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			[]byte{
				/* JoinNonce */
				0x42, 0xff, 0xff,
				/* NetID */
				0x42, 0xff, 0xff,
				/* DevAddr */
				0x42, 0xff, 0xff, 0xff,
				/* DLSettings */
				0xef,
				/* RxDelay */
				0x42,
				/* CFList */
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0,
				/* CFListType */
				0x1,
			},
		},
	} {
		t.Run(fmt.Sprintf("%d/%s", i, lorawanEncodingTestName(tc.Message)), func(t *testing.T) {
			a := assertions.New(t)

			b, err := tc.Message.MarshalLoRaWAN()
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)
			a.So(b, should.Resemble, tc.Bytes)

			b, err = tc.Message.AppendLoRaWAN(make([]byte, 0))
			a.So(err, should.BeNil)
			a.So(b, should.Resemble, tc.Bytes)

			msg := reflect.New(reflect.Indirect(reflect.ValueOf(tc.Message)).Type()).Interface().(lorawan.Unmarshaler)
			a.So(msg.UnmarshalLoRaWAN(b), should.BeNil)
			a.So(msg, should.Resemble, tc.Message)
		})
	}
}
