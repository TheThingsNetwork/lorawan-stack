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
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver"
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

func keyToBytes(key types.AES128Key) []byte { return key[:] }

func keyPtr(key types.AES128Key) *types.AES128Key { return &key }

func mustWrapKey(key types.AES128Key, kek []byte) []byte {
	return test.Must(crypto.WrapKey(key[:], kek)).([]byte)
}

func mustWrapAES128KeyWithKEK(ctx context.Context, key types.AES128Key, kekLabel string, kek types.AES128Key) *ttnpb.KeyEnvelope {
	return test.Must(cryptoutil.WrapAES128KeyWithKEK(ctx, key, kekLabel, kek)).(*ttnpb.KeyEnvelope)
}

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
					/** DevNonce **/ 0x00, 0x00,
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
					/** DevNonce **/ 0x00, 0x00,
					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				}
			},
			Assertion: errors.IsInvalidArgument,
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				redisClient, flush := test.NewRedis(ctx, "joinserver_test")
				defer flush()
				defer redisClient.Close()
				devReg := &redis.DeviceRegistry{Redis: redisClient, LockTTL: test.Delay << 10}
				if err := devReg.Init(ctx); !a.So(err, should.BeNil) {
					t.FailNow()
				}
				keyReg := &redis.KeyRegistry{Redis: redisClient, LockTTL: test.Delay << 10}
				if err := keyReg.Init(ctx); !a.So(err, should.BeNil) {
					t.FailNow()
				}
				aasReg, aasRegCloseFn := NewRedisApplicationActivationSettingRegistry(ctx)
				defer aasRegCloseFn()

				c := componenttest.NewComponent(t, &component.Config{})
				js := test.Must(joinserver.New(
					c,
					&joinserver.Config{
						ApplicationActivationSettings: aasReg,
						Devices:                       devReg,
						Keys:                          keyReg,
						JoinEUIPrefixes:               joinEUIPrefixes,
					},
				)).(*joinserver.JoinServer)
				componenttest.StartComponent(t, c)

				ctx = clusterauth.NewContext(ctx, nil)
				req := &ttnpb.JoinRequest{
					SelectedMacVersion: ttnpb.MAC_V1_1,
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
					DownlinkSettings: &ttnpb.DLSettings{
						OptNeg:      true,
						Rx1DrOffset: 0x7,
						Rx2Dr:       0xf,
					},
					RxDelay: 0x42,
				}
				if tc.Invalidate != nil {
					tc.Invalidate(req)
				}

				_, err := js.HandleJoin(ctx, req, joinserver.ClusterAuthorizer(ctx))
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
		Authorizer  joinserver.Authorizer

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
			Name:        "1.1.0/cluster auth/new device/no NwkKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsFailedPrecondition,
		},
		{
			Name:        "1.1.0/cluster auth/new device/unwrapped keys",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveNwkSEncKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			KeyVault: map[string][]byte{
				"ns:ns.test.org": {0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
				"as:as.test.org": {0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						KekLabel: "as:as.test.org",
						EncryptedKey: mustWrapKey(
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
						KekLabel: "ns:ns.test.org",
						EncryptedKey: mustWrapKey(
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
						KekLabel: "ns:ns.test.org",
						EncryptedKey: mustWrapKey(
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
						KekLabel: "ns:ns.test.org",
						EncryptedKey: mustWrapKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			KeyVault: map[string][]byte{
				"test-ns-kek": {0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
				"test-as-kek": {0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				ApplicationServerKekLabel: "test-as-kek",
				NetworkServerAddress:      nsAddr,
				NetworkServerKekLabel:     "test-ns-kek",
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{
				KekLabel: "test-aas-kek",
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						KekLabel: "test-as-kek",
						EncryptedKey: mustWrapKey(
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
						KekLabel: "test-ns-kek",
						EncryptedKey: mustWrapKey(
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
						KekLabel: "test-ns-kek",
						EncryptedKey: mustWrapKey(
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
						KekLabel: "test-ns-kek",
						EncryptedKey: mustWrapKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			KeyVault: map[string][]byte{
				"test-aas-kek-kek": {0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				"test-ns-kek":      {0x3f, 0x36, 0x7b, 0xa1, 0x16, 0x67, 0xd9, 0x8b, 0x89, 0x00, 0x47, 0x77, 0x84, 0xf6, 0xfe, 0x50, 0x56, 0x67, 0x12, 0xab, 0x71, 0x96, 0x04, 0x6b, 0x9f, 0x2b, 0xc2, 0x50, 0xdf, 0xc8, 0xc1, 0xa2},
				"test-as-kek":      {0xed, 0x8a, 0x2e, 0x97, 0xf6, 0x8e, 0xbb, 0x79, 0x4d, 0x96, 0x4b, 0xd6, 0x14, 0xbb, 0xbc, 0xf2, 0x25, 0xc3, 0x7d, 0x61, 0xa9, 0xfe, 0xd0, 0x83, 0x7b, 0x07, 0xc0, 0x5f, 0x02, 0x52, 0x3c, 0x8b},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				KekLabel: "test-aas-kek",
				Kek: mustWrapAES128KeyWithKEK(
					ctx,
					types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					"test-aas-kek-kek",
					types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				),
			},
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						KekLabel: "test-aas-kek",
						EncryptedKey: mustWrapKey(
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
						Key: keyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveNwkSEncKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				LastDevNonce: 0x2441,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveNwkSEncKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				LastDevNonce:  0x2441,
				LastJoinNonce: 0x42fffd,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveNwkSEncKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				LastDevNonce:  0x2442,
				LastJoinNonce: 0x42fffd,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.3/cluster auth/new device/provisioned with NwkKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					NwkKey: &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					},
					AppKey: &ttnpb.KeyEnvelope{
						Key: &appKey,
					},
				},
				LorawanVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			ApplicationActivationSettings: &ttnpb.ApplicationActivationSettings{},
			NextLastJoinNonce:             1,
			NextUsedDevNonces:             []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0_3,
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
					0x29, 0xa8, 0xe5, 0x7d,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetId:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      false,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
						0x7f,
						/* RxDelay */
						0x42,
						/* MIC */
						0x9a, 0x99, 0x1e, 0x72,
					})...),
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
				},
			},
		},
		{
			Name:        "1.0.3/cluster auth/new device/provisioned without NwkKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0_3,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0_2,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0_1,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2444},
				LastJoinNonce: 0x42fffd,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
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
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442, 0x2444},
				LastJoinNonce: 0x42fffd,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
					},
				},
			},
		},
		{
			Name: "1.0.0/interop auth/new device",
			ContextFunc: func(ctx context.Context) context.Context {
				return interop.NewContextWithNetworkServerAuthInfo(ctx, &interop.NetworkServerAuthInfo{
					NetID:     types.NetID{0x42, 0xff, 0xff},
					Addresses: []string{"*.test.org"},
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
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
				SessionKeys: &ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(crypto.DeriveLegacyAppSKey(
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
				return interop.NewContextWithNetworkServerAuthInfo(ctx, &interop.NetworkServerAuthInfo{
					Addresses: []string{nsAddr},
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/no NetID",
			ContextFunc: func(ctx context.Context) context.Context {
				return interop.NewContextWithNetworkServerAuthInfo(ctx, &interop.NetworkServerAuthInfo{
					Addresses: []string{nsAddr},
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsFailedPrecondition,
		},
		{
			Name: "1.0.0/address not authorized",
			ContextFunc: func(ctx context.Context) context.Context {
				return interop.NewContextWithNetworkServerAuthInfo(ctx, &interop.NetworkServerAuthInfo{
					Addresses: []string{"other.hostname.local"},
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsPermissionDenied,
		},
		{
			Name:        "1.0.0/repeated DevNonce",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/no payload",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
				NetId:              types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/not a join request payload",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
					},
					Payload: &ttnpb.Message_JoinAcceptPayload{},
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/unsupported LoRaWAN version",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
						Major: ttnpb.Major(10),
					},
					Payload: &ttnpb.Message_JoinRequestPayload{},
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/no JoinEUI",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/raw payload that can't be unmarshalled",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					0x23, 0x42, 0xff, 0xff, 0xaa, 0x42, 0x42, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				NetId: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:        "1.0.0/invalid MType",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				Ids: &ttnpb.EndDeviceIdentifiers{
					DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
					DeviceId:       "test-dev",
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
				SelectedMacVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
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
				DownlinkSettings: &ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DrOffset: 0x7,
					Rx2Dr:       0xf,
				},
				RxDelay: 0x42,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				redisClient, flush := test.NewRedis(ctx, "joinserver_test")
				defer flush()
				defer redisClient.Close()
				devReg := &redis.DeviceRegistry{Redis: redisClient, LockTTL: test.Delay << 10}
				if err := devReg.Init(ctx); !a.So(err, should.BeNil) {
					t.FailNow()
				}
				keyReg := &redis.KeyRegistry{Redis: redisClient, LockTTL: test.Delay << 10}
				if err := keyReg.Init(ctx); !a.So(err, should.BeNil) {
					t.FailNow()
				}
				aasReg, aasRegCloseFn := NewRedisApplicationActivationSettingRegistry(ctx)
				defer aasRegCloseFn()

				if tc.ApplicationActivationSettings != nil {
					_, err := aasReg.SetByID(ctx, tc.Device.Ids.ApplicationIds, nil, func(sets *ttnpb.ApplicationActivationSettings) (*ttnpb.ApplicationActivationSettings, []string, error) {
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
				js := test.Must(joinserver.New(
					c,
					&joinserver.Config{
						ApplicationActivationSettings: aasReg,
						Devices:                       devReg,
						Keys:                          keyReg,
						JoinEUIPrefixes:               joinEUIPrefixes,
					},
				)).(*joinserver.JoinServer)
				componenttest.StartComponent(t, c)

				pb := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

				start := time.Now()

				ret, err := devReg.SetByID(ctx, pb.Ids.ApplicationIds, pb.Ids.DeviceId,
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
				a.So(*ttnpb.StdTime(ret.CreatedAt), should.HappenAfter, start)
				a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.Resemble, ret.CreatedAt)
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
				a.So(res.SessionKeys.SessionKeyId, should.NotBeEmpty)
				expectedResp.SessionKeys.SessionKeyId = res.SessionKeys.SessionKeyId
				a.So(res, should.Resemble, expectedResp)

				retCtx, err := devReg.GetByEUI(ctx, *pb.Ids.JoinEui, *pb.Ids.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
				if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
					t.FailNow()
				}
				ret = retCtx.EndDevice
				a.So(ret.CreatedAt, should.Resemble, pb.CreatedAt)
				a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, *ttnpb.StdTime(pb.UpdatedAt))
				pb.UpdatedAt = ret.UpdatedAt
				pb.LastJoinNonce = tc.NextLastJoinNonce
				if tc.JoinRequest.SelectedMacVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					pb.UsedDevNonces = tc.NextUsedDevNonces
				} else {
					pb.LastDevNonce = tc.NextLastDevNonce
				}
				if !a.So(ret.Session, should.NotBeNil) {
					t.FailNow()
				}
				a.So([]time.Time{start, *ttnpb.StdTime(ret.GetSession().GetStartedAt()), time.Now()}, should.BeChronological)
				pb.Session = &ttnpb.Session{
					DevAddr:   tc.JoinRequest.DevAddr,
					Keys:      res.SessionKeys,
					StartedAt: ret.GetSession().GetStartedAt(),
				}
				pb.Ids.DevAddr = &tc.JoinRequest.DevAddr
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
	_, ctx := test.New(t)
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  joinserver.Authorizer

		GetByID     func(context.Context, types.EUI64, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.NwkSKeysResponse

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:        "Registry error",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, joinserver.ErrRegistryOperation.WithCause(errTest)) {
					t.FailNow()
				}
				return a.So(errors.IsUnknown(err), should.BeTrue)
			},
		},
		{
			Name:        "No SNwkSIntKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, joinserver.ErrNoSNwkSIntKey)
			},
		},
		{
			Name:        "No NwkSEncKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, joinserver.ErrNoNwkSEncKey)
			},
		},
		{
			Name:        "No FNwkSIntKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, joinserver.ErrNoFNwkSIntKey)
			},
		},
		{
			Name:        "Matching request",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: keyPtr(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key:      keyPtr(types.AES128Key{0x43, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel: "NwkSEncKey-kek",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      keyPtr(types.AES128Key{0x44, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel: "SNwkSIntKey-kek",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.NwkSKeysResponse{
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: keyPtr(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key:      keyPtr(types.AES128Key{0x43, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KekLabel: "NwkSEncKey-kek",
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key:      keyPtr(types.AES128Key{0x44, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KekLabel: "SNwkSIntKey-kek",
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				c := componenttest.NewComponent(t, &component.Config{})
				js := test.Must(joinserver.New(
					c,
					&joinserver.Config{
						Keys:    &joinserver.MockKeyRegistry{GetByIDFunc: tc.GetByID},
						Devices: &joinserver.MockDeviceRegistry{},
					},
				)).(*joinserver.JoinServer)
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
	_, ctx := test.New(t)
	errNotFound := errors.DefineNotFound("test_not_found", "not found")

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  joinserver.Authorizer

		GetKeyByID     func(context.Context, types.EUI64, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
		GetDeviceByEUI func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.ContextualEndDevice, error)
		KeyRequest     *ttnpb.SessionKeyRequest
		KeyResponse    *ttnpb.AppSKeyResponse

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:        "Registry error",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, joinserver.ErrRegistryOperation.WithCause(errNotFound)) {
					t.FailNow()
				}
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name:        "Missing AppSKey",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, joinserver.ErrNoAppSKey)
			},
		},
		{
			Name: "Address not authorized",
			ContextFunc: func(ctx context.Context) context.Context {
				return interop.NewContextWithApplicationServerAuthInfo(ctx, &interop.ApplicationServerAuthInfo{
					Addresses: []string{"other.hostname.local"},
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel:     "test-kek",
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No application rights",
			ContextFunc: func(ctx context.Context) context.Context {
				ctx = rights.NewContextWithAuthInfo(ctx, &ttnpb.AuthInfoResponse{})
				ctx = rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}): {
							Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ}, // Require READ_KEYS
						},
					},
				})
				return ctx
			},
			Authorizer: joinserver.ApplicationRightsAuthorizer(ctx),
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel:     "test-kek",
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
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: &ttnpb.ApplicationIdentifiers{
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name:        "Matching request/cluster auth",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel:     "test-kek",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: &ttnpb.KeyEnvelope{
					EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KekLabel:     "test-kek",
				},
			},
		},
		{
			Name: "Matching request/application auth",
			ContextFunc: func(ctx context.Context) context.Context {
				ctx = rights.NewContextWithAuthInfo(ctx, &ttnpb.AuthInfoResponse{})
				ctx = rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}): {
							Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS},
						},
					},
				})
				return ctx
			},
			Authorizer: joinserver.ApplicationRightsAuthorizer(ctx),
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel:     "test-kek",
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
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: &ttnpb.ApplicationIdentifiers{
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: &ttnpb.KeyEnvelope{
					EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KekLabel:     "test-kek",
				},
			},
		},
		{
			Name: "Matching request/interop auth/address ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return interop.NewContextWithApplicationServerAuthInfo(ctx, &interop.ApplicationServerAuthInfo{
					Addresses: []string{"as.test.org"},
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel:     "test-kek",
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
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: &ttnpb.ApplicationIdentifiers{
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
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: &ttnpb.KeyEnvelope{
					EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KekLabel:     "test-kek",
				},
			},
		},
		{
			Name: "Matching request/interop auth/custom ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return interop.NewContextWithApplicationServerAuthInfo(ctx, &interop.ApplicationServerAuthInfo{
					ASID: "test-as-id",
				})
			},
			Authorizer: joinserver.InteropAuthorizer,
			GetKeyByID: func(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KekLabel:     "test-kek",
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
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: &ttnpb.ApplicationIdentifiers{
								ApplicationId: "test-app",
							},
							DeviceId: "test-app",
						},
						ApplicationServerAddress: asAddr,
						ApplicationServerId:      "test-as-id",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				JoinEui:      types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				DevEui:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: &ttnpb.KeyEnvelope{
					EncryptedKey: keyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KekLabel:     "test-kek",
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				js := test.Must(joinserver.New(
					componenttest.NewComponent(t, &component.Config{}),
					&joinserver.Config{
						Keys:    &joinserver.MockKeyRegistry{GetByIDFunc: tc.GetKeyByID},
						Devices: &joinserver.MockDeviceRegistry{GetByEUIFunc: tc.GetDeviceByEUI},
					},
				)).(*joinserver.JoinServer)
				res, err := js.GetAppSKey(ctx, tc.KeyRequest, tc.Authorizer)

				if tc.ErrorAssertion != nil {
					if !tc.ErrorAssertion(t, err) {
						t.Fatalf("Received unexpected error: %s", err)
					}
					a.So(res, should.BeNil)
					return
				}

				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, tc.KeyResponse)
			},
		})
	}
}

func TestGetHomeNetID(t *testing.T) {
	_, ctx := test.New(t)
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name        string
		ContextFunc func(context.Context) context.Context
		Authorizer  joinserver.Authorizer

		GetByEUI      func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.ContextualEndDevice, error)
		JoinEUI       types.EUI64
		DevEUI        types.EUI64
		ResponseNetID *types.NetID
		ResponseNSID  *types.EUI64

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:        "Registry error",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			GetByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"net_id",
					"network_server_address",
				})
				return nil, errTest.New()
			},
			JoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, joinserver.ErrRegistryOperation.WithCause(errTest)) {
					t.FailNow()
				}
				return a.So(errors.IsUnknown(err), should.BeTrue)
			},
		},
		{
			Name:        "Matching request",
			ContextFunc: func(ctx context.Context) context.Context { return clusterauth.NewContext(ctx, nil) },
			Authorizer:  joinserver.ClusterAuthorizer(ctx),
			GetByEUI: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"net_id",
					"network_server_address",
				})
				return &ttnpb.ContextualEndDevice{
					Context: ctx,
					EndDevice: &ttnpb.EndDevice{
						NetId:                &types.NetID{0x42, 0xff, 0xff},
						NetworkServerAddress: nsAddr,
					},
				}, nil
			},
			JoinEUI:       types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:        types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ResponseNetID: &types.NetID{0x42, 0xff, 0xff},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx = tc.ContextFunc(ctx)

				js := test.Must(joinserver.New(
					componenttest.NewComponent(t, &component.Config{}),
					&joinserver.Config{
						Devices: &joinserver.MockDeviceRegistry{
							GetByEUIFunc: tc.GetByEUI,
						},
					},
				)).(*joinserver.JoinServer)
				homeNetwork, err := js.GetHomeNetwork(ctx, tc.JoinEUI, tc.DevEUI, tc.Authorizer)

				if tc.ErrorAssertion != nil {
					if !tc.ErrorAssertion(t, err) {
						t.Fatalf("Received unexpected error: %s", err)
					}
					a.So(homeNetwork, should.BeNil)
					return
				}

				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(homeNetwork.NetID, should.Resemble, tc.ResponseNetID)
				a.So(homeNetwork.NSID, should.Equal, tc.ResponseNSID)
			},
		})
	}
}

func TestJoinServerCleanup(t *testing.T) {
	a, ctx := test.New(t)

	appList := []*ttnpb.ApplicationIdentifiers{
		{ApplicationId: "app-1"},
		{ApplicationId: "app-2"},
		{ApplicationId: "app-3"},
		{ApplicationId: "app-4"},
	}

	deviceList := []*ttnpb.EndDevice{
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appList[0],
				DeviceId:       "dev-1",
				JoinEui:        eui64Ptr(types.EUI64{0x41, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x41, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appList[0],
				DeviceId:       "dev-2",
				JoinEui:        eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appList[1],
				DeviceId:       "dev-3",
				JoinEui:        eui64Ptr(types.EUI64{0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appList[3],
				DeviceId:       "dev-4",
				JoinEui:        eui64Ptr(types.EUI64{0x44, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x44, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appList[3],
				DeviceId:       "dev-5",
				JoinEui:        eui64Ptr(types.EUI64{0x45, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x45, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appList[3],
				DeviceId:       "dev-6",
				JoinEui:        eui64Ptr(types.EUI64{0x46, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x46, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
	}

	deviceRedisClient, devsFlush := test.NewRedis(ctx, "joinserver_test", "devices")
	defer devsFlush()
	defer deviceRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: deviceRedisClient, LockTTL: test.Delay << 10}
	if err := deviceRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	appAsRedisClient, appAsFlush := test.NewRedis(ctx, "joinserver_test", "application-activation-settings")
	defer appAsFlush()
	defer appAsRedisClient.Close()
	appAsRegistry := &redis.ApplicationActivationSettingRegistry{Redis: appAsRedisClient, LockTTL: test.Delay << 10}
	if err := appAsRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	for _, dev := range deviceList {
		ret, err := deviceRegistry.SetByID(ctx, dev.Ids.ApplicationIds, dev.Ids.DeviceId, []string{
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
		}, func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			return dev, []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
			}, nil
		})
		if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
			t.Fatalf("Failed to create device: %s", err)
		}
	}
	for _, app := range appList {
		ret, err := appAsRegistry.SetByID(ctx, app, []string{"application_server_id"}, func(stored *ttnpb.ApplicationActivationSettings) (*ttnpb.ApplicationActivationSettings, []string, error) {
			return &ttnpb.ApplicationActivationSettings{
				ApplicationServerId: "test",
			}, []string{"application_server_id"}, nil
		})
		if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
			t.Fatalf("Failed to create application activation settings entry: %s", err)
		}
	}
	// Mock IS application and device sets
	isApplicationSet := map[string]struct{}{
		unique.ID(ctx, appList[2]): {},
		unique.ID(ctx, appList[3]): {},
	}
	isDeviceSet := map[string]struct{}{
		unique.ID(ctx, deviceList[4].Ids): {},
		unique.ID(ctx, deviceList[5].Ids): {},
	}
	joinServerCleaner := &joinserver.RegistryCleaner{
		DevRegistry:   deviceRegistry,
		AppAsRegistry: appAsRegistry,
	}
	err := joinServerCleaner.RangeToLocalSet(ctx)
	a.So(err, should.BeNil)
	a.So(joinServerCleaner.LocalDeviceSet, should.HaveLength, 6)
	a.So(joinServerCleaner.LocalApplicationSet, should.HaveLength, 4)

	err = joinServerCleaner.CleanData(ctx, isDeviceSet, isApplicationSet)
	a.So(err, should.BeNil)
	joinServerCleaner.RangeToLocalSet(ctx)
	a.So(joinServerCleaner.LocalApplicationSet, should.HaveLength, 2)
	a.So(joinServerCleaner.LocalApplicationSet, should.ContainKey, unique.ID(ctx, appList[2]))
	a.So(joinServerCleaner.LocalApplicationSet, should.ContainKey, unique.ID(ctx, appList[3]))

	a.So(joinServerCleaner.LocalDeviceSet, should.HaveLength, 2)
	a.So(joinServerCleaner.LocalDeviceSet, should.ContainKey, unique.ID(ctx, deviceList[4].Ids))
	a.So(joinServerCleaner.LocalDeviceSet, should.ContainKey, unique.ID(ctx, deviceList[5].Ids))
}
