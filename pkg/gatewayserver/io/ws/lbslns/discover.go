// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package lbslns

import (
	"context"
	"encoding/json"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/id6"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var errEmptyGatewayEUI = errors.DefineFailedPrecondition("empty_gateway_eui", "empty gateway EUI")

// HandleConnectionInfo implements Formatter.
func (f *lbsLNS) HandleConnectionInfo(
	ctx context.Context,
	raw []byte,
	server io.Server,
	info ws.ServerInfo,
	assertAuth func(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error,
) []byte {
	var req DiscoverQuery

	if err := json.Unmarshal(raw, &req); err != nil {
		return logAndWrapDiscoverError(ctx, err, "Failed to parse discover query message")
	}
	if req.EUI.IsZero() {
		return logAndWrapDiscoverError(ctx, errEmptyGatewayEUI.New(), "Empty router EUI provided")
	}

	ids := &ttnpb.GatewayIdentifiers{
		Eui: req.EUI.EUI64.Bytes(),
	}

	filledCtx, ids, err := server.FillGatewayContext(ctx, ids)
	if err != nil {
		return logAndWrapDiscoverError(ctx, err, fmt.Sprintf("Failed to fetch gateway: %s", err.Error()))
	}
	ctx = filledCtx

	if err := assertAuth(ctx, ids); err != nil {
		return logAndWrapDiscoverError(ctx, err, fmt.Sprintf("Unauthorized"))
	}

	euiWithPrefix := fmt.Sprintf("eui-%s", types.MustEUI64(ids.Eui).OrZero().String())
	res := DiscoverResponse{
		EUI: req.EUI,
		Muxs: id6.EUI{
			Prefix: "muxs",
		},
		URI: fmt.Sprintf("%s://%s%s/%s", info.Scheme, info.Address, trafficEndPointPrefix, euiWithPrefix),
	}

	data, err := json.Marshal(res)
	if err != nil {
		return logAndWrapDiscoverError(ctx, err, "Router not provisioned")
	}
	return data
}

// logAndWrapDiscoverError logs the error text and wraps it into a DiscoverResponse.
func logAndWrapDiscoverError(ctx context.Context, err error, msg string) []byte {
	logger := log.FromContext(ctx)
	logger.WithError(err).Warn(msg)
	errMsg, err := json.Marshal(DiscoverResponse{Error: msg})
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal error message")
		return nil
	}
	return errMsg
}
