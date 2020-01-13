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

package cryptoservices_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	. "go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func aes128KeyPtr(key types.AES128Key) *types.AES128Key { return &key }
func eui64Ptr(eui types.EUI64) *types.EUI64             { return &eui }

func TestCryptoServices(t *testing.T) {
	ctx := test.Context()
	keyVault := cryptoutil.NewMemKeyVault(map[string][]byte{})
	memSvc := NewMemory(
		aes128KeyPtr(types.AES128Key{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
		aes128KeyPtr(types.AES128Key{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2}),
	)
	ids := ttnpb.EndDeviceIdentifiers{
		JoinEUI: eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		DevEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	s := grpc.NewServer()
	ttnpb.RegisterNetworkCryptoServiceServer(s, &mockNetworkRPCServer{memSvc, keyVault})
	ttnpb.RegisterApplicationCryptoServiceServer(s, &mockApplicationRPCServer{memSvc, keyVault})
	go s.Serve(lis)
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}

	for _, svc := range []Network{
		memSvc,
		NewNetworkRPCClient(conn, keyVault),
	} {
		t.Run(fmt.Sprintf("%T", svc), func(t *testing.T) {
			t.Run("JoinRequestMIC", func(t *testing.T) {
				for _, tc := range []struct {
					Version ttnpb.MACVersion
					Payload []byte
					Result  [4]byte
				}{
					{
						Version: ttnpb.MAC_V1_1,
						Payload: bytes.Repeat([]byte{0x1}, 19),
						Result:  [4]byte{0x21, 0x4d, 0x19, 0x7d},
					},
					{
						Version: ttnpb.MAC_V1_0_2,
						Payload: bytes.Repeat([]byte{0x1}, 19),
						Result:  [4]byte{0x87, 0x14, 0x9f, 0xd},
					},
					{
						Version: ttnpb.MAC_V1_0_1,
						Payload: bytes.Repeat([]byte{0x1}, 19),
						Result:  [4]byte{0x87, 0x14, 0x9f, 0xd},
					},
					{
						Version: ttnpb.MAC_V1_0,
						Payload: bytes.Repeat([]byte{0x1}, 19),
						Result:  [4]byte{0x87, 0x14, 0x9f, 0xd},
					},
				} {
					t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
						a := assertions.New(t)
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: ids,
						}
						res, err := svc.JoinRequestMIC(ctx, dev, tc.Version, tc.Payload)
						a.So(err, should.BeNil)
						a.So(res, should.Resemble, tc.Result)
					})
				}
			})

			t.Run("JoinAcceptMIC", func(t *testing.T) {
				for _, tc := range []struct {
					Version     ttnpb.MACVersion
					JoinReqType byte
					DevNonce    types.DevNonce
					Payload     []byte
					Result      [4]byte
				}{
					{
						Version:     ttnpb.MAC_V1_1,
						JoinReqType: 0xff,
						DevNonce:    types.DevNonce{0x1, 0x2},
						Payload:     bytes.Repeat([]byte{0x1}, 13),
						Result:      [4]byte{0x1, 0xdf, 0x1e, 0xff},
					},
					{
						Version:     ttnpb.MAC_V1_1,
						JoinReqType: 0x0,
						DevNonce:    types.DevNonce{0x1, 0x2},
						Payload:     bytes.Repeat([]byte{0x1}, 13),
						Result:      [4]byte{0xa, 0x9c, 0x88, 0x33},
					},
					{
						Version:     ttnpb.MAC_V1_1,
						JoinReqType: 0x1,
						DevNonce:    types.DevNonce{0x1, 0x2},
						Payload:     bytes.Repeat([]byte{0x1}, 13),
						Result:      [4]byte{0xae, 0x2d, 0xdc, 0xd1},
					},
					{
						Version:     ttnpb.MAC_V1_1,
						JoinReqType: 0x2,
						DevNonce:    types.DevNonce{0x1, 0x2},
						Payload:     bytes.Repeat([]byte{0x1}, 13),
						Result:      [4]byte{0x18, 0x32, 0x16, 0xb1},
					},
					{
						Version: ttnpb.MAC_V1_0_2,
						Payload: bytes.Repeat([]byte{0x1}, 13),
						Result:  [4]byte{0x3, 0x1b, 0x42, 0x0},
					},
					{
						Version: ttnpb.MAC_V1_0_1,
						Payload: bytes.Repeat([]byte{0x1}, 13),
						Result:  [4]byte{0x3, 0x1b, 0x42, 0x0},
					},
					{
						Version: ttnpb.MAC_V1_0,
						Payload: bytes.Repeat([]byte{0x1}, 13),
						Result:  [4]byte{0x3, 0x1b, 0x42, 0x0},
					},
				} {
					t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
						a := assertions.New(t)
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: ids,
						}
						res, err := svc.JoinAcceptMIC(ctx, dev, tc.Version, tc.JoinReqType, tc.DevNonce, tc.Payload)
						a.So(err, should.BeNil)
						a.So(res, should.Resemble, tc.Result)
					})
				}
			})

			t.Run("EncryptJoinAccept", func(t *testing.T) {
				for _, tc := range []struct {
					Version ttnpb.MACVersion
					Payload []byte
					Result  []byte
				}{
					{
						Version: ttnpb.MAC_V1_1,
						Payload: bytes.Repeat([]byte{0x1}, 16),
						Result:  []byte{0xbc, 0x6e, 0x2b, 0xaf, 0x23, 0xca, 0x1e, 0x66, 0xaa, 0xd7, 0xb3, 0x95, 0xc1, 0xd6, 0xa6, 0xa},
					},
					{
						Version: ttnpb.MAC_V1_0_2,
						Payload: bytes.Repeat([]byte{0x1}, 16),
						Result:  []byte{0xe3, 0xcd, 0xe2, 0x37, 0xc8, 0xf2, 0xd9, 0x7b, 0x8d, 0x79, 0xf9, 0x17, 0x1d, 0x4b, 0xda, 0xc1},
					},
					{
						Version: ttnpb.MAC_V1_0_1,
						Payload: bytes.Repeat([]byte{0x1}, 16),
						Result:  []byte{0xe3, 0xcd, 0xe2, 0x37, 0xc8, 0xf2, 0xd9, 0x7b, 0x8d, 0x79, 0xf9, 0x17, 0x1d, 0x4b, 0xda, 0xc1},
					},
					{
						Version: ttnpb.MAC_V1_0,
						Payload: bytes.Repeat([]byte{0x1}, 16),
						Result:  []byte{0xe3, 0xcd, 0xe2, 0x37, 0xc8, 0xf2, 0xd9, 0x7b, 0x8d, 0x79, 0xf9, 0x17, 0x1d, 0x4b, 0xda, 0xc1},
					},
				} {
					t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
						a := assertions.New(t)
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: ids,
						}
						res, err := svc.EncryptJoinAccept(ctx, dev, tc.Version, tc.Payload)
						a.So(err, should.BeNil)
						a.So(res, should.Resemble, tc.Result)
					})
				}
			})

			t.Run("EncryptRejoinAccept", func(t *testing.T) {
				for _, tc := range []struct {
					Version ttnpb.MACVersion
					Payload []byte
					Result  []byte
				}{
					{
						Version: ttnpb.MAC_V1_1,
						Payload: bytes.Repeat([]byte{0x1}, 16),
						Result:  []byte{0x61, 0x58, 0x25, 0x46, 0x6a, 0x90, 0xec, 0xce, 0xf5, 0xd1, 0xf1, 0xc5, 0xba, 0x56, 0x6b, 0xe7},
					},
				} {
					t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
						a := assertions.New(t)
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: ids,
						}
						res, err := svc.EncryptRejoinAccept(ctx, dev, tc.Version, tc.Payload)
						a.So(err, should.BeNil)
						a.So(res, should.Resemble, tc.Result)
					})
				}
			})

			t.Run("DeriveNwkSKeys", func(t *testing.T) {
				for _, tc := range []struct {
					Version     ttnpb.MACVersion
					JoinNonce   types.JoinNonce
					DevNonce    types.DevNonce
					NetID       types.NetID
					FNwkSIntKey types.AES128Key
					SNwkSIntKey types.AES128Key
					NwkSEncKey  types.AES128Key
				}{
					{
						Version:     ttnpb.MAC_V1_1,
						JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:    types.DevNonce{0x1, 0x2},
						FNwkSIntKey: types.AES128Key{0xf8, 0xd8, 0xb8, 0xb9, 0xb1, 0xec, 0x36, 0xe8, 0xb8, 0x10, 0x84, 0x29, 0xd3, 0xf7, 0x3d, 0xd2},
						SNwkSIntKey: types.AES128Key{0x72, 0xde, 0xab, 0x55, 0x40, 0x3, 0xd2, 0x29, 0xc, 0xec, 0x8, 0x6, 0x81, 0x71, 0x92, 0x5d},
						NwkSEncKey:  types.AES128Key{0x31, 0x87, 0x9c, 0xf0, 0x93, 0xc2, 0x41, 0x11, 0xe3, 0x99, 0x5, 0xc7, 0x72, 0x76, 0xbf, 0xd8},
					},
					{
						Version:     ttnpb.MAC_V1_0_3,
						JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:    types.DevNonce{0x1, 0x2},
						NetID:       types.NetID{0x1, 0x2, 0x3},
						FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
					},
					{
						Version:     ttnpb.MAC_V1_0_2,
						JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:    types.DevNonce{0x1, 0x2},
						NetID:       types.NetID{0x1, 0x2, 0x3},
						FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
					},
					{
						Version:     ttnpb.MAC_V1_0_1,
						JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:    types.DevNonce{0x1, 0x2},
						NetID:       types.NetID{0x1, 0x2, 0x3},
						FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
					},
					{
						Version:     ttnpb.MAC_V1_0,
						JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:    types.DevNonce{0x1, 0x2},
						NetID:       types.NetID{0x1, 0x2, 0x3},
						FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
					},
				} {
					t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
						a := assertions.New(t)
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: ids,
						}
						keys, err := svc.DeriveNwkSKeys(ctx, dev, tc.Version, tc.JoinNonce, tc.DevNonce, tc.NetID)
						a.So(err, should.BeNil)
						a.So(keys.FNwkSIntKey, should.Resemble, tc.FNwkSIntKey)
						a.So(keys.SNwkSIntKey, should.Resemble, tc.SNwkSIntKey)
						a.So(keys.NwkSEncKey, should.Resemble, tc.NwkSEncKey)
					})
				}
			})

			t.Run("NwkKey", func(t *testing.T) {
				a := assertions.New(t)
				key, err := svc.GetNwkKey(ctx, &ttnpb.EndDevice{EndDeviceIdentifiers: ids})
				a.So(err, should.BeNil)
				a.So(key, should.Resemble, &types.AES128Key{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1})
			})
		})
	}

	for _, svc := range []Application{
		memSvc,
		NewApplicationRPCClient(conn, keyVault),
	} {
		t.Run(fmt.Sprintf("%T", svc), func(t *testing.T) {
			t.Run("DeriveAppSKey", func(t *testing.T) {
				for _, tc := range []struct {
					Version   ttnpb.MACVersion
					JoinNonce types.JoinNonce
					DevNonce  types.DevNonce
					NetID     types.NetID
					AppSKey   types.AES128Key
				}{
					{
						Version:   ttnpb.MAC_V1_1,
						JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:  types.DevNonce{0x1, 0x2},
						AppSKey:   types.AES128Key{0x4, 0x30, 0x89, 0x5c, 0x7b, 0xa7, 0xb1, 0x51, 0xcf, 0x97, 0x36, 0x84, 0xf6, 0x22, 0xff, 0xc1},
					},
					{
						Version:   ttnpb.MAC_V1_0_3,
						JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:  types.DevNonce{0x1, 0x2},
						NetID:     types.NetID{0x1, 0x2, 0x3},
						AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
					},
					{
						Version:   ttnpb.MAC_V1_0_2,
						JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:  types.DevNonce{0x1, 0x2},
						NetID:     types.NetID{0x1, 0x2, 0x3},
						AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
					},
					{
						Version:   ttnpb.MAC_V1_0_1,
						JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:  types.DevNonce{0x1, 0x2},
						NetID:     types.NetID{0x1, 0x2, 0x3},
						AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
					},
					{
						Version:   ttnpb.MAC_V1_0,
						JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
						DevNonce:  types.DevNonce{0x1, 0x2},
						NetID:     types.NetID{0x1, 0x2, 0x3},
						AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
					},
				} {
					t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
						a := assertions.New(t)
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: ids,
						}
						appSKey, err := svc.DeriveAppSKey(ctx, dev, tc.Version, tc.JoinNonce, tc.DevNonce, tc.NetID)
						a.So(err, should.BeNil)
						a.So(appSKey, should.Resemble, tc.AppSKey)
					})
				}
			})

			t.Run("AppKey", func(t *testing.T) {
				a := assertions.New(t)
				key, err := svc.GetAppKey(ctx, &ttnpb.EndDevice{EndDeviceIdentifiers: ids})
				a.So(err, should.BeNil)
				a.So(key, should.Resemble, &types.AES128Key{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2})
			})
		})
	}
}

type mockNetworkRPCServer struct {
	Network Network
	crypto.KeyVault
}

func (s *mockNetworkRPCServer) JoinRequestMIC(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	mic, err := s.Network.JoinRequestMIC(ctx, dev, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: mic[:],
	}, nil
}

func (s *mockNetworkRPCServer) JoinAcceptMIC(ctx context.Context, req *ttnpb.JoinAcceptMICRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	mic, err := s.Network.JoinAcceptMIC(ctx, dev, req.LoRaWANVersion, byte(req.JoinRequestType), req.DevNonce, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: mic[:],
	}, nil
}

func (s *mockNetworkRPCServer) EncryptJoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	data, err := s.Network.EncryptJoinAccept(ctx, dev, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: data,
	}, nil
}

func (s *mockNetworkRPCServer) EncryptRejoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	data, err := s.Network.EncryptRejoinAccept(ctx, dev, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: data,
	}, nil
}

func (s *mockNetworkRPCServer) DeriveNwkSKeys(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.NwkSKeysResponse, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	nwkSKeys, err := s.Network.DeriveNwkSKeys(ctx, dev, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	res := &ttnpb.NwkSKeysResponse{}
	res.FNwkSIntKey, err = cryptoutil.WrapAES128Key(ctx, nwkSKeys.FNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.SNwkSIntKey, err = cryptoutil.WrapAES128Key(ctx, nwkSKeys.SNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.NwkSEncKey, err = cryptoutil.WrapAES128Key(ctx, nwkSKeys.NwkSEncKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *mockNetworkRPCServer) GetNwkKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	nwkKey, err := s.Network.GetNwkKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	env, err := cryptoutil.WrapAES128Key(ctx, *nwkKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &env, nil
}

type mockApplicationRPCServer struct {
	Application Application
	crypto.KeyVault
}

func (s *mockApplicationRPCServer) DeriveAppSKey(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.AppSKeyResponse, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	appSKey, err := s.Application.DeriveAppSKey(ctx, dev, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	res := &ttnpb.AppSKeyResponse{}
	res.AppSKey, err = cryptoutil.WrapAES128Key(ctx, appSKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *mockApplicationRPCServer) GetAppKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
	}
	appKey, err := s.Application.GetAppKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	env, err := cryptoutil.WrapAES128Key(ctx, *appKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &env, nil
}
