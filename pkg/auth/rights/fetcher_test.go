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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func fetchEntityRights(ctx context.Context, id string, f EntityFetcher) (res struct {
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
		res.AppRights, res.AppErr = f.ApplicationRights(ctx, ttnpb.ApplicationIdentifiers{ApplicationId: id})
		wg.Done()
	}()
	go func() {
		res.CliRights, res.CliErr = f.ClientRights(ctx, ttnpb.ClientIdentifiers{ClientId: id})
		wg.Done()
	}()
	go func() {
		res.GtwRights, res.GtwErr = f.GatewayRights(ctx, ttnpb.GatewayIdentifiers{GatewayId: id})
		wg.Done()
	}()
	go func() {
		res.OrgRights, res.OrgErr = f.OrganizationRights(ctx, ttnpb.OrganizationIdentifiers{OrganizationId: id})
		wg.Done()
	}()
	go func() {
		res.UsrRights, res.UsrErr = f.UserRights(ctx, ttnpb.UserIdentifiers{UserId: id})
		wg.Done()
	}()
	wg.Wait()
	return res
}

func fetchAuthInfo(ctx context.Context, f AuthInfoFetcher) (*ttnpb.AuthInfoResponse, error) {
	return f.AuthInfo(ctx)
}

type mockEntityAccessServer struct {
	ttnpb.EntityAccessServer
	*mockFetcher
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
	ttnpb.RegisterEntityAccessServer(srv, &mockEntityAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterApplicationAccessServer(srv, &mockApplicationAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterClientAccessServer(srv, &mockClientAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterGatewayAccessServer(srv, &mockGatewayAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterOrganizationAccessServer(srv, &mockOrganizationAccessServer{mockFetcher: &as.mockFetcher})
	ttnpb.RegisterUserAccessServer(srv, &mockUserAccessServer{mockFetcher: &as.mockFetcher})
	return srv
}

func (as *mockEntityAccessServer) AuthInfo(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.AuthInfoResponse, error) {
	as.authInfoCtx = ctx
	if as.authInfoError != nil {
		return nil, as.authInfoError
	}
	return as.authInfoResponse, nil
}

func (as *mockApplicationAccessServer) ListRights(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	as.applicationCtx, as.applicationIDs = ctx, *ids
	if as.applicationError != nil {
		return nil, as.applicationError
	}
	return as.applicationRights, nil
}

func (as *mockClientAccessServer) ListRights(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	as.clientCtx, as.clientIDs = ctx, *ids
	if as.clientError != nil {
		return nil, as.clientError
	}
	return as.clientRights, nil
}

func (as *mockGatewayAccessServer) ListRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	as.gatewayCtx, as.gatewayIDs = ctx, *ids
	if as.gatewayError != nil {
		return nil, as.gatewayError
	}
	return as.gatewayRights, nil
}

func (as *mockOrganizationAccessServer) ListRights(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	as.organizationCtx, as.organizationIDs = ctx, *ids
	if as.organizationError != nil {
		return nil, as.organizationError
	}
	return as.organizationRights, nil
}

func (as *mockUserAccessServer) ListRights(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	as.userCtx, as.userIDs = ctx, *ids
	if as.userError != nil {
		return nil, as.userError
	}
	return as.userRights, nil
}

func TestEntityFetcherFunc(t *testing.T) {
	a := assertions.New(t)

	var fetcher struct {
		mu     sync.Mutex
		ctx    []context.Context
		ids    []*ttnpb.EntityIdentifiers
		rights *ttnpb.Rights
		err    error
	}
	fetcher.err = errors.New("test err")
	f := EntityFetcherFunc(func(ctx context.Context, ids *ttnpb.EntityIdentifiers) (*ttnpb.Rights, error) {
		fetcher.mu.Lock()
		defer fetcher.mu.Unlock()
		fetcher.ctx = append(fetcher.ctx, ctx)
		fetcher.ids = append(fetcher.ids, ids)
		return fetcher.rights, fetcher.err
	})

	res := fetchEntityRights(test.Context(), "foo", f)
	a.So(res.AppErr, should.Resemble, fetcher.err)
	a.So(res.CliErr, should.Resemble, fetcher.err)
	a.So(res.GtwErr, should.Resemble, fetcher.err)
	a.So(res.OrgErr, should.Resemble, fetcher.err)
	a.So(res.UsrErr, should.Resemble, fetcher.err)

	if a.So(fetcher.ids, should.HaveLength, 5) {
		a.So(fetcher.ids, should.Contain, (&ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}).GetEntityIdentifiers())
		a.So(fetcher.ids, should.Contain, (&ttnpb.ClientIdentifiers{ClientId: "foo"}).GetEntityIdentifiers())
		a.So(fetcher.ids, should.Contain, (&ttnpb.GatewayIdentifiers{GatewayId: "foo"}).GetEntityIdentifiers())
		a.So(fetcher.ids, should.Contain, (&ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}).GetEntityIdentifiers())
		a.So(fetcher.ids, should.Contain, (&ttnpb.UserIdentifiers{UserId: "foo"}).GetEntityIdentifiers())
	}
}

func TestAuthInfoFetcherFunc(t *testing.T) {
	a := assertions.New(t)

	var fetcher struct {
		mu       sync.Mutex
		ctx      []context.Context
		authInfo *ttnpb.AuthInfoResponse
		err      error
	}
	fetcher.err = errors.New("test err")
	f := AuthInfoFetcherFunc(func(ctx context.Context) (*ttnpb.AuthInfoResponse, error) {
		fetcher.mu.Lock()
		defer fetcher.mu.Unlock()
		fetcher.ctx = append(fetcher.ctx, ctx)
		return fetcher.authInfo, fetcher.err
	})

	authInfo, err := fetchAuthInfo(test.Context(), f)
	a.So(err, should.Resemble, fetcher.err)
	a.So(authInfo, should.Resemble, fetcher.authInfo)
}

func TestAccessFetcher(t *testing.T) {
	a := assertions.New(t)

	is := &mockAccessServer{
		mockFetcher: mockFetcher{
			authInfoResponse: &ttnpb.AuthInfoResponse{
				UniversalRights: ttnpb.RightsFrom(ttnpb.Right_RIGHT_SEND_INVITES),
				IsAdmin:         true,
			},
			applicationRights:  ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
			clientRights:       ttnpb.RightsFrom(ttnpb.Right_RIGHT_CLIENT_ALL),
			gatewayRights:      ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_INFO),
			organizationRights: ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
			userRights:         ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO),
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
	_, unavailableAuthInfoErr := fetchAuthInfo(test.Context(), unavailableFetcher)
	a.So(errors.IsUnavailable(unavailableAuthInfoErr), should.BeTrue)
	unavailableEntityRes := fetchEntityRights(test.Context(), "foo", unavailableFetcher)
	a.So(errors.IsUnavailable(unavailableEntityRes.AppErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableEntityRes.CliErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableEntityRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableEntityRes.OrgErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableEntityRes.UsrErr), should.BeTrue)

	onlySecureFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return cc
	}, false)

	_, bgAuthInfoErr := fetchAuthInfo(test.Context(), onlySecureFetcher)
	a.So(errors.IsUnauthenticated(bgAuthInfoErr), should.BeTrue)
	bgEntityRes := fetchEntityRights(test.Context(), "foo", onlySecureFetcher)
	a.So(errors.IsUnauthenticated(bgEntityRes.AppErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgEntityRes.CliErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgEntityRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgEntityRes.OrgErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgEntityRes.UsrErr), should.BeTrue)

	md := metadata.Pairs("authorization", "Bearer token")
	if ctxMd, ok := metadata.FromIncomingContext(test.Context()); ok {
		md = metadata.Join(ctxMd, md)
	}
	authCtx := metadata.NewIncomingContext(test.Context(), md)

	_, authInfoErr := fetchAuthInfo(authCtx, onlySecureFetcher)
	a.So(errors.IsUnauthenticated(authInfoErr), should.BeTrue)
	authEntityRes := fetchEntityRights(authCtx, "foo", onlySecureFetcher)
	a.So(errors.IsUnauthenticated(authEntityRes.AppErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authEntityRes.CliErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authEntityRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authEntityRes.OrgErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authEntityRes.UsrErr), should.BeTrue)

	alsoInsecureFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return cc
	}, true)

	authInfoRes, authInfoErr := fetchAuthInfo(authCtx, alsoInsecureFetcher)
	a.So(authInfoErr, should.BeNil)
	a.So(authInfoRes, should.Resemble, is.mockFetcher.authInfoResponse)

	authEntityRes = fetchEntityRights(authCtx, "foo", alsoInsecureFetcher)
	a.So(authEntityRes.AppErr, should.BeNil)
	a.So(authEntityRes.CliErr, should.BeNil)
	a.So(authEntityRes.GtwErr, should.BeNil)
	a.So(authEntityRes.OrgErr, should.BeNil)
	a.So(authEntityRes.UsrErr, should.BeNil)

	a.So(authEntityRes.AppRights, should.Resemble, is.mockFetcher.applicationRights)
	a.So(authEntityRes.CliRights, should.Resemble, is.mockFetcher.clientRights)
	a.So(authEntityRes.GtwRights, should.Resemble, is.mockFetcher.gatewayRights)
	a.So(authEntityRes.OrgRights, should.Resemble, is.mockFetcher.organizationRights)
	a.So(authEntityRes.UsrRights, should.Resemble, is.mockFetcher.userRights)
}
