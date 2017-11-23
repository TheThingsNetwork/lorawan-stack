// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver

import (
	"encoding/binary"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// JoinServer implements the join server component.
//
// The join server exposes the NsJs and DeviceRegistry services.
type JoinServer struct {
	*component.Component
	*deviceregistry.RegistryRPC

	registry deviceregistry.Interface
	joinEUI  types.EUI64
}

type Config struct {
	Component *component.Component
	Registry  deviceregistry.Interface
	JoinEUI   types.EUI64
}

func New(conf *Config) *JoinServer {
	return &JoinServer{
		Component:   conf.Component,
		RegistryRPC: deviceregistry.NewRPC(conf.Component, conf.Registry),
		registry:    conf.Registry,
		joinEUI:     conf.JoinEUI,
	}
}

func checkMIC(key types.AES128Key, rawPayload []byte) error {
	if n := len(rawPayload); n != 23 {
		return errors.Errorf("Expected length of raw payload to be equal to 23, got %d", n)
	}
	computed, err := crypto.ComputeJoinRequestMIC(key, rawPayload[:19])
	if err != nil {
		return errors.NewWithCause("failed to compute MIC from payload", err)
	}
	for i := 0; i < 4; i++ {
		if computed[i] != rawPayload[19+i] {
			return ErrInvalidMIC.New(nil)
		}
	}
	return nil
}

func keyPointer(key types.AES128Key) *types.AES128Key {
	return &key
}

// HandleJoin is called by the network server to join a device
func (js *JoinServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if req == nil {
		return nil, ErrMissingJoinRequest.New(nil)
	}

	if req.EndDeviceIdentifiers.DevAddr == nil {
		return nil, ErrMissingDevAddr.New(nil)
	}
	devAddr := *req.EndDeviceIdentifiers.DevAddr

	rawPayload := req.GetRawPayload()
	if rawPayload != nil {
		if err := req.Payload.UnmarshalLoRaWAN(rawPayload); err != nil {
			return nil, ErrUnmarshalFailed.NewWithCause(nil, err)
		}
	}

	msg := req.GetPayload()
	if msg.GetMType() != ttnpb.MType_JOIN_REQUEST {
		return nil, ErrWrongPayloadType.New(errors.Attributes{
			"expected_value": ttnpb.MType_JOIN_REQUEST,
			"got_value":      req.Payload.MType,
		})
	}

	pld := msg.GetJoinRequestPayload()
	if pld == nil {
		return nil, ErrMissingPayload.New(nil)
	}

	if pld.DevEUI.IsZero() {
		return nil, ErrMissingDevEUI.New(nil)
	}
	if pld.JoinEUI.IsZero() {
		return nil, ErrMissingJoinEUI.New(nil)
	}
	if pld.JoinEUI != js.joinEUI {
		// TODO determine the cluster containing the device
		return nil, ErrForwardJoinRequest.New(nil)
	}

	devs, err := js.registry.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI: &pld.DevEUI,
	})
	if err != nil {
		return nil, err
	}
	var dev *deviceregistry.Device
	switch len(devs) {
	case 0:
		return nil, deviceregistry.ErrDeviceNotFound.New(errors.Attributes{
			"identifiers": pld.DevEUI,
		})
	case 1:
		dev = devs[0]
	default:
		return nil, deviceregistry.ErrTooManyDevices.New(errors.Attributes{
			"identifiers": pld.DevEUI,
		})
	}

	ke := dev.GetRootKeys().GetAppKey()
	if ke == nil {
		return nil, ErrAppKeyEnvelopeNotFound.New(nil)
	}
	if ke.Key == nil || ke.Key.IsZero() {
		return nil, ErrAppKeyNotFound.New(nil)
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
		return nil, ErrEncodeMHDRFailed.NewWithCause(nil, err)
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
		return nil, ErrEncodePayloadFailed.NewWithCause(nil, err)
	}

	dn := binary.LittleEndian.Uint16(pld.DevNonce[:])
	var resp *ttnpb.JoinResponse

	switch req.SelectedMacVersion {
	case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
		for _, used := range dev.UsedDevNonces {
			if dn == uint16(used) {
				return nil, ErrDevNonceReused.New(nil)
			}
		}

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
					KekLabel: "",
				},
				AppSKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveLegacyAppSKey(appKey, jn, req.NetID, pld.DevNonce)),
					KekLabel: "",
				},
			},
			Lifetime: nil,
		}
	case ttnpb.MAC_V1_1:
		if uint32(dn) < dev.NextDevNonce {
			return nil, ErrDevNonceTooSmall.New(nil)
		}
		dev.NextDevNonce = uint32(dn + 1)

		ke := dev.GetRootKeys().GetNwkKey()
		if ke == nil {
			return nil, ErrNwkKeyEnvelopeNotFound.New(nil)
		}
		if ke.Key == nil || ke.Key.IsZero() {
			return nil, ErrNwkKeyNotFound.New(nil)
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
					KekLabel: "",
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveSNwkSIntKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
					KekLabel: "",
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveNwkSEncKey(nwkKey, jn, pld.JoinEUI, pld.DevNonce)),
					KekLabel: "",
				},
				AppSKey: &ttnpb.KeyEnvelope{
					Key:      keyPointer(crypto.DeriveAppSKey(appKey, jn, pld.JoinEUI, pld.DevNonce)),
					KekLabel: "",
				},
			},
			Lifetime: nil,
		}
	default:
		return nil, ErrUnsupportedLoRaWANVersion.New(errors.Attributes{
			"lorawan_version": req.SelectedMacVersion,
		})
	}

	dev.UsedDevNonces = append(dev.UsedDevNonces, uint32(dn))
	dev.NextJoinNonce++
	go func() {
		if err := dev.Update(); err != nil {
			js.Component.Logger().WithField("device", dev).Error("Failed to update device")
		}
	}()
	return resp, nil
}

func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsJsServer(s, js)
	ttnpb.RegisterDeviceRegistryServer(s, js)
}

func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
}
