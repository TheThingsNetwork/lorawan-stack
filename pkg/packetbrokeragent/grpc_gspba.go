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
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type gsPbaServer struct {
	upstreamCh chan *ttnpb.GatewayUplinkMessage
}

var errForwarderDisabled = errors.DefineFailedPrecondition("forwarder_disabled", "Forwarder is disabled")

// PublishUplink is called by the Gateway Server when an uplink message arrives and needs to get forwarded to Packet Broker.
func (s *gsPbaServer) PublishUplink(ctx context.Context, up *ttnpb.GatewayUplinkMessage) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	if s.upstreamCh == nil {
		return nil, errForwarderDisabled.New()
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		up.CorrelationIDs,
		fmt.Sprintf("pba:uplink:%s", events.NewCorrelationID()),
	)...)
	up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.upstreamCh <- up:
		return ttnpb.Empty, nil
	}
}
