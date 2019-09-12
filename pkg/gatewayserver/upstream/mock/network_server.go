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

// Package mock provides mock implementation of necessary NS interfaces for testing.
package mock

import (
	"context"
	"net"

	types "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// NetworkServer is a mock NS for GS tests.
type NetworkServer struct {
	ttnpb.GsNsServer
	uplinkMessages []*ttnpb.UplinkMessage
}

// NewNS creates and starts an instance of the mock NS.
func NewNS(ctx context.Context) (*NetworkServer, string) {
	ns := &NetworkServer{
		uplinkMessages: make([]*ttnpb.UplinkMessage, 0),
	}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterGsNsServer(srv.Server, ns)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return ns, lis.Addr().String()
}

// HandleUplink implements ttnpb.GsNsServer
func (ns *NetworkServer) HandleUplink(ctx context.Context, up *ttnpb.UplinkMessage) (*types.Empty, error) {
	ns.uplinkMessages = append(ns.uplinkMessages, up)
	return &types.Empty{}, nil
}

// GetUpMessages returns the stored messages.
func (ns *NetworkServer) GetUpMessages() []*ttnpb.UplinkMessage {
	return ns.uplinkMessages
}
