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

// Package ns abstracts the V3 Network Server to the upstream.Handler interface.
package ns

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Handler is the upstream handler.
type Handler struct {
	ctx             context.Context
	name            string
	c               component.Component
	devAddrPrefixes []types.DevAddrPrefix
}

var errNotFound = errors.DefineNotFound("network_server_not_found", "network server not found for ids `ids`")

// NewHandler returns a new upstream handler.
func NewHandler(ctx context.Context, name string, c component.Component, devAddrPrefixes []types.DevAddrPrefix) *Handler {
	return &Handler{
		ctx:             ctx,
		name:            name,
		c:               c,
		devAddrPrefixes: devAddrPrefixes,
	}
}

// GetName implements upstream.Handler.
func (h *Handler) GetName() string {
	return h.name
}

// GetDevAddrPrefixes implements upstream.Handler.
func (h *Handler) GetDevAddrPrefixes() []types.DevAddrPrefix {
	return h.devAddrPrefixes
}

// Setup implements upstream.Handler.
func (h *Handler) Setup() error {
	// Not necessary for this upstream.
	return nil
}

// ConnectGateway implements upstream.Handler.
func (h *Handler) ConnectGateway(ctx context.Context, gtwConn *io.Connection) error {
	// Not necessary for this upstream.
	return nil
}

// DisconnectGateway implements upstream.Handler
func (h *Handler) DisconnectGateway(ctx context.Context, gatewayUID string) error {
	// Not necessary for this upstream.
	return nil
}

// HandleUp implements upstream.Handler.
func (h *Handler) HandleUp(ctx context.Context, gatewayUID string, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.GatewayUp) error {
	ns := h.c.GetPeerConn(ctx, ttnpb.PeerInfo_NETWORK_SERVER, ids)
	if ns == nil {
		return errNotFound.WithAttributes("ids", ids)
	}
	client := ttnpb.NewGsNsClient(ns.Conn())
	for _, up := range msg.UplinkMessages {
		_, err := client.HandleUplink(ctx, up, h.c.WithClusterAuth())
		if err != nil {
			return err
		}
	}
	return nil
}
