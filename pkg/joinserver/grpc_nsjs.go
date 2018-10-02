// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package joinserver provides a LoRaWAN 1.1-compliant Join Server implementation.
package joinserver

import (
	"context"
	"encoding/binary"
	"math"
	"time"

	"github.com/oklog/ulid"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

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
func (js *JoinServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (resp *ttnpb.JoinResponse, err error) {
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
		if req.SelectedMacVersion == ver {
			supported = true
			break
		}
	}
	if !supported {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes("version", req.SelectedMacVersion)
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
		if err = req.Payload.UnmarshalLoRaWAN(req.RawPayload); err != nil {
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
		rawPayload, err = req.Payload.MarshalLoRaWAN()
		if err != nil {
			return nil, errEncodePayload.WithCause(err)
		}
	}

	dev, err := js.devices.SetByEUI(ctx, pld.JoinEUI, pld.DevEUI, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
		match := false
		for _, p := range js.euiPrefixes {
			if p.Matches(pld.JoinEUI) {
				match = true
				break
			}
		}
		switch {
		case !match && dev.LoRaWANVersion == ttnpb.MAC_V1_0:
			return nil, errUnknownAppEUI
		case !match:
			// TODO: Determine the cluster containing the device.
			// https://github.com/TheThingsIndustries/ttn/issues/244
			return nil, errForwardJoinRequest
		}

		// Registered version is lower than selected.
		if dev.LoRaWANVersion.Compare(req.SelectedMacVersion) == -1 {
			return nil, errMACVersionMismatch.WithAttributes("registered", dev.LoRaWANVersion, "selected", req.SelectedMacVersion)
		}

		var b []byte
		if req.CFList == nil {
			b = make([]byte, 0, 17)
		} else {
			b = make([]byte, 0, 33)
		}

		b, err = (&ttnpb.MHDR{
			MType: ttnpb.MType_JOIN_ACCEPT,
			Major: req.Payload.Major,
		}).AppendLoRaWAN(b)
		if err != nil {
			return nil, errEncodePayload.WithCause(err)
		}

		var jn types.JoinNonce
		nb := make([]byte, 4)
		binary.BigEndian.PutUint32(nb, dev.NextJoinNonce)
		copy(jn[:], nb[1:])

		b, err = (&ttnpb.JoinAcceptPayload{
			NetID:      req.NetID,
			JoinNonce:  jn,
			CFList:     req.CFList,
			DevAddr:    *req.EndDeviceIdentifiers.DevAddr,
			DLSettings: req.DownlinkSettings,
			RxDelay:    req.RxDelay,
		}).AppendLoRaWAN(b)
		if err != nil {
			return nil, errDecodePayload.WithCause(err)
		}

		dn := binary.BigEndian.Uint16(pld.DevNonce[:])
		if !dev.ResetsJoinNonces {
			switch req.SelectedMacVersion {
			case ttnpb.MAC_V1_1:
				if uint32(dn) < dev.NextDevNonce {
					return nil, errDevNonceTooSmall
				}
				if dev.NextDevNonce == math.MaxUint32 {
					return nil, errDevNonceTooHigh
				}
				dev.NextDevNonce = uint32(dn + 1)
			case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
				for _, used := range dev.UsedDevNonces {
					if dn == uint16(used) {
						return nil, errReuseDevNonce
					}
				}
			default:
				panic("This statement is unreachable. Fix version check.")
			}
		}

		if dev.RootKeys.AppKey == nil || len(dev.RootKeys.AppKey.Key) == 0 {
			return nil, errNoAppKey
		}

		var appKey types.AES128Key
		if dev.RootKeys.AppKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(appKey[:], dev.RootKeys.AppKey.Key[:])

		js.entropyMu.Lock()
		skID, err := ulid.New(ulid.Timestamp(time.Now()), js.entropy)
		js.entropyMu.Unlock()
		if err != nil {
			return nil, errGenerateSessionKeyID
		}

		switch req.SelectedMacVersion {
		case ttnpb.MAC_V1_1:
			if dev.RootKeys.NwkKey == nil || len(dev.RootKeys.NwkKey.Key) == 0 {
				return nil, errNoNwkKey
			}

			var nwkKey types.AES128Key
			if dev.RootKeys.NwkKey.KEKLabel != "" {
				// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
				panic("Unsupported")
			}
			copy(nwkKey[:], dev.RootKeys.NwkKey.Key[:])

			if err := checkMIC(nwkKey, rawPayload); err != nil {
				return nil, errCheckMIC.WithCause(err)
			}

			mic, err := crypto.ComputeJoinAcceptMIC(crypto.DeriveJSIntKey(nwkKey, pld.DevEUI), 0xff, pld.JoinEUI, pld.DevNonce, b)
			if err != nil {
				return nil, errComputeMIC.WithCause(err)
			}

			enc, err := crypto.EncryptJoinAccept(nwkKey, append(b[1:], mic[:]...))
			if err != nil {
				return nil, errEncryptPayload.WithCause(err)
			}

			resp = &ttnpb.JoinResponse{
				RawPayload: append(b[:1], enc...),
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: skID.String(),
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      keyToBytes(crypto.DeriveFNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
						KEKLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      keyToBytes(crypto.DeriveSNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key:      keyToBytes(crypto.DeriveNwkSEncKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
						KEKLabel: "",
					},
					// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
					AppSKey: &ttnpb.KeyEnvelope{
						Key:      keyToBytes(crypto.DeriveAppSKey(appKey, jn, pld.JoinEUI, pld.DevNonce)),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			}

		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			if err := checkMIC(appKey, rawPayload); err != nil {
				return nil, errCheckMIC.WithCause(err)
			}

			mic, err := crypto.ComputeLegacyJoinAcceptMIC(appKey, b)
			if err != nil {
				return nil, errComputeMIC.WithCause(err)
			}

			enc, err := crypto.EncryptJoinAccept(appKey, append(b[1:], mic[:]...))
			if err != nil {
				return nil, errEncryptPayload.WithCause(err)
			}
			resp = &ttnpb.JoinResponse{
				RawPayload: append(b[:1], enc...),
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: skID.String(),
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      keyToBytes(crypto.DeriveLegacyNwkSKey(appKey, jn, req.NetID, pld.DevNonce)),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key:      keyToBytes(crypto.DeriveLegacyAppSKey(appKey, jn, req.NetID, pld.DevNonce)),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			}

		default:
			panic("This statement is unreachable. Fix version check.")
		}

		_, err = CreateKeys(ctx, js.keys, *dev.EndDeviceIdentifiers.DevEUI, &resp.SessionKeys)
		if err != nil {
			return nil, err
		}

		dev.UsedDevNonces = append(dev.UsedDevNonces, uint32(dn))
		dev.NextJoinNonce++
		dev.Session = &ttnpb.Session{
			StartedAt:   time.Now().UTC(),
			DevAddr:     *req.EndDeviceIdentifiers.DevAddr,
			SessionKeys: resp.SessionKeys,
		}
		return dev, nil
	})
	if err != nil {
		logger.WithFields(log.Fields(
			"join_eui", pld.JoinEUI,
			"dev_eui", pld.DevEUI,
		)).WithError(err).Error("Failed to update device")
		return nil, err
	}

	registerAcceptJoin(ctx, dev, req)
	return resp, nil
}

// GetNwkSKeys returns the NwkSKeys associated with session keys identified by the supplied request.
func (js *JoinServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if req.DevEUI.IsZero() {
		return nil, errInvalidRequest.WithCause(errNoDevEUI)
	}
	if req.SessionKeyID == "" {
		return nil, errInvalidRequest.WithCause(errNoSessionKeyID)
	}

	ks, err := js.keys.GetByID(ctx, req.DevEUI, req.SessionKeyID)
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
