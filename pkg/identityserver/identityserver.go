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
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
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
	"github.com/TheThingsNetwork/ttn/pkg/validate"
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

	// Specializers are the specializers used in the Identity Server.
	Specializers Specializers `name:"specializers" description:"IDs of used specializers for read-operations in the store."`

	// Hostname denotes the Identity Server hostname. It is used as issuer when
	// generating access tokens and API keys.
	Hostname string `name:"-"`
}

// Specializers contains the IDs of the specializers that will be used in
// read-operations in the store.
//
// If an empty value is provided it means no specializer will be used.
type Specializers struct {
	User         string `name:"user" description:"ID of the user specializer."`
	Application  string `name:"application" description:"ID of the application specializer."`
	Gateway      string `name:"gateway" description:"ID of the gateway specializer."`
	Client       string `name:"client" description:"ID of the client specializer."`
	Organization string `name:"organization" description:"ID of the organization specializer."`
}

// IdentityServer implements the Identity Server component behaviour.
type IdentityServer struct {
	// TODO: document things that need to be taken into account or to be done when updating the
	// rights list. See https://github.com/TheThingsIndustries/ttn/issues/724.
	*component.Component

	config Config

	store *store.Store
	email email.Provider

	logger log.Interface

	specializers struct {
		User         store.UserSpecializer
		Application  store.ApplicationSpecializer
		Gateway      store.GatewaySpecializer
		Client       store.ClientSpecializer
		Organization store.OrganizationSpecializer
	}

	*userService
	*applicationService
	*gatewayService
	*clientService
	*adminService
	*organizationService
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
		logger:    log,
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

	is.specializers.User, err = specializers.GetUser(config.Specializers.User)
	if err != nil {
		return nil, err
	}

	is.specializers.Application, err = specializers.GetApplication(config.Specializers.Application)
	if err != nil {
		return nil, err
	}

	is.specializers.Gateway, err = specializers.GetGateway(config.Specializers.Gateway)
	if err != nil {
		return nil, err
	}

	is.specializers.Client, err = specializers.GetClient(config.Specializers.Client)
	if err != nil {
		return nil, err
	}

	is.specializers.Organization, err = specializers.GetOrganization(config.Specializers.Organization)
	if err != nil {
		return nil, err
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

type InitialData struct {
	Settings ttnpb.IdentityServerSettings `name:"settings"`
	Admin    InitialAdminData             `name:"admin"`
	Console  InitialConsoleData           `name:"console"`
}

type InitialAdminData struct {
	UserID   string `name:"user-id" description:"User ID of the admin."`
	Email    string `name:"email" description:"Email of the admin."`
	Password string `name:"password" description:"Password of the admin."`
}

type InitialConsoleData struct {
	ClientSecret string `name:"client-secret" description:"console OAuth client secret"`
	RedirectURI  string `name:"redirect-uri" description:"console OAuth client redirect URI"`
}

// Validate returns error if InitialData is not valid.
func (d InitialData) Validate() error {
	return validate.All(
		validate.Field(d.Admin.UserID, validate.ID).DescribeFieldName("Admin User ID"),
		validate.Field(d.Admin.Password, validate.Required).DescribeFieldName("Admin password"),
		validate.Field(d.Admin.Email, validate.Email).DescribeFieldName("Admin email"),
		validate.Field(d.Console.ClientSecret, validate.Required).DescribeFieldName("Console client secret"),
		validate.Field(d.Console.RedirectURI, validate.Required).DescribeFieldName("Console redirect URI"),
	)
}

// Init initializes the Identity Server creating the database, applying the migrations to create
// the schema and inserting the initial given data. It fails if the database already exists.
func (is *IdentityServer) Init(data InitialData) error {
	err := data.Validate()
	if err != nil {
		return err
	}

	password, err := auth.Hash(data.Admin.Password)
	if err != nil {
		return err
	}

	// Returns error if database already exists.
	err = is.store.Init()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			is.store.Clean()
		}
	}()

	err = is.store.Transact(func(tx *store.Store) error {
		now := time.Now().UTC()

		err := tx.Users.Create(&ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: data.Admin.UserID,
				Email:  data.Admin.Email,
			},
			Name:              "Admin",
			Password:          password.String(),
			State:             ttnpb.STATE_APPROVED,
			Admin:             true,
			ValidatedAt:       timeValue(now),
			CreatedAt:         now,
			UpdatedAt:         now,
			PasswordUpdatedAt: now,
		})
		if err != nil {
			return err
		}
		is.logger.Infof("Created admin user with User ID `%s` and password `%s`", data.Admin.UserID, data.Admin.Password)

		err = tx.Clients.Create(&ttnpb.Client{
			ClientIdentifiers: ttnpb.ClientIdentifiers{
				ClientID: "console",
			},
			Description:       "The console is the official The Things Network web application.",
			Secret:            data.Console.ClientSecret,
			RedirectURI:       data.Console.RedirectURI,
			SkipAuthorization: true,
			State:             ttnpb.STATE_APPROVED,
			CreatorIDs: ttnpb.UserIdentifiers{
				UserID: data.Admin.UserID,
			},
			Grants:    []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_REFRESH_TOKEN},
			Rights:    ttnpb.AllRights(),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}

		return tx.Settings.Set(data.Settings)
	})

	return err
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
