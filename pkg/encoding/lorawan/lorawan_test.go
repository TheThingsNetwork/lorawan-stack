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

package lorawan_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMessageEncodingSymmetricity(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Message *ttnpb.Message
	}{
		{
			Name: "Uplink/Unconfirmed",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_UP,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x1, 0x2, 0x3, 0x4},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: test.DefaultDevAddr,
							FCtrl: &ttnpb.FCtrl{
								Adr:       true,
								AdrAckReq: true,
								ClassB:    true,
							},
							FCnt:  0x1234,
							FOpts: []byte{0x42, 0xff},
						},
						FPort:      42,
						FrmPayload: []byte{0x11, 0x22, 0x33},
					},
				},
			},
		},
		{
			Name: "Uplink/Confirmed",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_UP,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x4, 0x3, 0x2, 0x1},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: test.DefaultDevAddr,
							FCtrl: &ttnpb.FCtrl{
								Adr:       true,
								AdrAckReq: true,
								Ack:       true,
								ClassB:    true,
							},
							FCnt:  0x4321,
							FOpts: []byte{0xff, 0x42},
						},
						FPort:      42,
						FrmPayload: []byte{0x11, 0x22, 0x33, 0x44},
					},
				},
			},
		},
		{
			Name: "Downlink/Unconfirmed",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x1, 0x2, 0x3, 0x4},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: test.DefaultDevAddr,
							FCtrl: &ttnpb.FCtrl{
								FPending: true,
							},
							FCnt:  0x1234,
							FOpts: []byte{0x42, 0xff},
						},
						FPort:      42,
						FrmPayload: []byte{0x11, 0x22, 0x33},
					},
				},
			},
		},
		{
			Name: "Downlink/Confirmed",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x4, 0x3, 0x2, 0x1},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: test.DefaultDevAddr,
							FCtrl: &ttnpb.FCtrl{
								Ack: true,
							},
							FCnt:  0x4321,
							FOpts: []byte{0xff, 0x42},
						},
						FPort:      42,
						FrmPayload: []byte{0x11, 0x22, 0x33, 0x44},
					},
				},
			},
		},
		{
			Name: "JoinRequest",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_JOIN_REQUEST,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x1, 0x2, 0x3, 0x4},
				Payload: &ttnpb.Message_JoinRequestPayload{
					JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEui:  test.DefaultJoinEUI,
						DevEui:   test.DefaultDevEUI,
						DevNonce: test.DefaultDevNonce,
					},
				},
			},
		},
		{
			Name: "RejoinRequest/Type0",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_REJOIN_REQUEST,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x1, 0x2, 0x3, 0x4},
				Payload: &ttnpb.Message_RejoinRequestPayload{
					RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
						RejoinType: ttnpb.RejoinRequestType_CONTEXT,
						NetId:      test.DefaultNetID,
						DevEui:     test.DefaultDevEUI,
						RejoinCnt:  0xff42,
					},
				},
			},
		},
		{
			Name: "RejoinRequest/Type1",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_REJOIN_REQUEST,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x1, 0x2, 0x3, 0x4},
				Payload: &ttnpb.Message_RejoinRequestPayload{
					RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
						RejoinType: ttnpb.RejoinRequestType_SESSION,
						JoinEui:    test.DefaultJoinEUI,
						DevEui:     test.DefaultDevEUI,
						RejoinCnt:  0x42,
					},
				},
			},
		},
		{
			Name: "RejoinRequest/Type0",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_REJOIN_REQUEST,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Mic: []byte{0x1, 0x2, 0x3, 0x4},
				Payload: &ttnpb.Message_RejoinRequestPayload{
					RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
						RejoinType: ttnpb.RejoinRequestType_KEYS,
						NetId:      test.DefaultNetID,
						DevEui:     test.DefaultDevEUI,
						RejoinCnt:  0x42ff,
					},
				},
			},
		},
		{
			Name: "JoinAccept",
			Message: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_JOIN_ACCEPT,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_JoinAcceptPayload{
					JoinAcceptPayload: &ttnpb.JoinAcceptPayload{
						Encrypted: random.Bytes(16),
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			b, err := MarshalMessage(*tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(b, should.NotBeNil)

			ret, err := AppendMessage(make([]byte, 0), *tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(ret, should.Resemble, b)

			msg := &ttnpb.Message{}
			err = UnmarshalMessage(b, msg)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(msg, should.Resemble, tc.Message)
		})
	}
}

func TestLoRaWANEncodingRaw(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Message *ttnpb.Message
		Bytes   []byte
	}{
		{
			"JoinRequest",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: 0},
				Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
					JoinEui:  types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEui:   types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevNonce: types.DevNonce{0x42, 0xff},
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x00,

				/* MACPayload */
				/** JoinEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** DevNonce **/
				0xff, 0x42,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			"JoinAccept",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_JOIN_ACCEPT, Major: 0},
				Payload: &ttnpb.Message_JoinAcceptPayload{JoinAcceptPayload: &ttnpb.JoinAcceptPayload{
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
			"Uplink/Unconfirmed",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: 0},
				Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: &ttnpb.FCtrl{
							Adr:       true,
							AdrAckReq: false,
							Ack:       true,
							ClassB:    true,
							FPending:  false,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FrmPayload: []byte{0xfe, 0xff},
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x40,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0xff, 0xff, 0xff, 0x42,
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
			"Downlink/Unconfirmed",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_DOWN, Major: 0},
				Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: &ttnpb.FCtrl{
							Adr:       true,
							AdrAckReq: false,
							Ack:       true,
							ClassB:    false,
							FPending:  true,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FrmPayload: []byte{0xfe, 0xff},
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x60,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0xff, 0xff, 0xff, 0x42,
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
			"Downlink/Unconfirmed/no FPort",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_DOWN, Major: 0},
				Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: &ttnpb.FCtrl{
							Adr:       true,
							AdrAckReq: false,
							Ack:       true,
							ClassB:    false,
							FPending:  true,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x60,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0xff, 0xff, 0xff, 0x42,
				/*** FCtrl ***/
				0xb2,
				/*** FCnt ***/
				0x42, 0xff,
				/*** FOpts ***/
				0xfe, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			"Downlink/Confirmed",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_CONFIRMED_UP, Major: 0},
				Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: &ttnpb.FCtrl{
							Adr:       true,
							AdrAckReq: false,
							Ack:       true,
							ClassB:    true,
							FPending:  false,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FrmPayload: []byte{0xfe, 0xff},
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0x80,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0xff, 0xff, 0xff, 0x42,
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
			"Downlink/Confirmed",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_CONFIRMED_DOWN, Major: 0},
				Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						FCtrl: &ttnpb.FCtrl{
							Adr:       true,
							AdrAckReq: false,
							Ack:       true,
							ClassB:    false,
							FPending:  true,
						},
						FCnt:  0xff42,
						FOpts: []byte{0xfe, 0xff},
					},
					FPort:      0x42,
					FrmPayload: []byte{0xfe, 0xff},
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xa0,

				/* MACPayload */

				/** FHDR **/
				/*** DevAddr ***/
				0xff, 0xff, 0xff, 0x42,
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
			"RejoinRequest/Type0",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_REJOIN_REQUEST, Major: 0},
				Payload: &ttnpb.Message_RejoinRequestPayload{RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
					RejoinType: 0,
					NetId:      types.NetID{0x42, 0xff, 0xff},
					DevEui:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					RejoinCnt:  0xff42,
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xc0,

				/* MACPayload */
				/** RejoinType **/
				0x00,
				/** NetID **/
				0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** RejoinCnt **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			"RejoinRequest/Type1",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_REJOIN_REQUEST, Major: 0},
				Payload: &ttnpb.Message_RejoinRequestPayload{RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
					RejoinType: 1,
					JoinEui:    types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevEui:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					RejoinCnt:  0xff42,
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xc0,

				/* MACPayload */
				/** RejoinType **/
				0x01,
				/** JoinEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** RejoinCnt **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
		{
			"RejoinRequest/Type2",
			&ttnpb.Message{
				MHdr: &ttnpb.MHDR{MType: ttnpb.MType_REJOIN_REQUEST, Major: 0},
				Payload: &ttnpb.Message_RejoinRequestPayload{RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
					RejoinType: 2,
					NetId:      types.NetID{0x42, 0xff, 0xff},
					DevEui:     types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					RejoinCnt:  0xff42,
				}},
				Mic: []byte{0x42, 0xff, 0xff, 0xff},
			},
			[]byte{
				/* MHDR */
				0xc0,

				/* MACPayload */
				/** RejoinType **/
				0x02,
				/** NetID **/
				0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** RejoinCnt **/
				0x42, 0xff,

				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			b, err := MarshalMessage(*tc.Message)
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)
			a.So(b, should.Resemble, tc.Bytes)

			ret, err := AppendMessage(make([]byte, 0), *tc.Message)
			a.So(err, should.BeNil)
			a.So(ret, should.Resemble, b)

			msg := &ttnpb.Message{}
			err = UnmarshalMessage(b, msg)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(msg, should.Resemble, tc.Message)
		})
	}
}

func TestUnmarshalIdentifiers(t *testing.T) {
	devAddr := types.DevAddr{0x42, 0xff, 0xff, 0xff}
	devEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	joinEUI := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	for i, tc := range []struct {
		Bytes       []byte
		Identifiers *ttnpb.EndDeviceIdentifiers
	}{
		{
			[]byte{
				/* MHDR: Join-request */
				0x00,
				/* MACPayload */
				/** JoinEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
				/** DevNonce **/
				0xff, 0x42,
				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
			&ttnpb.EndDeviceIdentifiers{
				JoinEui: &joinEUI,
				DevEui:  &devEUI,
			},
		},
		{
			[]byte{
				/* MHDR: Confirmed up */
				0x80,
				/* MACPayload */
				/** FHDR **/
				/*** DevAddr ***/
				0xff, 0xff, 0xff, 0x42,
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
			&ttnpb.EndDeviceIdentifiers{
				DevAddr: &devAddr,
			},
		},
		{
			[]byte{
				/* MHDR: Rejoin-request */
				0xc0,
				/* MACPayload */
				/** RejoinType **/
				0x00,
				/** NetID **/
				0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
				/** RejoinCnt **/
				0x42, 0xff,
				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
			&ttnpb.EndDeviceIdentifiers{
				DevEui: &devEUI,
			},
		},
		{
			[]byte{
				/* MHDR: Rejoin-request */
				0xc0,
				/* MACPayload */
				/** RejoinType **/
				0x01,
				/** JoinEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
				/** RejoinCnt **/
				0x42, 0xff,
				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
			&ttnpb.EndDeviceIdentifiers{
				JoinEui: &joinEUI,
				DevEui:  &devEUI,
			},
		},
		{
			[]byte{
				/* MHDR: Rejoin-request */
				0xc0,
				/* MACPayload */
				/** RejoinType **/
				0x02,
				/** NetID **/
				0xff, 0xff, 0x42,
				/** DevEUI **/
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
				/** RejoinCnt **/
				0x42, 0xff,
				/* MIC */
				0x42, 0xff, 0xff, 0xff,
			},
			&ttnpb.EndDeviceIdentifiers{
				DevEui: &devEUI,
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			ids, err := GetUplinkMessageIdentifiers(tc.Bytes)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(&ids, should.Resemble, tc.Identifiers)
		})
	}
}

func TestUnmarshalResilience(t *testing.T) {
	for i, tc := range [][]byte{
		// Too little data: FHDR is at least 7 bytes.
		{
			byte(ttnpb.MType_UNCONFIRMED_UP)<<5 | byte(ttnpb.Major_LORAWAN_R1),
			0x01, 0x02,
		},
		// Too little data: FHDR is at least 7 bytes.
		{
			byte(ttnpb.MType_UNCONFIRMED_DOWN)<<5 | byte(ttnpb.Major_LORAWAN_R1),
			0x01, 0x02,
		},
		// Too little data: no join-request payload.
		{
			byte(ttnpb.MType_JOIN_REQUEST)<<5 | byte(ttnpb.Major_LORAWAN_R1),
		},
		// Too little data: too little join-request payload.
		{
			byte(ttnpb.MType_JOIN_REQUEST)<<5 | byte(ttnpb.Major_LORAWAN_R1),
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		},
		// Too little data: no rejoin-request type.
		{
			byte(ttnpb.MType_REJOIN_REQUEST)<<5 | byte(ttnpb.Major_LORAWAN_R1),
		},
		// Too little data: too little rejoin-request payload.
		{
			byte(ttnpb.MType_REJOIN_REQUEST)<<5 | byte(ttnpb.Major_LORAWAN_R1),
			0x02,
		},
		// Too little data: too little join-accept payload.
		{
			byte(ttnpb.MType_JOIN_ACCEPT)<<5 | byte(ttnpb.Major_LORAWAN_R1),
			0x01, 0x02, 0x03, 0x04,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			a.So(func() {
				var msg ttnpb.Message
				err := UnmarshalMessage(tc, &msg)
				a.So(err, should.NotBeNil)
			}, should.NotPanic)
			a.So(func() {
				_, err := GetUplinkMessageIdentifiers(tc)
				a.So(err, should.NotBeNil)
			}, should.NotPanic)
		})
	}

	t.Run("Downlink without FPort", func(t *testing.T) {
		a := assertions.New(t)
		downlink := &ttnpb.DownlinkMessage{Payload: &ttnpb.Message{}}
		payload := []byte{0x60, 0x29, 0x2e, 0x01, 0x26, 0x20, 0x03, 0x00, 0xd0, 0x36, 0x78, 0xbd}
		err := UnmarshalMessage(payload, downlink.Payload)
		a.So(err, should.BeNil)
	})
}

func TestMessageEncodingSymmetricityJoinAcceptPayload(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Message *ttnpb.JoinAcceptPayload
	}{
		{
			Name: "JoinAcceptPayload/CFList",
			Message: &ttnpb.JoinAcceptPayload{
				JoinNonce: test.DefaultJoinNonce,
				NetId:     test.DefaultNetID,
				DevAddr:   test.DefaultDevAddr,
				DlSettings: &ttnpb.DLSettings{
					Rx1DrOffset: ttnpb.DataRateOffset_DATA_RATE_OFFSET_2,
					Rx2Dr:       ttnpb.DATA_RATE_1,
					OptNeg:      false,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{
						867100,
						867300,
						867500,
						867700,
						867900,
					},
				},
			},
		},
		{
			Name: "JoinAcceptPayload/NoCFList",
			Message: &ttnpb.JoinAcceptPayload{
				JoinNonce: test.DefaultJoinNonce,
				NetId:     test.DefaultNetID,
				DevAddr:   test.DefaultDevAddr,
				DlSettings: &ttnpb.DLSettings{
					Rx1DrOffset: ttnpb.DataRateOffset_DATA_RATE_OFFSET_2,
					Rx2Dr:       ttnpb.DATA_RATE_1,
					OptNeg:      false,
				},
				RxDelay: ttnpb.RxDelay_RX_DELAY_5,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			b, err := MarshalJoinAcceptPayload(*tc.Message)
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)

			ret, err := AppendJoinAcceptPayload(make([]byte, 0), *tc.Message)
			a.So(err, should.BeNil)
			a.So(ret, should.Resemble, b)

			msg := &ttnpb.JoinAcceptPayload{}
			err = UnmarshalJoinAcceptPayload(b, msg)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(msg, should.Resemble, tc.Message)
		})
	}
}

func TestLoRaWANEncodingRawJoinAcceptPayload(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Message *ttnpb.JoinAcceptPayload
		Bytes   []byte
	}{
		{
			"JoinAcceptPayload/NoCFList",
			&ttnpb.JoinAcceptPayload{
				JoinNonce: types.JoinNonce{0x42, 0xff, 0xff},
				NetId:     types.NetID{0x42, 0xff, 0xff},
				DevAddr:   types.DevAddr{0x42, 0xff, 0xff, 0xff},
				DlSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			[]byte{
				/* JoinNonce */
				0xff, 0xff, 0x42,
				/* NetID */
				0xff, 0xff, 0x42,
				/* DevAddr */
				0xff, 0xff, 0xff, 0x42,
				/* DLSettings */
				0xef,
				/* RxDelay */
				0x42,
			},
		},
		{
			"JoinAcceptPayload/CFListFreq",
			&ttnpb.JoinAcceptPayload{
				JoinNonce: types.JoinNonce{0x42, 0xff, 0xff},
				NetId:     types.NetID{0x42, 0xff, 0xff},
				DevAddr:   types.DevAddr{0x42, 0xff, 0xff, 0xff},
				DlSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{0xffff42, 0xffffff, 0xffffff, 0xffffff},
				},
			},
			[]byte{
				/* JoinNonce */
				0xff, 0xff, 0x42,
				/* NetID */
				0xff, 0xff, 0x42,
				/* DevAddr */
				0xff, 0xff, 0xff, 0x42,
				/* DLSettings */
				0xef,
				/* RxDelay */
				0x42,
				/* CFList */
				0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0,
				/* CFListType */
				0x0,
			},
		},
		{
			"JoinAcceptPayload/CFListChMask",
			&ttnpb.JoinAcceptPayload{
				JoinNonce: types.JoinNonce{0x42, 0xff, 0xff},
				NetId:     types.NetID{0x42, 0xff, 0xff},
				DevAddr:   types.DevAddr{0x42, 0xff, 0xff, 0xff},
				DlSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x6,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
				CfList: &ttnpb.CFList{
					Type: ttnpb.CFListType_CHANNEL_MASKS,
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
				0xff, 0xff, 0x42,
				/* NetID */
				0xff, 0xff, 0x42,
				/* DevAddr */
				0xff, 0xff, 0xff, 0x42,
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
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			b, err := MarshalJoinAcceptPayload(*tc.Message)
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)
			a.So(b, should.Resemble, tc.Bytes)

			b, err = AppendJoinAcceptPayload(make([]byte, 0), *tc.Message)
			a.So(err, should.BeNil)
			a.So(b, should.Resemble, tc.Bytes)

			msg := &ttnpb.JoinAcceptPayload{}
			err = UnmarshalJoinAcceptPayload(b, msg)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(msg, should.Resemble, tc.Message)
		})
	}
}

func TestDeviceEIRPToFloat32(t *testing.T) {
	for _, tc := range []struct {
		Enum  ttnpb.DeviceEIRP
		Float float32
	}{
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_36, Float: 36},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_33, Float: 33},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_30, Float: 30},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_29, Float: 29},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_27, Float: 27},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_26, Float: 26},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_24, Float: 24},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_21, Float: 21},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_20, Float: 20},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_18, Float: 18},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_16, Float: 16},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_14, Float: 14},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_13, Float: 13},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_12, Float: 12},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_10, Float: 10},
		{Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_8, Float: 8},
	} {
		t.Run(fmt.Sprintf("%v", tc.Float), func(t *testing.T) {
			assertions.New(t).So(DeviceEIRPToFloat32(tc.Enum), should.Equal, tc.Float)
		})
	}
}

func TestFloat32ToDeviceEIRP(t *testing.T) {
	for _, tc := range []struct {
		Float float32
		Enum  ttnpb.DeviceEIRP
	}{
		{Float: 38, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_36},
		{Float: 37, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_36},
		{Float: 36, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_36},
		{Float: 35, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_33},
		{Float: 33, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_33},
		{Float: 30, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_30},
		{Float: 29, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_29},
		{Float: 27, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_27},
		{Float: 26, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_26},
		{Float: 24, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_24},
		{Float: 23, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_21},
		{Float: 22, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_21},
		{Float: 21, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_21},
		{Float: 20, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_20},
		{Float: 19, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_18},
		{Float: 18, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_18},
		{Float: 17, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_16},
		{Float: 16, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_16},
		{Float: 15, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_14},
		{Float: 14, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_14},
		{Float: 13, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_13},
		{Float: 12, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_12},
		{Float: 11, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_10},
		{Float: 10, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_10},
		{Float: 9, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_8},
		{Float: 8, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_8},
		{Float: 7, Enum: ttnpb.DeviceEIRP_DEVICE_EIRP_8},
	} {
		t.Run(fmt.Sprintf("%v", tc.Float), func(t *testing.T) {
			assertions.New(t).So(Float32ToDeviceEIRP(tc.Float), should.Equal, tc.Enum)
		})
	}
}

func TestADRAckLimitExponentToUint32(t *testing.T) {
	for _, tc := range []struct {
		Enum ttnpb.ADRAckLimitExponent
		Uint uint32
	}{
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768, Uint: 32768},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16384, Uint: 16384},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8192, Uint: 8192},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4096, Uint: 4096},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2048, Uint: 2048},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1024, Uint: 1024},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_512, Uint: 512},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_256, Uint: 256},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_128, Uint: 128},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64, Uint: 64},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32, Uint: 32},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16, Uint: 16},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8, Uint: 8},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4, Uint: 4},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2, Uint: 2},
		{Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1, Uint: 1},
	} {
		t.Run(fmt.Sprintf("%v", tc.Uint), func(t *testing.T) {
			assertions.New(t).So(ADRAckLimitExponentToUint32(tc.Enum), should.Equal, tc.Uint)
		})
	}
}

func TestUint32ToADRAckLimitExponent(t *testing.T) {
	for _, tc := range []struct {
		Uint uint32
		Enum ttnpb.ADRAckLimitExponent
	}{
		{Uint: 32769, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768},
		{Uint: 32768, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768},
		{Uint: 32767, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16384},
		{Uint: 16384, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16384},
		{Uint: 16383, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8192},
		{Uint: 8192, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8192},
		{Uint: 8191, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4096},
		{Uint: 4097, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4096},
		{Uint: 4096, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4096},
		{Uint: 4095, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2048},
		{Uint: 2049, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2048},
		{Uint: 2048, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2048},
		{Uint: 2047, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1024},
		{Uint: 1024, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1024},
		{Uint: 1023, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_512},
		{Uint: 512, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_512},
		{Uint: 511, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_256},
		{Uint: 256, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_256},
		{Uint: 255, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_128},
		{Uint: 128, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_128},
		{Uint: 127, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64},
		{Uint: 64, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64},
		{Uint: 63, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32},
		{Uint: 32, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32},
		{Uint: 31, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16},
		{Uint: 16, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16},
		{Uint: 15, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8},
		{Uint: 9, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8},
		{Uint: 8, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8},
		{Uint: 7, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4},
		{Uint: 6, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4},
		{Uint: 5, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4},
		{Uint: 4, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4},
		{Uint: 3, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2},
		{Uint: 2, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2},
		{Uint: 1, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1},
		{Uint: 0, Enum: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1},
	} {
		t.Run(fmt.Sprintf("%v", tc.Uint), func(t *testing.T) {
			assertions.New(t).So(Uint32ToADRAckLimitExponent(tc.Uint), should.Equal, tc.Enum)
		})
	}
}

func TestADRAckDelayExponentToUint32(t *testing.T) {
	for _, tc := range []struct {
		Enum ttnpb.ADRAckDelayExponent
		Uint uint32
	}{
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32768, Uint: 32768},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16384, Uint: 16384},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8192, Uint: 8192},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4096, Uint: 4096},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048, Uint: 2048},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024, Uint: 1024},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_512, Uint: 512},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_256, Uint: 256},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_128, Uint: 128},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_64, Uint: 64},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32, Uint: 32},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16, Uint: 16},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8, Uint: 8},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4, Uint: 4},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2, Uint: 2},
		{Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1, Uint: 1},
	} {
		t.Run(fmt.Sprintf("%v", tc.Uint), func(t *testing.T) {
			assertions.New(t).So(ADRAckDelayExponentToUint32(tc.Enum), should.Equal, tc.Uint)
		})
	}
}

func TestUint32ToADRAckDelayExponent(t *testing.T) {
	for _, tc := range []struct {
		Uint uint32
		Enum ttnpb.ADRAckDelayExponent
	}{
		{Uint: 32769, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32768},
		{Uint: 32768, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32768},
		{Uint: 32767, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16384},
		{Uint: 16384, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16384},
		{Uint: 16383, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8192},
		{Uint: 8192, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8192},
		{Uint: 8191, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4096},
		{Uint: 4097, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4096},
		{Uint: 4096, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4096},
		{Uint: 4095, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048},
		{Uint: 2049, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048},
		{Uint: 2048, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048},
		{Uint: 2047, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024},
		{Uint: 1024, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024},
		{Uint: 1023, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_512},
		{Uint: 512, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_512},
		{Uint: 511, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_256},
		{Uint: 256, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_256},
		{Uint: 255, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_128},
		{Uint: 128, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_128},
		{Uint: 127, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_64},
		{Uint: 64, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_64},
		{Uint: 63, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32},
		{Uint: 32, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32},
		{Uint: 31, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16},
		{Uint: 16, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16},
		{Uint: 15, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8},
		{Uint: 9, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8},
		{Uint: 8, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8},
		{Uint: 7, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4},
		{Uint: 6, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4},
		{Uint: 5, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4},
		{Uint: 4, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4},
		{Uint: 3, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2},
		{Uint: 2, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2},
		{Uint: 1, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1},
		{Uint: 0, Enum: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1},
	} {
		t.Run(fmt.Sprintf("%v", tc.Uint), func(t *testing.T) {
			assertions.New(t).So(Uint32ToADRAckDelayExponent(tc.Uint), should.Equal, tc.Enum)
		})
	}
}
