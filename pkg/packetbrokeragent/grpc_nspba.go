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

package packetbrokeragent

import (
	"context"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type nsPbaServer struct {
	ttnpb.UnimplementedNsPbaServer

	contextDecoupler contextDecoupler
	downstreamCh     chan *downlinkMessage
	frequencyPlans   GetFrequencyPlansStore
}

// PublishDownlink is called by the Network Server when a downlink message needs to get scheduled via Packet Broker.
func (s *nsPbaServer) PublishDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*emptypb.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, down.CorrelationIds...)
	ctx = appendDownlinkCorrelationID(ctx)
	down.CorrelationIds = events.CorrelationIDsFromContext(ctx)

	fps, err := s.frequencyPlans(ctx)
	if err != nil {
		return nil, err
	}

	msg, token, err := toPBDownlink(ctx, down, fps)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to convert outgoing downlink message")
		return nil, err
	}

	ctxMsg := &downlinkMessage{
		Context:          s.contextDecoupler.FromRequestContext(ctx),
		agentUplinkToken: token,
		DownlinkMessage:  msg,
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.downstreamCh <- ctxMsg:
		return ttnpb.Empty, nil
	}
}
