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

package rights

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Fetcher interface for rights fetching.
type Fetcher interface {
	ApplicationRights(context.Context, ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error)
	ClientRights(context.Context, ttnpb.ClientIdentifiers) (*ttnpb.Rights, error)
	GatewayRights(context.Context, ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error)
	OrganizationRights(context.Context, ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error)
	UserRights(context.Context, ttnpb.UserIdentifiers) (*ttnpb.Rights, error)
}

// FetcherFunc is a function that implements the Fetcher interface.
//
// A FetcherFunc that returns all Application rights for any Application,
// would look like this:
//
//    fetcher := rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
//    	rights := ttnpb.AllApplicationRights // Instead this usually comes from an identity server or a database.
//    	return &rights, nil
//    })
//
type FetcherFunc func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error)

// ApplicationRights implements the Fetcher interface.
func (f FetcherFunc) ApplicationRights(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	rights, err := f(ctx, ids)
	registerRightsFetch(ctx, "application", rights, err)
	return rights, err
}

// ClientRights implements the Fetcher interface.
func (f FetcherFunc) ClientRights(ctx context.Context, ids ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	rights, err := f(ctx, ids)
	registerRightsFetch(ctx, "client", rights, err)
	return rights, err
}

// GatewayRights implements the Fetcher interface.
func (f FetcherFunc) GatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	rights, err := f(ctx, ids)
	registerRightsFetch(ctx, "gateway", rights, err)
	return rights, err
}

// OrganizationRights implements the Fetcher interface.
func (f FetcherFunc) OrganizationRights(ctx context.Context, ids ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	rights, err := f(ctx, ids)
	registerRightsFetch(ctx, "organization", rights, err)
	return rights, err
}

// UserRights implements the Fetcher interface.
func (f FetcherFunc) UserRights(ctx context.Context, ids ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	rights, err := f(ctx, ids)
	registerRightsFetch(ctx, "user", rights, err)
	return rights, err
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

var errNoFetcher = errors.DefineInternal("no_fetcher", "no fetcher found in context")

// NewAccessFetcher returns a new rights fetcher that fetches from the Access role returned by getConn.
// The allowInsecure argument indicates whether it's allowed to send credentials over connections without TLS.
func NewAccessFetcher(getConn func(ctx context.Context) *grpc.ClientConn, allowInsecure bool) Fetcher {
	return &accessFetcher{getConn: getConn, allowInsecure: allowInsecure}
}

type accessFetcher struct {
	getConn       func(ctx context.Context) *grpc.ClientConn
	allowInsecure bool
}

var errNoISConn = errors.DefineUnavailable("no_identity_server_conn", "no connection to Identity Server")

func (f accessFetcher) ApplicationRights(ctx context.Context, appID ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, f.allowInsecure)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewApplicationAccessClient(cc).ListRights(ctx, &appID, callOpt)
	registerRightsFetch(ctx, "application", rights, err)
	if err != nil {
		return nil, err
	}
	return rights, nil
}

func (f accessFetcher) ClientRights(ctx context.Context, clientID ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, f.allowInsecure)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewClientAccessClient(cc).ListRights(ctx, &clientID, callOpt)
	registerRightsFetch(ctx, "client", rights, err)
	if err != nil {
		return nil, err
	}
	return rights, nil
}

func (f accessFetcher) GatewayRights(ctx context.Context, gtwID ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, f.allowInsecure)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewGatewayAccessClient(cc).ListRights(ctx, &gtwID, callOpt)
	registerRightsFetch(ctx, "gateway", rights, err)
	if err != nil {
		return nil, err
	}
	return rights, nil
}

func (f accessFetcher) OrganizationRights(ctx context.Context, orgID ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, f.allowInsecure)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewOrganizationAccessClient(cc).ListRights(ctx, &orgID, callOpt)
	registerRightsFetch(ctx, "organization", rights, err)
	if err != nil {
		return nil, err
	}
	return rights, nil
}

func (f accessFetcher) UserRights(ctx context.Context, userID ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	cc := f.getConn(ctx)
	if cc == nil {
		return nil, errNoISConn
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, f.allowInsecure)
	if err != nil {
		return nil, err
	}
	rights, err := ttnpb.NewUserAccessClient(cc).ListRights(ctx, &userID, callOpt)
	registerRightsFetch(ctx, "user", rights, err)
	if err != nil {
		return nil, err
	}
	return rights, nil
}
