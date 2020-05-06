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
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type nsPbaServer struct {
	downstreamCh chan *ttnpb.DownlinkMessage
}

var errHomeNetworkDisabled = errors.DefineFailedPrecondition("home_network_disabled", "Home Network is disabled")

// PublishDownlink is called by the Network Server when a downlink message needs to get scheduled via Packet Broker.
func (s *nsPbaServer) PublishDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	if s.downstreamCh == nil {
		return nil, errHomeNetworkDisabled.New()
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		down.CorrelationIDs,
		fmt.Sprintf("pba:downlink:%s", events.NewCorrelationID()),
	)...)
	down.CorrelationIDs = events.CorrelationIDsFromContext(ctx)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.downstreamCh <- down:
		return ttnpb.Empty, nil
	}
}
