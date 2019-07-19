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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		res.AppErr = RequireApplication(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: id}, ttnpb.RIGHT_APPLICATION_INFO)
		wg.Done()
	}()
	go func() {
		res.CliErr = RequireClient(ctx, ttnpb.ClientIdentifiers{ClientID: id}, ttnpb.RIGHT_CLIENT_ALL)
		wg.Done()
	}()
	go func() {
		res.GtwErr = RequireGateway(ctx, ttnpb.GatewayIdentifiers{GatewayID: id}, ttnpb.RIGHT_GATEWAY_INFO)
		wg.Done()
	}()
	go func() {
		res.OrgErr = RequireOrganization(ctx, ttnpb.OrganizationIdentifiers{OrganizationID: id}, ttnpb.RIGHT_ORGANIZATION_INFO)
		wg.Done()
	}()
	go func() {
		res.UsrErr = RequireUser(ctx, ttnpb.UserIdentifiers{UserID: id}, ttnpb.RIGHT_USER_INFO)
		wg.Done()
	}()
	wg.Wait()
	return
}

func TestRequire(t *testing.T) {
	a := assertions.New(t)

	a.So(func() {
		RequireApplication(test.Context(), ttnpb.ApplicationIdentifiers{}, ttnpb.RIGHT_APPLICATION_INFO)
	}, should.Panic)
	a.So(func() {
		RequireClient(test.Context(), ttnpb.ClientIdentifiers{}, ttnpb.RIGHT_CLIENT_ALL)
	}, should.Panic)
	a.So(func() {
		RequireGateway(test.Context(), ttnpb.GatewayIdentifiers{}, ttnpb.RIGHT_GATEWAY_INFO)
	}, should.Panic)
	a.So(func() {
		RequireOrganization(test.Context(), ttnpb.OrganizationIdentifiers{}, ttnpb.RIGHT_ORGANIZATION_INFO)
	}, should.Panic)
	a.So(func() {
		RequireUser(test.Context(), ttnpb.UserIdentifiers{}, ttnpb.RIGHT_USER_INFO)
	}, should.Panic)

	fooCtx := test.Context()
	fooCtx = NewContext(fooCtx, Rights{
		ApplicationRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}): ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO),
		},
		ClientRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.ClientIdentifiers{ClientID: "foo"}): ttnpb.RightsFrom(ttnpb.RIGHT_CLIENT_ALL),
		},
		GatewayRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.GatewayIdentifiers{GatewayID: "foo"}): ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_INFO),
		},
		OrganizationRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}): ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_INFO),
		},
		UserRights: map[string]*ttnpb.Rights{
			unique.ID(fooCtx, ttnpb.UserIdentifiers{UserID: "foo"}): ttnpb.RightsFrom(ttnpb.RIGHT_USER_INFO),
		},
	})

	fooRes := requireRights(fooCtx, "foo")
	a.So(fooRes.AppErr, should.BeNil)
	a.So(fooRes.CliErr, should.BeNil)
	a.So(fooRes.GtwErr, should.BeNil)
	a.So(fooRes.OrgErr, should.BeNil)
	a.So(fooRes.UsrErr, should.BeNil)

	mockErr := errors.New("mock")
	errFetchCtx := NewContextWithFetcher(test.Context(), &mockFetcher{
		applicationError:  mockErr,
		clientError:       mockErr,
		gatewayError:      mockErr,
		organizationError: mockErr,
		userError:         mockErr,
	})
	errFetchRes := requireRights(errFetchCtx, "foo")
	a.So(errFetchRes.AppErr, should.Resemble, mockErr)
	a.So(errFetchRes.CliErr, should.Resemble, mockErr)
	a.So(errFetchRes.GtwErr, should.Resemble, mockErr)
	a.So(errFetchRes.OrgErr, should.Resemble, mockErr)
	a.So(errFetchRes.UsrErr, should.Resemble, mockErr)

	errPermissionDenied := status.New(codes.PermissionDenied, "permission denied").Err()
	permissionDeniedFetchCtx := NewContextWithFetcher(test.Context(), &mockFetcher{
		applicationError:  errPermissionDenied,
		clientError:       errPermissionDenied,
		gatewayError:      errPermissionDenied,
		organizationError: errPermissionDenied,
		userError:         errPermissionDenied,
	})
	permissionDeniedRes := requireRights(permissionDeniedFetchCtx, "foo")
	a.So(errors.IsPermissionDenied(permissionDeniedRes.AppErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedRes.CliErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedRes.GtwErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedRes.OrgErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedRes.UsrErr), should.BeTrue)
}
