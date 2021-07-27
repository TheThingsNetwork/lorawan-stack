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

package joinserver_test

import (
	"context"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	joinEUIPrefixes = []types.EUI64Prefix{
		{EUI64: types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 42},
		{EUI64: types.EUI64{0x10, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 12},
		{EUI64: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, Length: 56},
	}
	nwkKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	appKey = types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nsAddr = "ns.test.org:1234"
	asAddr = "as.test.org:1234"
)

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func mustEncryptJoinAccept(key types.AES128Key, pld []byte) []byte {
	return test.Must(crypto.EncryptJoinAccept(key, pld)).([]byte)
}

func TestInvalidJoinRequests(t *testing.T) {
	_, ctx := test.New(t)

	for _, tc := range []struct {
		Name       string
		Invalidate func(*ttnpb.JoinRequest)
		Assertion  func(error) bool
	}{
		{
			// Baseline test: the join-request is valid; the end device is not found.
			// Other test cases modify the join-request to trigger InvalidArgument errors.
			Name:      "Device not found",
			Assertion: errors.IsNotFound,
		},
		{
			Name: "Invalid MType",
			Invalidate: func(up *ttnpb.JoinRequest) {
				// Confirmed uplink message
				up.RawPayload = []byte{0x80, 0x6c, 0xbf, 0xab, 0x1d, 0x80, 0x00, 0x00, 0x02, 0xfc, 0x7f, 0xa8, 0x0c, 0x17, 0xd5, 0x5b, 0x85, 0x5c, 0xa8, 0xa0, 0x14}
			},
			Assertion: errors.IsInvalidArgument,
		},
		{
			Name: "No payload",
			Invalidate: func(up *ttnpb.JoinRequest) {
				up.RawPayload = nil
			},
			Assertion: errors.IsInvalidArgument,
		},
		{
			Name: "Unknown JoinEUI",
			Invalidate: func(up *ttnpb.JoinRequest) {
				up.RawPayload = []byte{
					/* MHDR */
					0x00,
					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,
					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				}
			},
			Assertion: errors.IsInvalidArgument,
		},
		{
			Name: "Empty DevEUI",
			Invalidate: func(up *ttnpb.JoinRequest) {
				up.RawPayload = []byte{
					/* MHDR */
					0x00,
					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					/** DevNonce **/
					0x00, 0x00,
					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				}
			},
			Assertion: errors.IsInvalidArgument,
		},
	} {
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				redisClient, flush := test.NewRedis(ctx, "joinserver_test")
				defer flush()
				defer redisClient.Close()
				devReg := &redis.DeviceRegistry{Redis: redisClient}
				keyReg := &redis.KeyRegistry{Redis: redisClient}
				aasReg, aasRegCloseFn := NewRedisApplicationActivationSettingRegistry(ctx)
				defer aasRegCloseFn()

				c := componenttest.NewComponent(t, &component.Config{})
				js := test.Must(New(
					c,
					&Config{
						ApplicationActivationSettings: aasReg,
						Devices:                       devReg,
						Keys:                          keyReg,
						JoinEUIPrefixes:               joinEUIPrefixes,
					},
				)).(*JoinServer)
				componenttest.StartComponent(t, c)

				ctx = clusterauth.NewContext(ctx, nil)
				req := &ttnpb.JoinRequest{
					SelectedMACVersion: ttnpb.MAC_V1_1,
					RawPayload: []byte{
						/* MHDR */
						0x00,
						/* MACPayload */
						/** JoinEUI **/
						0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
						/** DevEUI **/
						0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
						/** DevNonce **/
						0x00, 0x00,
						/* MIC */
						0x55, 0x17, 0x54, 0x8e,
					},
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
					NetId:   types.NetID{0x42, 0xff, 0xff},
					DownlinkSettings: ttnpb.DLSettings{
						OptNeg:      true,
						Rx1DROffset: 0x7,
						Rx2DR:       0xf,
					},
					RxDelay: 0x42,
				}
				if tc.Invalidate != nil {
					tc.Invalidate(req)
				}

				_, err := js.HandleJoin(ctx, req, ClusterAuthorizer)
				a.So(tc.Assertion(err), should.BeTrue)
			},
		})
	}
}

func TestHandleJoin(t *testing.T) {
	_, ctx := test.New(t)

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  Authorizer

		KeyVault                      map[string][]byte
		Device                        *ttnpb.EndDevice
		ApplicationActivationSettings *ttnpb.ApplicationActivationSettings

		NextLastDevNonce  uint32
		NextLastJoinNonce uint32
		NextUsedDevNonces []uint32

		JoinRequest  *ttnpb.JoinRequest
		JoinResponse *ttnpb.JoinResponse

		ErrorAssertion func(error) bool
	}{
		{
			Name:        "1.1.0/cluster auth/new device/unwrapped keys",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xeb, 0xcd, 0x74, 0x59,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
				},
			},
		},
		{
			Name:        "1.1.0/cluster auth/new device/wrapped keys/addr KEKs",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			KeyVault: map[string][]byte{
				"ns:ns.test.org": {0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
				"as:as.test.org": {0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:           ttnpb.MAC_V1_1,
				ApplicationServerAddress: asAddr,
				NetworkServerAddress:     nsAddr,
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xeb, 0xcd, 0x74, 0x59,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						KEKLabel: "as:as.test.org",
						EncryptedKey: MustWrapKey(
							crypto.DeriveAppSKey(
								appKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
						),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						KEKLabel: "ns:ns.test.org",
						EncryptedKey: MustWrapKey(
							crypto.DeriveSNwkSIntKey(
								nwkKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
						),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						KEKLabel: "ns:ns.test.org",
						EncryptedKey: MustWrapKey(
							crypto.DeriveFNwkSIntKey(
								nwkKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
						),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						KEKLabel: "ns:ns.test.org",
						EncryptedKey: MustWrapKey(
							crypto.DeriveNwkSEncKey(
								nwkKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
						),
					},
				},
			},
		},
		{
			Name:        "1.1.0/cluster auth/new device/wrapped keys/custom device KEKs",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			KeyVault: map[string][]byte{
				"test-ns-kek": {0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
				"test-as-kek": {0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:            ttnpb.MAC_V1_1,
				ApplicationServerAddress:  asAddr,
				ApplicationServerKEKLabel: "test-as-kek",
				NetworkServerAddress:      nsAddr,
				NetworkServerKEKLabel:     "test-ns-kek",
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{
				KEKLabel: "test-aas-kek",
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xeb, 0xcd, 0x74, 0x59,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						KEKLabel: "test-as-kek",
						EncryptedKey: MustWrapKey(
							crypto.DeriveAppSKey(
								appKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
						),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						KEKLabel: "test-ns-kek",
						EncryptedKey: MustWrapKey(
							crypto.DeriveSNwkSIntKey(
								nwkKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
						),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						KEKLabel: "test-ns-kek",
						EncryptedKey: MustWrapKey(
							crypto.DeriveFNwkSIntKey(
								nwkKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
						),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						KEKLabel: "test-ns-kek",
						EncryptedKey: MustWrapKey(
							crypto.DeriveNwkSEncKey(
								nwkKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
						),
					},
				},
			},
		},
		{
			Name:        "1.1.0/cluster auth/new device/wrapped keys/custom AAS KEK",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			KeyVault: map[string][]byte{
				"test-aas-kek-kek": {0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				"test-ns-kek":      {0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
				"test-as-kek":      {0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion: ttnpb.MAC_V1_1,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{
				KEKLabel: "test-aas-kek",
				KEK: MustWrapAES128KeyWithKEK(
					ctx,
					types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					"test-aas-kek-kek",
					types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				),
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xeb, 0xcd, 0x74, 0x59,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						KEKLabel: "test-aas-kek",
						EncryptedKey: MustWrapKey(
							crypto.DeriveAppSKey(
								appKey,
								types.JoinNonce{0x00, 0x00, 0x01},
								types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								types.DevNonce{0x00, 0x00},
							),
							[]byte{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
				},
			},
		},
		{
			Name:        "1.1.0/existing device/dev nonce reset",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				LastDevNonce: 0x2441,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
				ResetsJoinNonces:     true,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xeb, 0xcd, 0x74, 0x59,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
				},
			},
		},
		{
			Name:        "1.1.0/cluster auth/existing device",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				LastDevNonce:  0x2441,
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastDevNonce:              0x2442,
			NextLastJoinNonce:             0x42fffe,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x6e, 0x54, 0x1b, 0x37,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0xfe, 0xff, 0x42,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc8, 0xf7, 0x62, 0xf4,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
				},
			},
		},
		{
			Name:        "1.1.0/DevNonce too small",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				LastDevNonce:  0x2442,
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastDevNonce:  0x2442,
			NextLastJoinNonce: 0x42fffd,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x6e, 0x54, 0x1b, 0x37,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.3/cluster auth/new device",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0_3,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			NextUsedDevNonces:             []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0_3,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
				},
			},
		},
		{
			Name:        "1.0.2/cluster auth/new device",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0_2,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			NextUsedDevNonces:             []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0_2,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
				},
			},
		},
		{
			Name:        "1.0.1/cluster auth/new device",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0_1,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			NextUsedDevNonces:             []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
				},
			},
		},
		{
			Name:        "1.0.0/cluster auth/new device",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			NextUsedDevNonces:             []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
				},
			},
		},
		{
			Name:        "1.0.0/cluster auth/existing device",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2444},
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             0x42fffe,
			NextUsedDevNonces:             []uint32{23, 41, 42, 52, 0x2442, 0x2444},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0xed, 0x8b, 0xd2, 0x24,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0xfe, 0xff, 0x42,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xf8, 0x4a, 0x11, 0x8e,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
				},
			},
		},
		{
			Name:        "1.0.0/cluster auth/existing device/nonce reuse",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442, 0x2444},
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
				ResetsJoinNonces:     true,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             0x42fffe,
			NextUsedDevNonces:             []uint32{23, 41, 42, 52, 0x2442, 0x2444},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0xed, 0x8b, 0xd2, 0x24,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0xfe, 0xff, 0x42,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xf8, 0x4a, 0x11, 0x8e,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
				},
			},
		},
		{
			Name: "1.0.0/TLS client auth/new device",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: "*.test.org",
				})
			},
			Authorizer: X509DNAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetId:                &types.NetID{0x42, 0xff, 0xff},
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			NextUsedDevNonces:             []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20,
				},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
				},
			},
		},
		{
			Name: "1.0.0/NetID mismatch",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: nsAddr,
				})
			},
			Authorizer: X509DNAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetId:                &types.NetID{0x42, 0xff, 0xff},
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			NextUsedDevNonces: []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0x42, 0x42},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/no NetID",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: nsAddr,
				})
			},
			Authorizer: X509DNAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			NextUsedDevNonces: []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsFailedPrecondition,
		},
		{
			Name: "1.0.0/address not authorized",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: "other.hostname.local",
				})
			},
			Authorizer: X509DNAuthorizer,
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetId:                &types.NetID{0x42, 0xff, 0xff},
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			NextUsedDevNonces: []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsPermissionDenied,
		},
		{
			Name:        "1.0.0/repeated DevNonce",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0xed, 0x8b, 0xd2, 0x24,
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/no payload",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				NetId:              types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/not a join request payload",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
					},
					Payload: &ttnpb.Message_JoinAcceptPayload{},
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/unsupported LoRaWAN version",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
						Major: ttnpb.Major(10),
					},
					Payload: &ttnpb.Message_JoinRequestPayload{},
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/no JoinEUI",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
						Major: ttnpb.Major_LORAWAN_R1,
					},
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							DevEui: types.EUI64{0x27, 0x00, 0x00, 0x00, 0x00, 0xab, 0xaa, 0x00},
						},
					},
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/raw payload that can't be unmarshalled",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					0x23, 0x42, 0xff, 0xff, 0xaa, 0x42, 0x42, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/invalid MType",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
					},
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							DevEui:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							JoinEui: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				redisClient, flush := test.NewRedis(ctx, "joinserver_test")
				defer flush()
				defer redisClient.Close()
				devReg := &redis.DeviceRegistry{Redis: redisClient}
				keyReg := &redis.KeyRegistry{Redis: redisClient}
				aasReg, aasRegCloseFn := NewRedisApplicationActivationSettingRegistry(ctx)
				defer aasRegCloseFn()

				if tc.ApplicationActivationSettings != nil {
					_, err := aasReg.SetByID(ctx, tc.Device.ApplicationIdentifiers, nil, func(sets *ttnpb.ApplicationActivationSettings) (*ttnpb.ApplicationActivationSettings, []string, error) {
						if sets != nil {
							panic("Application activation setting registry is not empty")
						}
						return tc.ApplicationActivationSettings, ttnpb.ApplicationActivationSettingsFieldPathsTopLevel, nil
					})
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to set application activation settings: %s", err)
					}
				}

				c := componenttest.NewComponent(t, &component.Config{
					ServiceBase: config.ServiceBase{
						KeyVault: config.KeyVault{
							Provider: "static",
							Static:   tc.KeyVault,
						},
					},
				})
				js := test.Must(New(
					c,
					&Config{
						ApplicationActivationSettings: aasReg,
						Devices:                       devReg,
						Keys:                          keyReg,
						JoinEUIPrefixes:               joinEUIPrefixes,
					},
				)).(*JoinServer)
				componenttest.StartComponent(t, c)

				pb := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

				start := time.Now()

				ret, err := devReg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceId,
					[]string{
						"application_server_address",
						"application_server_id",
						"application_server_kek_label",
						"created_at",
						"last_dev_nonce",
						"last_join_nonce",
						"lorawan_version",
						"net_id",
						"network_server_address",
						"network_server_kek_label",
						"provisioner_id",
						"provisioning_data",
						"root_keys",
						"resets_join_nonces",
						"updated_at",
						"used_dev_nonces",
					},
					func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
						if !a.So(stored, should.BeNil) {
							t.Fatal("Registry is not empty")
						}
						return CopyEndDevice(pb), []string{
							"application_server_address",
							"application_server_id",
							"application_server_kek_label",
							"ids.application_ids",
							"ids.dev_eui",
							"ids.device_id",
							"ids.join_eui",
							"last_dev_nonce",
							"last_join_nonce",
							"lorawan_version",
							"net_id",
							"network_server_address",
							"network_server_kek_label",
							"provisioner_id",
							"provisioning_data",
							"root_keys",
							"resets_join_nonces",
							"used_dev_nonces",
						}, nil
					},
				)
				if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
					t.Fatalf("Failed to create device: %s", err)
				}
				a.So(ret.CreatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
				pb.CreatedAt = ret.CreatedAt
				pb.UpdatedAt = ret.UpdatedAt
				a.So(ret, should.HaveEmptyDiff, pb)

				res, err := js.HandleJoin(ctx, deepcopy.Copy(tc.JoinRequest).(*ttnpb.JoinRequest), tc.Authorizer)
				if tc.ErrorAssertion != nil {
					if !a.So(err, should.BeError) || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
						t.Fatalf("Received an unexpected error: %s", err)
					}
					a.So(res, should.BeNil)
					return
				}

				if !a.So(err, should.BeNil) || !a.So(res, should.NotBeNil) {
					t.FailNow()
				}
				expectedResp := deepcopy.Copy(tc.JoinResponse).(*ttnpb.JoinResponse)
				a.So(res.SessionKeyID, should.NotBeEmpty)
				expectedResp.SessionKeyID = res.SessionKeyID
				a.So(res, should.Resemble, expectedResp)

				retCtx, err := devReg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEui, *pb.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
				if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
					t.FailNow()
				}
				ret = retCtx.EndDevice
				a.So(ret.CreatedAt, should.Equal, pb.CreatedAt)
				a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
				pb.UpdatedAt = ret.UpdatedAt
				pb.LastJoinNonce = tc.NextLastJoinNonce
				if tc.JoinRequest.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					pb.UsedDevNonces = tc.NextUsedDevNonces
				} else {
					pb.LastDevNonce = tc.NextLastDevNonce
				}
				if !a.So(ret.Session, should.NotBeNil) {
					t.FailNow()
				}
				a.So([]time.Time{start, ret.GetSession().GetStartedAt(), time.Now()}, should.BeChronological)
				pb.Session = &ttnpb.Session{
					DevAddr:     tc.JoinRequest.DevAddr,
					SessionKeys: res.SessionKeys,
					StartedAt:   ret.GetSession().GetStartedAt(),
				}
				pb.DevAddr = &tc.JoinRequest.DevAddr
				a.So(ret, should.HaveEmptyDiff, pb)

				res, err = js.HandleJoin(ctx, deepcopy.Copy(tc.JoinRequest).(*ttnpb.JoinRequest), tc.Authorizer)
				if !tc.Device.ResetsJoinNonces {
					a.So(err, should.BeError)
					a.So(res, should.BeNil)
				} else {
					a.So(err, should.BeNil)
					a.So(res, should.NotBeNil)
				}
			},
		})
	}
}

func TestGetNwkSKeys(t *testing.T) {
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  Authorizer

		GetByID     func(context.Context, types.EUI64, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.NwkSKeysResponse

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:        "Registry error",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return nil, errTest.New()
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, ErrRegistryOperation.WithCause(errTest)) {
					t.FailNow()
				}
				return a.So(errors.IsUnknown(err), should.BeTrue)
			},
		},
		{
			Name:        "No SNwkSIntKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					FNwkSIntKey: test.DefaultFNwkSIntKeyEnvelopeWrapped,
					NwkSEncKey:  test.DefaultNwkSEncKeyEnvelopeWrapped,
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoSNwkSIntKey)
			},
		},
		{
			Name:        "No NwkSEncKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					FNwkSIntKey: test.DefaultFNwkSIntKeyEnvelopeWrapped,
					SNwkSIntKey: test.DefaultSNwkSIntKeyEnvelopeWrapped,
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoNwkSEncKey)
			},
		},
		{
			Name:        "No FNwkSIntKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					SNwkSIntKey: test.DefaultSNwkSIntKeyEnvelopeWrapped,
					NwkSEncKey:  test.DefaultNwkSEncKeyEnvelopeWrapped,
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoFNwkSIntKey)
			},
		},
		{
			Name:        "Matching request",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key:      KeyPtr(types.AES128Key{0x43, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel: "NwkSEncKey-kek",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      KeyPtr(types.AES128Key{0x44, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel: "SNwkSIntKey-kek",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.NwkSKeysResponse{
				FNwkSIntKey: ttnpb.KeyEnvelope{
					Key: KeyPtr(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
				},
				NwkSEncKey: ttnpb.KeyEnvelope{
					Key:      KeyPtr(types.AES128Key{0x43, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel: "NwkSEncKey-kek",
				},
				SNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      KeyPtr(types.AES128Key{0x44, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel: "SNwkSIntKey-kek",
				},
			},
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				c := componenttest.NewComponent(t, &component.Config{})
				js := test.Must(New(
					c,
					&Config{
						Keys:    &MockKeyRegistry{GetByIDFunc: tc.GetByID},
						Devices: &MockDeviceRegistry{},
					},
				)).(*JoinServer)
				componenttest.StartComponent(t, c)
				res, err := js.GetNwkSKeys(ctx, tc.KeyRequest, tc.Authorizer)

				if tc.ErrorAssertion != nil {
					if !tc.ErrorAssertion(t, err) {
						t.Errorf("Received unexpected error: %s", err)
					}
					a.So(res, should.BeNil)
					return
				}

				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.KeyResponse)
			},
		})
	}
}

func TestGetAppSKey(t *testing.T) {
	errNotFound := errors.DefineNotFound("test_not_found", "not found")

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  Authorizer

		GetKeyByID     func(context.Context, types.EUI64, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
		GetDeviceByEUI func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.ContextualEndDevice, error)
		KeyRequest     *ttnpb.SessionKeyRequest
		KeyResponse    *ttnpb.AppSKeyResponse

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:        "Registry error",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return nil, errNotFound.New()
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, ErrRegistryOperation.WithCause(errNotFound)) {
					t.FailNow()
				}
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name:        "Missing AppSKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoAppSKey)
			},
		},
		{
			Name: "Address not authorized",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: "other.hostname.local",
				})
			},
			Authorizer: X509DNAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			GetDeviceByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"application_server_address",
					"application_server_id",
				})
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						ApplicationServerAddress: asAddr,
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrCallerNotAuthorized)
			},
		},
		{
			Name: "No application rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_DEVICES_READ}, // Require READ_KEYS
						},
					},
				})
			},
			Authorizer: ApplicationRightsAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			GetDeviceByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.BeEmpty)
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationId: "test-app",
							},
							DeviceId: "test-app",
						},
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name:        "Matching request/cluster auth",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel:     "test-kek",
				},
			},
		},
		{
			Name: "Matching request/application auth",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS},
						},
					},
				})
			},
			Authorizer: ApplicationRightsAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			GetDeviceByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.BeEmpty)
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationId: "test-app",
							},
							DeviceId: "test-app",
						},
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel:     "test-kek",
				},
			},
		},
		{
			Name: "Matching request/TLS client auth/address ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: "as.test.org",
				})
			},
			Authorizer: X509DNAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			GetDeviceByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"application_server_address",
					"application_server_id",
				})
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationId: "test-app",
							},
							DeviceId: "test-app",
						},
						ApplicationServerAddress: asAddr,
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel:     "test-kek",
				},
			},
		},
		{
			Name: "Matching request/TLS client auth/custom ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return auth.NewContextWithX509DN(ctx, pkix.Name{
					CommonName: "test-as-id",
				})
			},
			Authorizer: X509DNAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			GetDeviceByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"application_server_address",
					"application_server_id",
				})
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationId: "test-app",
							},
							DeviceId: "test-app",
						},
						ApplicationServerAddress: asAddr,
						ApplicationServerID:      "test-as-id",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel:     "test-kek",
				},
			},
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				js := test.Must(New(
					componenttest.NewComponent(t, &component.Config{}),
					&Config{
						Keys:    &MockKeyRegistry{GetByIDFunc: tc.GetKeyByID},
						Devices: &MockDeviceRegistry{GetByEUIFunc: tc.GetDeviceByEUI},
					},
				)).(*JoinServer)
				res, err := js.GetAppSKey(ctx, tc.KeyRequest, tc.Authorizer)

				if tc.ErrorAssertion != nil {
					if !tc.ErrorAssertion(t, err) {
						t.Errorf("Received unexpected error: %s", err)
					}
					a.So(res, should.BeNil)
					return
				}

				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.KeyResponse)
			},
		})
	}
}

func TestGetHomeNetID(t *testing.T) {
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  Authorizer

		GetByEUI func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.ContextualEndDevice, error)
		JoinEUI  types.EUI64
		DevEUI   types.EUI64
		Response *types.NetID

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:        "Registry error",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"net_id",
				})
				return nil, errTest.New()
			},
			JoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, ErrRegistryOperation.WithCause(errTest)) {
					t.FailNow()
				}
				return a.So(errors.IsUnknown(err), should.BeTrue)
			},
		},
		{
			Name:        "Matching request",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  ClusterAuthorizer,
			GetByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"net_id",
				})
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						NetId:                &types.NetID{0x42, 0xff, 0xff},
						NetworkServerAddress: nsAddr,
					},
				}, nil
			},
			JoinEUI:  types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:   types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			Response: &types.NetID{0x42, 0xff, 0xff},
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				js := test.Must(New(
					componenttest.NewComponent(t, &component.Config{}),
					&Config{
						Devices: &MockDeviceRegistry{
							GetByEUIFunc: tc.GetByEUI,
						},
					},
				)).(*JoinServer)
				netID, err := js.GetHomeNetID(ctx, tc.JoinEUI, tc.DevEUI, tc.Authorizer)

				if tc.ErrorAssertion != nil {
					if !tc.ErrorAssertion(t, err) {
						t.Errorf("Received unexpected error: %s", err)
					}
					a.So(netID, should.BeNil)
					return
				}

				a.So(err, should.BeNil)
				a.So(netID, should.Resemble, tc.Response)
			},
		})
	}
}
