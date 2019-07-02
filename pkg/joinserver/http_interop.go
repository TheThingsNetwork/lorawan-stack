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
	"time"

	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type interopHandler interface {
	HandleJoin(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetHomeNetID(context.Context, types.EUI64, types.EUI64) (*types.NetID, error)
	GetAppSKey(context.Context, *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error)
}

type interopServer struct {
	JS interopHandler
}

func (srv interopServer) JoinRequest(ctx context.Context, in *interop.JoinReq) (*interop.JoinAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	var cfList *ttnpb.CFList
	if len(in.CFList) > 0 {
		cfList = new(ttnpb.CFList)
		if err := lorawan.UnmarshalCFList(in.CFList, cfList); err != nil {
			return nil, interop.ErrMalformedMessage.WithCause(err)
		}
	}
	var dlSettings ttnpb.DLSettings
	if err := lorawan.UnmarshalDLSettings(in.DLSettings, &dlSettings); err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}

	req := &ttnpb.JoinRequest{
		RawPayload:         in.PHYPayload,
		DevAddr:            types.DevAddr(in.DevAddr),
		SelectedMACVersion: ttnpb.MACVersion(in.MACVersion),
		NetID:              types.NetID(in.SenderID),
		DownlinkSettings:   dlSettings,
		RxDelay:            in.RxDelay,
		CFList:             cfList,
	}
	if err := req.ValidateFields(
		"raw_payload",
		"dev_addr",
		"selected_mac_version",
		"net_id",
		"downlink_settings",
		"rx_delay",
		"cf_list",
	); err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}

	res, err := srv.JS.HandleJoin(ctx, req)
	if err != nil {
		switch {
		case errors.Resemble(err, errDecodePayload),
			errors.Resemble(err, errWrongPayloadType),
			errors.Resemble(err, errNoDevEUI),
			errors.Resemble(err, errNoJoinEUI):
			return nil, interop.ErrMalformedMessage.WithCause(err)
		case errors.Resemble(err, errAddressNotAuthorized):
			return nil, interop.ErrActivation.WithCause(err)
		case errors.Resemble(err, errMICMismatch):
			return nil, interop.ErrMIC.WithCause(err)
		case errors.Resemble(err, errRegistryOperation):
			if errors.IsNotFound(errors.Cause(err)) {
				return nil, interop.ErrUnknownDevEUI.WithCause(err)
			}
		}
		return nil, interop.ErrJoinReq.WithCause(err)
	}

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	ans := &interop.JoinAns{
		JsNsMessageHeader: header,
		PHYPayload:        interop.Buffer(res.RawPayload),
		Result: interop.Result{
			ResultCode: interop.ResultSuccess,
		},
		Lifetime:     uint32(res.Lifetime / time.Second),
		AppSKey:      (*interop.KeyEnvelope)(res.AppSKey),
		SessionKeyID: interop.Buffer(res.SessionKeyID),
	}
	if ttnpb.MACVersion(in.MACVersion).Compare(ttnpb.MAC_V1_1) < 0 {
		ans.NwkSKey = (*interop.KeyEnvelope)(res.FNwkSIntKey)
	} else {
		ans.FNwkSIntKey = (*interop.KeyEnvelope)(res.FNwkSIntKey)
		ans.SNwkSIntKey = (*interop.KeyEnvelope)(res.SNwkSIntKey)
		ans.NwkSEncKey = (*interop.KeyEnvelope)(res.NwkSEncKey)
	}
	return ans, nil
}

func (srv interopServer) HomeNSRequest(ctx context.Context, in *interop.HomeNSReq) (*interop.HomeNSAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	netID, err := srv.JS.GetHomeNetID(ctx, types.EUI64(in.ReceiverID), types.EUI64(in.DevEUI))
	if err != nil {
		return nil, err
	}
	if netID == nil {
		return nil, interop.ErrActivation
	}

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	return &interop.HomeNSAns{
		JsNsMessageHeader: header,
		Result: interop.Result{
			ResultCode: interop.ResultSuccess,
		},
		HNSID:  interop.NetID(*netID),
		HNetID: interop.NetID(*netID),
	}, nil
}

func (srv interopServer) AppSKeyRequest(ctx context.Context, in *interop.AppSKeyReq) (*interop.AppSKeyAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	req := &ttnpb.SessionKeyRequest{
		JoinEUI:      types.EUI64(in.ReceiverID),
		DevEUI:       types.EUI64(in.DevEUI),
		SessionKeyID: in.SessionKeyID,
	}
	if err := req.ValidateFields(
		"join_eui",
		"dev_eui",
		"session_key_id",
	); err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}

	res, err := srv.JS.GetAppSKey(ctx, req)
	if err != nil {
		switch {
		case errors.Resemble(err, errAddressNotAuthorized):
			return nil, interop.ErrActivation.WithCause(err)
		case errors.Resemble(err, errRegistryOperation):
			if errors.IsNotFound(errors.Cause(err)) {
				return nil, interop.ErrUnknownDevEUI.WithCause(err)
			}
		}
		return nil, err
	}

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	return &interop.AppSKeyAns{
		JsAsMessageHeader: header,
		Result: interop.Result{
			ResultCode: interop.ResultSuccess,
		},
		DevEUI:       in.DevEUI,
		AppSKey:      interop.KeyEnvelope(res.AppSKey),
		SessionKeyID: in.SessionKeyID,
	}, nil
}
