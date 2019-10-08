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

package cryptoservices

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

type networkRPCClient struct {
	Client ttnpb.NetworkCryptoServiceClient
	crypto.KeyVault
	callOpts []grpc.CallOption
}

// NewNetworkRPCClient returns a network service which uses a gRPC service on the given connection.
func NewNetworkRPCClient(cc *grpc.ClientConn, keyVault crypto.KeyVault, callOpts ...grpc.CallOption) Network {
	return &networkRPCClient{
		Client:   ttnpb.NewNetworkCryptoServiceClient(cc),
		KeyVault: keyVault,
		callOpts: callOpts,
	}
}

func (s *networkRPCClient) JoinRequestMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) (mic [4]byte, err error) {
	res, err := s.Client.JoinRequestMIC(ctx, &ttnpb.CryptoServicePayloadRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		LoRaWANVersion:       version,
		Payload:              payload,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		return
	}
	copy(mic[:], res.Payload)
	return
}

func (s *networkRPCClient) JoinAcceptMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) (mic [4]byte, err error) {
	res, err := s.Client.JoinAcceptMIC(ctx, &ttnpb.JoinAcceptMICRequest{
		CryptoServicePayloadRequest: ttnpb.CryptoServicePayloadRequest{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			LoRaWANVersion:       version,
			Payload:              payload,
			ProvisionerID:        dev.ProvisionerID,
			ProvisioningData:     dev.ProvisioningData,
		},
		JoinRequestType: ttnpb.RejoinType(joinReqType),
		DevNonce:        dn,
	}, s.callOpts...)
	if err != nil {
		return
	}
	copy(mic[:], res.Payload)
	return
}

func (s *networkRPCClient) EncryptJoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	res, err := s.Client.EncryptJoinAccept(ctx, &ttnpb.CryptoServicePayloadRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		LoRaWANVersion:       version,
		Payload:              payload,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func (s *networkRPCClient) EncryptRejoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	res, err := s.Client.EncryptRejoinAccept(ctx, &ttnpb.CryptoServicePayloadRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		LoRaWANVersion:       version,
		Payload:              payload,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func (s *networkRPCClient) DeriveNwkSKeys(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (NwkSKeys, error) {
	keys, err := s.Client.DeriveNwkSKeys(ctx, &ttnpb.DeriveSessionKeysRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		LoRaWANVersion:       version,
		JoinNonce:            jn,
		DevNonce:             dn,
		NetID:                nid,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		return NwkSKeys{}, err
	}
	var res NwkSKeys
	res.FNwkSIntKey, err = cryptoutil.UnwrapAES128Key(ctx, keys.FNwkSIntKey, s.KeyVault)
	if err != nil {
		return NwkSKeys{}, err
	}
	res.SNwkSIntKey, err = cryptoutil.UnwrapAES128Key(ctx, keys.SNwkSIntKey, s.KeyVault)
	if err != nil {
		return NwkSKeys{}, err
	}
	res.NwkSEncKey, err = cryptoutil.UnwrapAES128Key(ctx, keys.NwkSEncKey, s.KeyVault)
	if err != nil {
		return NwkSKeys{}, err
	}
	return res, nil
}

func (s *networkRPCClient) GetNwkKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	key, err := s.Client.GetNwkKey(ctx, &ttnpb.GetRootKeysRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		if errors.IsFailedPrecondition(err) {
			return nil, nil
		}
		return nil, err
	}
	plain, err := cryptoutil.UnwrapAES128Key(ctx, *key, s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &plain, err
}

type applicationRPCClient struct {
	Client ttnpb.ApplicationCryptoServiceClient
	crypto.KeyVault
	callOpts []grpc.CallOption
}

// NewApplicationRPCClient returns an application service which uses a gRPC service on the given connection.
func NewApplicationRPCClient(cc *grpc.ClientConn, keyVault crypto.KeyVault, callOpts ...grpc.CallOption) Application {
	return &applicationRPCClient{
		Client:   ttnpb.NewApplicationCryptoServiceClient(cc),
		KeyVault: keyVault,
		callOpts: callOpts,
	}
}

func (s *applicationRPCClient) DeriveAppSKey(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	res, err := s.Client.DeriveAppSKey(ctx, &ttnpb.DeriveSessionKeysRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		LoRaWANVersion:       version,
		JoinNonce:            jn,
		DevNonce:             dn,
		NetID:                nid,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		return types.AES128Key{}, err
	}
	return cryptoutil.UnwrapAES128Key(ctx, res.AppSKey, s.KeyVault)
}

func (s *applicationRPCClient) GetAppKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	key, err := s.Client.GetAppKey(ctx, &ttnpb.GetRootKeysRequest{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		ProvisionerID:        dev.ProvisionerID,
		ProvisioningData:     dev.ProvisioningData,
	}, s.callOpts...)
	if err != nil {
		if errors.IsFailedPrecondition(err) {
			return nil, nil
		}
		return nil, err
	}
	plain, err := cryptoutil.UnwrapAES128Key(ctx, *key, s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &plain, err
}
