// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/sendgrid"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
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
	// DSN is the Data Source Name used by the store to connect to the database.
	DSN string `name:"dsn" description:"Data Source Name to connect to the database"`

	// Hostname denotes the Identity Server hostname. It is used as issuer when
	// generating access tokens and API keys.
	Hostname string `name:"hostname" description:"Hostname where this server is running. Used as issuer when generating access tokens and API keys"`

	// RecreateDatabase denotes if the database is recreated when the store is initialized.
	// WARNING: it will erase all the previous data
	RecreateDatabase bool `name:"recreate-database" description:"Recreates the database when the server is initialized. WARNING: it deletes all previous data"`

	// OrganizationName is the display name of the organization that runs the network.
	// e.g. The Things Network
	OrganizationName string `name:"organization-name" description:"The name of the organization who is in behalf of this server"`

	// PublicURL is the public url this server will use to serve content such as
	// email content. e.g. https://www.thethingsnetwork.org
	PublicURL string `name:"public-url" description:"Public URL this server uses to serve content such as email content"`

	// SendGridAPIKey is the API key issued by SendGrid to send emails using its service.
	SendGridAPIKey string `name:"sendgrid-api-key" description:"SendGrid API Key. If left blank the mock email provider will be used"`

	// defaultSettings are the default settings within the tenant loaded in the store
	// when it first-time initialized.
	defaultSettings *ttnpb.IdentityServerSettings
}

// Option is the type for options of the Identity Server.
type Option func(*IdentityServer)

// WithEmailProvider replaces the default (mock) email provider.
func WithEmailProvider(provider email.Provider) Option {
	return func(is *IdentityServer) {
		is.email = provider
	}
}

// WithUserFactory replaces the default user ttnpb.User factory.
func WithUserFactory(factory store.UserFactory) Option {
	return func(is *IdentityServer) {
		is.factories.user = factory
	}
}

var defaultUserFactory = func() store.User {
	return &ttnpb.User{}
}

// WithApplicationFactory replaces the default application ttnpb.Application factory.
func WithApplicationFactory(factory store.ApplicationFactory) Option {
	return func(is *IdentityServer) {
		is.factories.application = factory
	}
}

var defaultApplicationFactory = func() store.Application {
	return &ttnpb.Application{}
}

// WithGatewayFactory replaces the default gateway ttnpb.Gateway factory.
func WithGatewayFactory(factory store.GatewayFactory) Option {
	return func(is *IdentityServer) {
		is.factories.gateway = factory
	}
}

var defaultGatewayFactory = func() store.Gateway {
	return &ttnpb.Gateway{}
}

// WithClientFactory replaces the default client ttnpb.Client factory.
func WithClientFactory(factory store.ClientFactory) Option {
	return func(is *IdentityServer) {
		is.factories.client = factory
	}
}

var defaultClientFactory = func() store.Client {
	return &ttnpb.Client{}
}

// WithDefaultSettings replaces the default settings that are loaded when the
// store is first-time initialized.
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

// New returns a new IdentityServer.
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

	if len(config.SendGridAPIKey) != 0 {
		opts = append(opts, WithEmailProvider(sendgrid.New(comp.Logger(), config.SendGridAPIKey, sendgrid.SenderAddress(config.OrganizationName, fmt.Sprintf("noreply@%s", config.Hostname)))))
	}

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
