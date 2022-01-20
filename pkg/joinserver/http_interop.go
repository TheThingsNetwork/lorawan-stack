// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type interopHandler interface {
	HandleJoin(context.Context, *ttnpb.JoinRequest, Authorizer) (*ttnpb.JoinResponse, error)
	GetHomeNetwork(context.Context, types.EUI64, types.EUI64, Authorizer) (*EndDeviceHomeNetwork, error)
	GetAppSKey(context.Context, *ttnpb.SessionKeyRequest, Authorizer) (*ttnpb.AppSKeyResponse, error)
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
	dlSettings := &ttnpb.DLSettings{}
	if err := lorawan.UnmarshalDLSettings(in.DLSettings, dlSettings); err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}

	req := &ttnpb.JoinRequest{
		RawPayload:         in.PHYPayload,
		DevAddr:            types.DevAddr(in.DevAddr),
		SelectedMacVersion: ttnpb.MACVersion(in.MACVersion),
		NetId:              types.NetID(in.SenderID),
		DownlinkSettings:   dlSettings,
		RxDelay:            in.RxDelay,
		CfList:             cfList,
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

	res, err := srv.JS.HandleJoin(ctx, req, InteropAuthorizer)
	if err != nil {
		switch {
		case errors.Resemble(err, errDecodePayload),
			errors.Resemble(err, errWrongPayloadType),
			errors.Resemble(err, errNoDevEUI),
			errors.Resemble(err, errNoJoinEUI):
			return nil, interop.ErrMalformedMessage.WithCause(err)
		case errors.IsPermissionDenied(err):
			return nil, interop.ErrActivation.WithCause(err)
		case errors.Resemble(err, errMICMismatch):
			return nil, interop.ErrMIC.WithCause(err)
		case errors.IsNotFound(err):
			return nil, interop.ErrUnknownDevEUI.WithCause(err)
		}
		return nil, interop.ErrJoinReq.WithCause(err)
	}

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	ans := &interop.JoinAns{
		JsNsMessageHeader: interop.JsNsMessageHeader{
			MessageHeader: header,
			SenderID:      in.ReceiverID,
			ReceiverID:    in.SenderID,
			ReceiverNSID:  in.SenderNSID,
		},
		PHYPayload: interop.Buffer(res.RawPayload),
		Result: interop.Result{
			ResultCode: interop.ResultSuccess,
		},
		AppSKey:      (*interop.KeyEnvelope)(res.SessionKeys.AppSKey),
		SessionKeyID: interop.Buffer(res.SessionKeys.SessionKeyId),
		Lifetime:     uint32(ttnpb.StdDurationOrZero(res.Lifetime) / time.Second),
	}
	if ttnpb.MACVersion(in.MACVersion).Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
		ans.NwkSKey = (*interop.KeyEnvelope)(res.SessionKeys.FNwkSIntKey)
	} else {
		ans.FNwkSIntKey = (*interop.KeyEnvelope)(res.SessionKeys.FNwkSIntKey)
		ans.SNwkSIntKey = (*interop.KeyEnvelope)(res.SessionKeys.SNwkSIntKey)
		ans.NwkSEncKey = (*interop.KeyEnvelope)(res.SessionKeys.NwkSEncKey)
	}
	return ans, nil
}

func (srv interopServer) HomeNSRequest(ctx context.Context, in *interop.HomeNSReq) (*interop.TTIHomeNSAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	homeNetwork, err := srv.JS.GetHomeNetwork(ctx, types.EUI64(in.ReceiverID), types.EUI64(in.DevEUI), InteropAuthorizer)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, interop.ErrUnknownDevEUI.WithCause(err)
		}
		return nil, err
	}
	if homeNetwork.NetID == nil {
		return nil, interop.ErrActivation.New()
	}

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	ans := &interop.TTIHomeNSAns{
		HomeNSAns: interop.HomeNSAns{
			JsNsMessageHeader: interop.JsNsMessageHeader{
				MessageHeader: header,
				SenderID:      in.ReceiverID,
				ReceiverID:    in.SenderID,
				ReceiverNSID:  in.SenderNSID,
			},
			Result: interop.Result{
				ResultCode: interop.ResultSuccess,
			},
			HNetID: interop.NetID(*homeNetwork.NetID),
		},
		HTenantID:  homeNetwork.TenantID,
		HNSAddress: homeNetwork.NetworkServerAddress,
	}
	if homeNetwork.NSID != nil && in.ProtocolVersion.SupportsNSID() {
		ans.HNSID = (*interop.EUI64)(homeNetwork.NSID)
	}
	return ans, nil
}

func (srv interopServer) AppSKeyRequest(ctx context.Context, in *interop.AppSKeyReq) (*interop.AppSKeyAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	req := &ttnpb.SessionKeyRequest{
		JoinEui:      types.EUI64(in.ReceiverID),
		DevEui:       types.EUI64(in.DevEUI),
		SessionKeyId: in.SessionKeyID,
	}
	if err := req.ValidateFields(
		"join_eui",
		"dev_eui",
		"session_key_id",
	); err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}

	res, err := srv.JS.GetAppSKey(ctx, req, InteropAuthorizer)
	if err != nil {
		switch {
		case errors.IsPermissionDenied(err):
			return nil, interop.ErrActivation.WithCause(err)
		case errors.IsNotFound(err):
			return nil, interop.ErrUnknownDevEUI.WithCause(err)
		}
		return nil, err
	}

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	return &interop.AppSKeyAns{
		JsAsMessageHeader: interop.JsAsMessageHeader{
			MessageHeader: header,
			SenderID:      in.ReceiverID,
			ReceiverID:    in.SenderID,
		},
		Result: interop.Result{
			ResultCode: interop.ResultSuccess,
		},
		DevEUI:       in.DevEUI,
		AppSKey:      interop.KeyEnvelope(*res.AppSKey),
		SessionKeyID: in.SessionKeyID,
	}, nil
}
