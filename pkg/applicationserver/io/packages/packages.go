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

package packages

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

type server struct {
	ctx context.Context

	io       io.Server
	registry Registry

	handlers map[string]ApplicationPackageHandler
}

// Server is an application packages frontend.
type Server interface {
	rpcserver.Registerer
	NewSubscription() *io.Subscription
}

// New returns an application packages server wrapping the given registries.
func New(ctx context.Context, io io.Server, registry Registry) (Server, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/packages")
	s := &server{
		ctx:      ctx,
		io:       io,
		registry: registry,
		handlers: make(map[string]ApplicationPackageHandler),
	}
	for _, p := range registeredPackages {
		s.handlers[p.Name] = p.new(io, registry)
	}
	return s, nil
}

func (s *server) handleUp(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	switch up := msg.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		association, err := s.registry.Get(ctx, ttnpb.ApplicationPackageAssociationIdentifiers{
			EndDeviceIdentifiers: msg.EndDeviceIdentifiers,
			FPort:                up.UplinkMessage.FPort,
		}, []string{
			"data",
			"ids.end_device_ids",
			"ids.f_port",
			"package_name",
		})
		if errors.IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
		if handler, ok := s.handlers[association.PackageName]; ok {
			return handler.HandleUp(ctx, association, msg)
		}
		return errNotImplemented.WithAttributes("name", association.PackageName)
	}
	return nil
}

// Roles implements the rpcserver.Registerer interface.
func (s *server) Roles() []ttnpb.ClusterRole {
	return nil
}

// RegisterServices registers the services of the registered application packages.
func (s *server) RegisterServices(gs *grpc.Server) {
	ttnpb.RegisterApplicationPackageRegistryServer(gs, s)
	for _, subsystem := range s.handlers {
		subsystem.RegisterServices(gs)
	}
}

// RegisterHandlers registers the handlers of the registered application packages.
func (s *server) RegisterHandlers(rs *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterApplicationPackageRegistryHandler(s.ctx, rs, conn)
	for _, subsystem := range s.handlers {
		subsystem.RegisterHandlers(rs, conn)
	}
}

// NewSubscription creates a new default subscription for upstream application packages traffic.
func (s *server) NewSubscription() *io.Subscription {
	sub := io.NewSubscription(s.ctx, "applicationpackages", nil)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case up := <-sub.Up():
				if err := s.handleUp(up.Context, up.ApplicationUp); err != nil {
					log.FromContext(s.ctx).WithError(err).Warn("Failed to handle message")
				}
			}
		}
	}()
	return sub
}
