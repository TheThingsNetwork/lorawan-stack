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

package gatewayserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	errNotTxRequest = errors.DefineInvalidArgument("not_tx_request", "downlink message is not a Tx request")
	errSchedulePath = errors.Define("schedule_path", "failed to schedule on path `{gateway_uid}`")
	errSchedule     = errors.DefineAborted("schedule", "failed to schedule")
)

// ScheduleDownlink instructs the Gateway Server to schedule a downlink message request.
// This method returns an error if the downlink path cannot be found, if the requested parameters are invalid for the
// gateway's frequency plan or if there is no transmission window available because of scheduling conflicts or regional
// limitations such as duty-cycle and dwell time.
func (gs *GatewayServer) ScheduleDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*types.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	request := down.GetRequest()
	if request == nil {
		return nil, errNotTxRequest
	}
	paths := request.DownlinkPaths

	logger := log.FromContext(ctx)
	var details []interface{}
	for _, path := range paths {
		uid := unique.ID(ctx, path.GatewayIdentifiers)
		val, ok := gs.connections.Load(uid)
		if !ok {
			details = append(details, errNotConnected.WithAttributes("gateway_uid", uid))
			continue
		}
		conn := val.(*io.Connection)
		// Provide only the connection's downlink path as required by SendDown.
		request.DownlinkPaths = []*ttnpb.TxRequest_DownlinkPath{path}
		if err := conn.SendDown(down); err != nil {
			logger.WithField("gateway_uid", uid).WithError(err).Debug("Failed to schedule on path")
			details = append(details, errSchedulePath.WithCause(err).WithAttributes("gateway_uid", uid))
			continue
		}
		ctx = events.ContextWithCorrelationID(ctx, events.CorrelationIDsFromContext(conn.Context())...)
		registerSendDownlink(ctx, conn.Gateway(), down)
		return &types.Empty{}, nil
	}

	return nil, errSchedule.WithDetails(details...)
}
