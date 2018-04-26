// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package identityserver

import (
	"net/url"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/sendgrid"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
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
	DefaultSettings ttnpb.IdentityServerSettings `name:"default-settings" description:"Default settings that are loaded when the is first starts"`

	// Specializers are the specializers used in the Identity Server.
	Specializers Specializers `name:"-"`

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
	*organizationService
}

// Specializers are the specializers to be used in the Identity Server.
type Specializers struct {
	User         store.UserSpecializer
	Application  store.ApplicationSpecializer
	Gateway      store.GatewaySpecializer
	Client       store.ClientSpecializer
	Organization store.OrganizationSpecializer
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

	config.Hostname, err = hostname(config.PublicURL)
	if err != nil {
		return nil, err
	}

	is.userService = &userService{is}
	is.applicationService = &applicationService{is}
	is.gatewayService = &gatewayService{is}
	is.clientService = &clientService{is}
	is.adminService = &adminService{is}
	is.organizationService = &organizationService{is}

	if config.Sendgrid != nil && config.Sendgrid.APIKey != "" {
		is.email = sendgrid.New(log, *config.Sendgrid)
	} else {
		log.Warn("No sendgrid API key configured, not sending emails")
		is.email = mock.New()
	}

	hooks.RegisterUnaryHook("/ttn.v3.IsUser", authorizationDataHookName, is.authorizationDataUnaryHook())
	hooks.RegisterUnaryHook("/ttn.v3.IsApplication", authorizationDataHookName, is.authorizationDataUnaryHook())
	hooks.RegisterUnaryHook("/ttn.v3.IsGateway", authorizationDataHookName, is.authorizationDataUnaryHook())
	hooks.RegisterUnaryHook("/ttn.v3.IsClient", authorizationDataHookName, is.authorizationDataUnaryHook())
	hooks.RegisterUnaryHook("/ttn.v3.IsOrganization", authorizationDataHookName, is.authorizationDataUnaryHook())

	c.RegisterGRPC(is)

	return is, nil
}

func hostname(u string) (string, error) {
	p, err := url.Parse(u)
	if err != nil {
		return "", errors.Errorf("Could not parse PublicURL %s", u)
	}

	return p.Hostname(), nil
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
		if err = is.store.Settings.Set(is.config.DefaultSettings); err != nil {
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
	ttnpb.RegisterIsOrganizationServer(s, is.organizationService)
}

// RegisterHandlers registers gRPC handlers.
func (is *IdentityServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterIsUserHandler(is.Context(), s, conn)
	ttnpb.RegisterIsApplicationHandler(is.Context(), s, conn)
	ttnpb.RegisterIsGatewayHandler(is.Context(), s, conn)
	ttnpb.RegisterIsClientHandler(is.Context(), s, conn)
	ttnpb.RegisterIsAdminHandler(is.Context(), s, conn)
	ttnpb.RegisterIsOrganizationHandler(is.Context(), s, conn)
}

// Roles returns the roles that the identity server fulfils
func (is *IdentityServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_IDENTITY_SERVER}
}
