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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// NetworkCryptoService performs network layer cryptographic operations.
type NetworkCryptoService interface {
	JoinRequestMIC(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([4]byte, error)
	JoinAcceptMIC(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, joinReqType byte, dn types.DevNonce, payload []byte) ([4]byte, error)
	EncryptJoinAccept(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([]byte, error)
	EncryptRejoinAccept(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, payload []byte) ([]byte, error)
	DeriveNwkSKeys(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (fNwkSIntKey, sNwkSIntKey, nwkSEncKey types.AES128Key, err error)
}

// ApplicationCryptoService performs application layer cryptographic operations.
type ApplicationCryptoService interface {
	DeriveAppSKey(ctx context.Context, ids ttnpb.CryptoServiceEndDeviceIdentifiers, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (appSKey types.AES128Key, err error)
}
