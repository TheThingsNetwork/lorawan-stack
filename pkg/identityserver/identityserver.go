// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"net/url"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/sendgrid"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// Config defines the needed parameters to start the Identity Server.
type Config struct {
	// DatabaseURI is the database connection URI; e.g. "postgres://root@localhost:26257/is_development?sslmode=disable"
	DatabaseURI string `name:"database-uri" description:"URI of the database to connect at"`

	// OrganizationName is the display name of the organization that runs the network.
	// e.g. The Things Network
	OrganizationName string `name:"organization-name" description:"The name of the organization who is in behalf of this server"`

	// PublicURL is the public url this server will use to serve content such as
	// email content. e.g. https://www.thethingsnetwork.org
	PublicURL string `name:"public-url" description:"Public URL this server uses to serve content such as email content"`

	// Sendgrid is the sendgrid config.
	Sendgrid *sendgrid.Config `name:"sendgrid"`

	// defaultSettings are the default settings within the tenant loaded in the storewhen it first-time initialized.
	DefaultSettings *ttnpb.IdentityServerSettings `name:"default-settings" description:"Default settings that are loaded when the is first starts"`

	// Factories are the factories used in the identity server
	Factories Factories `name:"-"`

	// Hostname denotes the Identity Server hostname. It is used as issuer when
	// generating access tokens and API keys.
	Hostname string `name:"-"`
}

// IdentityServer implements the Identity Server component behaviour.
type IdentityServer struct {
	*component.Component

	config Config

	store *sql.Store
	email email.Provider

	*userService
	*applicationService
	*gatewayService
	*clientService
	*adminService
}

// Factories are the factories to be used in the identity server.
type Factories struct {
	User        store.UserFactory
	Application store.ApplicationFactory
	Gateway     store.GatewayFactory
	Client      store.ClientFactory
}

// New returns a new IdentityServer.
func New(c *component.Component, config Config) (*IdentityServer, error) {
	log := log.FromContext(c.Context()).WithField("namespace", "is")
	store, err := sql.Open(config.DatabaseURI)
	if err != nil {
		return nil, err
	}

	is := &IdentityServer{
		Component: c,
		store:     store,
		config:    config,
	}

	config.Hostname = hostname(config.PublicURL)

	is.userService = &userService{is}
	is.applicationService = &applicationService{is}
	is.gatewayService = &gatewayService{is}
	is.clientService = &clientService{is}
	is.adminService = &adminService{is}

	if config.Sendgrid != nil && config.Sendgrid.APIKey != "" {
		is.email = sendgrid.New(log, *config.Sendgrid)
	} else {
		log.Warn("No sendgrid API key configured, not sending emails")
		is.email = mock.New()
	}

	c.RegisterGRPC(is)

	return is, nil
}

func hostname(u string) string {
	p, err := url.Parse(u)
	if err != nil {
		panic(err)
	}

	return p.Hostname()
}

// Init initializes the store and sets the default settings in case they aren't.
func (is *IdentityServer) Init() error {
	err := is.store.Init()
	if err != nil {
		return err
	}

	// set default settings if these are not set yet
	_, err = is.store.Settings.Get()
	if sql.ErrSettingsNotFound.Describes(err) {
		if err := is.store.Settings.Set(is.config.DefaultSettings); err != nil {
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
	ttnpb.RegisterIsUserServer(s, is.userService)
	ttnpb.RegisterIsApplicationServer(s, is.applicationService)
	ttnpb.RegisterIsGatewayServer(s, is.gatewayService)
	ttnpb.RegisterIsClientServer(s, is.clientService)
	ttnpb.RegisterIsAdminServer(s, is.adminService)
}

// RegisterHandlers registers gRPC handlers.
func (is *IdentityServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {}

// Roles returns the roles that the identity server fulfils
func (is *IdentityServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_IDENTITY_SERVER}
}
