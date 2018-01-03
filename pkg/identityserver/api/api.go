// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package api

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// GRPC implements the multiple gRPC server interfaces of the Identity Server.
type GRPC struct {
	store *sql.Store
	email email.Provider

	factories struct {
		user        store.UserFactory
		application store.ApplicationFactory
		gateway     store.GatewayFactory
		client      store.ClientFactory
	}
}

// GRPCOption is a initialization option of the GRPC type in its constructor.
type GRPCOption func(*GRPC)

// WithEmailProvider allows to replace the default (mock) email provider.
func WithEmailProvider(provider email.Provider) GRPCOption {
	return func(g *GRPC) {
		g.email = provider
	}
}

// WithUserFactory allows to replace the default user ttnpb.User factory.
func WithUserFactory(factory store.UserFactory) GRPCOption {
	return func(g *GRPC) {
		g.factories.user = factory
	}
}

var defaultUserFactory = func() types.User {
	return &ttnpb.User{}
}

// WithApplicationFactory allows to replace the default application ttnpb.Application factory.
func WithApplicationFactory(factory store.ApplicationFactory) GRPCOption {
	return func(g *GRPC) {
		g.factories.application = factory
	}
}

var defaultApplicationFactory = func() types.Application {
	return &ttnpb.Application{}
}

// WithGatewayFactory allows to replace the default gateway ttnpb.Gateway factory.
func WithGatewayFactory(factory store.GatewayFactory) GRPCOption {
	return func(g *GRPC) {
		g.factories.gateway = factory
	}
}

var defaultGatewayFactory = func() types.Gateway {
	return &ttnpb.Gateway{}
}

// WithClientFactory allows to replace the default client ttnpb.Client factory.
func WithClientFactory(factory store.ClientFactory) GRPCOption {
	return func(g *GRPC) {
		g.factories.client = factory
	}
}

var defaultClientFactory = func() types.Client {
	return &ttnpb.Client{}
}

// NewGRPC returns a new gRPC API implementation of the Identity Server.
func NewGRPC(store *sql.Store, opts ...GRPCOption) *GRPC {
	grpc := &GRPC{
		store: store,
		email: mock.New(),
	}
	grpc.factories.user = defaultUserFactory
	grpc.factories.application = defaultApplicationFactory
	grpc.factories.gateway = defaultGatewayFactory
	grpc.factories.client = defaultClientFactory

	for _, opt := range opts {
		opt(grpc)
	}

	return grpc
}
