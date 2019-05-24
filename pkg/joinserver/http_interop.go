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

func (srv interopServer) JoinRequest(ctx context.Context, req *interop.JoinReq) (*interop.JoinAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	var cfList *ttnpb.CFList
	if len(req.CFList) > 0 {
		cfList = new(ttnpb.CFList)
		if err := lorawan.UnmarshalCFList(req.CFList, cfList); err != nil {
			return nil, interop.ErrMalformedMessage.WithCause(err)
		}
	}
	var dlSettings ttnpb.DLSettings
	if err := lorawan.UnmarshalDLSettings(req.DLSettings, &dlSettings); err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}

	res, err := srv.JS.HandleJoin(ctx, &ttnpb.JoinRequest{
		RawPayload:         req.PHYPayload,
		DevAddr:            types.DevAddr(req.DevAddr),
		SelectedMACVersion: ttnpb.MACVersion(req.MACVersion),
		NetID:              types.NetID(req.SenderID),
		DownlinkSettings:   dlSettings,
		RxDelay:            req.RxDelay,
		CFList:             cfList,
	})
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

	header, err := req.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	ans := &interop.JoinAns{
		JsNsMessageHeader: header,
		PHYPayload:        interop.Buffer(res.RawPayload),
		Result:            interop.ResultSuccess,
		Lifetime:          uint32(res.Lifetime / time.Second),
		AppSKey:           (*interop.KeyEnvelope)(res.AppSKey),
		SessionKeyID:      interop.Buffer(res.SessionKeyID),
	}
	if ttnpb.MACVersion(req.MACVersion).Compare(ttnpb.MAC_V1_1) < 0 {
		ans.NwkSKey = (*interop.KeyEnvelope)(res.FNwkSIntKey)
	} else {
		ans.FNwkSIntKey = (*interop.KeyEnvelope)(res.FNwkSIntKey)
		ans.SNwkSIntKey = (*interop.KeyEnvelope)(res.SNwkSIntKey)
		ans.NwkSEncKey = (*interop.KeyEnvelope)(res.NwkSEncKey)
	}
	return ans, nil
}

func (srv interopServer) HomeNSRequest(ctx context.Context, req *interop.HomeNSReq) (*interop.HomeNSAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	netID, err := srv.JS.GetHomeNetID(ctx, types.EUI64(req.ReceiverID), types.EUI64(req.DevEUI))
	if err != nil {
		return nil, err
	}
	if netID == nil {
		return nil, interop.ErrActivation
	}

	header, err := req.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	return &interop.HomeNSAns{
		JsNsMessageHeader: header,
		HNSID:             interop.NetID(*netID),
		HNetID:            interop.NetID(*netID),
	}, nil
}

func (srv interopServer) AppSKeyRequest(ctx context.Context, req *interop.AppSKeyReq) (*interop.AppSKeyAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "joinserver/interop")

	res, err := srv.JS.GetAppSKey(ctx, &ttnpb.SessionKeyRequest{
		JoinEUI:      types.EUI64(req.ReceiverID),
		DevEUI:       types.EUI64(req.DevEUI),
		SessionKeyID: req.SessionKeyID,
	})
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

	header, err := req.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	return &interop.AppSKeyAns{
		JsAsMessageHeader: header,
		Result:            interop.ResultSuccess,
		DevEUI:            req.DevEUI,
		AppSKey:           interop.KeyEnvelope(res.AppSKey),
		SessionKeyID:      req.SessionKeyID,
	}, nil
}
