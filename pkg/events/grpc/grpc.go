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

// Package grpc contains an implementation of the EventsServer, which is used to
// stream all events published for a set of identifiers.
package grpc

import (
	"context"
	"os"
	"time"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights/rightsutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const workersPerCPU = 2

// NewEventsServer returns a new EventsServer on the given PubSub.
func NewEventsServer(ctx context.Context, pubsub events.PubSub) *EventsServer {
	return &EventsServer{
		ctx:    ctx,
		pubsub: pubsub,
	}
}

// EventsServer streams events from a PubSub over gRPC.
type EventsServer struct {
	ctx    context.Context
	pubsub events.PubSub
}

var errNoIdentifiers = errors.DefineInvalidArgument("no_identifiers", "no identifiers")

// Stream implements the EventsServer interface.
func (srv *EventsServer) Stream(req *ttnpb.StreamEventsRequest, stream ttnpb.Events_StreamServer) error {
	if len(req.Identifiers) == 0 {
		return errNoIdentifiers
	}
	ctx := stream.Context()

	if err := rights.RequireAny(ctx, req.Identifiers...); err != nil {
		return err
	}

	ch := make(events.Channel, 8)
	handler := events.ContextHandler(ctx, ch)
	if err := srv.pubsub.Subscribe(ctx, "", req.Identifiers, handler); err != nil {
		return err
	}

	if req.Tail > 0 || req.After != nil {
		warning.Add(ctx, "Historical events not implemented")
	}

	if err := stream.SendHeader(metadata.MD{}); err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if err := stream.Send(&ttnpb.Event{
		UniqueID:       events.NewCorrelationID(),
		Name:           "events.stream.start",
		Time:           time.Now().UTC(),
		Identifiers:    req.Identifiers,
		Origin:         hostname,
		CorrelationIDs: events.CorrelationIDsFromContext(ctx),
	}); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-ch:
			isVisible, err := rightsutil.EventIsVisible(ctx, evt)
			if err != nil {
				if err := rights.RequireAny(ctx, req.Identifiers...); err != nil {
					return err
				}
				log.FromContext(ctx).WithError(err).Warn("Failed to check event visibility")
				continue
			}
			if !isVisible {
				continue
			}
			proto, err := events.Proto(evt)
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to convert event to proto")
				continue
			}
			if err := stream.Send(proto); err != nil {
				return err
			}
		}
	}
}

// Roles implements rpcserver.Registerer.
func (srv *EventsServer) Roles() []ttnpb.ClusterRole {
	return nil
}

// RegisterServices implements rpcserver.Registerer.
func (srv *EventsServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEventsServer(s, srv)
}

// RegisterHandlers implements rpcserver.Registerer.
func (srv *EventsServer) RegisterHandlers(s *grpc_runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEventsHandler(srv.ctx, s, conn)
}
