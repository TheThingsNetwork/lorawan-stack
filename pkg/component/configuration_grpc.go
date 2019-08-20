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

package component

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// NewConfigurationServer returns a new ConfigurationServer on top of the given component.
func NewConfigurationServer(c *Component) *ConfigurationServer {
	return &ConfigurationServer{component: c}
}

// ConfigurationServer implements the Configuration RPC service.
type ConfigurationServer struct {
	component *Component
}

// Roles implements the rpcserver.Registerer interface. It just returns nil.
func (c *ConfigurationServer) Roles() []ttnpb.ClusterRole { return nil }

// RegisterServices registers the Configuration service.
func (c *ConfigurationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterConfigurationServer(s, c)
}

// RegisterHandlers registers the Configuration service handler.
func (c *ConfigurationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterConfigurationHandler(c.component.Context(), s, conn)
}

// ListFrequencyPlans implements the Configuration service's ListFrequencyPlans RPC.
func (c *ConfigurationServer) ListFrequencyPlans(ctx context.Context, req *ttnpb.ListFrequencyPlansRequest) (*ttnpb.ListFrequencyPlansResponse, error) {
	return frequencyplans.NewRPCServer(c.component.FrequencyPlans).ListFrequencyPlans(ctx, req)
}
