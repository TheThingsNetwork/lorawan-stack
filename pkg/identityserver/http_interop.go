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

package identityserver

import (
	"context"
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type interopServer struct {
	*IdentityServer
	interop.Authorizer
}

func (srv *interopServer) hNSID(ctx context.Context, dev *ttnpb.EndDevice) string {
	hNSID := dev.NetworkServerAddress
	if tid := srv.configFromContext(ctx).Network.TenantID; tid != "" {
		hNSID = fmt.Sprintf("%s@%s", tid, hNSID)
	}
	return hNSID
}

func (srv *interopServer) HomeNSRequest(ctx context.Context, in *interop.HomeNSReq) (*interop.HomeNSAns, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "identityserver/interop")
	if err := srv.RequireAuthorized(ctx); err != nil {
		return nil, err
	}

	ids := &ttnpb.EndDeviceIdentifiers{
		JoinEui: (*types.EUI64)(&in.ReceiverID),
		DevEui:  (*types.EUI64)(&in.DevEUI),
	}

	var dev *ttnpb.EndDevice
	err := srv.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = store.GetEndDeviceStore(db).GetEndDevice(ctx, ids, &pbtypes.FieldMask{
			Paths: []string{"network_server_address"},
		})
		return err
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, interop.ErrUnknownDevEUI.WithCause(err)
		}
		return nil, err
	}

	homeNetID := srv.configFromContext(ctx).Network.NetID
	hNSID := srv.hNSID(ctx, dev)

	header, err := in.AnswerHeader()
	if err != nil {
		return nil, interop.ErrMalformedMessage.WithCause(err)
	}
	ans := &interop.HomeNSAns{
		JsNsMessageHeader: interop.JsNsMessageHeader{
			MessageHeader: header,
			SenderID:      in.ReceiverID,
			ReceiverID:    in.SenderID,
			ReceiverNSID:  in.SenderNSID,
		},
		Result: interop.Result{
			ResultCode: interop.ResultSuccess,
		},
		HNetID: interop.NetID(homeNetID),
	}
	if in.ProtocolVersion.SupportsNSID() {
		ans.HNSID = &hNSID
	}
	return ans, nil
}
