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
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// GetKeysFunc returns the root keys for the given end device.
type GetKeysFunc func(ctx context.Context, dev *ttnpb.EndDevice) (nwkKey, appKey *types.AES128Key, err error)

type mem struct {
	getKeys GetKeysFunc
}

// NewMemory returns a network and application service using the given fixed root keys,
// performing cryptograhpic operations in memory.
func NewMemory(nwkKey, appKey *types.AES128Key) NetworkApplication {
	return &mem{
		getKeys: func(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, *types.AES128Key, error) {
			return nwkKey, appKey, nil
		},
	}
}

// NewPerDevice returns a network and application service using per-device root keys,
// performing cryptograhpic operations in memory.
func NewPerDevice(getKeys GetKeysFunc) NetworkApplication {
	return &mem{getKeys: getKeys}
}

var (
	errNoNwkKey = errors.DefineCorruption("no_nwk_key", "no NwkKey specified")
	errNoAppKey = errors.DefineCorruption("no_app_key", "no AppKey specified")
)

func (d *mem) nwkKey(ctx context.Context, dev *ttnpb.EndDevice) (types.AES128Key, error) {
	// It is assumed that the underlying key getter provides the fallback to AppKey if applicable
	// (e.g. for 1.0.x devices). If the key getter returns no NwkKey, an error is returned here.
	nwkKey, _, err := d.getKeys(ctx, dev)
	if err != nil {
		return types.AES128Key{}, err
	}
	if nwkKey == nil {
		return types.AES128Key{}, errNoNwkKey.New()
	}
	return *nwkKey, nil
}

func (d *mem) appKey(ctx context.Context, dev *ttnpb.EndDevice) (types.AES128Key, error) {
	_, appKey, err := d.getKeys(ctx, dev)
	if err != nil {
		return types.AES128Key{}, err
	}
	if appKey == nil {
		return types.AES128Key{}, errNoAppKey.New()
	}
	return *appKey, nil
}

// JoinRequestMIC implements NetworkApplication.
func (d *mem) JoinRequestMIC(
	ctx context.Context, dev *ttnpb.EndDevice, _ ttnpb.MACVersion, payload []byte,
) (res [4]byte, err error) {
	nwkKey, err := d.nwkKey(ctx, dev)
	if err != nil {
		return [4]byte{}, err
	}
	return crypto.ComputeJoinRequestMIC(nwkKey, payload)
}

var (
	errNoDevEUI  = errors.DefineCorruption("no_dev_eui", "no DevEUI specified")
	errNoJoinEUI = errors.DefineCorruption("no_join_eui", "no JoinEUI specified")
)

// JoinAcceptMIC implements NetworkApplication.
func (d *mem) JoinAcceptMIC(
	ctx context.Context,
	dev *ttnpb.EndDevice,
	version ttnpb.MACVersion,
	joinReqType byte,
	dn types.DevNonce,
	payload []byte,
) ([4]byte, error) {
	if dev.Ids == nil || len(dev.Ids.JoinEui) == 0 {
		return [4]byte{}, errNoJoinEUI.New()
	}
	if types.MustEUI64(dev.Ids.DevEui).OrZero().IsZero() {
		return [4]byte{}, errNoDevEUI.New()
	}
	nwkKey, err := d.nwkKey(ctx, dev)
	if err != nil {
		return [4]byte{}, err
	}
	switch {
	case macspec.UseNwkKey(version):
		jsIntKey := crypto.DeriveJSIntKey(nwkKey, types.MustEUI64(dev.Ids.DevEui).OrZero())
		return crypto.ComputeJoinAcceptMIC(jsIntKey, joinReqType, types.MustEUI64(dev.Ids.JoinEui).OrZero(), dn, payload)
	default:
		return crypto.ComputeLegacyJoinAcceptMIC(nwkKey, payload)
	}
}

// EncryptJoinAccept implements NetworkApplication.
func (d *mem) EncryptJoinAccept(
	ctx context.Context, dev *ttnpb.EndDevice, _ ttnpb.MACVersion, payload []byte,
) ([]byte, error) {
	nwkKey, err := d.nwkKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	return crypto.EncryptJoinAccept(nwkKey, payload)
}

// EncryptRejoinAccept implements NetworkApplication.
func (d *mem) EncryptRejoinAccept(
	ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte,
) ([]byte, error) {
	if !macspec.UseNwkKey(version) {
		panic("This statement is unreachable. Please version check.")
	}
	if dev.Ids == nil || len(dev.Ids.JoinEui) == 0 {
		return nil, errNoJoinEUI.New()
	}
	if types.MustEUI64(dev.Ids.DevEui).OrZero().IsZero() {
		return nil, errNoDevEUI.New()
	}
	nwkKey, err := d.nwkKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	jsEncKey := crypto.DeriveJSEncKey(nwkKey, types.MustEUI64(dev.Ids.DevEui).OrZero())
	return crypto.EncryptJoinAccept(jsEncKey, payload)
}

// DeriveNwkSKeys implements NetworkApplication.
func (d *mem) DeriveNwkSKeys(
	ctx context.Context,
	dev *ttnpb.EndDevice,
	version ttnpb.MACVersion,
	jn types.JoinNonce,
	dn types.DevNonce,
	nid types.NetID,
) (NwkSKeys, error) {
	if dev.Ids == nil || len(dev.Ids.JoinEui) == 0 {
		return NwkSKeys{}, errNoJoinEUI.New()
	}
	if types.MustEUI64(dev.Ids.DevEui).OrZero().IsZero() {
		return NwkSKeys{}, errNoDevEUI.New()
	}
	nwkKey, err := d.nwkKey(ctx, dev)
	if err != nil {
		return NwkSKeys{}, err
	}
	switch {
	case macspec.UseNwkKey(version):
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveFNwkSIntKey(nwkKey, jn, types.MustEUI64(dev.Ids.JoinEui).OrZero(), dn),
			SNwkSIntKey: crypto.DeriveSNwkSIntKey(nwkKey, jn, types.MustEUI64(dev.Ids.JoinEui).OrZero(), dn),
			NwkSEncKey:  crypto.DeriveNwkSEncKey(nwkKey, jn, types.MustEUI64(dev.Ids.JoinEui).OrZero(), dn),
		}, nil
	default:
		return NwkSKeys{
			FNwkSIntKey: crypto.DeriveLegacyNwkSKey(nwkKey, jn, nid, dn),
		}, nil
	}
}

// GetNwkKey implements NetworkApplication.
func (d *mem) GetNwkKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	nwkKey, err := d.nwkKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	return &nwkKey, nil
}

// DeriveAppSKey implements NetworkApplication.
func (d *mem) DeriveAppSKey(
	ctx context.Context,
	dev *ttnpb.EndDevice,
	version ttnpb.MACVersion,
	jn types.JoinNonce,
	dn types.DevNonce,
	nid types.NetID,
) (types.AES128Key, error) {
	if dev.Ids == nil || len(dev.Ids.JoinEui) == 0 {
		return types.AES128Key{}, errNoJoinEUI.New()
	}
	if types.MustEUI64(dev.Ids.DevEui).OrZero().IsZero() {
		return types.AES128Key{}, errNoDevEUI.New()
	}
	appKey, err := d.appKey(ctx, dev)
	if err != nil {
		return types.AES128Key{}, err
	}

	switch {
	case macspec.UseNwkKey(version):
		return crypto.DeriveAppSKey(appKey, jn, types.MustEUI64(dev.Ids.JoinEui).OrZero(), dn), nil
	default:
		return crypto.DeriveLegacyAppSKey(appKey, jn, nid, dn), nil
	}
}

// GetAppKey implements NetworkApplication.
func (d *mem) GetAppKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	appKey, err := d.appKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	return &appKey, nil
}
