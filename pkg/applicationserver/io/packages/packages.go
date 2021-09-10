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
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
)

const namespace = "applicationserver/io/packages"

type server struct {
	ctx context.Context

	server   io.Server
	registry Registry

	handlers map[string]ApplicationPackageHandler
	pools    map[string]workerpool.WorkerPool
}

// Server is an application packages frontend.
type Server interface {
	rpcserver.Registerer
}

func createPackagePoolHandler(name string, handler ApplicationPackageHandler, timeout time.Duration) workerpool.HandlerFactory {
	h := func(ctx context.Context, item interface{}) {
		associatedUp := item.(*associatedApplicationUp)
		pair, up := associatedUp.pair, associatedUp.up

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := handler.HandleUp(ctx, pair.defaultAssociation, pair.association, up); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
			registerMessageFailed(name, err)
			return
		}
		registerMessageProcessed(name)
	}
	return workerpool.StaticHandlerFactory(h)
}

func createFrontendPoolHandler(fanOut func(context.Context, *ttnpb.ApplicationUp) error) workerpool.HandlerFactory {
	h := func(ctx context.Context, item interface{}) {
		up := item.(*ttnpb.ApplicationUp)

		if err := fanOut(ctx, up); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to fanout message")
			registerMessageFailed("frontend-fanout", err)
		}
	}
	return workerpool.StaticHandlerFactory(h)
}

func startFrontendSubscriber(ctx context.Context, as io.Server, sub *io.Subscription, submit func(context.Context, interface{}) error) {
	as.StartTask(&component.TaskConfig{
		Context: ctx,
		ID:      "run_application_packages",
		Func: func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-sub.Context().Done():
					return sub.Context().Err()
				case up := <-sub.Up():
					ctx := log.NewContextWithField(up.Context, "namespace", namespace)
					if err := submit(ctx, up.ApplicationUp); err != nil {
						log.FromContext(ctx).WithError(err).Warn("Failed to submit message")
						registerMessageFailed("frontend-submit", err)
					}
				}
			}
		},
		Restart: component.TaskRestartOnFailure,
		Backoff: component.DefaultTaskBackoffConfig,
	})
}

// New returns an application packages server wrapping the given registries and handlers.
func New(ctx context.Context, as io.Server, registry Registry, handlers map[string]ApplicationPackageHandler, workers int, timeout time.Duration) (Server, error) {
	ctx = log.NewContextWithField(ctx, "namespace", namespace)
	s := &server{
		ctx:      ctx,
		server:   as,
		registry: registry,
		handlers: handlers,
		pools:    make(map[string]workerpool.WorkerPool),
	}

	for name, handler := range handlers {
		wp, err := workerpool.NewWorkerPool(workerpool.Config{
			Component:     as,
			Context:       ctx,
			Name:          fmt.Sprintf("application_packages_%v", name),
			CreateHandler: createPackagePoolHandler(name, handler, timeout),
			MaxWorkers:    workers,
		})
		if err != nil {
			return nil, err
		}
		s.pools[name] = wp
	}

	sub, err := as.Subscribe(ctx, "applicationpackages", nil, false)
	if err != nil {
		return nil, err
	}

	wp, err := workerpool.NewWorkerPool(workerpool.Config{
		Component:     as,
		Context:       ctx,
		Name:          "application_packages_fanout",
		CreateHandler: createFrontendPoolHandler(s.fanOut),
		MaxWorkers:    workers,
	})
	if err != nil {
		return nil, err
	}

	startFrontendSubscriber(ctx, as, sub, wp.Publish)
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

func (s *server) findAssociations(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (associationsMap, error) {
	paths := []string{
		"data",
		"ids",
		"package_name",
	}
	associations, err := s.registry.ListAssociations(ctx, ids, paths)
	if err != nil {
		return nil, err
	}
	defaults, err := s.registry.ListDefaultAssociations(ctx, ids.ApplicationIdentifiers, paths)
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

func (s *server) fanOut(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	associations, err := s.findAssociations(ctx, msg.EndDeviceIdentifiers)
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
			registerMessageFailed(fmt.Sprintf("publish-%v", name), err)
		}
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
