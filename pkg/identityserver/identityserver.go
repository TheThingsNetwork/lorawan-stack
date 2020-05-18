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

package identityserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Postgres database driver.
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// IdentityServer implements the Identity Server component.
//
// The Identity Server exposes the Registry and Access services for Applications,
// OAuth clients, Gateways, Organizations and Users.
type IdentityServer struct {
	*component.Component
	ctx            context.Context
	config         *Config
	db             *gorm.DB
	redis          *redis.Client
	emailTemplates *email.TemplateRegistry
	oauth          oauth.Server
}

// Context returns the context of the Identity Server.
func (is *IdentityServer) Context() context.Context {
	return is.ctx
}

// SetRedisCache configures the given redis instance for caching.
func (is *IdentityServer) SetRedisCache(redis *redis.Client) {
	is.redis = redis
}

type ctxKeyType struct{}

var ctxKey ctxKeyType

func (is *IdentityServer) configFromContext(ctx context.Context) *Config {
	if config, ok := ctx.Value(ctxKey).(*Config); ok {
		return config
	}
	return is.config
}

var errDBNeedsMigration = errors.Define("db_needs_migration", "the database needs to be migrated")

// New returns new *IdentityServer.
func New(c *component.Component, config *Config) (is *IdentityServer, err error) {
	is = &IdentityServer{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "identityserver"),
		config:    config,
	}
	is.db, err = store.Open(is.Context(), is.config.DatabaseURI)
	if err != nil {
		return nil, err
	}
	if c.LogDebug() {
		is.db = is.db.Debug()
	}
	if err = store.Check(is.db); err != nil {
		return nil, errDBNeedsMigration.WithCause(err)
	}
	go func() {
		<-is.Context().Done()
		is.db.Close()
	}()

	is.emailTemplates, err = is.initEmailTemplates(is.Context())
	if err != nil {
		return nil, err
	}

	is.config.OAuth.CSRFAuthKey = is.GetBaseConfig(is.Context()).HTTP.Cookie.HashKey
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
		ctx = is.withRequestAccessCache(ctx)
		ctx = rights.NewContextWithFetcher(ctx, is)
		ctx = rights.NewContextWithCache(ctx)
		return ctx
	})

	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("identityserver")},
		{cluster.HookName, c.ClusterAuthUnaryHook()},
	} {
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.ApplicationRegistry", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.ApplicationAccess", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.ClientRegistry", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.ClientAccess", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.EndDeviceRegistry", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.GatewayRegistry", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.GatewayAccess", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.OrganizationRegistry", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.OrganizationAccess", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.UserRegistry", hook.name, hook.middleware)
		hooks.RegisterUnaryHook("/ttn.lorawan.v3.UserAccess", hook.name, hook.middleware)
	}
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.EntityAccess", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("identityserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.EntityAccess", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.OAuthAuthorizationRegistry", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("identityserver"))

	c.RegisterGRPC(is)
	c.RegisterWeb(is.oauth)

	return is, nil
}

func (is *IdentityServer) withDatabase(ctx context.Context, f func(*gorm.DB) error) error {
	return store.Transact(ctx, is.db, f)
}

// RegisterServices registers services provided by is at s.
func (is *IdentityServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEntityAccessServer(s, &entityAccess{IdentityServer: is})
	ttnpb.RegisterApplicationRegistryServer(s, &applicationRegistry{IdentityServer: is})
	ttnpb.RegisterApplicationAccessServer(s, &applicationAccess{IdentityServer: is})
	ttnpb.RegisterClientRegistryServer(s, &clientRegistry{IdentityServer: is})
	ttnpb.RegisterClientAccessServer(s, &clientAccess{IdentityServer: is})
	ttnpb.RegisterEndDeviceRegistryServer(s, &endDeviceRegistry{IdentityServer: is})
	ttnpb.RegisterGatewayRegistryServer(s, &gatewayRegistry{IdentityServer: is})
	ttnpb.RegisterGatewayAccessServer(s, &gatewayAccess{IdentityServer: is})
	ttnpb.RegisterOrganizationRegistryServer(s, &organizationRegistry{IdentityServer: is})
	ttnpb.RegisterOrganizationAccessServer(s, &organizationAccess{IdentityServer: is})
	ttnpb.RegisterUserRegistryServer(s, &userRegistry{IdentityServer: is})
	ttnpb.RegisterUserAccessServer(s, &userAccess{IdentityServer: is})
	ttnpb.RegisterUserInvitationRegistryServer(s, &invitationRegistry{IdentityServer: is})
	ttnpb.RegisterEntityRegistrySearchServer(s, &registrySearch{IdentityServer: is})
	ttnpb.RegisterEndDeviceRegistrySearchServer(s, &registrySearch{IdentityServer: is})
	ttnpb.RegisterOAuthAuthorizationRegistryServer(s, &oauthRegistry{IdentityServer: is})
	ttnpb.RegisterContactInfoRegistryServer(s, &contactInfoRegistry{IdentityServer: is})
}

// RegisterHandlers registers gRPC handlers.
func (is *IdentityServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEntityAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterApplicationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterApplicationAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterClientRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterClientAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterEndDeviceRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterOrganizationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterOrganizationAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterUserRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterUserAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterUserInvitationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterEntityRegistrySearchHandler(is.Context(), s, conn)
	ttnpb.RegisterEndDeviceRegistrySearchHandler(is.Context(), s, conn)
	ttnpb.RegisterOAuthAuthorizationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterContactInfoRegistryHandler(is.Context(), s, conn)
}

// Roles returns the roles that the Identity Server fulfills.
func (is *IdentityServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_ACCESS, ttnpb.ClusterRole_ENTITY_REGISTRY}
}

func (is *IdentityServer) getMembershipStore(ctx context.Context, db *gorm.DB) store.MembershipStore {
	s := store.GetMembershipStore(db)
	if is.redis != nil {
		if membershipTTL := is.configFromContext(ctx).AuthCache.MembershipTTL; membershipTTL > 0 {
			s = store.GetMembershipCache(s, is.redis, membershipTTL)
		}
	}
	return s
}
