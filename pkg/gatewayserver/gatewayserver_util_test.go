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

package gatewayserver_test

import (
	"context"
	"crypto/rand"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var testRights = []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK, ttnpb.Right_RIGHT_GATEWAY_STATUS_READ}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

func randomJoinRequestPayload(joinEUI, devEUI types.EUI64) []byte {
	var nwkKey types.AES128Key
	rand.Read(nwkKey[:])
	var devNonce types.DevNonce
	rand.Read(devNonce[:])

	msg := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: ttnpb.MType_JOIN_REQUEST,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			Payload: &ttnpb.Message_JoinRequestPayload{
				JoinRequestPayload: &ttnpb.JoinRequestPayload{
					JoinEui:  joinEUI.Bytes(),
					DevEui:   devEUI.Bytes(),
					DevNonce: devNonce.Bytes(),
				},
			},
		},
	}
	buf, err := lorawan.MarshalMessage(msg.Payload)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeJoinRequestMIC(nwkKey, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}

func randomUpDataPayload(devAddr types.DevAddr, fPort uint32, size int) []byte {
	var fNwkSIntKey, sNwkSIntKey, appSKey types.AES128Key
	rand.Read(fNwkSIntKey[:])
	rand.Read(sNwkSIntKey[:])
	rand.Read(appSKey[:])

	pld := &ttnpb.MACPayload{
		FHdr: &ttnpb.FHDR{
			DevAddr: devAddr.Bytes(),
			FCnt:    42,
		},
		FPort:      fPort,
		FrmPayload: random.Bytes(size),
	}
	buf, err := crypto.EncryptUplink(appSKey, devAddr, pld.FHdr.FCnt, pld.FrmPayload)
	if err != nil {
		panic(err)
	}
	pld.FrmPayload = buf

	msg := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			Payload: &ttnpb.Message_MacPayload{
				MacPayload: pld,
			},
		},
	}
	buf, err = lorawan.MarshalMessage(msg.Payload)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, 0, 5, 0, devAddr, pld.FHdr.FCnt, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}

func randomDownDataPayload(devAddr types.DevAddr, fPort uint32, size int) []byte {
	var sNwkSIntKey, appSKey types.AES128Key
	rand.Read(sNwkSIntKey[:])
	rand.Read(appSKey[:])

	pld := &ttnpb.MACPayload{
		FHdr: &ttnpb.FHDR{
			DevAddr: devAddr.Bytes(),
			FCnt:    42,
		},
		FPort:      fPort,
		FrmPayload: random.Bytes(size),
	}
	buf, err := crypto.EncryptDownlink(appSKey, devAddr, pld.FHdr.FCnt, pld.FrmPayload)
	if err != nil {
		panic(err)
	}
	pld.FrmPayload = buf

	msg := &ttnpb.Message{
		MHdr: &ttnpb.MHDR{
			MType: ttnpb.MType_UNCONFIRMED_DOWN,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MacPayload{
			MacPayload: pld,
		},
	}
	buf, err = lorawan.MarshalMessage(msg)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeDownlinkMIC(sNwkSIntKey, devAddr, 0, pld.FHdr.FCnt, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}
