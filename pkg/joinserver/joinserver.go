// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var supportedMACVersions = [...]ttnpb.MACVersion{
	ttnpb.MAC_V1_0,
	ttnpb.MAC_V1_0_1,
	ttnpb.MAC_V1_0_2,
	ttnpb.MAC_V1_1,
}

// JoinServer implements the join server component.
//
// The join server exposes the NsJs and DeviceRegistry services.
type JoinServer struct {
	*component.Component
	*deviceregistry.RegistryRPC

	registry    deviceregistry.Interface
	euiPrefixes []types.EUI64Prefix
}

// Config represents the JoinServer configuration.
type Config struct {
	Component       *component.Component
	Registry        deviceregistry.Interface
	JoinEUIPrefixes []types.EUI64Prefix
}

// New returns new *JoinServer.
func New(conf *Config) *JoinServer {
	return &JoinServer{
		Component:   conf.Component,
		RegistryRPC: deviceregistry.NewRPC(conf.Component, conf.Registry),
		registry:    conf.Registry,
		euiPrefixes: conf.JoinEUIPrefixes,
	}
}

func keyPointer(key types.AES128Key) *types.AES128Key {
	return &key
}

func checkMIC(key types.AES128Key, rawPayload []byte) error {
	if n := len(rawPayload); n != 23 {
		return errors.Errorf("Expected length of raw payload to be equal to 23, got %d", n)
	}
	computed, err := crypto.ComputeJoinRequestMIC(key, rawPayload[:19])
	if err != nil {
		return ErrMICComputeFailed.New(nil)
	}
	for i := 0; i < 4; i++ {
		if computed[i] != rawPayload[19+i] {
			return ErrMICMismatch.New(nil)
		}
	}
	return nil
}

// HandleJoin is called by the network server to join a device
func (js *JoinServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (resp *ttnpb.JoinResponse, err error) {
	supported := false
	for _, v := range supportedMACVersions {
		supported = v == req.SelectedMacVersion
		if supported {
			break
		}
	}
	if !supported {
		return nil, ErrUnsupportedLoRaWANMACVersion.New(errors.Attributes{
			"version": req.SelectedMacVersion,
		})
	}

	if req.EndDeviceIdentifiers.DevAddr == nil {
		return nil, ErrMissingDevAddr.New(nil)
	}
	devAddr := *req.EndDeviceIdentifiers.DevAddr

	rawPayload := req.GetRawPayload()
	if req.Payload.Payload == nil {
		if rawPayload == nil {
			return nil, ErrMissingPayload.New(nil)
		}
		if err := req.Payload.UnmarshalLoRaWAN(rawPayload); err != nil {
			return nil, ErrUnmarshalFailed.NewWithCause(nil, err)
		}
	}

	msg := req.GetPayload()
	if msg.GetMajor() != ttnpb.Major_LORAWAN_R1 {
		return nil, ErrUnsupportedLoRaWANMajorVersion.New(errors.Attributes{
			"major": msg.GetMajor(),
		})
	}
	if msg.GetMType() != ttnpb.MType_JOIN_REQUEST {
		return nil, ErrWrongPayloadType.New(errors.Attributes{
			"type": req.Payload.MType,
		})
	}

	pld := msg.GetJoinRequestPayload()
	if pld == nil {
		return nil, ErrMissingJoinRequest.New(nil)
	}

	if pld.DevEUI.IsZero() {
		return nil, ErrMissingDevEUI.New(nil)
	}
	if pld.JoinEUI.IsZero() {
		return nil, ErrMissingJoinEUI.New(nil)
	}

	if rawPayload == nil {
		rawPayload, err = req.Payload.MarshalLoRaWAN()
		if err != nil {
			panic(errors.NewWithCause("Failed to marshal join request payload", err))
		}
	}

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(js.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI:  &pld.DevEUI,
		JoinEUI: &pld.JoinEUI,
	})
	if err != nil {
		return nil, err
	}

	if rpcmetadata.FromIncomingContext(ctx).NetAddress != dev.NetworkServerAddress {
		return nil, ErrAddressMismatch.New(errors.Attributes{
			"component": "network server",
		})
	}

	if dev.LoRaWANVersion != ttnpb.MAC_V1_0 {
		match := false
		for _, px := range js.euiPrefixes {
			if px.Matches(pld.JoinEUI) {
				match = true
			}
		}
		if !match {
			// TODO determine the cluster containing the device
			// https://github.com/TheThingsIndustries/ttn/issues/244
			return nil, ErrForwardJoinRequest.NewWithCause(nil, deviceregistry.ErrDeviceNotFound.New(errors.Attributes{
				"id": &ttnpb.EndDeviceIdentifiers{JoinEUI: &pld.JoinEUI}}))
		}
	}

	// Registered version is lower than selected.
	if dev.LoRaWANVersion.Compare(req.SelectedMacVersion) == -1 {
		return nil, ErrMACVersionMismatch.New(errors.Attributes{
			"registered": dev.LoRaWANVersion,
			"selected":   req.SelectedMacVersion,
		})
	}

	ke := dev.GetRootKeys().GetAppKey()
	if ke == nil {
		return nil, ErrCorruptRegistry.NewWithCause(nil, ErrAppKeyEnvelopeNotFound.New(nil))
	}
	if ke.Key == nil || ke.Key.IsZero() {
		return nil, ErrCorruptRegistry.NewWithCause(nil, ErrAppKeyNotFound.New(nil))
	}
	appKey := *ke.Key

	var b []byte
	if req.CFList == nil {
		b = make([]byte, 0, 17)
	} else {
		b = make([]byte, 0, 33)
	}

	b, err = (&ttnpb.MHDR{
		MType: ttnpb.MType_JOIN_ACCEPT,
		Major: msg.GetMajor(),
	}).AppendLoRaWAN(b)
	if err != nil {
		panic(errors.NewWithCause("Failed to encode join accept MHDR", err))
	}

	var jn types.JoinNonce
	bb := make([]byte, 4)
	binary.LittleEndian.PutUint32(bb, dev.NextJoinNonce)
	copy(jn[:], bb)

	b, err = (&ttnpb.JoinAcceptPayload{
		NetID:      req.NetID,
		JoinNonce:  jn,
		CFList:     req.CFList,
		DevAddr:    devAddr,
		DLSettings: req.DownlinkSettings,
		RxDelay:    req.RxDelay,
	}).AppendLoRaWAN(b)
	if err != nil {
		panic(errors.NewWithCause("Failed to encode join accept MAC payload", err))
	}

	dn := binary.LittleEndian.Uint16(pld.DevNonce[:])
	if !dev.GetDisableJoinNonceCheck() {
		switch req.SelectedMacVersion {
		case ttnpb.MAC_V1_1:
			if uint32(dn) < dev.NextDevNonce {
				return nil, ErrDevNonceTooSmall.New(nil)
			}
			if dev.NextDevNonce == math.MaxUint32 {
				return nil, ErrDevNonceTooHigh.New(nil)
			}
			dev.NextDevNonce = uint32(dn + 1)
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			for _, used := range dev.UsedDevNonces {
				if dn == uint16(used) {
					return nil, ErrDevNonceReused.New(nil)
				}
			}
		default:
			panic("This statement is unreachable. Fix version check.")
		}
	}

	switch req.SelectedMacVersion {
	case ttnpb.MAC_V1_1:
		ke := dev.GetRootKeys().GetNwkKey()
		if ke == nil {
			return nil, ErrCorruptRegistry.NewWithCause(nil, ErrNwkKeyEnvelopeNotFound.New(nil))
		}
		if ke.Key == nil || ke.Key.IsZero() {
			return nil, ErrCorruptRegistry.NewWithCause(nil, ErrNwkKeyNotFound.New(nil))
		}
		nwkKey := *ke.Key

		if err := checkMIC(nwkKey, rawPayload); err != nil {
			return nil, ErrMICCheckFailed.NewWithCause(nil, err)
		}

		mic, err := crypto.ComputeJoinAcceptMIC(crypto.DeriveJSIntKey(nwkKey, pld.DevEUI), 0xff, pld.JoinEUI, pld.DevNonce, b)
		if err != nil {
			return nil, ErrComputeJoinAcceptMIC.NewWithCause(nil, err)
		}

		enc, err := crypto.EncryptJoinAccept(nwkKey, append(b[1:], mic[:]...))
		if err != nil {
			return nil, ErrEncryptPayloadFailed.NewWithCause(nil, err)
		}
		resp = &ttnpb.JoinResponse{
			RawPayload: append(b[:1], enc...),
			SessionKeys: ttnpb.SessionKeys{
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveFNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
					KEKLabel: "",
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveSNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
					KEKLabel: "",
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveNwkSEncKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
					KEKLabel: "",
				},
				// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
				AppSKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveAppSKey(appKey, jn, pld.JoinEUI, pld.DevNonce)),
					KEKLabel: "",
				},
			},
			Lifetime: nil,
		}
	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		if err := checkMIC(appKey, rawPayload); err != nil {
			return nil, ErrMICCheckFailed.NewWithCause(nil, err)
		}

		mic, err := crypto.ComputeLegacyJoinAcceptMIC(appKey, b)
		if err != nil {
			return nil, ErrComputeJoinAcceptMIC.NewWithCause(nil, err)
		}

		enc, err := crypto.EncryptJoinAccept(appKey, append(b[1:], mic[:]...))
		if err != nil {
			return nil, ErrEncryptPayloadFailed.NewWithCause(nil, err)
		}
		resp = &ttnpb.JoinResponse{
			RawPayload: append(b[:1], enc...),
			SessionKeys: ttnpb.SessionKeys{
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveLegacyNwkSKey(appKey, jn, req.NetID, pld.DevNonce)),
					KEKLabel: "",
				},
				AppSKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveLegacyAppSKey(appKey, jn, req.NetID, pld.DevNonce)),
					KEKLabel: "",
				},
			},
			Lifetime: nil,
		}
	default:
		panic("This statement is unreachable. Fix version check.")
	}

	dev.UsedDevNonces = append(dev.UsedDevNonces, uint32(dn))
	dev.NextJoinNonce++
	dev.EndDevice.Session = &ttnpb.Session{
		StartedAt:   time.Now().UTC(),
		DevAddr:     &devAddr,
		SessionKeys: resp.SessionKeys,
	}
	if err := dev.Update(); err != nil {
		js.Component.Logger().WithField("device", dev).WithError(err).Error("Failed to update device")
	}
	return resp, nil
}

func (js *JoinServer) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	if req.DevEUI.IsZero() {
		return nil, ErrMissingDevEUI.New(nil)
	}
	if req.SessionKeyID == "" {
		return nil, ErrMissingSessionKeyID.New(nil)
	}

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(js.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI: &req.DevEUI,
	})
	if err != nil {
		return nil, err
	}

	if rpcmetadata.FromIncomingContext(ctx).NetAddress != dev.ApplicationServerAddress {
		return nil, ErrAddressMismatch.New(errors.Attributes{
			"component": "application server",
		})
	}

	s := dev.GetSession()
	if s == nil {
		return nil, ErrNoSession.New(nil)
	}
	if s.GetSessionKeyID() != req.GetSessionKeyID() {
		s = dev.GetSessionFallback()
		if s == nil || s.GetSessionKeyID() != req.GetSessionKeyID() {
			return nil, ErrSessionKeyIDMismatch.New(nil)
		}
	}

	appSKey := s.GetAppSKey()
	if appSKey == nil {
		return nil, ErrAppSKeyEnvelopeNotFound.New(nil)
	}
	// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
	return &ttnpb.AppSKeyResponse{
		AppSKey: *appSKey,
	}, nil
}

func (js *JoinServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if req.DevEUI.IsZero() {
		return nil, ErrMissingDevEUI.New(nil)
	}
	if req.SessionKeyID == "" {
		return nil, ErrMissingSessionKeyID.New(nil)
	}

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(js.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI: &req.DevEUI,
	})
	if err != nil {
		return nil, err
	}

	if rpcmetadata.FromIncomingContext(ctx).NetAddress != dev.NetworkServerAddress {
		return nil, ErrAddressMismatch.New(errors.Attributes{
			"component": "network server",
		})
	}

	s := dev.GetSession()
	if s == nil {
		return nil, ErrNoSession.New(nil)
	}
	if s.GetSessionKeyID() != req.GetSessionKeyID() {
		s = dev.GetSessionFallback()
		if s == nil || s.GetSessionKeyID() != req.GetSessionKeyID() {
			return nil, ErrSessionKeyIDMismatch.New(nil)
		}
	}

	nwkSEncKey := s.GetNwkSEncKey()
	if nwkSEncKey == nil {
		return nil, ErrNwkSEncKeyEnvelopeNotFound.New(nil)
	}
	fNwkSIntKey := s.GetFNwkSIntKey()
	if fNwkSIntKey == nil {
		return nil, ErrFNwkSIntKeyEnvelopeNotFound.New(nil)
	}
	sNwkSIntKey := s.GetSNwkSIntKey()
	if sNwkSIntKey == nil {
		return nil, ErrSNwkSIntKeyEnvelopeNotFound.New(nil)
	}
	// TODO: Encrypt key with AS KEK https://github.com/TheThingsIndustries/ttn/issues/271
	return &ttnpb.NwkSKeysResponse{
		NwkSEncKey:  *nwkSEncKey,
		FNwkSIntKey: *fNwkSIntKey,
		SNwkSIntKey: *sNwkSIntKey,
	}, nil
}

// RegisterServices registers services provided by js at s.
func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsJsServer(s, js)
	ttnpb.RegisterDeviceRegistryServer(s, js)
}

// RegisterHandlers registers gRPC handlers.
func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterDeviceRegistryHandler(js.Context(), s, conn)
}
