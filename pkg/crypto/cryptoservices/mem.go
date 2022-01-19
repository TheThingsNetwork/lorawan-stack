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

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
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

func (d *mem) JoinRequestMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) (res [4]byte, err error) {
	if d.nwkKey == nil {
		return [4]byte{}, errNoNwkKey.New()
	}
	return crypto.ComputeJoinRequestMIC(*d.nwkKey, payload)
}

var (
	errNoDevEUI  = errors.DefineCorruption("no_dev_eui", "no DevEUI specified")
	errNoJoinEUI = errors.DefineCorruption("no_join_eui", "no JoinEUI specified")
)

func (d *mem) JoinAcceptMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) ([4]byte, error) {
	if dev.Ids == nil || dev.Ids.JoinEui == nil {
		return [4]byte{}, errNoJoinEUI.New()
	}
	if dev.Ids.DevEui == nil || dev.Ids.DevEui.IsZero() {
		return [4]byte{}, errNoDevEUI.New()
	}
	if d.nwkKey == nil {
		return [4]byte{}, errNoNwkKey.New()
	}
	switch {
	case version.Compare(ttnpb.MACVersion_MAC_V1_1) >= 0:
		jsIntKey := crypto.DeriveJSIntKey(*d.nwkKey, *dev.Ids.DevEui)
		return crypto.ComputeJoinAcceptMIC(jsIntKey, joinReqType, *dev.Ids.JoinEui, dn, payload)
	default:
		return crypto.ComputeLegacyJoinAcceptMIC(*d.nwkKey, payload)
	}
}

func (d *mem) EncryptJoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	if d.nwkKey == nil {
		return nil, errNoNwkKey.New()
	}
	return crypto.EncryptJoinAccept(*d.nwkKey, payload)
}

func (d *mem) EncryptRejoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	if version.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
		panic("This statement is unreachable. Please version check.")
	}
	if dev.Ids == nil || dev.Ids.JoinEui == nil {
		return nil, errNoJoinEUI.New()
	}
	if dev.Ids.DevEui == nil || dev.Ids.DevEui.IsZero() {
		return nil, errNoDevEUI.New()
	}
	if d.nwkKey == nil {
		return nil, errNoNwkKey.New()
	}
	jsEncKey := crypto.DeriveJSEncKey(*d.nwkKey, *dev.Ids.DevEui)
	return crypto.EncryptJoinAccept(jsEncKey, payload)
}

func (d *mem) DeriveNwkSKeys(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (NwkSKeys, error) {
	if dev.Ids == nil || dev.Ids.JoinEui == nil {
		return NwkSKeys{}, errNoJoinEUI.New()
	}
	if dev.Ids.DevEui == nil || dev.Ids.DevEui.IsZero() {
		return NwkSKeys{}, errNoDevEUI.New()
	}
	if d.nwkKey == nil {
		return NwkSKeys{}, errNoNwkKey.New()
	}
	switch {
	case version.Compare(ttnpb.MACVersion_MAC_V1_1) >= 0:
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveFNwkSIntKey(*d.nwkKey, jn, *dev.Ids.JoinEui, dn),
			SNwkSIntKey: crypto.DeriveSNwkSIntKey(*d.nwkKey, jn, *dev.Ids.JoinEui, dn),
			NwkSEncKey:  crypto.DeriveNwkSEncKey(*d.nwkKey, jn, *dev.Ids.JoinEui, dn),
		}, nil
	default:
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveLegacyNwkSKey(*d.nwkKey, jn, nid, dn),
		}, nil
	}
}

func (d *mem) GetNwkKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	if d.nwkKey == nil {
		return nil, errNoNwkKey.New()
	}
	return d.nwkKey, nil
}

var errNoAppKey = errors.DefineCorruption("no_app_key", "no AppKey specified")

func (d *mem) DeriveAppSKey(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	if dev.Ids == nil || dev.Ids.JoinEui == nil {
		return types.AES128Key{}, errNoJoinEUI.New()
	}
	if dev.Ids.DevEui == nil || dev.Ids.DevEui.IsZero() {
		return types.AES128Key{}, errNoDevEUI.New()
	}
	if d.appKey == nil {
		return types.AES128Key{}, errNoAppKey.New()
	}

	switch {
	case version.Compare(ttnpb.MACVersion_MAC_V1_1) >= 0:
		return crypto.DeriveAppSKey(*d.appKey, jn, *dev.Ids.JoinEui, dn), nil
	default:
		return crypto.DeriveLegacyAppSKey(*d.appKey, jn, nid, dn), nil
	}
}

func (d *mem) GetAppKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	if d.appKey == nil {
		return nil, errNoAppKey.New()
	}
	return d.appKey, nil
}
