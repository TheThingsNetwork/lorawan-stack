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
	"encoding/binary"
	"math"
	"sort"
	"time"

	"github.com/oklog/ulid"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
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
	ttnpb.MAC_V1_1,
}

func keyToBytes(k types.AES128Key) []byte {
	return k[:]
}

func checkMIC(key types.AES128Key, rawPayload []byte) error {
	if n := len(rawPayload); n != 23 {
		return errPayloadLengthMismatch.WithAttributes("length", n)
	}
	computed, err := crypto.ComputeJoinRequestMIC(key, rawPayload[:19])
	if err != nil {
		return errComputeMIC
	}
	for i := 0; i < 4; i++ {
		if computed[i] != rawPayload[19+i] {
			return errMICMismatch
		}
	}
	return nil
}

// HandleJoin is called by the Network Server to join a device.
func (srv nsJsServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (res *ttnpb.JoinResponse, err error) {
	// TODO: Authorize using client TLS and application rights (https://github.com/TheThingsIndustries/lorawan-stack/issues/244)
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

	if req.EndDeviceIdentifiers.DevAddr == nil {
		return nil, errNoDevAddr
	}

	if req.GetPayload().GetPayload() == nil {
		if req.RawPayload == nil {
			return nil, errNoPayload
		}
		if req.Payload == nil {
			req.Payload = &ttnpb.Message{}
		}
		if err = lorawan.UnmarshalMessage(req.RawPayload, req.Payload); err != nil {
			return nil, errDecodePayload.WithCause(err)
		}
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

	rawPayload := req.RawPayload
	if rawPayload == nil {
		rawPayload, err = lorawan.MarshalMessage(*req.Payload)
		if err != nil {
			return nil, errEncodePayload.WithCause(err)
		}
	}

	dev, err := srv.JS.devices.SetByEUI(ctx, pld.JoinEUI, pld.DevEUI,
		[]string{
			"last_dev_nonce",
			"last_join_nonce",
			"resets_join_nonces",
			"root_keys",
			"used_dev_nonces",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			paths := make([]string, 0, 3)

			match := false
			for _, p := range srv.JS.euiPrefixes {
				if p.Matches(pld.JoinEUI) {
					match = true
					break
				}
			}
			switch {
			case !match && req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0:
				return nil, nil, errUnknownAppEUI
			case !match:
				// TODO: Determine the cluster containing the device.
				// https://github.com/TheThingsIndustries/ttn/issues/244
				return nil, nil, errForwardJoinRequest
			}

			dn := uint32(binary.BigEndian.Uint16(pld.DevNonce[:]))
			switch req.SelectedMACVersion {
			case ttnpb.MAC_V1_1:
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
			case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
				i := sort.Search(len(dev.UsedDevNonces), func(i int) bool { return dev.UsedDevNonces[i] >= dn })
				if i < len(dev.UsedDevNonces) && dev.UsedDevNonces[i] == dn {
					return nil, nil, errReuseDevNonce
				}
				dev.UsedDevNonces = append(dev.UsedDevNonces, 0)
				copy(dev.UsedDevNonces[i+1:], dev.UsedDevNonces[i:])
				dev.UsedDevNonces[i] = dn
				paths = append(paths, "used_dev_nonces")
			default:
				panic("This statement is unreachable. Fix version check.")
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
				DevAddr:    *req.EndDeviceIdentifiers.DevAddr,
				DLSettings: req.DownlinkSettings,
				RxDelay:    req.RxDelay,
			})
			if err != nil {
				return nil, nil, errEncodePayload.WithCause(err)
			}

			if dev.RootKeys.AppKey == nil {
				return nil, nil, errNoAppKey
			}

			var appKey types.AES128Key
			if dev.RootKeys.AppKey.KEKLabel != "" {
				// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
				panic("Unsupported")
			}
			copy(appKey[:], dev.RootKeys.AppKey.Key[:])

			srv.JS.entropyMu.Lock()
			skID, err := ulid.New(ulid.Timestamp(time.Now()), srv.JS.entropy)
			srv.JS.entropyMu.Unlock()
			if err != nil {
				return nil, nil, errGenerateSessionKeyID
			}

			switch req.SelectedMACVersion {
			case ttnpb.MAC_V1_1:
				if dev.RootKeys.NwkKey == nil {
					return nil, nil, errNoNwkKey
				}

				var nwkKey types.AES128Key
				if dev.RootKeys.NwkKey.KEKLabel != "" {
					// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
					panic("Unsupported")
				}
				copy(nwkKey[:], dev.RootKeys.NwkKey.Key[:])

				if err := checkMIC(nwkKey, rawPayload); err != nil {
					return nil, nil, errCheckMIC.WithCause(err)
				}

				mic, err := crypto.ComputeJoinAcceptMIC(crypto.DeriveJSIntKey(nwkKey, pld.DevEUI), 0xff, pld.JoinEUI, pld.DevNonce, b)
				if err != nil {
					return nil, nil, errComputeMIC.WithCause(err)
				}

				enc, err := crypto.EncryptJoinAccept(nwkKey, append(b[1:], mic[:]...))
				if err != nil {
					return nil, nil, errEncryptPayload.WithCause(err)
				}

				res = &ttnpb.JoinResponse{
					RawPayload: append(b[:1], enc...),
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: skID[:],
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							// TODO: Encrypt key with NS KEK https://github.com/TheThingsIndustries/ttn/issues/271
							Key:      keyToBytes(crypto.DeriveFNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
							KEKLabel: "",
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							// TODO: Encrypt key with NS KEK https://github.com/TheThingsIndustries/ttn/issues/271
							Key:      keyToBytes(crypto.DeriveSNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
							KEKLabel: "",
						},
						NwkSEncKey: &ttnpb.KeyEnvelope{
							// TODO: Encrypt key with NS KEK https://github.com/TheThingsIndustries/ttn/issues/271
							Key:      keyToBytes(crypto.DeriveNwkSEncKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
							KEKLabel: "",
						},
						AppSKey: &ttnpb.KeyEnvelope{
							// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
							Key:      keyToBytes(crypto.DeriveAppSKey(appKey, jn, pld.JoinEUI, pld.DevNonce)),
							KEKLabel: "",
						},
					},
					Lifetime: 0,
				}

			case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
				if err := checkMIC(appKey, rawPayload); err != nil {
					return nil, nil, errCheckMIC.WithCause(err)
				}

				mic, err := crypto.ComputeLegacyJoinAcceptMIC(appKey, b)
				if err != nil {
					return nil, nil, errComputeMIC.WithCause(err)
				}

				enc, err := crypto.EncryptJoinAccept(appKey, append(b[1:], mic[:]...))
				if err != nil {
					return nil, nil, errEncryptPayload.WithCause(err)
				}

				res = &ttnpb.JoinResponse{
					RawPayload: append(b[:1], enc...),
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: skID[:],
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							// TODO: Encrypt key with NS KEK https://github.com/TheThingsIndustries/ttn/issues/271
							Key:      keyToBytes(crypto.DeriveLegacyNwkSKey(appKey, jn, req.NetID, pld.DevNonce)),
							KEKLabel: "",
						},
						AppSKey: &ttnpb.KeyEnvelope{
							// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
							Key:      keyToBytes(crypto.DeriveLegacyAppSKey(appKey, jn, req.NetID, pld.DevNonce)),
							KEKLabel: "",
						},
					},
					Lifetime: 0,
				}

			default:
				panic("This statement is unreachable. Fix version check.")
			}

			_, err = CreateKeys(ctx, srv.JS.keys, *dev.EndDeviceIdentifiers.DevEUI, &res.SessionKeys)
			if err != nil {
				return nil, nil, err
			}

			dev.Session = &ttnpb.Session{
				StartedAt:   time.Now().UTC(),
				DevAddr:     *req.EndDeviceIdentifiers.DevAddr,
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
	// TODO: Authorize using client TLS and application rights (https://github.com/TheThingsIndustries/lorawan-stack/issues/244)
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
