// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/api"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// IdentityServer implements the Identity Server component behaviour.
type IdentityServer struct {
	*component.Component
	*api.GRPC

	store  *sql.Store
	config *Config
}

// Config defines the needed parameters to start the Identity Server.
type Config struct {
	config.ServiceBase
	DSN              string
	Hostname         string
	RecreateDatabase bool
}

// New retrieves a new IdentityServer.
func New(logger log.Stack, config *Config) (*IdentityServer, error) {
	store, err := sql.Open(config.DSN)
	if err != nil {
		return nil, err
	}

	return &IdentityServer{
		Component: component.New(logger, &component.Config{
			ServiceBase:          config.ServiceBase,
			TokenKeyInfoProvider: &tokenKeyProvider{store},
		}),
		GRPC:   api.NewGRPC(store),
		store:  store,
		config: config,
	}, nil
}

// Start initializes the store, sets the default settings in case they are not
// set and starts the registered gRPC services.
func (is *IdentityServer) Start() error {
	err := is.store.Init(is.config.RecreateDatabase)
	if err != nil {
		return err
	}

	// set default settings if these are not set yet
	_, err = is.store.Settings.Get()
	if sql.ErrSettingsNotFound.Describes(err) {
		if err := is.store.Settings.Set(defaultSettings); err != nil {
			return err
		}
	}
	if !sql.ErrSettingsNotFound.Describes(err) && err != nil {
		return err
	}

	return is.Component.Start()
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
