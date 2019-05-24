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
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"sort"
	"time"

	"github.com/oklog/ulid"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type nsJsServer struct {
	JS *JoinServer
}

var supportedMACVersions = [...]ttnpb.MACVersion{
	ttnpb.MAC_V1_0,
	ttnpb.MAC_V1_0_1,
	ttnpb.MAC_V1_0_2,
	ttnpb.MAC_V1_0_3,
	ttnpb.MAC_V1_1,
}

// HandleJoin is called by the Network Server to join a device.
func (srv nsJsServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (res *ttnpb.JoinResponse, err error) {
	// TODO: Authorize using client TLS and application rights (https://github.com/TheThingsNetwork/lorawan-stack/issues/4)
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx)
	defer func() {
		if err != nil {
			registerRejectJoin(ctx, req, err)
		}
	}()

	supported := false
	for _, ver := range supportedMACVersions {
		if req.SelectedMACVersion == ver {
			supported = true
			break
		}
	}
	if !supported {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes("version", req.SelectedMACVersion)
	}

	req.Payload = &ttnpb.Message{}
	if err = lorawan.UnmarshalMessage(req.RawPayload, req.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}

	if req.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes("version", req.Payload.Major)
	}
	if req.Payload.MType != ttnpb.MType_JOIN_REQUEST {
		return nil, errWrongPayloadType.WithAttributes("type", req.Payload.MType)
	}

	pld := req.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, errNoJoinRequest
	}
	if pld.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	if pld.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}

	match := false
	for _, p := range srv.JS.euiPrefixes {
		if p.Matches(pld.JoinEUI) {
			match = true
			break
		}
	}
	switch {
	case !match && req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0:
		return nil, errUnknownAppEUI
	case !match:
		// TODO: Determine the cluster containing the device.
		// https://github.com/TheThingsNetwork/lorawan-stack/issues/4
		return nil, errForwardJoinRequest
	}

	dev, err := srv.JS.devices.SetByEUI(ctx, pld.JoinEUI, pld.DevEUI,
		[]string{
			"last_dev_nonce",
			"last_join_nonce",
			"resets_join_nonces",
			"root_keys",
			"used_dev_nonces",
			"provisioner_id",
			"provisioning_data",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			paths := make([]string, 0, 3)

			dn := uint32(binary.BigEndian.Uint16(pld.DevNonce[:]))
			if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_0_3) >= 0 {
				if (dn != 0 || dev.LastDevNonce != 0 || dev.LastJoinNonce != 0) && !dev.ResetsJoinNonces {
					if dn <= dev.LastDevNonce {
						return nil, nil, errDevNonceTooSmall
					}
					if dn == math.MaxUint32 {
						return nil, nil, errDevNonceTooHigh
					}
				}
				dev.LastDevNonce = dn
				paths = append(paths, "last_dev_nonce")
			} else {
				i := sort.Search(len(dev.UsedDevNonces), func(i int) bool { return dev.UsedDevNonces[i] >= dn })
				if i < len(dev.UsedDevNonces) && dev.UsedDevNonces[i] == dn {
					return nil, nil, errReuseDevNonce
				}
				dev.UsedDevNonces = append(dev.UsedDevNonces, 0)
				copy(dev.UsedDevNonces[i+1:], dev.UsedDevNonces[i:])
				dev.UsedDevNonces[i] = dn
				paths = append(paths, "used_dev_nonces")
			}

			var b []byte
			if req.CFList == nil {
				b = make([]byte, 0, 17)
			} else {
				b = make([]byte, 0, 33)
			}
			b, err = lorawan.AppendMHDR(b, ttnpb.MHDR{
				MType: ttnpb.MType_JOIN_ACCEPT,
				Major: req.Payload.Major,
			})
			if err != nil {
				return nil, nil, errEncodePayload.WithCause(err)
			}

			if dev.LastJoinNonce >= 1<<24-1 {
				return nil, nil, errJoinNonceTooHigh
			}
			dev.LastJoinNonce++
			paths = append(paths, "last_join_nonce")

			var jn types.JoinNonce
			nb := make([]byte, 4)
			binary.BigEndian.PutUint32(nb, dev.LastJoinNonce)
			copy(jn[:], nb[1:])

			b, err = lorawan.AppendJoinAcceptPayload(b, ttnpb.JoinAcceptPayload{
				NetID:      req.NetID,
				JoinNonce:  jn,
				CFList:     req.CFList,
				DevAddr:    req.DevAddr,
				DLSettings: req.DownlinkSettings,
				RxDelay:    req.RxDelay,
			})
			if err != nil {
				return nil, nil, errEncodePayload.WithCause(err)
			}

			srv.JS.entropyMu.Lock()
			skID, err := ulid.New(ulid.Timestamp(time.Now()), srv.JS.entropy)
			srv.JS.entropyMu.Unlock()
			if err != nil {
				return nil, nil, errGenerateSessionKeyID
			}

			cs := srv.JS.GetPeer(ctx, ttnpb.PeerInfo_CRYPTO_SERVER, dev.EndDeviceIdentifiers)

			var networkCryptoService cryptoservices.Network
			if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) >= 0 && dev.RootKeys != nil && dev.RootKeys.NwkKey != nil {
				// LoRaWAN 1.1 and higher use a NwkKey.
				nwkKey, err := cryptoutil.UnwrapAES128Key(*dev.RootKeys.NwkKey, srv.JS.KeyVault)
				if err != nil {
					return nil, nil, err
				}
				networkCryptoService = cryptoservices.NewMemory(&nwkKey, nil)
			} else if cs != nil {
				networkCryptoService = cryptoservices.NewNetworkRPCClient(cs.Conn(), srv.JS.KeyVault, srv.JS.WithClusterAuth())
			}

			var applicationCryptoService cryptoservices.Application
			if dev.RootKeys != nil && dev.RootKeys.AppKey != nil {
				appKey, err := cryptoutil.UnwrapAES128Key(*dev.RootKeys.AppKey, srv.JS.KeyVault)
				if err != nil {
					return nil, nil, err
				}
				applicationCryptoService = cryptoservices.NewMemory(nil, &appKey)
				if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					// LoRaWAN 1.0.x use the AppKey for network security operations.
					networkCryptoService = cryptoservices.NewMemory(nil, &appKey)
				}
			} else if cs != nil {
				applicationCryptoService = cryptoservices.NewApplicationRPCClient(cs.Conn(), srv.JS.KeyVault, srv.JS.WithClusterAuth())
			}
			if networkCryptoService == nil {
				return nil, nil, errNoNwkKey
			}
			if applicationCryptoService == nil {
				return nil, nil, errNoAppKey
			}

			cryptoDev := &ttnpb.EndDevice{}
			if err := cryptoDev.SetFields(dev, "ids", "provisioner_id", "provisioning_data"); err != nil {
				return nil, nil, err
			}
			reqMIC, err := networkCryptoService.JoinRequestMIC(ctx, cryptoDev, req.SelectedMACVersion, req.RawPayload[:19])
			if err != nil {
				return nil, nil, errComputeMIC.WithCause(err)
			}
			if !bytes.Equal(reqMIC[:], req.RawPayload[19:]) {
				return nil, nil, errMICMismatch
			}
			resMIC, err := networkCryptoService.JoinAcceptMIC(ctx, cryptoDev, req.SelectedMACVersion, 0xff, pld.DevNonce, b)
			if err != nil {
				return nil, nil, errComputeMIC.WithCause(err)
			}
			enc, err := networkCryptoService.EncryptJoinAccept(ctx, cryptoDev, req.SelectedMACVersion, append(b[1:], resMIC[:]...))
			if err != nil {
				return nil, nil, errEncryptPayload.WithCause(err)
			}
			nwkSKeys, err := networkCryptoService.DeriveNwkSKeys(ctx, cryptoDev, req.SelectedMACVersion, jn, pld.DevNonce, req.NetID)
			if err != nil {
				return nil, nil, errDeriveNwkSKeys.WithCause(err)
			}
			appSKey, err := applicationCryptoService.DeriveAppSKey(ctx, cryptoDev, req.SelectedMACVersion, jn, pld.DevNonce, req.NetID)
			if err != nil {
				return nil, nil, errDeriveAppSKey.WithCause(err)
			}

			sessionKeys := ttnpb.SessionKeys{
				SessionKeyID: skID[:],
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					// TODO: Encrypt key with NS KEK https://github.com/TheThingsNetwork/lorawan-stack/issues/5
					Key: &nwkSKeys.FNwkSIntKey,
				},
				AppSKey: &ttnpb.KeyEnvelope{
					// TODO: Encrypt key with AS KEK https://github.com/TheThingsNetwork/lorawan-stack/issues/5
					Key: &appSKey,
				},
			}
			if req.SelectedMACVersion == ttnpb.MAC_V1_1 {
				sessionKeys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					// TODO: Encrypt key with NS KEK https://github.com/TheThingsNetwork/lorawan-stack/issues/5
					Key: &nwkSKeys.SNwkSIntKey,
				}
				sessionKeys.NwkSEncKey = &ttnpb.KeyEnvelope{
					// TODO: Encrypt key with NS KEK https://github.com/TheThingsNetwork/lorawan-stack/issues/5
					Key: &nwkSKeys.NwkSEncKey,
				}
			}

			res = &ttnpb.JoinResponse{
				RawPayload:  append(b[:1], enc...),
				SessionKeys: sessionKeys,
			}
			_, err = srv.JS.keys.SetByID(ctx, *dev.DevEUI, res.SessionKeys.SessionKeyID,
				[]string{
					"session_key_id",
					"f_nwk_s_int_key",
					"s_nwk_s_int_key",
					"nwk_s_enc_key",
					"app_s_key",
				},
				func(stored *ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error) {
					if stored != nil {
						return nil, nil, errDuplicateIdentifiers
					}
					return &res.SessionKeys, []string{
						"session_key_id",
						"f_nwk_s_int_key",
						"s_nwk_s_int_key",
						"nwk_s_enc_key",
						"app_s_key",
					}, nil
				},
			)
			if err != nil {
				return nil, nil, err
			}

			dev.Session = &ttnpb.Session{
				StartedAt:   time.Now().UTC(),
				DevAddr:     req.DevAddr,
				SessionKeys: res.SessionKeys,
			}
			paths = append(paths, "session")

			return dev, paths, nil
		})
	if err != nil {
		logger.WithFields(log.Fields(
			"join_eui", pld.JoinEUI,
			"dev_eui", pld.DevEUI,
		)).WithError(err).Error("Failed to update device")
		return nil, err
	}

	registerAcceptJoin(ctx, dev, req)
	return res, nil
}

// GetNwkSKeys returns the NwkSKeys associated with session keys identified by the supplied request.
func (srv nsJsServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	// TODO: Authorize using client TLS and application rights (https://github.com/TheThingsNetwork/lorawan-stack/issues/4)
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ks, err := srv.JS.keys.GetByID(ctx, req.DevEUI, req.SessionKeyID,
		[]string{
			"f_nwk_s_int_key",
			"nwk_s_enc_key",
			"s_nwk_s_int_key",
		},
	)
	if err != nil {
		return nil, errRegistryOperation.WithCause(err)
	}

	if ks.NwkSEncKey == nil {
		return nil, errNoNwkSEncKey
	}
	if ks.FNwkSIntKey == nil {
		return nil, errNoFNwkSIntKey
	}
	if ks.SNwkSIntKey == nil {
		return nil, errNoSNwkSIntKey
	}

	return &ttnpb.NwkSKeysResponse{
		NwkSEncKey:  *ks.NwkSEncKey,
		FNwkSIntKey: *ks.FNwkSIntKey,
		SNwkSIntKey: *ks.SNwkSIntKey,
	}, nil
}
