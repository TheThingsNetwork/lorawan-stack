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

package joinserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type rpcNetworkCryptoService struct {
	Client ttnpb.NetworkCryptoServiceClient
	crypto.KeyVault
}

func (s *rpcNetworkCryptoService) JoinRequestMIC(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) (mic [4]byte, err error) {
	res, err := s.Client.JoinRequestMIC(ctx, &ttnpb.CryptoServicePayloadRequest{
		CryptoServiceEndDeviceIdentifiers: ids,
		Payload:                           payload,
	})
	if err != nil {
		return
	}
	copy(mic[:], res.Payload)
	return
}

func (s *rpcNetworkCryptoService) JoinAcceptMIC(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, joinReqType byte, dn types.DevNonce, payload []byte) (mic [4]byte, err error) {
	res, err := s.Client.JoinAcceptMIC(ctx, &ttnpb.JoinAcceptMICRequest{
		CryptoServicePayloadRequest: ttnpb.CryptoServicePayloadRequest{
			CryptoServiceEndDeviceIdentifiers: ids,
			Payload:                           payload,
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

func (s *rpcNetworkCryptoService) EncryptJoinAccept(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([]byte, error) {
	res, err := s.Client.EncryptJoinAccept(ctx, &ttnpb.CryptoServicePayloadRequest{
		CryptoServiceEndDeviceIdentifiers: ids,
		Payload:                           payload,
	})
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func (s *rpcNetworkCryptoService) EncryptRejoinAccept(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([]byte, error) {
	res, err := s.Client.EncryptRejoinAccept(ctx, &ttnpb.CryptoServicePayloadRequest{
		CryptoServiceEndDeviceIdentifiers: ids,
		Payload:                           payload,
	})
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func (s *rpcNetworkCryptoService) DeriveNwkSKeys(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (fNwkSIntKey, sNwkSIntKey, nwkSEncKey types.AES128Key, err error) {
	res, err := s.Client.DeriveNwkSKeys(ctx, &ttnpb.DeriveSessionKeysRequest{
		CryptoServiceEndDeviceIdentifiers: ids,
		JoinNonce:                         jn,
		DevNonce:                          dn,
		NetID:                             nid,
	})
	if err != nil {
		return
	}
	fNwkSIntKey, err = cryptoutil.UnwrapAES128Key(res.FNwkSIntKey, s.KeyVault)
	if err != nil {
		return
	}
	sNwkSIntKey, err = cryptoutil.UnwrapAES128Key(res.SNwkSIntKey, s.KeyVault)
	if err != nil {
		return
	}
	nwkSEncKey, err = cryptoutil.UnwrapAES128Key(res.NwkSEncKey, s.KeyVault)
	if err != nil {
		return
	}
	return
}

type rpcApplicationCryptoService struct {
	Client ttnpb.ApplicationCryptoServiceClient
	crypto.KeyVault
}

func (s *rpcApplicationCryptoService) DeriveAppSKey(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (appSKey types.AES128Key, err error) {
	res, err := s.Client.DeriveAppSKey(ctx, &ttnpb.DeriveSessionKeysRequest{
		CryptoServiceEndDeviceIdentifiers: ids,
		JoinNonce:                         jn,
		DevNonce:                          dn,
		NetID:                             nid,
	})
	if err != nil {
		return
	}
	appSKey, err = cryptoutil.UnwrapAES128Key(res.AppSKey, s.KeyVault)
	if err != nil {
		return
	}
	return
}
