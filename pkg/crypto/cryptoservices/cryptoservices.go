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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// NwkSKeys contains network session keys.
type NwkSKeys struct {
	FNwkSIntKey,
	SNwkSIntKey,
	NwkSEncKey types.AES128Key
}

// Network performs network layer cryptographic operations.
type Network interface {
	JoinRequestMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([4]byte, error)
	JoinAcceptMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) ([4]byte, error)
	EncryptJoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error)
	EncryptRejoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error)
	DeriveNwkSKeys(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (NwkSKeys, error)
	// GetNwkKey returns the NwkKey of the given end device.
	// If the implementation does not expose root keys, this method returns nil, nil.
	GetNwkKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error)
}

// Application performs application layer cryptographic operations.
type Application interface {
	DeriveAppSKey(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error)
	// GetAppKey returns the AppKey of the given end device.
	// If the implementation does not expose root keys, this method returns nil, nil.
	GetAppKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error)
}

// NetworkApplication is an interface that combines Network and Application.
type NetworkApplication interface {
	Network
	Application
}
