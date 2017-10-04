// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb_test

import (
	fmt "fmt"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func lorawanEncodingTestName(msg *Message) string {
	switch msg.MType {
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
		return fmt.Sprintf("RejoinRequest%d", msg.GetRejoinRequestPayload().RejoinType)
	}
	panic("unreachable")
}

func TestLoRaWANEncodingRandomized(t *testing.T) {
	r := test.Randy
	for _, expected := range []*Message{
		NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), false),
		NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), true),
		NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), false),
		NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), true),
		NewPopulatedMessageJoinRequest(test.Randy),
		NewPopulatedMessageJoinAccept(test.Randy, false),
		NewPopulatedMessageRejoinRequest(test.Randy, 0),
		NewPopulatedMessageRejoinRequest(test.Randy, 1),
		NewPopulatedMessageRejoinRequest(test.Randy, 2),
	} {
		t.Run(lorawanEncodingTestName(expected), func(t *testing.T) {
			a := assertions.New(t)

			b, err := expected.MarshalLoRaWAN()
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)

			pld := &Message{}
			a.So(pld.UnmarshalLoRaWAN(b), should.BeNil)
			a.So(pld, should.Resemble, expected)

			ret, err := pld.AppendLoRaWAN(make([]byte, 0))
			a.So(err, should.BeNil)
			a.So(ret, should.Resemble, b)
		})
	}
}

func TestLoRaWANEncodingRaw(t *testing.T) {
	for _, tc := range []struct {
		Message *Message
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
	} {
		t.Run(lorawanEncodingTestName(tc.Message), func(t *testing.T) {
			a := assertions.New(t)

			b, err := tc.Message.MarshalLoRaWAN()
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)
			a.So(b, should.Resemble, tc.Bytes)

			msg := &Message{}
			a.So(msg.UnmarshalLoRaWAN(b), should.BeNil)
			a.So(msg, should.Resemble, tc.Message)

			ret, err := tc.Message.AppendLoRaWAN(make([]byte, 0))
			a.So(err, should.BeNil)
			a.So(ret, should.Resemble, tc.Bytes)
		})
	}
}
