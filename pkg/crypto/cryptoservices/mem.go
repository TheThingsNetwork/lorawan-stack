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
)

type mem struct {
	ttnpb.RootKeys
	crypto.KeyVault
}

// NewMemory returns a network and application service using the given root keys and key vault.
func NewMemory(keys ttnpb.RootKeys, keyVault crypto.KeyVault) NetworkApplication {
	return &mem{
		RootKeys: keys,
		KeyVault: keyVault,
	}
}

var errNoNwkKey = errors.DefineCorruption("no_nwk_key", "no NwkKey specified")

func (d *mem) getNwkKey(version ttnpb.MACVersion) (types.AES128Key, error) {
	switch version {
	case ttnpb.MAC_V1_1:
		if d.NwkKey == nil {
			return types.AES128Key{}, errNoNwkKey
		}
		return cryptoutil.UnwrapAES128Key(*d.NwkKey, d.KeyVault)
	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		if d.AppKey == nil {
			return types.AES128Key{}, errNoAppKey
		}
		return cryptoutil.UnwrapAES128Key(*d.AppKey, d.KeyVault)
	default:
		panic("This statement is unreachable. Please version check.")
	}
}

func (d *mem) JoinRequestMIC(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, payload []byte) (res [4]byte, err error) {
	key, err := d.getNwkKey(version)
	if err != nil {
		return
	}
	return crypto.ComputeJoinRequestMIC(key, payload)
}

var (
	errNoDevEUI  = errors.DefineCorruption("no_dev_eui", "no DevEUI specified")
	errNoJoinEUI = errors.DefineCorruption("no_join_eui", "no JoinEUI specified")
)

func (d *mem) JoinAcceptMIC(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) (res [4]byte, err error) {
	if ids.JoinEUI == nil || ids.JoinEUI.IsZero() {
		return [4]byte{},

			errNoJoinEUI
	}
	if ids.DevEUI == nil || ids.DevEUI.IsZero() {
		return [4]byte{},

			errNoDevEUI
	}
	key, err := d.getNwkKey(version)
	if err != nil {
		return
	}
	switch version {
	case ttnpb.MAC_V1_1:
		jsIntKey := crypto.DeriveJSIntKey(key, *ids.DevEUI)
		return crypto.ComputeJoinAcceptMIC(jsIntKey, joinReqType, *ids.JoinEUI, dn, payload)
	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		return crypto.ComputeLegacyJoinAcceptMIC(key, payload)
	default:
		panic("This statement is unreachable. Please version check.")
	}
}

func (d *mem) EncryptJoinAccept(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	key, err := d.getNwkKey(version)
	if err != nil {
		return nil, err
	}
	return crypto.EncryptJoinAccept(key, payload)
}

func (d *mem) EncryptRejoinAccept(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	if version != ttnpb.MAC_V1_1 {
		panic("This statement is unreachable. Please version check.")
	}
	if ids.JoinEUI == nil || ids.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}
	if ids.DevEUI == nil || ids.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	nwkKey, err := cryptoutil.UnwrapAES128Key(*d.NwkKey, d.KeyVault)
	if err != nil {
		return nil, err
	}
	jsEncKey := crypto.DeriveJSEncKey(nwkKey, *ids.DevEUI)
	return crypto.EncryptJoinAccept(jsEncKey, payload)
}

func (d *mem) DeriveNwkSKeys(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (NwkSKeys, error) {
	if ids.JoinEUI == nil || ids.JoinEUI.IsZero() {
		return NwkSKeys{}, errNoJoinEUI
	}
	if ids.DevEUI == nil || ids.DevEUI.IsZero() {
		return NwkSKeys{}, errNoDevEUI
	}
	switch version {
	case ttnpb.MAC_V1_1:
		if d.NwkKey == nil {
			return NwkSKeys{}, errNoNwkKey
		}
		nwkKey, err := cryptoutil.UnwrapAES128Key(*d.NwkKey, d.KeyVault)
		if err != nil {
			return NwkSKeys{}, err
		}
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveFNwkSIntKey(nwkKey, jn, *ids.JoinEUI, dn),
			SNwkSIntKey: crypto.DeriveSNwkSIntKey(nwkKey, jn, *ids.JoinEUI, dn),
			NwkSEncKey:  crypto.DeriveNwkSEncKey(nwkKey, jn, *ids.JoinEUI, dn),
		}, nil

	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		if d.AppKey == nil {
			return NwkSKeys{}, errNoAppKey
		}
		appKey, err := cryptoutil.UnwrapAES128Key(*d.AppKey, d.KeyVault)
		if err != nil {
			return NwkSKeys{}, err
		}
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveLegacyNwkSKey(appKey, jn, nid, dn),
		}, nil

	default:
		panic("This statement is unreachable. Fix version check.")
	}
}

var errNoAppKey = errors.DefineCorruption("no_app_key", "no AppKey specified")

func (d *mem) DeriveAppSKey(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	if ids.JoinEUI == nil || ids.JoinEUI.IsZero() {
		return types.AES128Key{}, errNoJoinEUI
	}
	if ids.DevEUI == nil || ids.DevEUI.IsZero() {
		return types.AES128Key{}, errNoDevEUI
	}
	if d.AppKey == nil {
		return types.AES128Key{}, errNoAppKey
	}
	appKey, err := cryptoutil.UnwrapAES128Key(*d.AppKey, d.KeyVault)
	if err != nil {
		return types.AES128Key{}, err
	}

	switch version {
	case ttnpb.MAC_V1_1:
		return crypto.DeriveAppSKey(appKey, jn, *ids.JoinEUI, dn), nil

	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		return crypto.DeriveLegacyAppSKey(appKey, jn, nid, dn), nil

	default:
		panic("This statement is unreachable. Fix version check.")
	}
}
