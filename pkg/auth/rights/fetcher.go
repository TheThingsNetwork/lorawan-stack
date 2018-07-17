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

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Fetcher interface for rights fetching.
type Fetcher interface {
	ApplicationRights(context.Context, ttnpb.ApplicationIdentifiers) ([]ttnpb.Right, error)
	GatewayRights(context.Context, ttnpb.GatewayIdentifiers) ([]ttnpb.Right, error)
	OrganizationRights(context.Context, ttnpb.OrganizationIdentifiers) ([]ttnpb.Right, error)
}

// FetcherFunc is a function that implements the Fetcher interface.
type FetcherFunc func(ctx context.Context, ids ttnpb.Identifiers) ([]ttnpb.Right, error)

// ApplicationRights implements the Fetcher interface.
func (f FetcherFunc) ApplicationRights(ctx context.Context, ids ttnpb.ApplicationIdentifiers) ([]ttnpb.Right, error) {
	return f(ctx, ids)
}

// GatewayRights implements the Fetcher interface.
func (f FetcherFunc) GatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers) ([]ttnpb.Right, error) {
	return f(ctx, ids)
}

// OrganizationRights implements the Fetcher interface.
func (f FetcherFunc) OrganizationRights(ctx context.Context, ids ttnpb.OrganizationIdentifiers) ([]ttnpb.Right, error) {
	return f(ctx, ids)
}

type fetcherKeyType struct{}

var fetcherKey fetcherKeyType

// NewContextWithFetcher returns a new context with the given rights fetcher.
func NewContextWithFetcher(ctx context.Context, fetcher Fetcher) context.Context {
	return context.WithValue(ctx, fetcherKey, fetcher)
}

func fetcherFromContext(ctx context.Context) (Fetcher, bool) {
	if fetcher, ok := ctx.Value(fetcherKey).(Fetcher); ok {
		return fetcher, true
	}
	return nil, false
}

// NewIdentityServerFetcher returns a new rights fetcher that fetches from the identity server returned by getConn.
// The allowInsecure argument indicates whether it's allowed to send credentials over connections without TLS.
func NewIdentityServerFetcher(getConn func(ctx context.Context) *grpc.ClientConn, allowInsecure bool) Fetcher {
	return &identityServerFetcher{getConn: getConn, allowInsecure: allowInsecure}
}

type identityServerFetcher struct {
	getConn       func(ctx context.Context) *grpc.ClientConn
	allowInsecure bool
}

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "no authentication found in call headers")

var errNoISConn = errors.DefineUnavailable("no_identity_server_conn", "no connection to identity server")

func (f identityServerFetcher) forwardAuth(ctx context.Context) (context.Context, rpcmetadata.MD, error) {
	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType == "" || md.AuthValue == "" {
		return ctx, md, errUnauthenticated
	}
	md.AllowInsecure = f.allowInsecure
	return md.ToOutgoingContext(ctx), md, nil
}

func (f identityServerFetcher) ApplicationRights(ctx context.Context, appID ttnpb.ApplicationIdentifiers) ([]ttnpb.Right, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	ctx, md, err := f.forwardAuth(ctx)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewIsApplicationClient(cc).ListApplicationRights(ctx, &appID, grpc.PerRPCCredentials(md))
	if err != nil {
		return nil, err
	}
	return rights.Rights, nil
}

func (f identityServerFetcher) GatewayRights(ctx context.Context, gtwID ttnpb.GatewayIdentifiers) ([]ttnpb.Right, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	ctx, md, err := f.forwardAuth(ctx)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewIsGatewayClient(cc).ListGatewayRights(ctx, &gtwID, grpc.PerRPCCredentials(md))
	if err != nil {
		return nil, err
	}
	return rights.Rights, nil
}

func (f identityServerFetcher) OrganizationRights(ctx context.Context, orgID ttnpb.OrganizationIdentifiers) ([]ttnpb.Right, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	ctx, md, err := f.forwardAuth(ctx)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewIsOrganizationClient(cc).ListOrganizationRights(ctx, &orgID, grpc.PerRPCCredentials(md))
	if err != nil {
		return nil, err
	}
	return rights.Rights, nil
}
