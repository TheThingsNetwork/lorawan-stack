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

package crypto_test

import (
	"fmt"
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestJoinAcceptEncryption(t *testing.T) {
	a := assertions.New(t)

	_, err := EncryptJoinAccept(types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, nil)
	a.So(err, should.NotBeNil)
	_, err = DecryptJoinAccept(types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, nil)
	a.So(err, should.NotBeNil)

	for i, tc := range []struct {
		Key                  types.AES128Key
		Decrypted, Encrypted []byte
	}{
		{
			Key: types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Decrypted: []byte{
				/* JoinNonce */
				0x03, 0x02, 0x01,
				/* NetID */
				0x03, 0x02, 0x01,
				/* DevAddr */
				0x04, 0x03, 0x02, 0x01,
				/* DLSettings */
				0x00,
				/* RxDelay */
				0x01,
				/* MIC */
				0x32, 0xf5, 0x4a, 0xb3,
			},
			Encrypted: []byte{0xc9, 0xfb, 0xb2, 0x59, 0xe1, 0x16, 0x49, 0x09, 0x6a, 0x56, 0x8a, 0x9e, 0x3b, 0x71, 0x17, 0xc3},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			a := assertions.New(t)

			key := deepcopy.Copy(tc.Key).(types.AES128Key)
			dec := deepcopy.Copy(tc.Decrypted).([]byte)
			enc, err := EncryptJoinAccept(key, dec)
			a.So(err, should.BeNil)
			a.So(dec, should.Resemble, tc.Decrypted)
			a.So(enc, should.Resemble, tc.Encrypted)
			a.So(key, should.Resemble, tc.Key)

			key = deepcopy.Copy(tc.Key).(types.AES128Key)
			enc = deepcopy.Copy(tc.Encrypted).([]byte)
			dec, err = DecryptJoinAccept(tc.Key, tc.Encrypted)
			a.So(err, should.BeNil)
			a.So(dec, should.Resemble, tc.Decrypted)
			a.So(enc, should.Resemble, tc.Encrypted)
			a.So(key, should.Resemble, tc.Key)
		})
	}
}

func TestComputeJoinRequestMIC(t *testing.T) {
	a := assertions.New(t)

	_, err := ComputeJoinRequestMIC(types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, nil)
	a.So(err, should.NotBeNil)

	for i, tc := range []struct {
		Key     types.AES128Key
		Payload []byte
		MIC     [4]byte
	}{
		{
			Key: types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Payload: []byte{
				/* MHDR */
				0b000_000_00,
				/* Join-Request */
				/** JoinEUI **/
				0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
				/** DevEUI **/
				0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
				/** DevNonce **/
				0x02, 0x01,
			},
			MIC: [4]byte{0xe6, 0xe1, 0x0c, 0x55},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			a := assertions.New(t)

			key := deepcopy.Copy(tc.Key).(types.AES128Key)
			pld := deepcopy.Copy(tc.Payload).([]byte)
			mic, err := ComputeJoinRequestMIC(key, pld)
			a.So(err, should.BeNil)
			a.So(mic, should.Equal, tc.MIC)
			a.So(key, should.Resemble, tc.Key)
		})
	}
}

func TestComputeRejoinRequestMIC(t *testing.T) {
	a := assertions.New(t)

	_, err := ComputeRejoinRequestMIC(types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, nil)
	a.So(err, should.NotBeNil)

	for _, tc := range []struct {
		Key     types.AES128Key
		Payload []byte
		MIC     [4]byte
	}{
		{
			Key: types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Payload: []byte{
				/* MHDR */
				0b110_000_00,
				/* Rejoin-Request */
				/** RejoinType **/
				0x00,
				/** NetID **/
				0x03, 0x02, 0x01,
				/** DevEUI **/
				0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
				/** RJcount0 **/
				0x02, 0x01,
			},
			MIC: [4]byte{0x11, 0xda, 0x47, 0xbd},
		},
		{
			Key: types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Payload: []byte{
				/* MHDR */
				0b110_000_00,
				/* Rejoin-Request */
				/** RejoinType **/
				0x01,
				/** JoinEUI **/
				0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
				/** DevEUI **/
				0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
				/** RJcount1 **/
				0x02, 0x01,
			},
			MIC: [4]byte{0x67, 0xf9, 0xab, 0xe7},
		},
		{
			Key: types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Payload: []byte{
				/* MHDR */
				0b110_000_00,
				/* Rejoin-Request */
				/** RejoinType **/
				0x02,
				/** NetID **/
				0x03, 0x02, 0x01,
				/** DevEUI **/
				0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
				/** RJcount0 **/
				0x02, 0x01,
			},
			MIC: [4]byte{0x2b, 0x45, 0x9c, 0xf0},
		},
	} {
		t.Run(fmt.Sprintf("Type %d", tc.Payload[1]), func(t *testing.T) {
			a := assertions.New(t)

			key := deepcopy.Copy(tc.Key).(types.AES128Key)
			pld := deepcopy.Copy(tc.Payload).([]byte)
			mic, err := ComputeRejoinRequestMIC(key, pld)
			a.So(err, should.BeNil)
			a.So(mic, should.Equal, tc.MIC)
			a.So(key, should.Resemble, tc.Key)
		})
	}
}

func TestComputeLegacyJoinAcceptMIC(t *testing.T) {
	a := assertions.New(t)

	_, err := ComputeLegacyJoinAcceptMIC(types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, nil)
	a.So(err, should.NotBeNil)

	for i, tc := range []struct {
		Key     types.AES128Key
		Payload []byte
		MIC     [4]byte
	}{
		{
			Key: types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Payload: []byte{
				/* MHDR */
				0b001_000_00,
				/** AppNonce **/
				0x03, 0x02, 0x01,
				/** NetID **/
				0x03, 0x02, 0x01,
				/** DevAddr **/
				0x04, 0x03, 0x02, 0x01,
				/** DLSettings **/
				0x00,
				/** RxDelay **/
				0x01,
			},
			MIC: [4]byte{0x32, 0xf5, 0x4a, 0xb3},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			a := assertions.New(t)

			key := deepcopy.Copy(tc.Key).(types.AES128Key)
			pld := deepcopy.Copy(tc.Payload).([]byte)
			mic, err := ComputeLegacyJoinAcceptMIC(key, pld)
			a.So(err, should.BeNil)
			a.So(mic, should.Equal, tc.MIC)
			a.So(key, should.Resemble, tc.Key)
		})
	}
}

func TestComputeJoinAcceptMIC(t *testing.T) {
	a := assertions.New(t)

	_, err := ComputeJoinAcceptMIC(types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 0xff, types.EUI64{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}, types.DevNonce{0x02, 0x01}, nil)
	a.So(err, should.NotBeNil)

	for _, tc := range []struct {
		Name     string
		Key      types.AES128Key
		Type     byte
		JoinEUI  types.EUI64
		DevNonce types.DevNonce
		Payload  []byte
		MIC      [4]byte
	}{
		{
			Name:     "Join-request accept/no CFList",
			Key:      types.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Type:     0xff,
			JoinEUI:  types.EUI64{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01},
			DevNonce: types.DevNonce{0x02, 0x01},
			Payload: []byte{
				/* MHDR */
				0b001_000_00,
				/** JoinNonce **/
				0x03, 0x02, 0x01,
				/** NetID **/
				0x03, 0x02, 0x01,
				/** DevAddr **/
				0x04, 0x03, 0x02, 0x01,
				/** DLSettings **/
				0x00,
				/** RxDelay **/
				0x01,
			},
			MIC: [4]byte{0x48, 0xe9, 0xbe, 0x5f},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			key := deepcopy.Copy(tc.Key).(types.AES128Key)
			joinEUI := deepcopy.Copy(tc.JoinEUI).(types.EUI64)
			devNonce := deepcopy.Copy(tc.DevNonce).(types.DevNonce)
			pld := deepcopy.Copy(tc.Payload).([]byte)
			mic, err := ComputeJoinAcceptMIC(key, tc.Type, joinEUI, devNonce, pld)
			a.So(err, should.BeNil)
			a.So(mic, should.Equal, tc.MIC)
			a.So(key, should.Resemble, tc.Key)
			a.So(joinEUI, should.Equal, tc.JoinEUI)
			a.So(devNonce, should.Equal, tc.DevNonce)
			a.So(pld, should.Resemble, tc.Payload)
		})
	}
}
