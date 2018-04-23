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

package rights

import (
	"context"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

type applicationIdentifiersGetters interface {
	GetApplicationID() string
}

type gatewayIdentifiersGetters interface {
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

type componentConnector struct {
	*component.Component

	tags     []string
	shardKey []byte
}

func (c componentConnector) Get(context.Context) *grpc.ClientConn {
	peer := c.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, c.tags, c.shardKey)
	if peer == nil {
		return nil
	}

	return peer.Conn()
}

// ConnectorFromComponent returns an IdentityServerConnector tied to a component.
// The tags and shard key are then used to select the identity server to use.
func ConnectorFromComponent(c *component.Component, tags []string, shardKey []byte) IdentityServerConnector {
	return componentConnector{
		Component: c,
		tags:      tags,
		shardKey:  shardKey,
	}
}

// Config is the type that configures the rights hook.
type Config struct {
	// TTL is the duration that entries will remain in the cache before being
	// garbage collected. If the value is not set (i.e. 0) caching will be disabled.
	TTL time.Duration `name:"ttl" description:"How long to cache client authorizations before re-validating with the identity server. Valid example values: 30s, 5m or 1h."`
}

// Hook implements a gRPC unary hook that preloads in the context the rights
// based on the authorization value in the request metadata with the resource
// that is being trying to be accessed to.
type Hook struct {
	ctx               context.Context
	logger            log.Interface
	conn              IdentityServerConnector
	applicationsCache cache
	gatewaysCache     cache
}

// New returns a new hook instance. ctx is a cancelable context
// used to stop the garbage collector of the TTL cache if it has been set.
func New(ctx context.Context, conn IdentityServerConnector, config Config) (*Hook, error) {
	if conn == nil {
		return nil, errors.New("An Identity Server connection provider must be given")
	}

	h := &Hook{
		ctx:  log.NewContextWithField(ctx, "hook", "rights"),
		conn: conn,
	}
	h.logger = log.FromContext(ctx)

	if config.TTL == time.Duration(0) {
		h.logger.Warn("Not setting up the TTL cache as the TTL value was not set in the config")
		h.applicationsCache = new(noopCache)
		h.gatewaysCache = new(noopCache)
	} else {
		h.applicationsCache = newTTLCache(h.ctx, config.TTL)
		h.gatewaysCache = newTTLCache(h.ctx, config.TTL)
	}

	return h, nil
}

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

			conn := h.conn.Get(ctx)
			if conn == nil {
				return nil, errors.New("No Identity Server to connect to")
			}

			if m, ok := req.(applicationIdentifiersGetters); ok {
				appIDs := new(ttnpb.ApplicationIdentifiers)
				appIDs.ApplicationID = m.GetApplicationID()

				if !appIDs.IsZero() {
					key := fmt.Sprintf("%s:%s", md.AuthValue, appIDs.UniqueID(ctx))

					rights, err := h.applicationsCache.GetOrFetch(key, func() (rights []ttnpb.Right, err error) {
						resp, err := ttnpb.NewIsApplicationClient(conn).ListApplicationRights(ctx, appIDs, grpc.PerRPCCredentials(md))
						if err != nil {
							return nil, errors.NewWithCause(err, "Failed to fetch application rights")
						}

						return resp.Rights, nil
					})
					if err != nil {
						return nil, err
					}

					return next(NewContext(ctx, rights), req)
				}
			}

			if m, ok := req.(gatewayIdentifiersGetters); ok {
				gtwIDs := new(ttnpb.GatewayIdentifiers)
				gtwIDs.GatewayID = m.GetGatewayID()

				if !gtwIDs.IsZero() {
					key := fmt.Sprintf("%s:%s", md.AuthValue, gtwIDs.UniqueID(ctx))

					rights, err := h.gatewaysCache.GetOrFetch(key, func() (rights []ttnpb.Right, err error) {
						resp, err := ttnpb.NewIsGatewayClient(conn).ListGatewayRights(ctx, gtwIDs, grpc.PerRPCCredentials(md))
						if err != nil {
							return nil, errors.NewWithCause(err, "Failed to fetch gateway rights")
						}

						return resp.Rights, nil
					})
					if err != nil {
						return nil, err
					}

					return next(NewContext(ctx, rights), req)
				}
			}

			return next(ctx, req)
		}
	}
}
