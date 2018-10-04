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

// Package grpc contains an implementation of the EventsServer, which is used to
// stream all events published for a set of identifiers.
package grpc

import (
	"context"
	"runtime"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

const workersPerCPU = 2

// NewEventsServer returns a new EventsServer on the given PubSub.
func NewEventsServer(ctx context.Context, pubsub events.PubSub) *EventsServer {
	srv := &EventsServer{
		ctx:    ctx,
		events: make(events.Channel, 256),
		filter: events.NewIdentifierFilter(),
	}

	hander := events.ContextHandler(ctx, srv.events)
	pubsub.Subscribe("**", hander)
	go func() {
		<-ctx.Done()
		pubsub.Unsubscribe("**", hander)
		close(srv.events)
	}()

	for i := 0; i < runtime.NumCPU()*workersPerCPU; i++ {
		go func() {
			for evt := range srv.events {
				proto, err := events.Proto(evt)
				if err != nil {
					return
				}
				srv.filter.Notify(marshaledEvent{
					Event: evt,
					proto: proto,
				})
			}
		}()
	}

	return srv
}

type marshaledEvent struct {
	events.Event
	proto *ttnpb.Event
}

// EventsServer streams events from a PubSub over gRPC.
type EventsServer struct {
	ctx    context.Context
	events events.Channel
	filter events.IdentifierFilter
}

// Stream implements the EventsServer interface.
func (srv *EventsServer) Stream(ids *ttnpb.CombinedIdentifiers, stream ttnpb.Events_StreamServer) (err error) {
	ctx := stream.Context()

	for _, applicationIDs := range ids.ApplicationIDs {
		if err = rights.RequireApplication(ctx, *applicationIDs, ttnpb.RIGHT_APPLICATION_ALL); err != nil {
			return err
		}
	}
	for _, clientIDs := range ids.ClientIDs {
		if err = rights.RequireClient(ctx, *clientIDs, ttnpb.RIGHT_CLIENT_ALL); err != nil {
			return err
		}
	}
	for _, deviceIDs := range ids.DeviceIDs {
		if err = rights.RequireApplication(ctx, deviceIDs.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_ALL); err != nil {
			return err
		}
	}
	for _, gatewayIDs := range ids.GatewayIDs {
		if err = rights.RequireGateway(ctx, *gatewayIDs, ttnpb.RIGHT_GATEWAY_ALL); err != nil {
			return err
		}
	}
	for _, organizationIDs := range ids.OrganizationIDs {
		if err = rights.RequireOrganization(ctx, *organizationIDs, ttnpb.RIGHT_ORGANIZATION_ALL); err != nil {
			return err
		}
	}
	for _, userIDs := range ids.UserIDs {
		if err = rights.RequireUser(ctx, *userIDs, ttnpb.RIGHT_USER_ALL); err != nil {
			return err
		}
	}

	ch := make(events.Channel, 8)
	handler := events.ContextHandler(ctx, ch)
	srv.filter.Subscribe(ctx, ids, handler)
	defer srv.filter.Unsubscribe(ctx, ids, handler)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-ch:
			marshaled := evt.(marshaledEvent)
			if err := stream.Send(marshaled.proto); err != nil {
				return err
			}
		}
	}
}

// Roles implements rpcserver.Registerer.
func (srv *EventsServer) Roles() []ttnpb.PeerInfo_Role {
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
