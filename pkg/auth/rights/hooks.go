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

// Package rights implements a gRPC unary hook that preloads rights in the context.
package rights

import (
	"context"
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/pkg/config"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

type organizationIDGetter interface {
	GetOrganizationID() string
}

type applicationIDGetter interface {
	GetApplicationID() string
}

type gatewayIDGetter interface {
	GetGatewayID() string
}

// HookName denotes the unique name that components should use to register this hook.
const HookName = "rights-fetcher"

// IdentityServerConnector is the interface that provides a method to get a
// gRPC connection to an Identity Server, tipically within the cluster.
type IdentityServerConnector interface {
	// Get retrieves a gRPC connection to an Identity Server.
	// The context of the current request is passed by argument.
	Get(context.Context) *grpc.ClientConn
}

// Config is the type that configures the rights hook.
type Config struct {
	config.Rights

	// AllowInsecure enables transmission of credentials over insecure connections.
	AllowInsecure bool
}

// Hook implements a gRPC unary hook that preloads in the context the rights
// based on the authorization value in the request metadata with the resource
// that is being trying to be accessed to.
//
// On an incoming RPC:
//
// 1. The hook looks up whether the request message identifies one of the entity Getter interface,
// defined in the package.
//
// 2. Using the IdentityServerConnector, the hook retrieves the rights associated to the authorization value of the request.
//
// 3. The resulting rights are put in the context using NewContext, and can be retrieved using FromContext.
type Hook struct {
	ctx                context.Context
	logger             log.Interface
	config             Config
	connector          IdentityServerConnector
	organizationsCache cache
	applicationsCache  cache
	gatewaysCache      cache
}

// New returns a new instance of a hook that preloads rights in the context.
//
// If a TTL cache is set in the config, a garbage collector runs until ctx is cancelled.
func New(ctx context.Context, connector IdentityServerConnector, config Config) (*Hook, error) {
	if connector == nil {
		return nil, errors.New("An Identity Server connection provider must be given")
	}

	h := &Hook{
		ctx:       log.NewContextWithField(ctx, "hook", "rights"),
		config:    config,
		connector: connector,
	}
	h.logger = log.FromContext(h.ctx)

	if config.TTL == time.Duration(0) {
		h.logger.Warn("No rights cache TTL configured, not caching rights")
		h.organizationsCache = new(noopCache)
		h.applicationsCache = new(noopCache)
		h.gatewaysCache = new(noopCache)
	} else {
		h.organizationsCache = newTTLCache(h.ctx, config.TTL)
		h.applicationsCache = newTTLCache(h.ctx, config.TTL)
		h.gatewaysCache = newTTLCache(h.ctx, config.TTL)
	}

	return h, nil
}

var errNoIdentityServer = errors.DefineInternal(
	"no_identity_server",
	"no Identity Server to fetch rights from",
)

var errFetchFailed = errors.Define(
	"rights_fetch_failed",
	"failed to fetch rights from identity server",
)

var errEmptyIdentifiers = errors.DefineInternal(
	"missing_rights_identifiers",
	"missing identifiers in RPC argument",
)

// UnaryHook returns an unary handler middleware which loads in the context
// the rights that the authorization data has to the application or gateway
// that is being trying to be accessed to.
func (h *Hook) UnaryHook() hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			md := rpcmetadata.FromIncomingContext(ctx)

			// If the request does not include an authorization value we can just skip
			// as the hook wont be able to make a call to the Identity Server.
			if md.AuthValue == "" {
				return next(ctx, req)
			}

			conn := h.connector.Get(ctx)
			if conn == nil {
				return nil, errNoIdentityServer
			}

			var (
				rights []ttnpb.Right
				err    error
			)

			if m, ok := req.(organizationIDGetter); ok {
				orgIDs := &ttnpb.OrganizationIdentifiers{
					OrganizationID: m.GetOrganizationID(),
				}

				if !orgIDs.IsZero() {
					defer func() { registerRightsRequest(ctx, "organization", rights, err) }()

					uid := orgIDs.UniqueID(ctx)
					key := fmt.Sprintf("%s:%s", md.AuthValue, uid)

					rights, err = h.organizationsCache.GetOrFetch(key, func() (rights []ttnpb.Right, err error) {
						defer func() { registerRightsFetch(ctx, "organization", rights, err) }()
						md.AllowInsecure = h.config.AllowInsecure
						resp, err := ttnpb.NewIsOrganizationClient(conn).ListOrganizationRights(ctx, orgIDs, grpc.PerRPCCredentials(md))
						switch {
						case err == nil:
							return resp.Rights, nil
						case errors.IsInvalidArgument(err) || errors.IsUnauthenticated(err) || errors.IsPermissionDenied(err):
							return nil, err
						default:
							return nil, errFetchFailed.WithCause(err)
						}
					})
					if err != nil {
						return nil, err
					}

					return next(newContext(ctx, uid, rights), req)
				}
			}

			if m, ok := req.(applicationIDGetter); ok {
				appIDs := &ttnpb.ApplicationIdentifiers{
					ApplicationID: m.GetApplicationID(),
				}

				if !appIDs.IsZero() {
					defer func() { registerRightsRequest(ctx, "application", rights, err) }()

					uid := appIDs.UniqueID(ctx)
					key := fmt.Sprintf("%s:%s", md.AuthValue, uid)

					rights, err = h.applicationsCache.GetOrFetch(key, func() (rights []ttnpb.Right, err error) {
						defer func() { registerRightsFetch(ctx, "application", rights, err) }()
						md.AllowInsecure = h.config.AllowInsecure
						resp, err := ttnpb.NewIsApplicationClient(conn).ListApplicationRights(ctx, appIDs, grpc.PerRPCCredentials(md))
						switch {
						case err == nil:
							return resp.Rights, nil
						case errors.IsInvalidArgument(err) || errors.IsUnauthenticated(err) || errors.IsPermissionDenied(err):
							return nil, err
						default:
							return nil, errFetchFailed.WithCause(err)
						}
					})
					if err != nil {
						return nil, err
					}

					return next(newContext(ctx, uid, rights), req)
				}
			}

			if m, ok := req.(gatewayIDGetter); ok {
				gtwIDs := &ttnpb.GatewayIdentifiers{
					GatewayID: m.GetGatewayID(),
				}

				if !gtwIDs.IsZero() {
					defer func() { registerRightsRequest(ctx, "gateway", rights, err) }()

					uid := gtwIDs.UniqueID(ctx)
					key := fmt.Sprintf("%s:%s", md.AuthValue, uid)

					rights, err = h.gatewaysCache.GetOrFetch(key, func() (rights []ttnpb.Right, err error) {
						defer func() { registerRightsFetch(ctx, "gateway", rights, err) }()
						md.AllowInsecure = h.config.AllowInsecure
						resp, err := ttnpb.NewIsGatewayClient(conn).ListGatewayRights(ctx, gtwIDs, grpc.PerRPCCredentials(md))
						switch {
						case err == nil:
							return resp.Rights, nil
						case errors.IsInvalidArgument(err) || errors.IsUnauthenticated(err) || errors.IsPermissionDenied(err):
							return nil, err
						default:
							return nil, errFetchFailed.WithCause(err)
						}
					})
					if err != nil {
						return nil, err
					}

					return next(newContext(ctx, uid, rights), req)
				}
			}

			return nil, errEmptyIdentifiers
		}
	}
}
