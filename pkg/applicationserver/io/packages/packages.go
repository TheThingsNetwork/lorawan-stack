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
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
)

const namespace = "applicationserver/io/packages"

type server struct {
	ttnpb.UnimplementedApplicationPackageRegistryServer

	ctx context.Context

	server   io.Server
	registry Registry

	handlers map[string]ApplicationPackageHandler
	pools    map[string]workerpool.WorkerPool[*associatedApplicationUp]
}

// Server is an application packages frontend.
type Server interface {
	rpcserver.ServiceRegisterer
	web.Registerer
}

func createPackagePoolHandler(
	name string, handler ApplicationPackageHandler, timeout time.Duration,
) workerpool.Handler[*associatedApplicationUp] {
	h := func(ctx context.Context, associatedUp *associatedApplicationUp) {
		pair, up := associatedUp.pair, associatedUp.up

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := handler.HandleUp(ctx, pair.defaultAssociation, pair.association, up); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
			registerMessageFailed(ctx, name, err)
			return
		}
		registerMessageProcessed(ctx, name)
	}
	return h
}

// New returns an application packages server wrapping the given registries and handlers.
func New(ctx context.Context, as io.Server, registry Registry, handlers map[string]ApplicationPackageHandler, workers int, timeout time.Duration) (Server, error) {
	ctx = log.NewContextWithField(ctx, "namespace", namespace)
	s := &server{
		ctx:      ctx,
		server:   as,
		registry: registry,
		handlers: handlers,
		pools:    make(map[string]workerpool.WorkerPool[*associatedApplicationUp]),
	}
	for name, handler := range handlers {
		s.pools[name] = workerpool.NewWorkerPool(workerpool.Config[*associatedApplicationUp]{
			Component:  as,
			Context:    ctx,
			Name:       fmt.Sprintf("application_packages_%v", name),
			Handler:    createPackagePoolHandler(name, handler, timeout),
			MaxWorkers: workers,
		})
	}
	sub, err := as.Subscribe(ctx, "applicationpackages", nil, false)
	if err != nil {
		return nil, err
	}
	wp := workerpool.NewWorkerPool(workerpool.Config[*ttnpb.ApplicationUp]{
		Component: as,
		Context:   ctx,
		Name:      "application_packages_fanout",
		Handler:   workerpool.HandlerFromUplinkHandler(s.handleUp),
	})
	sub.Pipe(ctx, as, "application_packages", wp.Publish)
	return s, nil
}

type associationsPair struct {
	defaultAssociation *ttnpb.ApplicationPackageDefaultAssociation
	association        *ttnpb.ApplicationPackageAssociation
}

type associatedApplicationUp struct {
	pair *associationsPair
	up   *ttnpb.ApplicationUp
}

type associationsMap map[string]*associationsPair

func (s *server) findAssociations(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (associationsMap, error) {
	paths := []string{
		"data",
		"ids",
		"package_name",
	}
	associations, err := s.registry.ListAssociations(ctx, ids, paths)
	if err != nil {
		return nil, err
	}
	defaults, err := s.registry.ListDefaultAssociations(ctx, ids.ApplicationIds, paths)
	if err != nil {
		return nil, err
	}
	m := make(associationsMap)
	for _, association := range associations {
		m[association.PackageName] = &associationsPair{
			association: association,
		}
	}
	for _, defaultAssociation := range defaults {
		if pair, ok := m[defaultAssociation.PackageName]; ok {
			pair.defaultAssociation = defaultAssociation
		} else {
			m[defaultAssociation.PackageName] = &associationsPair{
				defaultAssociation: defaultAssociation,
			}
		}
	}
	return m, nil
}

func (s *server) handleUp(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	ctx = log.NewContextWithField(ctx, "namespace", namespace)
	associations, err := s.findAssociations(ctx, msg.EndDeviceIds)
	if err != nil {
		return err
	}
	for name, pair := range associations {
		pool, ok := s.pools[name]
		if !ok {
			continue
		}
		ctx := log.NewContextWithField(ctx, "package", name)
		if err := pool.Publish(ctx, &associatedApplicationUp{
			pair: pair,
			up:   msg,
		}); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
			registerMessageFailed(ctx, name, err)
		}
	}
	return nil
}

// RegisterServices implements the rpcserver.ServiceRegisterer interface.
func (s *server) RegisterServices(gs *grpc.Server) {
	ttnpb.RegisterApplicationPackageRegistryServer(gs, s)
	for _, subsystem := range s.handlers {
		if subsystem, ok := subsystem.(rpcserver.ServiceRegisterer); ok {
			subsystem.RegisterServices(gs)
		}
	}
}

// RegisterHandlers implements the rpcserver.ServiceRegisterer interface.
func (s *server) RegisterHandlers(rs *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterApplicationPackageRegistryHandler(s.ctx, rs, conn)
	for _, subsystem := range s.handlers {
		if subsystem, ok := subsystem.(rpcserver.ServiceRegisterer); ok {
			subsystem.RegisterHandlers(rs, conn)
		}
	}
}

// RegisterRoutes implements the web.Registerer interface.
func (s *server) RegisterRoutes(ws *web.Server) {
	for _, subsystem := range s.handlers {
		if subsystem, ok := subsystem.(web.Registerer); ok {
			subsystem.RegisterRoutes(ws)
		}
	}
}
