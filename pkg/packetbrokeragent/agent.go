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

package packetbrokeragent

import (
	"context"
	"crypto/tls"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config configures Packet Broker clients.
type Config struct {
	Address     string
	Certificate *tls.Certificate
}

// Agent implements the Packet Broker Agent component, acting as Home Network.
//
// Agent exposes the NsGs interface for scheduling downlink.
type Agent struct {
	*component.Component
	ctx context.Context

	grpc struct {
		nsGs *nsGsServer
	}
}

// New returns a new Packet Broker Agent.
func New(c *component.Component, conf *Config) (*Agent, error) {
	a := &Agent{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "packetbroker/agent"),
	}

	c.RegisterGRPC(a)
	return a, nil
}

// Context returns the context.
func (a *Agent) Context() context.Context {
	return a.ctx
}

// Roles returns the Packet Broker Agent cluster role.
func (a *Agent) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_PACKET_BROKER_AGENT}
}

// RegisterServices registers services provided by a at s.
func (a *Agent) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsGsServer(s, a.grpc.nsGs)
}

// RegisterHandlers registers gRPC handlers.
func (a *Agent) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
}
