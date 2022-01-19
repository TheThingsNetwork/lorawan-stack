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
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requireAuthInfo(ctx context.Context) (res struct {
	UniversalErr error
	IsAdminErr   error
}) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		res.UniversalErr = RequireUniversal(ctx, ttnpb.Right_RIGHT_SEND_INVITES)
		wg.Done()
	}()
	go func() {
		res.IsAdminErr = RequireIsAdmin(ctx)
		wg.Done()
	}()
	wg.Wait()
	return res
}

func requireRights(ctx context.Context, id string) (res struct {
	AppErr error
	CliErr error
	GtwErr error
	OrgErr error
	UsrErr error
}) {
	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		res.AppErr = RequireApplication(ctx, ttnpb.ApplicationIdentifiers{ApplicationId: id}, ttnpb.Right_RIGHT_APPLICATION_INFO)
		wg.Done()
	}()
	go func() {
		res.CliErr = RequireClient(ctx, ttnpb.ClientIdentifiers{ClientId: id}, ttnpb.Right_RIGHT_CLIENT_ALL)
		wg.Done()
	}()
	go func() {
		res.GtwErr = RequireGateway(ctx, ttnpb.GatewayIdentifiers{GatewayId: id}, ttnpb.Right_RIGHT_GATEWAY_INFO)
		wg.Done()
	}()
	go func() {
		res.OrgErr = RequireOrganization(ctx, ttnpb.OrganizationIdentifiers{OrganizationId: id}, ttnpb.Right_RIGHT_ORGANIZATION_INFO)
		wg.Done()
	}()
	go func() {
		res.UsrErr = RequireUser(ctx, ttnpb.UserIdentifiers{UserId: id}, ttnpb.Right_RIGHT_USER_INFO)
		wg.Done()
	}()
	wg.Wait()
	return res
}

func TestRequire(t *testing.T) {
	a := assertions.New(t)

	a.So(func() {
		RequireUniversal(test.Context(), ttnpb.Right_RIGHT_SEND_INVITES)
	}, should.Panic)
	a.So(func() {
		RequireIsAdmin(test.Context())
	}, should.Panic)
	a.So(func() {
		RequireApplication(test.Context(), ttnpb.ApplicationIdentifiers{}, ttnpb.Right_RIGHT_APPLICATION_INFO)
	}, should.Panic)
	a.So(func() {
		RequireClient(test.Context(), ttnpb.ClientIdentifiers{}, ttnpb.Right_RIGHT_CLIENT_ALL)
	}, should.Panic)
	a.So(func() {
		RequireGateway(test.Context(), ttnpb.GatewayIdentifiers{}, ttnpb.Right_RIGHT_GATEWAY_INFO)
	}, should.Panic)
	a.So(func() {
		RequireOrganization(test.Context(), ttnpb.OrganizationIdentifiers{}, ttnpb.Right_RIGHT_ORGANIZATION_INFO)
	}, should.Panic)
	a.So(func() {
		RequireUser(test.Context(), ttnpb.UserIdentifiers{}, ttnpb.Right_RIGHT_USER_INFO)
	}, should.Panic)

	fooCtx := test.Context()
	fooCtx = NewContext(fooCtx, Rights{
		ApplicationRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
		},
		ClientRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.ClientIdentifiers{ClientId: "foo"}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_CLIENT_ALL),
		},
		GatewayRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.GatewayIdentifiers{GatewayId: "foo"}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_INFO),
		},
		OrganizationRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		},
		UserRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.UserIdentifiers{UserId: "foo"}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO),
		},
	})
	fooCtx = NewContextWithAuthInfo(fooCtx, &ttnpb.AuthInfoResponse{
		UniversalRights: ttnpb.RightsFrom(ttnpb.Right_RIGHT_SEND_INVITES),
		IsAdmin:         true,
	})

	fooAuthInfoRes := requireAuthInfo(fooCtx)
	a.So(fooAuthInfoRes.UniversalErr, should.BeNil)
	a.So(fooAuthInfoRes.IsAdminErr, should.BeNil)
	fooEntityRes := requireRights(fooCtx, "foo")
	a.So(fooEntityRes.AppErr, should.BeNil)
	a.So(fooEntityRes.CliErr, should.BeNil)
	a.So(fooEntityRes.GtwErr, should.BeNil)
	a.So(fooEntityRes.OrgErr, should.BeNil)
	a.So(fooEntityRes.UsrErr, should.BeNil)

	mockErr := errors.New("mock")
	errFetchCtx := NewContextWithFetcher(test.Context(), &mockFetcher{
		authInfoError:     mockErr,
		applicationError:  mockErr,
		clientError:       mockErr,
		gatewayError:      mockErr,
		organizationError: mockErr,
		userError:         mockErr,
	})
	errFetchAuthInfoRes := requireAuthInfo(errFetchCtx)
	a.So(errFetchAuthInfoRes.UniversalErr, should.Resemble, mockErr)
	a.So(errFetchAuthInfoRes.IsAdminErr, should.Resemble, mockErr)
	errFetchEntityRes := requireRights(errFetchCtx, "foo")
	a.So(errFetchEntityRes.AppErr, should.Resemble, mockErr)
	a.So(errFetchEntityRes.CliErr, should.Resemble, mockErr)
	a.So(errFetchEntityRes.GtwErr, should.Resemble, mockErr)
	a.So(errFetchEntityRes.OrgErr, should.Resemble, mockErr)
	a.So(errFetchEntityRes.UsrErr, should.Resemble, mockErr)

	errPermissionDenied := status.New(codes.PermissionDenied, "permission denied").Err()
	permissionDeniedFetchCtx := NewContextWithFetcher(test.Context(), &mockFetcher{
		authInfoError:     errPermissionDenied,
		applicationError:  errPermissionDenied,
		clientError:       errPermissionDenied,
		gatewayError:      errPermissionDenied,
		organizationError: errPermissionDenied,
		userError:         errPermissionDenied,
	})
	permissionDeniedAuthInfoRes := requireAuthInfo(permissionDeniedFetchCtx)
	a.So(errors.IsPermissionDenied(permissionDeniedAuthInfoRes.UniversalErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedAuthInfoRes.IsAdminErr), should.BeTrue)
	permissionDeniedEntityRes := requireRights(permissionDeniedFetchCtx, "foo")
	a.So(errors.IsPermissionDenied(permissionDeniedEntityRes.AppErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedEntityRes.CliErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedEntityRes.GtwErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedEntityRes.OrgErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedEntityRes.UsrErr), should.BeTrue)
}
