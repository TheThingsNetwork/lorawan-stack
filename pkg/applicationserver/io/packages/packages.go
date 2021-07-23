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
	"google.golang.org/grpc"
)

const namespace = "applicationserver/io/packages"

type server struct {
	ctx context.Context

	server   io.Server
	registry Registry

	handlers      map[string]ApplicationPackageHandler
	subscriptions map[string]*io.Subscription
}

// Server is an application packages frontend.
type Server interface {
	rpcserver.Registerer
}

func startPackageWorker(ctx context.Context, as io.Server, name string, handler ApplicationPackageHandler, sub *io.Subscription, id int, timeout time.Duration) {
	handleUp := func(ctx context.Context, defAssoc *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation, up *ttnpb.ApplicationUp) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler.HandleUp(ctx, defAssoc, assoc, up)
	}
	as.StartTask(&component.TaskConfig{
		Context: ctx,
		ID:      fmt.Sprintf("run_application_packages_%v_%v", name, id),
		Func: func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-sub.Context().Done():
					return sub.Context().Err()
				case up := <-sub.Up():
					ctx := up.Context
					pair := associationsPairFromContext(ctx)
					if err := handleUp(ctx, pair.defaultAssociation, pair.association, up.ApplicationUp); err != nil {
						log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
						registerMessageFailed(name, err)
						continue
					}
					registerMessageProcessed(name)
				}
			}
		},
		Restart: component.TaskRestartOnFailure,
		Backoff: component.DefaultTaskBackoffConfig,
	})
}

func startFrontendWorker(ctx context.Context, as io.Server, sub *io.Subscription, handler func(context.Context, *ttnpb.ApplicationUp) error) {
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
					if err := handler(ctx, up.ApplicationUp); err != nil {
						log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
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
		ctx:           ctx,
		server:        as,
		registry:      registry,
		handlers:      handlers,
		subscriptions: make(map[string]*io.Subscription),
	}
	for name, handler := range handlers {
		s.subscriptions[name] = io.NewSubscription(ctx, name, nil)
		for id := 0; id < workers; id++ {
			startPackageWorker(ctx, as, name, handler, s.subscriptions[name], id, timeout)
		}
	}
	sub, err := as.Subscribe(ctx, "applicationpackages", nil, false)
	if err != nil {
		return nil, err
	}
	startFrontendWorker(ctx, as, sub, s.handleUp)
	return s, nil
}

type associationsPair struct {
	defaultAssociation *ttnpb.ApplicationPackageDefaultAssociation
	association        *ttnpb.ApplicationPackageAssociation
}

type associationsPairCtxKeyType struct{}

var associationsPairCtxKey = &associationsPairCtxKeyType{}

func contextWithAssociationsPair(ctx context.Context, pair *associationsPair) context.Context {
	return context.WithValue(ctx, associationsPairCtxKey, pair)
}

func associationsPairFromContext(ctx context.Context) *associationsPair {
	if val, ok := ctx.Value(associationsPairCtxKey).(*associationsPair); ok {
		return val
	}
	return nil
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

func (s *server) handleUp(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	associations, err := s.findAssociations(ctx, msg.EndDeviceIdentifiers)
	if err != nil {
		return err
	}
	for name, pair := range associations {
		sub, ok := s.subscriptions[name]
		if !ok {
			continue
		}
		ctx := log.NewContextWithField(ctx, "package", name)
		ctx = contextWithAssociationsPair(ctx, pair)
		if err := sub.Publish(ctx, msg); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
			registerMessageFailed(name, err)
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
