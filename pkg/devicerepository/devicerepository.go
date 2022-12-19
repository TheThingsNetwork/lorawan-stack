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

// Package devicerepository integrates with the Device Repository.
package devicerepository

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// DeviceRepository implements the Device Repository component.
//
// The Device Repository component exposes the DeviceRepository service.
type DeviceRepository struct {
	ttnpb.UnimplementedDeviceRepositoryServer

	*component.Component
	ctx context.Context

	config *Config

	store store.Store
}

// New returns a new *DeviceRepository.
func New(c *component.Component, conf *Config) (*DeviceRepository, error) {
	dr := &DeviceRepository{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "devicerepository"),
		config:    conf,

		store: conf.Store.Store,
	}

	c.RegisterGRPC(dr)

	c.GRPC.RegisterUnaryHook(
		"/ttn.lorawan.v3.DeviceRepository",
		rpclog.NamespaceHook,
		rpclog.UnaryNamespaceHook("devicerepository"),
	)
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.DeviceRepository", cluster.HookName, c.ClusterAuthUnaryHook())
	return dr, nil
}

// Context returns the context of the Device Repository.
func (dr *DeviceRepository) Context() context.Context {
	return dr.ctx
}

// Roles returns the roles that the Device Repository fulfills.
func (*DeviceRepository) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_DEVICE_REPOSITORY}
}

// RegisterServices registers services provided by dr at s.
func (dr *DeviceRepository) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterDeviceRepositoryServer(s, dr)
}

// RegisterHandlers registers gRPC handlers.
func (dr *DeviceRepository) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterDeviceRepositoryHandler(dr.Context(), s, conn) //nolint:errcheck
}
