// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"bytes"
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type relayKeyService struct {
	devices DeviceRegistry
	keys    crypto.KeyService
}

var _ mac.RelayKeyService = (*relayKeyService)(nil)

// BatchDeriveRootWorSKey implements mac.RelayKeyService.
func (r *relayKeyService) BatchDeriveRootWorSKey(
	ctx context.Context, appID *ttnpb.ApplicationIdentifiers, deviceIDs []string, sessionKeyIDs [][]byte,
) (devAddrs []*types.DevAddr, keys []*types.AES128Key, err error) {
	if len(deviceIDs) != len(sessionKeyIDs) {
		panic("device IDs and session key IDs must have the same length")
	}
	if len(deviceIDs) == 0 {
		return nil, nil, nil
	}
	devices, err := r.devices.BatchGetByID(
		ctx,
		appID,
		deviceIDs,
		[]string{
			"pending_session.dev_addr",
			"pending_session.keys.nwk_s_enc_key.encrypted_key",
			"pending_session.keys.nwk_s_enc_key.kek_label",
			"pending_session.keys.nwk_s_enc_key.key",
			"pending_session.keys.session_key_id",
			"session.dev_addr",
			"session.keys.nwk_s_enc_key.encrypted_key",
			"session.keys.nwk_s_enc_key.kek_label",
			"session.keys.nwk_s_enc_key.key",
			"session.keys.session_key_id",
		},
	)
	if err != nil {
		return nil, nil, err
	}
	devAddrs, keys = make([]*types.DevAddr, len(deviceIDs)), make([]*types.AES128Key, len(deviceIDs))
	for i, dev := range devices {
		var devAddr types.DevAddr
		var keyEnvelope *ttnpb.KeyEnvelope
		switch {
		case dev.GetPendingSession().GetKeys().GetNwkSEncKey() != nil &&
			bytes.Equal(dev.PendingSession.Keys.SessionKeyId, sessionKeyIDs[i]):
			copy(devAddr[:], dev.PendingSession.DevAddr)
			keyEnvelope = dev.PendingSession.Keys.NwkSEncKey
		case dev.GetSession().GetKeys().GetNwkSEncKey() != nil &&
			bytes.Equal(dev.Session.Keys.SessionKeyId, sessionKeyIDs[i]):
			copy(devAddr[:], dev.Session.DevAddr)
			keyEnvelope = dev.Session.Keys.NwkSEncKey
		default:
			continue
		}
		key, err := cryptoutil.UnwrapAES128Key(ctx, keyEnvelope, r.keys)
		if err != nil {
			return nil, nil, err
		}
		key = crypto.DeriveRootWorSKey(key)
		devAddrs[i], keys[i] = &devAddr, &key
	}
	return devAddrs, keys, nil
}

func (ns *NetworkServer) relayKeyService() mac.RelayKeyService {
	return &relayKeyService{ns.devices, ns.KeyService()}
}
