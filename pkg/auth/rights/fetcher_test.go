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
	"net"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func fetchRights(ctx context.Context, id string, f Fetcher) (res struct {
	AppRights *ttnpb.Rights
	AppErr    error
	CliRights *ttnpb.Rights
	CliErr    error
	GtwRights *ttnpb.Rights
	GtwErr    error
	OrgRights *ttnpb.Rights
	OrgErr    error
	UsrRights *ttnpb.Rights
	UsrErr    error
}) {
	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		res.AppRights, res.AppErr = f.ApplicationRights(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: id})
		wg.Done()
	}()
	go func() {
		res.CliRights, res.CliErr = f.ClientRights(ctx, ttnpb.ClientIdentifiers{ClientID: id})
		wg.Done()
	}()
	go func() {
		res.GtwRights, res.GtwErr = f.GatewayRights(ctx, ttnpb.GatewayIdentifiers{GatewayID: id})
		wg.Done()
	}()
	go func() {
		res.OrgRights, res.OrgErr = f.OrganizationRights(ctx, ttnpb.OrganizationIdentifiers{OrganizationID: id})
		wg.Done()
	}()
	go func() {
		res.UsrRights, res.UsrErr = f.UserRights(ctx, ttnpb.UserIdentifiers{UserID: id})
		wg.Done()
	}()
	wg.Wait()
	return
}

type mockApplicationAccessServer struct {
	ttnpb.ApplicationAccessServer
	*mockFetcher
}
type mockClientAccessServer struct {
	ttnpb.ClientAccessServer
	*mockFetcher
}
type mockGatewayAccessServer struct {
	ttnpb.GatewayAccessServer
	*mockFetcher
}
type mockOrganizationAccessServer struct {
	ttnpb.OrganizationAccessServer
	*mockFetcher
}
type mockUserAccessServer struct {
	ttnpb.UserAccessServer
	*mockFetcher
}

type mockAccessServer struct {
	mockFetcher
}

func (as *mockAccessServer) Server() *grpc.Server {
	srv := grpc.NewServer()
	ttnpb.RegisterApplicationAccessServer(srv, mockApplicationAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterClientAccessServer(srv, mockClientAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterGatewayAccessServer(srv, mockGatewayAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterOrganizationAccessServer(srv, mockOrganizationAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterUserAccessServer(srv, mockUserAccessServer{mockFetcher: &as.mockFetcher})
	return srv
}

func (as mockApplicationAccessServer) ListRights(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	as.applicationCtx, as.applicationIDs = ctx, *ids
	if as.applicationError != nil {
		return nil, as.applicationError
	}
	return as.applicationRights, nil
}

func (as mockClientAccessServer) ListRights(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	as.clientCtx, as.clientIDs = ctx, *ids
	if as.clientError != nil {
		return nil, as.clientError
	}
	return as.clientRights, nil
}

func (as mockGatewayAccessServer) ListRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	as.gatewayCtx, as.gatewayIDs = ctx, *ids
	if as.gatewayError != nil {
		return nil, as.gatewayError
	}
	return as.gatewayRights, nil
}

func (as mockOrganizationAccessServer) ListRights(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	as.organizationCtx, as.organizationIDs = ctx, *ids
	if as.organizationError != nil {
		return nil, as.organizationError
	}
	return as.organizationRights, nil
}

func (as mockUserAccessServer) ListRights(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	as.userCtx, as.userIDs = ctx, *ids
	if as.userError != nil {
		return nil, as.userError
	}
	return as.userRights, nil
}

func TestFetcherFunc(t *testing.T) {
	a := assertions.New(t)

	var fetcher struct {
		mu     sync.Mutex
		ctx    []context.Context
		ids    []ttnpb.Identifiers
		rights *ttnpb.Rights
		err    error
	}
	fetcher.err = errors.New("test err")
	f := FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
		fetcher.mu.Lock()
		defer fetcher.mu.Unlock()
		fetcher.ctx = append(fetcher.ctx, ctx)
		fetcher.ids = append(fetcher.ids, ids)
		return fetcher.rights, fetcher.err
	})

	res := fetchRights(test.Context(), "foo", f)
	a.So(res.AppErr, should.Resemble, fetcher.err)
	a.So(res.CliErr, should.Resemble, fetcher.err)
	a.So(res.GtwErr, should.Resemble, fetcher.err)
	a.So(res.OrgErr, should.Resemble, fetcher.err)
	a.So(res.UsrErr, should.Resemble, fetcher.err)

	if a.So(fetcher.ids, should.HaveLength, 5) {
		a.So(fetcher.ids, should.Contain, ttnpb.ApplicationIdentifiers{ApplicationID: "foo"})
		a.So(fetcher.ids, should.Contain, ttnpb.ClientIdentifiers{ClientID: "foo"})
		a.So(fetcher.ids, should.Contain, ttnpb.GatewayIdentifiers{GatewayID: "foo"})
		a.So(fetcher.ids, should.Contain, ttnpb.OrganizationIdentifiers{OrganizationID: "foo"})
		a.So(fetcher.ids, should.Contain, ttnpb.UserIdentifiers{UserID: "foo"})
	}
}

func TestAccessFetcher(t *testing.T) {
	a := assertions.New(t)

	is := &mockAccessServer{
		mockFetcher: mockFetcher{
			applicationRights:  ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO),
			clientRights:       ttnpb.RightsFrom(ttnpb.RIGHT_CLIENT_ALL),
			gatewayRights:      ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_INFO),
			organizationRights: ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_INFO),
			userRights:         ttnpb.RightsFrom(ttnpb.RIGHT_USER_INFO),
		},
	}
	srv := is.Server()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)

	cc, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}

	// Identity Server not available, return Unavailable error.
	unavailableFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return nil
	}, false)
	unavailableRes := fetchRights(test.Context(), "foo", unavailableFetcher)
	a.So(errors.IsUnavailable(unavailableRes.AppErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableRes.CliErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableRes.OrgErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableRes.UsrErr), should.BeTrue)

	onlySecureFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return cc
	}, false)

	bgRes := fetchRights(test.Context(), "foo", onlySecureFetcher)
	a.So(errors.IsUnauthenticated(bgRes.AppErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgRes.CliErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgRes.OrgErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgRes.UsrErr), should.BeTrue)

	md := metadata.Pairs("authorization", "Bearer token")
	if ctxMd, ok := metadata.FromIncomingContext(test.Context()); ok {
		md = metadata.Join(ctxMd, md)
	}
	authCtx := metadata.NewIncomingContext(test.Context(), md)

	authRes := fetchRights(authCtx, "foo", onlySecureFetcher)
	a.So(errors.IsUnauthenticated(authRes.AppErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authRes.CliErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authRes.OrgErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authRes.UsrErr), should.BeTrue)

	alsoInsecureFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return cc
	}, true)

	authRes = fetchRights(authCtx, "foo", alsoInsecureFetcher)
	a.So(authRes.AppErr, should.BeNil)
	a.So(authRes.CliErr, should.BeNil)
	a.So(authRes.GtwErr, should.BeNil)
	a.So(authRes.OrgErr, should.BeNil)
	a.So(authRes.UsrErr, should.BeNil)

	a.So(authRes.AppRights, should.Resemble, is.mockFetcher.applicationRights)
	a.So(authRes.CliRights, should.Resemble, is.mockFetcher.clientRights)
	a.So(authRes.GtwRights, should.Resemble, is.mockFetcher.gatewayRights)
	a.So(authRes.OrgRights, should.Resemble, is.mockFetcher.organizationRights)
	a.So(authRes.UsrRights, should.Resemble, is.mockFetcher.userRights)
}
