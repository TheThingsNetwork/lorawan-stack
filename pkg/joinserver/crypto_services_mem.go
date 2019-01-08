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

type MemCryptoService struct {
	crypto.KeyVault
	ttnpb.RootKeys
}

func (d *MemCryptoService) getNwkKey(ids ttnpb.CryptoServiceEndDeviceIdentifiers) (key types.AES128Key, err error) {
	switch ids.LoRaWANVersion {
	case ttnpb.MAC_V1_1:
		if d.NwkKey == nil {
			err = errNoNwkKey
			return
		}
		key, err = cryptoutil.UnwrapAES128Key(*d.NwkKey, d.KeyVault)
		return

	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		if d.AppKey == nil {
			err = errNoAppKey
			return
		}
		key, err = cryptoutil.UnwrapAES128Key(*d.AppKey, d.KeyVault)
		return

	default:
		panic("This statement is unreachable. Please version check.")
	}
}

func (d *MemCryptoService) JoinRequestMIC(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) (res [4]byte, err error) {
	key, err := d.getNwkKey(ids)
	if err != nil {
		return
	}
	return crypto.ComputeJoinRequestMIC(key, payload)
}

func (d *MemCryptoService) JoinAcceptMIC(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, joinReqType byte, dn types.DevNonce, payload []byte) (res [4]byte, err error) {
	key, err := d.getNwkKey(ids)
	if err != nil {
		return
	}
	switch ids.LoRaWANVersion {
	case ttnpb.MAC_V1_1:
		jsIntKey := crypto.DeriveJSIntKey(key, ids.DevEUI)
		return crypto.ComputeJoinAcceptMIC(jsIntKey, joinReqType, ids.JoinEUI, dn, payload)
	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		return crypto.ComputeLegacyJoinAcceptMIC(key, payload)
	default:
		panic("This statement is unreachable. Please version check.")
	}
}

func (d *MemCryptoService) EncryptJoinAccept(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([]byte, error) {
	key, err := d.getNwkKey(ids)
	if err != nil {
		return nil, err
	}
	return crypto.EncryptJoinAccept(key, payload)
}

func (d *MemCryptoService) EncryptRejoinAccept(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([]byte, error) {
	if ids.LoRaWANVersion != ttnpb.MAC_V1_1 {
		panic("This statement is unreachable. Please version check.")
	}
	if d.NwkKey == nil {
		return nil, errNoNwkKey
	}
	nwkKey, err := cryptoutil.UnwrapAES128Key(*d.NwkKey, d.KeyVault)
	if err != nil {
		return nil, err
	}
	jsEncKey := crypto.DeriveJSEncKey(nwkKey, ids.DevEUI)
	return crypto.EncryptJoinAccept(jsEncKey, payload)
}

func (d *MemCryptoService) DeriveNwkSKeys(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (fNwkSIntKey, sNwkSIntKey, nwkSEncKey types.AES128Key, err error) {
	switch ids.LoRaWANVersion {
	case ttnpb.MAC_V1_1:
		if d.NwkKey == nil {
			err = errNoNwkKey
			return
		}
		var nwkKey types.AES128Key
		nwkKey, err = cryptoutil.UnwrapAES128Key(*d.NwkKey, d.KeyVault)
		if err != nil {
			return
		}
		fNwkSIntKey = crypto.DeriveFNwkSIntKey(nwkKey, jn, ids.JoinEUI, dn)
		sNwkSIntKey = crypto.DeriveSNwkSIntKey(nwkKey, jn, ids.JoinEUI, dn)
		nwkSEncKey = crypto.DeriveNwkSEncKey(nwkKey, jn, ids.JoinEUI, dn)

	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		if d.AppKey == nil {
			err = errNoAppKey
			return
		}
		var appKey types.AES128Key
		appKey, err = cryptoutil.UnwrapAES128Key(*d.AppKey, d.KeyVault)
		if err != nil {
			return
		}
		fNwkSIntKey = crypto.DeriveLegacyNwkSKey(appKey, jn, nid, dn)

	default:
		panic("This statement is unreachable. Fix version check.")
	}
	return
}

func (d *MemCryptoService) DeriveAppSKey(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (appSKey types.AES128Key, err error) {
	if d.AppKey == nil {
		err = errNoAppKey
		return
	}
	var appKey types.AES128Key
	appKey, err = cryptoutil.UnwrapAES128Key(*d.AppKey, d.KeyVault)
	if err != nil {
		return
	}

	switch ids.LoRaWANVersion {
	case ttnpb.MAC_V1_1:
		appSKey = crypto.DeriveAppSKey(appKey, jn, ids.JoinEUI, dn)

	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		appSKey = crypto.DeriveLegacyAppSKey(appKey, jn, nid, dn)

	default:
		panic("This statement is unreachable. Fix version check.")
	}
	return
}
