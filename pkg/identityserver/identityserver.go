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

// Package identityserver handles the database operations for The Things Stack.
package identityserver

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/account"
	account_store "go.thethings.network/lorawan-stack/v3/pkg/account/store"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	oauth_store "go.thethings.network/lorawan-stack/v3/pkg/oauth/store"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpctracer"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
	"google.golang.org/grpc"
)

// IdentityServer implements the Identity Server component.
//
// The Identity Server exposes the Registry and Access services for Applications,
// OAuth clients, Gateways, Organizations and Users.
type IdentityServer struct {
	ttnpb.UnimplementedIsServer

	*component.Component
	ctx    context.Context
	config *Config
	db     *sql.DB

	store store.TransactionalStore

	redis   *redis.Client
	account account.Server
	oauth   oauth.Server

	telemetryQueue telemetry.TaskQueue
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

// GenerateCSPString returns a Content-Security-Policy header value
// for OAuth and Account app template.
func GenerateCSPString(config *oauth.Config, nonce string) string {
	baseURLs := webui.RewriteSchemes(
		webui.WebsocketSchemeRewrites,
		config.UI.StackConfig.IS.BaseURL,
	)
	return webui.ContentSecurityPolicy{
		ConnectionSource: append([]string{
			"'self'",
			config.UI.SentryDSN,
			"gravatar.com",
			"www.gravatar.com",
		}, baseURLs...),
		StyleSource: []string{
			"'self'",
			config.UI.AssetsBaseURL,
			config.UI.BrandingBaseURL,
			"'unsafe-inline'",
		},
		ScriptSource: []string{
			"'self'",
			config.UI.AssetsBaseURL,
			config.UI.BrandingBaseURL,
			"'unsafe-eval'",
			"'strict-dynamic'",
			fmt.Sprintf("'nonce-%s'", nonce),
		},
		BaseURI: []string{
			"'self'",
		},
		FrameAncestors: []string{
			"'none'",
		},
	}.Clean().String()
}

type accountAppStore struct {
	store.TransactionalStore
}

// Transact implements account_store.Interface.
func (as *accountAppStore) Transact(ctx context.Context, f func(context.Context, account_store.Interface) error) error {
	return as.TransactionalStore.Transact(ctx, func(ctx context.Context, st store.Store) error { return f(ctx, st) })
}

type oauthAppStore struct {
	store.TransactionalStore
}

// Transact implements oauth_store.Interface.
func (as *oauthAppStore) Transact(ctx context.Context, f func(context.Context, oauth_store.Interface) error) error {
	return as.TransactionalStore.Transact(ctx, func(ctx context.Context, st store.Store) error { return f(ctx, st) })
}

var errDBNeedsMigration = errors.Define("db_needs_migration", "the database needs to be migrated")

// New returns new *IdentityServer.
func New(c *component.Component, config *Config) (is *IdentityServer, err error) {
	ctx := tracer.NewContextWithTracer(c.Context(), tracerNamespace)

	is = &IdentityServer{
		Component:      c,
		ctx:            log.NewContextWithField(ctx, "namespace", logNamespace),
		config:         config,
		telemetryQueue: config.TelemetryQueue,
	}

	if err := is.setupStore(); err != nil {
		return nil, err
	}

	is.config.OAuth.CSRFAuthKey = is.GetBaseConfig(is.Context()).HTTP.Cookie.HashKey
	is.config.OAuth.UI.FrontendConfig.EnableUserRegistration = is.config.UserRegistration.Enabled
	is.oauth, err = oauth.NewServer(c, &oauthAppStore{is.store}, is.config.OAuth, GenerateCSPString)
	if err != nil {
		return nil, err
	}

	is.account, err = account.NewServer(c, &accountAppStore{is.store}, is.config.OAuth, GenerateCSPString)
	if err != nil {
		return nil, err
	}

	c.AddContextFiller(func(ctx context.Context) context.Context {
		ctx = is.withRequestAccessCache(ctx)
		ctx = rights.NewContextWithFetcher(ctx, is)
		ctx = rights.NewContextWithCache(ctx)
		return ctx
	})

	// Tasks initialization.
	if err := is.initializeTelemetryTasks(is.Context()); err != nil {
		return nil, err
	}

	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpctracer.TracerHook, rpctracer.UnaryTracerHook(tracerNamespace)},
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook(logNamespace)},
		{cluster.HookName, c.ClusterAuthUnaryHook()},
	} {
		for _, filter := range []string{
			"/ttn.lorawan.v3.Is",
			"/ttn.lorawan.v3.EntityAccess",
			"/ttn.lorawan.v3.ApplicationRegistry",
			"/ttn.lorawan.v3.ApplicationAccess",
			"/ttn.lorawan.v3.ClientRegistry",
			"/ttn.lorawan.v3.ClientAccess",
			"/ttn.lorawan.v3.EndDeviceRegistry",
			"/ttn.lorawan.v3.EndDeviceBatchRegistry",
			"/ttn.lorawan.v3.GatewayRegistry",
			"/ttn.lorawan.v3.GatewayAccess",
			"/ttn.lorawan.v3.GatewayBatchRegistry",
			"/ttn.lorawan.v3.GatewayBatchAccess",
			"/ttn.lorawan.v3.OrganizationRegistry",
			"/ttn.lorawan.v3.OrganizationAccess",
			"/ttn.lorawan.v3.UserRegistry",
			"/ttn.lorawan.v3.UserAccess",
			"/ttn.lorawan.v3.UserSessionRegistry",
			"/ttn.lorawan.v3.NotificationService",
		} {
			c.GRPC.RegisterUnaryHook(filter, hook.name, hook.middleware)
		}
	}
	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpctracer.TracerHook, rpctracer.UnaryTracerHook(tracerNamespace)},
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook(logNamespace)},
	} {
		for _, filter := range []string{
			"/ttn.lorawan.v3.UserInvitationRegistry",
			"/ttn.lorawan.v3.EntityRegistrySearch",
			"/ttn.lorawan.v3.EndDeviceRegistrySearch",
			"/ttn.lorawan.v3.ContactInfoRegistry",
			"/ttn.lorawan.v3.OAuthAuthorizationRegistry",
		} {
			c.GRPC.RegisterUnaryHook(filter, hook.name, hook.middleware)
		}
	}

	c.RegisterGRPC(is)
	c.RegisterWeb(is.oauth)
	c.RegisterWeb(is.account)
	c.RegisterInterop(is)

	return is, nil
}

// RegisterServices registers services provided by is at s.
func (is *IdentityServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterIsServer(s, is)
	ttnpb.RegisterEntityAccessServer(s, &entityAccess{IdentityServer: is})
	ttnpb.RegisterApplicationRegistryServer(s, &applicationRegistry{IdentityServer: is})
	ttnpb.RegisterApplicationAccessServer(s, &applicationAccess{IdentityServer: is})
	ttnpb.RegisterClientRegistryServer(s, &clientRegistry{IdentityServer: is})
	ttnpb.RegisterClientAccessServer(s, &clientAccess{IdentityServer: is})
	ttnpb.RegisterEndDeviceRegistryServer(s, &endDeviceRegistry{IdentityServer: is})
	ttnpb.RegisterGatewayRegistryServer(s, &gatewayRegistry{IdentityServer: is})
	ttnpb.RegisterGatewayAccessServer(s, &gatewayAccess{IdentityServer: is})
	ttnpb.RegisterGatewayBatchRegistryServer(s, &gatewayBatchRegistry{IdentityServer: is})
	ttnpb.RegisterGatewayBatchAccessServer(s, &gatewayBatchAccess{IdentityServer: is})
	ttnpb.RegisterOrganizationRegistryServer(s, &organizationRegistry{IdentityServer: is})
	ttnpb.RegisterOrganizationAccessServer(s, &organizationAccess{IdentityServer: is})
	ttnpb.RegisterUserRegistryServer(s, &userRegistry{IdentityServer: is})
	ttnpb.RegisterUserAccessServer(s, &userAccess{IdentityServer: is})
	ttnpb.RegisterUserSessionRegistryServer(s, &userSessionRegistry{IdentityServer: is})
	ttnpb.RegisterUserInvitationRegistryServer(s, &invitationRegistry{IdentityServer: is})
	ttnpb.RegisterEntityRegistrySearchServer(s, &registrySearch{IdentityServer: is})
	ttnpb.RegisterEndDeviceRegistrySearchServer(s, &registrySearch{IdentityServer: is})
	ttnpb.RegisterOAuthAuthorizationRegistryServer(s, &oauthRegistry{IdentityServer: is})
	ttnpb.RegisterContactInfoRegistryServer(s, &contactInfoRegistry{IdentityServer: is})
	ttnpb.RegisterNotificationServiceServer(s, &notificationRegistry{IdentityServer: is})
	ttnpb.RegisterEndDeviceBatchRegistryServer(s, &endDeviceBatchRegistry{IdentityServer: is})
}

// RegisterHandlers registers gRPC handlers.
func (is *IdentityServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterIsHandler(is.Context(), s, conn)
	ttnpb.RegisterEntityAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterApplicationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterApplicationAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterClientRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterClientAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterEndDeviceRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterGatewayBatchRegistryHandler(is.Context(), s, conn) // nolint:errcheck
	ttnpb.RegisterGatewayBatchAccessHandler(is.Context(), s, conn)   // nolint:errcheck
	ttnpb.RegisterOrganizationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterOrganizationAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterUserRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterUserAccessHandler(is.Context(), s, conn)
	ttnpb.RegisterUserSessionRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterUserInvitationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterEntityRegistrySearchHandler(is.Context(), s, conn)
	ttnpb.RegisterEndDeviceRegistrySearchHandler(is.Context(), s, conn)
	ttnpb.RegisterOAuthAuthorizationRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterContactInfoRegistryHandler(is.Context(), s, conn)
	ttnpb.RegisterNotificationServiceHandler(is.Context(), s, conn)
	ttnpb.RegisterEndDeviceBatchRegistryHandler(is.Context(), s, conn) // nolint:errcheck
}

// RegisterInterop registers the LoRaWAN Backend Interfaces interoperability services.
func (is *IdentityServer) RegisterInterop(srv *interop.Server) {
	srv.RegisterIS(&interopServer{IdentityServer: is})
}

// Roles returns the roles that the Identity Server fulfills.
func (*IdentityServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_ACCESS, ttnpb.ClusterRole_ENTITY_REGISTRY}
}

var softDeleteFieldMask = []string{"deleted_at"}

var errRestoreWindowExpired = errors.DefineFailedPrecondition("restore_window_expired", "this entity can no longer be restored")

// Close closes the Identity Server database connections and the underlying component.
func (is *IdentityServer) Close() {
	is.db.Close()
	is.Component.Close()
}
