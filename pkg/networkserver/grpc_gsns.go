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

// Package networkserver provides a LoRaWAN 1.1-compliant Network Server implementation.
package networkserver

import (
	"context"
	"fmt"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// HandleUplink is called by the Gateway Server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, up *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		up.CorrelationIDs,
		fmt.Sprintf("ns:uplink:%s", events.NewCorrelationID()),
	)...)
	up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)

	up.ReceivedAt = time.Now()

	logger := log.FromContext(ctx)

	if up.Payload.Payload == nil {
		if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
			return nil, errDecodePayload.WithCause(err)
		}
	}

	if up.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"major", up.Payload.Major,
		)
	}

	acc, stopDedup, ok := ns.deduplicateUplink(ctx, up)
	if ok {
		registerReceiveUplinkDuplicate(ctx, up)
		return ttnpb.Empty, nil
	}
	registerReceiveUplink(ctx, up)

	defer func(up *ttnpb.UplinkMessage) {
		<-ns.collectionDone(ctx, up)
		stopDedup()
	}(up)

	up = deepcopy.Copy(up).(*ttnpb.UplinkMessage)
	switch up.Payload.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return ttnpb.Empty, ns.handleUplink(ctx, up, acc)
	case ttnpb.MType_JOIN_REQUEST:
		return ttnpb.Empty, ns.handleJoin(ctx, up, acc)
	case ttnpb.MType_REJOIN_REQUEST:
		return ttnpb.Empty, ns.handleRejoin(ctx, up, acc)
	default:
		logger.Error("Unmatched MType")
		return ttnpb.Empty, nil
	}
}
