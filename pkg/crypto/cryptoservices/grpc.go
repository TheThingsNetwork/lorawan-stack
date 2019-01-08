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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

type networkRPCClient struct {
	Client ttnpb.NetworkCryptoServiceClient
	crypto.KeyVault
}

// NewNetworkRPCClient returns a network service which uses a gRPC service on the given connection.
func NewNetworkRPCClient(cc *grpc.ClientConn, keyVault crypto.KeyVault) Network {
	return &networkRPCClient{
		Client:   ttnpb.NewNetworkCryptoServiceClient(cc),
		KeyVault: keyVault,
	}
}

func (s *networkRPCClient) JoinRequestMIC(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, payload []byte) (mic [4]byte, err error) {
	res, err := s.Client.JoinRequestMIC(ctx, &ttnpb.CryptoServicePayloadRequest{
		EndDeviceIdentifiers: ids,
		LoRaWANVersion:       version,
		Payload:              payload,
	})
	if err != nil {
		return
	}
	copy(mic[:], res.Payload)
	return
}

func (s *networkRPCClient) JoinAcceptMIC(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) (mic [4]byte, err error) {
	res, err := s.Client.JoinAcceptMIC(ctx, &ttnpb.JoinAcceptMICRequest{
		CryptoServicePayloadRequest: ttnpb.CryptoServicePayloadRequest{
			EndDeviceIdentifiers: ids,
			LoRaWANVersion:       version,
			Payload:              payload,
		},
		JoinRequestType: uint32(joinReqType),
		DevNonce:        dn,
	})
	if err != nil {
		return
	}
	copy(mic[:], res.Payload)
	return
}

func (s *networkRPCClient) EncryptJoinAccept(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	res, err := s.Client.EncryptJoinAccept(ctx, &ttnpb.CryptoServicePayloadRequest{
		EndDeviceIdentifiers: ids,
		LoRaWANVersion:       version,
		Payload:              payload,
	})
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func (s *networkRPCClient) EncryptRejoinAccept(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	res, err := s.Client.EncryptRejoinAccept(ctx, &ttnpb.CryptoServicePayloadRequest{
		EndDeviceIdentifiers: ids,
		LoRaWANVersion:       version,
		Payload:              payload,
	})
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func (s *networkRPCClient) DeriveNwkSKeys(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (NwkSKeys, error) {
	keys, err := s.Client.DeriveNwkSKeys(ctx, &ttnpb.DeriveSessionKeysRequest{
		EndDeviceIdentifiers: ids,
		LoRaWANVersion:       version,
		JoinNonce:            jn,
		DevNonce:             dn,
		NetID:                nid,
	})
	if err != nil {
		return NwkSKeys{}, err
	}
	var res NwkSKeys
	res.FNwkSIntKey, err = cryptoutil.UnwrapAES128Key(keys.FNwkSIntKey, s.KeyVault)
	if err != nil {
		return NwkSKeys{}, err
	}
	res.SNwkSIntKey, err = cryptoutil.UnwrapAES128Key(keys.SNwkSIntKey, s.KeyVault)
	if err != nil {
		return NwkSKeys{}, err
	}
	res.NwkSEncKey, err = cryptoutil.UnwrapAES128Key(keys.NwkSEncKey, s.KeyVault)
	if err != nil {
		return NwkSKeys{}, err
	}
	return res, nil
}

type applicationRPCClient struct {
	Client ttnpb.ApplicationCryptoServiceClient
	crypto.KeyVault
}

// NewApplicationRPCClient returns an application service which uses a gRPC service on the given connection.
func NewApplicationRPCClient(cc *grpc.ClientConn, keyVault crypto.KeyVault) Application {
	return &applicationRPCClient{
		Client:   ttnpb.NewApplicationCryptoServiceClient(cc),
		KeyVault: keyVault,
	}
}

func (s *applicationRPCClient) DeriveAppSKey(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	res, err := s.Client.DeriveAppSKey(ctx, &ttnpb.DeriveSessionKeysRequest{
		EndDeviceIdentifiers: ids,
		LoRaWANVersion:       version,
		JoinNonce:            jn,
		DevNonce:             dn,
		NetID:                nid,
	})
	if err != nil {
		return types.AES128Key{}, err
	}
	return cryptoutil.UnwrapAES128Key(res.AppSKey, s.KeyVault)
}

type networkRPCServer struct {
	Network Network
	crypto.KeyVault
}

// NewNetworkRPCServer returns a ttnpb.NetworkCryptoServiceServer using the given service and key vault.
func NewNetworkRPCServer(network Network, keyVault crypto.KeyVault) ttnpb.NetworkCryptoServiceServer {
	return &networkRPCServer{
		Network:  network,
		KeyVault: keyVault,
	}
}

func (s *networkRPCServer) JoinRequestMIC(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	mic, err := s.Network.JoinRequestMIC(ctx, req.EndDeviceEUIIdentifiers, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: mic[:],
	}, nil
}

func (s *networkRPCServer) JoinAcceptMIC(ctx context.Context, req *ttnpb.JoinAcceptMICRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	mic, err := s.Network.JoinAcceptMIC(ctx, req.EndDeviceEUIIdentifiers, req.LoRaWANVersion, byte(req.JoinRequestType), req.DevNonce, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: mic[:],
	}, nil
}

func (s *networkRPCServer) EncryptJoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	data, err := s.Network.EncryptJoinAccept(ctx, req.EndDeviceEUIIdentifiers, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: data,
	}, nil
}

func (s *networkRPCServer) EncryptRejoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	data, err := s.Network.EncryptRejoinAccept(ctx, req.EndDeviceEUIIdentifiers, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: data,
	}, nil
}

func (s *networkRPCServer) DeriveNwkSKeys(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.NwkSKeysResponse, error) {
	nwkSKeys, err := s.Network.DeriveNwkSKeys(ctx, req.EndDeviceEUIIdentifiers, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	res := &ttnpb.NwkSKeysResponse{}
	res.FNwkSIntKey, err = cryptoutil.WrapAES128Key(nwkSKeys.FNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.SNwkSIntKey, err = cryptoutil.WrapAES128Key(nwkSKeys.SNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.NwkSEncKey, err = cryptoutil.WrapAES128Key(nwkSKeys.NwkSEncKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type applicationRPCServer struct {
	Application Application
	crypto.KeyVault
}

// NewApplicationRPCServer returns a ttnpb.ApplicationCryptoServiceServer using the given service and key vault.
func NewApplicationRPCServer(application Application, keyVault crypto.KeyVault) ttnpb.ApplicationCryptoServiceServer {
	return &applicationRPCServer{
		Application: application,
		KeyVault:    keyVault,
	}
}

func (s *applicationRPCServer) DeriveAppSKey(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.AppSKeyResponse, error) {
	appSKey, err := s.Application.DeriveAppSKey(ctx, req.EndDeviceEUIIdentifiers, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	res := &ttnpb.AppSKeyResponse{}
	res.AppSKey, err = cryptoutil.WrapAES128Key(appSKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}
