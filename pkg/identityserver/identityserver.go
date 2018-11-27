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
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Postgres database driver.
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/oauth"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config for the Identity Server
type Config struct {
	DatabaseURI string       `name:"database-uri" description:"Database connection URI"`
	OAuth       oauth.Config `name:"oauth"`
}

// IdentityServer implements the Identity Server component.
//
// The Identity Server exposes the Registry and Access services for Applications,
// OAuth Clients, Gateways, Organizations and Users.
type IdentityServer struct {
	*component.Component
	config *Config
	db     *gorm.DB
	oauth  oauth.Server
}

// New returns new *IdentityServer.
func New(c *component.Component, config *Config) (is *IdentityServer, err error) {
	is = &IdentityServer{
		Component: c,
		config:    config,
	}
	is.db, err = gorm.Open("postgres", is.config.DatabaseURI)
	if err != nil {
		return nil, err
	}
	if c.LogDebug() {
		is.db = is.db.Debug()
	}
	err = store.Check(is.db)
	if err != nil {
		return nil, err
	}
	go func() {
		<-is.Context().Done()
		is.db.Close()
	}()

	is.oauth = oauth.NewServer(is.Context(), struct {
		store.UserStore
		store.UserSessionStore
		store.ClientStore
		store.OAuthStore
	}{
		UserStore:        store.GetUserStore(is.db),
		UserSessionStore: store.GetUserSessionStore(is.db),
		ClientStore:      store.GetClientStore(is.db),
		OAuthStore:       store.GetOAuthStore(is.db),
	}, is.config.OAuth)

	c.AddContextFiller(func(ctx context.Context) context.Context {
		return rights.NewContextWithFetcher(ctx, is)
	})

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.ApplicationRegistry", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.ApplicationAccess", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.ClientRegistry", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.ClientAccess", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GatewayRegistry", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GatewayAccess", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.OrganizationRegistry", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.OrganizationAccess", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.UserRegistry", rights.HookName, rights.Hook)
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.UserAccess", rights.HookName, rights.Hook)

	c.RegisterGRPC(is)
	c.RegisterWeb(is.oauth)

	return is, nil
}

func (is *IdentityServer) withDatabase(ctx context.Context, f func(*gorm.DB) error) error {
	return store.Transact(ctx, is.db, f)
}

// RegisterServices registers services provided by is at s.
func (is *IdentityServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterApplicationRegistryServer(s, &applicationRegistry{IdentityServer: is})
	ttnpb.RegisterApplicationAccessServer(s, &applicationAccess{IdentityServer: is})
	ttnpb.RegisterClientRegistryServer(s, &clientRegistry{IdentityServer: is})
	ttnpb.RegisterClientAccessServer(s, &clientAccess{IdentityServer: is})
	ttnpb.RegisterGatewayRegistryServer(s, &gatewayRegistry{IdentityServer: is})
	ttnpb.RegisterGatewayAccessServer(s, &gatewayAccess{IdentityServer: is})
	ttnpb.RegisterOrganizationRegistryServer(s, &organizationRegistry{IdentityServer: is})
	ttnpb.RegisterOrganizationAccessServer(s, &organizationAccess{IdentityServer: is})
	ttnpb.RegisterUserRegistryServer(s, &userRegistry{IdentityServer: is})
	ttnpb.RegisterUserAccessServer(s, &userAccess{IdentityServer: is})
}

// RegisterHandlers registers gRPC handlers.
func (is *IdentityServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterApplicationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterApplicationAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterClientRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterClientAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterOrganizationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterOrganizationAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterUserRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterUserAccessHandler(is.Context(), s, conn)
}

// Roles returns the roles that the Identity Server fulfills.
func (is *IdentityServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_ACCESS, ttnpb.PeerInfo_ENTITY_REGISTRY}
}
