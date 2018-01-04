// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// IdentityServer implements the Identity Server component behaviour.
type IdentityServer struct {
	*component.Component

	config *Config

	store *sql.Store
	email email.Provider

	factories struct {
		user        store.UserFactory
		application store.ApplicationFactory
		gateway     store.GatewayFactory
		client      store.ClientFactory
	}
}

// Config defines the needed parameters to start the Identity Server.
type Config struct {
	DSN              string
	Hostname         string
	RecreateDatabase bool

	defaultSettings *ttnpb.IdentityServerSettings
}

type Option func(*IdentityServer)

// WithEmailProvider allows to replace the default (mock) email provider.
func WithEmailProvider(provider email.Provider) Option {
	return func(is *IdentityServer) {
		is.email = provider
	}
}

// WithUserFactory allows to replace the default user ttnpb.User factory.
func WithUserFactory(factory store.UserFactory) Option {
	return func(is *IdentityServer) {
		is.factories.user = factory
	}
}

var defaultUserFactory = func() types.User {
	return &ttnpb.User{}
}

// WithApplicationFactory allows to replace the default application ttnpb.Application factory.
func WithApplicationFactory(factory store.ApplicationFactory) Option {
	return func(is *IdentityServer) {
		is.factories.application = factory
	}
}

var defaultApplicationFactory = func() types.Application {
	return &ttnpb.Application{}
}

// WithGatewayFactory allows to replace the default gateway ttnpb.Gateway factory.
func WithGatewayFactory(factory store.GatewayFactory) Option {
	return func(is *IdentityServer) {
		is.factories.gateway = factory
	}
}

var defaultGatewayFactory = func() types.Gateway {
	return &ttnpb.Gateway{}
}

// WithClientFactory allows to replace the default client ttnpb.Client factory.
func WithClientFactory(factory store.ClientFactory) Option {
	return func(is *IdentityServer) {
		is.factories.client = factory
	}
}

var defaultClientFactory = func() types.Client {
	return &ttnpb.Client{}
}

// WithDefaultSettings allows to replace the default settings that are loaded
// when the store is first-time initialized.
func WithDefaultSettings(settings *ttnpb.IdentityServerSettings) Option {
	return func(is *IdentityServer) {
		is.config.defaultSettings = settings
	}
}

var defaultOptions = []Option{
	WithEmailProvider(mock.New()),
	WithUserFactory(defaultUserFactory),
	WithApplicationFactory(defaultApplicationFactory),
	WithGatewayFactory(defaultGatewayFactory),
	WithClientFactory(defaultClientFactory),
	WithDefaultSettings(defaultSettings),
}

// New retrieves a new IdentityServer.
func New(comp *component.Component, config *Config, opts ...Option) (*IdentityServer, error) {
	store, err := sql.Open(config.DSN)
	if err != nil {
		return nil, err
	}

	is := &IdentityServer{
		Component: comp,
		store:     store,
		config:    config,
	}

	opts = append(defaultOptions, opts...)

	for _, opt := range opts {
		opt(is)
	}

	return is, nil
}

// Start initializes the store, sets the default settings in case they are not
// set and starts the registered gRPC services.
func (is *IdentityServer) Start() error {
	err := is.start()
	if err != nil {
		return err
	}

	return is.Component.Start()
}

// start inits all the IdentityServer stuff that is not related with the base component.
func (is *IdentityServer) start() error {
	err := is.store.Init(is.config.RecreateDatabase)
	if err != nil {
		return err
	}

	// set default settings if these are not set yet
	_, err = is.store.Settings.Get()
	if sql.ErrSettingsNotFound.Describes(err) {
		if err := is.store.Settings.Set(is.config.defaultSettings); err != nil {
			return err
		}
	}
	if !sql.ErrSettingsNotFound.Describes(err) && err != nil {
		return err
	}

	return nil
}

// RegisterServices registers services provided by is at s.
func (is *IdentityServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterIsUserServer(s, is)
	ttnpb.RegisterIsApplicationServer(s, is)
	ttnpb.RegisterIsGatewayServer(s, is)
	ttnpb.RegisterIsClientServer(s, is)
	ttnpb.RegisterIsSettingsServer(s, is)
}

// RegisterHandlers registers gRPC handlers.
func (is *IdentityServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {}
