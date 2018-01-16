// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

// RightsFetcher is a specialized Identity Server gRPC client that can be used
// for listing the rights that either an API key or access token has to an
// application or gateway. Tipically used by network components to check that
// a request with credentials is authorized to access a certain protected resource.
//
// Optionally an expirable cache can be set to the client so the responses are
// cached in order to reduce the load in the Identity Server.
//
// RightsFetcher is safe to be used concurrently with different credentials.
type RightsFetcher struct {
	applications      ttnpb.IsApplicationClient
	applicationsCache cache

	gateways      ttnpb.IsGatewayClient
	gatewaysCache cache
}

// Option is the type that configures a RightsFetcher.
type Option func(*RightsFetcher)

// WithTTLCache sets to the RightsFetcher a TTL cache with the given ttl.
func WithTTLCache(ttl time.Duration) Option {
	return func(r *RightsFetcher) {
		r.applicationsCache = newTTLCache(ttl)
		r.gatewaysCache = newTTLCache(ttl)
	}
}

// New creates a new RightsFetcher with a noop cache.
func New(conn *grpc.ClientConn, opts ...Option) *RightsFetcher {
	client := &RightsFetcher{
		applications:      ttnpb.NewIsApplicationClient(conn),
		applicationsCache: new(noopCache),
		gateways:          ttnpb.NewIsGatewayClient(conn),
		gatewaysCache:     new(noopCache),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// ListApplicationRights returns the rights the caller has to an application.
// Either from the cache or fetching them by making the gRPC call.
func (r *RightsFetcher) ListApplicationRights(ctx context.Context, req *ttnpb.ApplicationIdentifier, creds *rpcmetadata.MD) (*ttnpb.ListApplicationRightsResponse, error) {
	rights, err := r.applicationsCache.GetOrFetch(authorization(creds), req.ApplicationID, func() ([]ttnpb.Right, error) {
		resp, err := r.applications.ListApplicationRights(ctx, req, grpc.PerRPCCredentials(creds))
		if err != nil {
			return nil, err
		}

		return resp.Rights, nil
	})

	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationRightsResponse{
		Rights: rights,
	}, nil
}

// ListGatewayRights returns the rights the caller has to a gateway.
// Either from the cache or fetching them by making the gRPC call.
func (r *RightsFetcher) ListGatewayRights(ctx context.Context, req *ttnpb.GatewayIdentifier, creds *rpcmetadata.MD) (*ttnpb.ListGatewayRightsResponse, error) {
	rights, err := r.gatewaysCache.GetOrFetch(authorization(creds), req.GatewayID, func() ([]ttnpb.Right, error) {
		resp, err := r.gateways.ListGatewayRights(ctx, req, grpc.PerRPCCredentials(creds))
		if err != nil {
			return nil, err
		}

		return resp.Rights, nil
	})

	if err != nil {
		return nil, err
	}

	return &ttnpb.ListGatewayRightsResponse{
		Rights: rights,
	}, nil
}

func authorization(creds *rpcmetadata.MD) string {
	m, _ := creds.GetRequestMetadata(context.Background())
	if m == nil {
		return ""
	}

	return m["authorization"]
}
