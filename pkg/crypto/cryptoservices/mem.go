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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type mem struct {
	nwkKey,
	appKey *types.AES128Key
}

// NewMemory returns a network and application service using the given root keys and key vault.
func NewMemory(nwkKey, appKey *types.AES128Key) NetworkApplication {
	return &mem{
		nwkKey: nwkKey,
		appKey: appKey,
	}
}

var errNoNwkKey = errors.DefineCorruption("no_nwk_key", "no NwkKey specified")

func (d *mem) getNwkKey(version ttnpb.MACVersion) (*types.AES128Key, error) {
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		if d.nwkKey == nil {
			return nil, errNoNwkKey
		}
		return d.nwkKey, nil
	default:
		if d.appKey == nil {
			return nil, errNoAppKey
		}
		return d.appKey, nil
	}
}

func (d *mem) JoinRequestMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) (res [4]byte, err error) {
	key, err := d.getNwkKey(version)
	if err != nil {
		return
	}
	if key == nil {
		return [4]byte{}, errNoNwkKey
	}
	return crypto.ComputeJoinRequestMIC(*key, payload)
}

var (
	errNoDevEUI  = errors.DefineCorruption("no_dev_eui", "no DevEUI specified")
	errNoJoinEUI = errors.DefineCorruption("no_join_eui", "no JoinEUI specified")
)

func (d *mem) JoinAcceptMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) ([4]byte, error) {
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return [4]byte{}, errNoJoinEUI
	}
	if dev.DevEUI == nil || dev.DevEUI.IsZero() {
		return [4]byte{}, errNoDevEUI
	}
	key, err := d.getNwkKey(version)
	if err != nil {
		return [4]byte{}, err
	}
	if key == nil {
		return [4]byte{}, errNoNwkKey
	}
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		jsIntKey := crypto.DeriveJSIntKey(*key, *dev.DevEUI)
		return crypto.ComputeJoinAcceptMIC(jsIntKey, joinReqType, *dev.JoinEUI, dn, payload)
	default:
		return crypto.ComputeLegacyJoinAcceptMIC(*key, payload)
	}
}

func (d *mem) EncryptJoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	key, err := d.getNwkKey(version)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errNoNwkKey
	}
	return crypto.EncryptJoinAccept(*key, payload)
}

func (d *mem) EncryptRejoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	if version.Compare(ttnpb.MAC_V1_1) < 0 {
		panic("This statement is unreachable. Please version check.")
	}
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}
	if dev.DevEUI == nil || dev.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	if d.nwkKey == nil {
		return nil, errNoNwkKey
	}
	jsEncKey := crypto.DeriveJSEncKey(*d.nwkKey, *dev.DevEUI)
	return crypto.EncryptJoinAccept(jsEncKey, payload)
}

func (d *mem) DeriveNwkSKeys(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (NwkSKeys, error) {
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return NwkSKeys{}, errNoJoinEUI
	}
	if dev.DevEUI == nil || dev.DevEUI.IsZero() {
		return NwkSKeys{}, errNoDevEUI
	}
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		if d.nwkKey == nil {
			return NwkSKeys{}, errNoNwkKey
		}
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveFNwkSIntKey(*d.nwkKey, jn, *dev.JoinEUI, dn),
			SNwkSIntKey: crypto.DeriveSNwkSIntKey(*d.nwkKey, jn, *dev.JoinEUI, dn),
			NwkSEncKey:  crypto.DeriveNwkSEncKey(*d.nwkKey, jn, *dev.JoinEUI, dn),
		}, nil

	default:
		if d.appKey == nil {
			return NwkSKeys{}, errNoAppKey
		}
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveLegacyNwkSKey(*d.appKey, jn, nid, dn),
		}, nil
	}
}

func (d *mem) GetNwkKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	if d.nwkKey == nil {
		return nil, errNoNwkKey
	}
	return d.nwkKey, nil
}

var errNoAppKey = errors.DefineCorruption("no_app_key", "no AppKey specified")

func (d *mem) DeriveAppSKey(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return types.AES128Key{}, errNoJoinEUI
	}
	if dev.DevEUI == nil || dev.DevEUI.IsZero() {
		return types.AES128Key{}, errNoDevEUI
	}
	if d.appKey == nil {
		return types.AES128Key{}, errNoAppKey
	}

	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		return crypto.DeriveAppSKey(*d.appKey, jn, *dev.JoinEUI, dn), nil
	default:
		return crypto.DeriveLegacyAppSKey(*d.appKey, jn, nid, dn), nil
	}
}

func (d *mem) GetAppKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	if d.appKey == nil {
		return nil, errNoAppKey
	}
	return d.appKey, nil
}
