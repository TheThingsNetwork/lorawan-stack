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
	"net"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func fetchRights(ctx context.Context, id string, f Fetcher) (res struct {
	AppRights *ttnpb.Rights
	AppErr    error
	GtwRights *ttnpb.Rights
	GtwErr    error
	OrgRights *ttnpb.Rights
	OrgErr    error
}) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		res.AppRights, res.AppErr = f.ApplicationRights(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: id})
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
	wg.Wait()
	return
}

type mockIdentityServer struct {
	ttnpb.ApplicationAccessServer
	ttnpb.GatewayAccessServer
	ttnpb.OrganizationAccessServer
	mockFetcher
}

func (is *mockIdentityServer) Server() *grpc.Server {
	srv := grpc.NewServer()
	ttnpb.RegisterApplicationAccessServer(srv, is)
	ttnpb.RegisterGatewayAccessServer(srv, is)
	ttnpb.RegisterOrganizationAccessServer(srv, is)
	return srv
}

func (is *mockIdentityServer) ListApplicationRights(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	is.applicationCtx, is.applicationIDs = ctx, *ids
	if is.applicationError != nil {
		return nil, is.applicationError
	}
	return is.applicationRights, nil
}

func (is *mockIdentityServer) ListGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	is.gatewayCtx, is.gatewayIDs = ctx, *ids
	if is.gatewayError != nil {
		return nil, is.gatewayError
	}
	return is.gatewayRights, nil
}

func (is *mockIdentityServer) ListOrganizationRights(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	is.organizationCtx, is.organizationIDs = ctx, *ids
	if is.organizationError != nil {
		return nil, is.organizationError
	}
	return is.organizationRights, nil
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
	a.So(res.GtwErr, should.Resemble, fetcher.err)
	a.So(res.OrgErr, should.Resemble, fetcher.err)

	a.So(fetcher.ids, should.HaveLength, 3)
	a.So(ttnpb.CombineIdentifiers(fetcher.ids...), should.Resemble, &ttnpb.CombinedIdentifiers{
		ApplicationIDs:  []*ttnpb.ApplicationIdentifiers{{ApplicationID: "foo"}},
		GatewayIDs:      []*ttnpb.GatewayIdentifiers{{GatewayID: "foo"}},
		OrganizationIDs: []*ttnpb.OrganizationIdentifiers{{OrganizationID: "foo"}},
	})
}

func TestAccessFetcher(t *testing.T) {
	a := assertions.New(t)

	is := &mockIdentityServer{
		mockFetcher: mockFetcher{
			applicationRights:  ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO),
			gatewayRights:      ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_INFO),
			organizationRights: ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_INFO),
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
	a.So(errors.IsUnavailable(unavailableRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnavailable(unavailableRes.OrgErr), should.BeTrue)

	onlySecureFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return cc
	}, false)

	bgRes := fetchRights(test.Context(), "foo", onlySecureFetcher)
	a.So(errors.IsUnauthenticated(bgRes.AppErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(bgRes.OrgErr), should.BeTrue)

	md := metadata.Pairs("authorization", "Bearer token")
	if ctxMd, ok := metadata.FromIncomingContext(test.Context()); ok {
		md = metadata.Join(ctxMd, md)
	}
	authCtx := metadata.NewIncomingContext(test.Context(), md)

	authRes := fetchRights(authCtx, "foo", onlySecureFetcher)
	a.So(errors.IsUnauthenticated(authRes.AppErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authRes.GtwErr), should.BeTrue)
	a.So(errors.IsUnauthenticated(authRes.OrgErr), should.BeTrue)

	alsoInsecureFetcher := NewAccessFetcher(func(context.Context) *grpc.ClientConn {
		return cc
	}, true)

	authRes = fetchRights(authCtx, "foo", alsoInsecureFetcher)
	a.So(authRes.AppErr, should.BeNil)
	a.So(authRes.GtwErr, should.BeNil)
	a.So(authRes.OrgErr, should.BeNil)

	a.So(authRes.AppRights, should.Resemble, is.mockFetcher.applicationRights)
	a.So(authRes.GtwRights, should.Resemble, is.mockFetcher.gatewayRights)
	a.So(authRes.OrgRights, should.Resemble, is.mockFetcher.organizationRights)
}
