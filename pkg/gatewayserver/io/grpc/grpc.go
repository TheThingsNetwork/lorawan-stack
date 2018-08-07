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

package grpc

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

type impl struct {
	server io.Server
}

// New returns a new gRPC frontend.
func New(server io.Server) ttnpb.GtwGsServer {
	return &impl{server}
}

var errConnect = errors.Define("connect", "failed to connect gateway `{gateway_uid}`")

// Link the gateway to the Gateway Server. The authentication information will
// be used to determine the gateway ID. If no authentication information is present,
// this gateway may not be used for downlink.
func (s *impl) Link(link ttnpb.GtwGs_LinkServer) (err error) {
	ctx := log.NewContextWithField(link.Context(), "namespace", "io/grpc")

	id := ttnpb.GatewayIdentifiers{
		GatewayID: rpcmetadata.FromIncomingContext(ctx).ID,
	}
	if err = validate.ID(id.GetGatewayID()); err != nil {
		return
	}
	if err = rights.RequireGateway(ctx, id, ttnpb.RIGHT_GATEWAY_LINK); err != nil {
		return
	}

	uid := unique.ID(ctx, id)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)
	logger := log.FromContext(ctx)
	conn, err := s.server.Connect(ctx, id)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return errConnect.WithCause(err).WithAttributes("gateway_uid", uid)
	}
	logger.Info("Connected")
	if err = s.server.ClaimDownlink(ctx, id); err != nil {
		logger.WithError(err).Warn("Failed to claim downlink")
		return
	}

	go func() {
		for {
			select {
			case <-conn.Context().Done():
				return
			case down := <-conn.Down():
				msg := &ttnpb.GatewayDown{
					DownlinkMessage: down,
				}
				logger.Info("Sending downlink message")
				if err := link.Send(msg); err != nil {
					logger.WithError(err).Warn("Failed to send message")
					continue
				}
			}
		}
	}()

	for {
		msg, err := link.Recv()
		if err != nil {
			if !errors.IsCanceled(err) {
				logger.WithError(err).Warn("Link failed")
			}
			return err
		}

		logger.WithFields(log.Fields(
			"has_status", msg.GatewayStatus != nil,
			"uplink_count", len(msg.UplinkMessages),
		)).Debug("Received message")

		for _, up := range msg.UplinkMessages {
			if err := conn.HandleUp(up); err != nil {
				logger.WithError(err).Warn("Failed to handle uplink message")
			}
		}
		if msg.GatewayStatus != nil {
			if err := conn.HandleStatus(msg.GatewayStatus); err != nil {
				logger.WithError(err).Warn("Failed to handle status message")
			}
		}
	}
}

func (s *impl) GetFrequencyPlan(ctx context.Context, req *ttnpb.GetFrequencyPlanRequest) (*ttnpb.FrequencyPlan, error) {
	return s.server.GetFrequencyPlan(ctx, req.FrequencyPlanID)
}
