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

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
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

// JoinServer implements the Join Server component.
//
// The Join Server exposes the NsJs and DeviceRegistry services.
type JoinServer struct {
	*component.Component
	*deviceregistry.RegistryRPC

	registry    deviceregistry.Interface
	euiPrefixes []types.EUI64Prefix
}

// Config represents the JoinServer configuration.
type Config struct {
	Registry        deviceregistry.Interface `name:"-"`
	JoinEUIPrefixes []types.EUI64Prefix      `name:"join-eui-prefix" description:"JoinEUI prefixes handled by this JS"`
}

// New returns new *JoinServer.
func New(c *component.Component, conf *Config, rpcOptions ...deviceregistry.RPCOption) (*JoinServer, error) {
	rpcOptions = append(rpcOptions, deviceregistry.ForComponents(ttnpb.PeerInfo_JOIN_SERVER))
	registryRPC, err := deviceregistry.NewRPC(c, conf.Registry, rpcOptions...)
	if err != nil {
		return nil, err
	}

	js := &JoinServer{
		Component:   c,
		RegistryRPC: registryRPC,
		registry:    conf.Registry,
		euiPrefixes: conf.JoinEUIPrefixes,
	}
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	c.RegisterGRPC(js)
	return js, nil
}

func checkMIC(key types.AES128Key, rawPayload []byte) error {
	if n := len(rawPayload); n != 23 {
		return errPayloadLengthMismatch.WithAttributes("length", n)
	}
	computed, err := crypto.ComputeJoinRequestMIC(key, rawPayload[:19])
	if err != nil {
		return errMICComputeFailed
	}
	for i := 0; i < 4; i++ {
		if computed[i] != rawPayload[19+i] {
			return errMICMismatch
		}
	}
	return nil
}

// HandleJoin is called by the Network Server to join a device
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
		return nil, errMissingDevAddr
	}

	if req.GetPayload().GetPayload() == nil {
		if req.RawPayload == nil {
			return nil, errMissingPayload
		}
		if req.Payload == nil {
			req.Payload = &ttnpb.Message{}
		}
		if err = req.Payload.UnmarshalLoRaWAN(req.RawPayload); err != nil {
			return nil, errUnmarshalPayloadFailed.WithCause(err)
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
		return nil, errMissingJoinRequest
	}

	if pld.DevEUI.IsZero() {
		return nil, errMissingDevEUI
	}
	if pld.JoinEUI.IsZero() {
		return nil, errMissingJoinEUI
	}

	rawPayload := req.RawPayload
	if rawPayload == nil {
		rawPayload, err = req.Payload.MarshalLoRaWAN()
		if err != nil {
			return nil, errMarshalPayloadFailed.WithCause(err)
		}
	}

	dev, err := deviceregistry.FindByIdentifiers(js.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI:  &pld.DevEUI,
		JoinEUI: &pld.JoinEUI,
	})
	if err != nil {
		return nil, err
	}

	if rpcmetadata.FromIncomingContext(ctx).NetAddress != dev.NetworkServerAddress {
		return nil, errAddressMismatch.WithAttributes("component", "Network Server")
	}

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
		// TODO determine the cluster containing the device
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
		return nil, errMarshalPayloadFailed.WithCause(err)
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
		return nil, errMarshalPayloadFailed.WithCause(err)
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
					return nil, errDevNonceReused
				}
			}
		default:
			panic("This statement is unreachable. Fix version check.")
		}
	}

	if dev.RootKeys.AppKey == nil || len(dev.RootKeys.AppKey.Key) == 0 {
		return nil, errMissingAppKey
	}

	var appKey types.AES128Key
	if dev.RootKeys.AppKey.KEKLabel != "" {
		// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
		panic("Unsupported")
	}
	copy(appKey[:], dev.RootKeys.AppKey.Key[:])

	switch req.SelectedMacVersion {
	case ttnpb.MAC_V1_1:
		if dev.RootKeys.NwkKey == nil || len(dev.RootKeys.NwkKey.Key) == 0 {
			return nil, errMissingNwkKey
		}

		var nwkKey types.AES128Key
		if dev.RootKeys.NwkKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(nwkKey[:], dev.RootKeys.NwkKey.Key[:])

		if err := checkMIC(nwkKey, rawPayload); err != nil {
			return nil, errMICCheckFailed.WithCause(err)
		}

		mic, err := crypto.ComputeJoinAcceptMIC(crypto.DeriveJSIntKey(nwkKey, pld.DevEUI), 0xff, pld.JoinEUI, pld.DevNonce, b)
		if err != nil {
			return nil, errMICComputeFailed.WithCause(err)
		}

		enc, err := crypto.EncryptJoinAccept(nwkKey, append(b[1:], mic[:]...))
		if err != nil {
			return nil, errEncryptPayloadFailed.WithCause(err)
		}

		resp = &ttnpb.JoinResponse{
			RawPayload: append(b[:1], enc...),
			SessionKeys: ttnpb.SessionKeys{
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
			return nil, errMICCheckFailed.WithCause(err)
		}

		mic, err := crypto.ComputeLegacyJoinAcceptMIC(appKey, b)
		if err != nil {
			return nil, errMICComputeFailed.WithCause(err)
		}

		enc, err := crypto.EncryptJoinAccept(appKey, append(b[1:], mic[:]...))
		if err != nil {
			return nil, errEncryptPayloadFailed.WithCause(err)
		}
		resp = &ttnpb.JoinResponse{
			RawPayload: append(b[:1], enc...),
			SessionKeys: ttnpb.SessionKeys{
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

	dev.UsedDevNonces = append(dev.UsedDevNonces, uint32(dn))
	dev.NextJoinNonce++
	dev.EndDevice.Session = &ttnpb.Session{
		StartedAt:   time.Now().UTC(),
		DevAddr:     *req.EndDeviceIdentifiers.DevAddr,
		SessionKeys: resp.SessionKeys,
	}
	if err := dev.Store(); err != nil {
		logger.WithFields(log.Fields(
			"dev_eui", dev.EndDeviceIdentifiers.DevEUI,
			"join_eui", dev.EndDeviceIdentifiers.JoinEUI,
			"application_id", dev.EndDeviceIdentifiers.ApplicationID,
			"device_id", dev.EndDeviceIdentifiers.DeviceID,
		)).WithError(err).Error("Failed to update device")
	}
	registerAcceptJoin(ctx, dev.EndDevice, req)
	return resp, nil
}

// GetAppSKey returns the AppSKey associated with device specified by the supplied request.
func (js *JoinServer) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	if req.DevEUI.IsZero() {
		return nil, errMissingDevEUI
	}
	if req.SessionKeyID == "" {
		return nil, errMissingSessionKeyID
	}

	dev, err := deviceregistry.FindByIdentifiers(js.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI: &req.DevEUI,
	})
	if err != nil {
		return nil, err
	}

	if rpcmetadata.FromIncomingContext(ctx).NetAddress != dev.ApplicationServerAddress {
		return nil, errAddressMismatch.WithAttributes("component", "Application Server")
	}

	if dev.Session == nil {
		return nil, errNoSession
	}
	if dev.Session.SessionKeyID != req.SessionKeyID {
		dev.Session = dev.SessionFallback
		if dev.Session == nil || dev.Session.SessionKeyID != req.SessionKeyID {
			return nil, errSessionKeyIDMismatch
		}
	}

	if dev.Session.AppSKey == nil {
		return nil, errMissingAppSKey
	}
	// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
	return &ttnpb.AppSKeyResponse{
		AppSKey: *dev.Session.AppSKey,
	}, nil
}

// GetNwkSKeys returns the NwkSKeys associated with device specified by the supplied request.
func (js *JoinServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if req.DevEUI.IsZero() {
		return nil, errMissingDevEUI
	}
	if req.SessionKeyID == "" {
		return nil, errMissingSessionKeyID
	}

	dev, err := deviceregistry.FindByIdentifiers(js.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI: &req.DevEUI,
	})
	if err != nil {
		return nil, err
	}

	if rpcmetadata.FromIncomingContext(ctx).NetAddress != dev.NetworkServerAddress {
		return nil, errAddressMismatch.WithAttributes("component", "Network Server")
	}

	if dev.Session == nil {
		return nil, errNoSession
	}
	if dev.Session.SessionKeyID != req.SessionKeyID {
		dev.Session = dev.SessionFallback
		if dev.Session == nil || dev.Session.SessionKeyID != req.SessionKeyID {
			return nil, errSessionKeyIDMismatch
		}
	}

	if dev.Session.NwkSEncKey == nil {
		return nil, errMissingNwkSEncKey
	}
	if dev.Session.FNwkSIntKey == nil {
		return nil, errMissingFNwkSIntKey
	}
	if dev.Session.SNwkSIntKey == nil {
		return nil, errMissingSNwkSIntKey
	}
	// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
	return &ttnpb.NwkSKeysResponse{
		NwkSEncKey:  *dev.Session.NwkSEncKey,
		FNwkSIntKey: *dev.Session.FNwkSIntKey,
		SNwkSIntKey: *dev.Session.SNwkSIntKey,
	}, nil
}

// Roles of the gRPC service
func (js *JoinServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_JOIN_SERVER}
}

// RegisterServices registers services provided by js at s.
func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsJsServer(s, js)
	ttnpb.RegisterJsDeviceRegistryServer(s, js)
}

// RegisterHandlers registers gRPC handlers.
func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterJsDeviceRegistryHandler(js.Context(), s, conn)
}
