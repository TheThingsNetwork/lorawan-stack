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
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Postgres database driver.
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/email/sendgrid"
	"go.thethings.network/lorawan-stack/pkg/email/smtp"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/oauth"
	"go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config for the Identity Server
type Config struct {
	DatabaseURI      string `name:"database-uri" description:"Database connection URI"`
	UserRegistration struct {
		Invitation struct {
			Required bool          `name:"required" description:"Require invitations for new users"`
			TokenTTL time.Duration `name:"token-ttl" description:"TTL of user invitation tokens"`
		} `name:"invitation"`
		ContactInfoValidation struct {
			Required bool `name:"required" description:"Require contact info validation for new users"`
		} `name:"contact-info-validation"`
		AdminApproval struct {
			Required bool `name:"required" description:"Require admin approval for new users"`
		} `name:"admin-approval"`
		PasswordRequirements struct {
			MinLength    int `name:"min-length" description:"Minimum password length"`
			MinUppercase int `name:"min-uppercase" description:"Minimum number of uppercase letters"`
			MinDigits    int `name:"min-digits" description:"Minimum number of digits"`
			MinSpecial   int `name:"min-special" description:"Minimum number of special characters"`
		} `name:"password-requirements"`
	} `name:"user-registration"`
	AuthCache struct {
		MembershipTTL time.Duration `name:"membership-ttl" description:"TTL of membership caches"`
	} `name:"auth-cache"`
	OAuth          oauth.Config `name:"oauth"`
	ProfilePicture struct {
		UseGravatar bool   `name:"use-gravatar" description:"Use Gravatar fallback for users without profile picture"`
		Bucket      string `name:"bucket" description:"Bucket used for storing profile pictures"`
		BucketURL   string `name:"bucket-url" description:"Base URL for public bucket access"`
	} `name:"profile-picture"`
	Email struct {
		email.Config `name:",squash"`
		SendGrid     sendgrid.Config `name:"sendgrid"`
		SMTP         smtp.Config     `name:"smtp"`
	} `name:"email"`
}

// IdentityServer implements the Identity Server component.
//
// The Identity Server exposes the Registry and Access services for Applications,
// OAuth clients, Gateways, Organizations and Users.
type IdentityServer struct {
	*component.Component
	config *Config
	db     *gorm.DB
	oauth  oauth.Server

	redis *redis.Client
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

// New returns new *IdentityServer.
func New(c *component.Component, config *Config) (is *IdentityServer, err error) {
	is = &IdentityServer{
		Component: c,
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
		ctx = is.withRequestAccessCache(ctx)
		ctx = rights.NewContextWithFetcher(ctx, is)
		return ctx
	})

	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("identityserver")},
		{cluster.HookName, c.ClusterAuthUnaryHook()},
		{rights.HookName, rights.Hook},
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
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.OAuthAuthorizationRegistry", rights.HookName, rights.Hook)

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
	ttnpb.RegisterEntityRegistrySearchServer(s, &registrySearch{IdentityServer: is, adminOnly: true})
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
	ttnpb.RegisterOAuthAuthorizationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterContactInfoRegistryHandler(is.Context(), s, conn)
}

// Roles returns the roles that the Identity Server fulfills.
func (is *IdentityServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_ACCESS, ttnpb.PeerInfo_ENTITY_REGISTRY}
}
