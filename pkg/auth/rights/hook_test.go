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

func TestHook(t *testing.T) {
	mockErr := errors.New("mock")
	errPermissionDenied := status.New(codes.PermissionDenied, "permission denied").Err()

	for name, tt := range map[string]struct {
		Fetcher     *mockFetcher
		Req         interface{}
		ShouldPanic bool
		Call        func(a *assertions.Assertion, ctx context.Context, req interface{})
		ErrCheck    func(error) bool
	}{
		"Panic:NoFetcher": {
			ShouldPanic: true,
		},
		"Application:ISUnavailable": {
			Fetcher: &mockFetcher{
				applicationError: mockErr,
			},
			Req:      &ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
			ErrCheck: func(err error) bool { return err.Error() == mockErr.Error() },
		},
		"Client:ISUnavailable": {
			Fetcher: &mockFetcher{
				clientError: mockErr,
			},
			Req:      &ttnpb.ClientIdentifiers{ClientID: "foo"},
			ErrCheck: func(err error) bool { return err.Error() == mockErr.Error() },
		},
		"Gateway:ISUnavailable": {
			Fetcher: &mockFetcher{
				gatewayError: mockErr,
			},
			Req:      &ttnpb.GatewayIdentifiers{GatewayID: "foo"},
			ErrCheck: func(err error) bool { return err.Error() == mockErr.Error() },
		},
		"Organization:ISUnavailable": {
			Fetcher: &mockFetcher{
				organizationError: mockErr,
			},
			Req:      &ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
			ErrCheck: func(err error) bool { return err.Error() == mockErr.Error() },
		},
		"AllPermissionDenied,EmptyRights": {
			Fetcher: &mockFetcher{
				applicationError:  errPermissionDenied,
				clientError:       errPermissionDenied,
				gatewayError:      errPermissionDenied,
				organizationError: errPermissionDenied,
				userError:         errPermissionDenied,
			},
			Req: &ttnpb.CombinedIdentifiers{
				EntityIdentifiers: []*ttnpb.EntityIdentifiers{
					ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}.EntityIdentifiers(),
					ttnpb.ClientIdentifiers{ClientID: "foo"}.EntityIdentifiers(),
					ttnpb.GatewayIdentifiers{GatewayID: "foo"}.EntityIdentifiers(),
					ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}.EntityIdentifiers(),
					ttnpb.UserIdentifiers{UserID: "foo"}.EntityIdentifiers(),
				},
			},
			Call: func(a *assertions.Assertion, ctx context.Context, req interface{}) {
				rights, ok := FromContext(ctx)
				a.So(ok, should.BeTrue)
				a.So(rights.ApplicationRights[unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "foo"})], should.BeNil)
				a.So(rights.ClientRights[unique.ID(ctx, ttnpb.ClientIdentifiers{ClientID: "foo"})], should.BeNil)
				a.So(rights.GatewayRights[unique.ID(ctx, ttnpb.GatewayIdentifiers{GatewayID: "foo"})], should.BeNil)
				a.So(rights.OrganizationRights[unique.ID(ctx, ttnpb.OrganizationIdentifiers{OrganizationID: "foo"})], should.BeNil)
				a.So(rights.UserRights[unique.ID(ctx, ttnpb.UserIdentifiers{UserID: "foo"})], should.BeNil)
			},
		},
		"AllInfoRights": {
			Fetcher: &mockFetcher{
				applicationRights:  ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO),
				clientRights:       ttnpb.RightsFrom(ttnpb.RIGHT_CLIENT_ALL),
				gatewayRights:      ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_INFO),
				organizationRights: ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_INFO),
				userRights:         ttnpb.RightsFrom(ttnpb.RIGHT_USER_INFO),
			},
			Req: &ttnpb.CombinedIdentifiers{
				EntityIdentifiers: []*ttnpb.EntityIdentifiers{
					ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}.EntityIdentifiers(),
					ttnpb.ClientIdentifiers{ClientID: "foo"}.EntityIdentifiers(),
					ttnpb.GatewayIdentifiers{GatewayID: "foo"}.EntityIdentifiers(),
					ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}.EntityIdentifiers(),
					ttnpb.UserIdentifiers{UserID: "foo"}.EntityIdentifiers(),
				},
			},
			Call: func(a *assertions.Assertion, ctx context.Context, req interface{}) {
				rights, ok := FromContext(ctx)
				a.So(ok, should.BeTrue)
				if a.So(rights.ApplicationRights, should.HaveLength, 1) {
					a.So(rights.IncludesApplicationRights(unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}), ttnpb.RIGHT_APPLICATION_INFO), should.BeTrue)
				}
				if a.So(rights.ClientRights, should.HaveLength, 1) {
					a.So(rights.IncludesClientRights(unique.ID(ctx, ttnpb.ClientIdentifiers{ClientID: "foo"}), ttnpb.RIGHT_CLIENT_ALL), should.BeTrue)
				}
				if a.So(rights.GatewayRights, should.HaveLength, 1) {
					a.So(rights.IncludesGatewayRights(unique.ID(ctx, ttnpb.GatewayIdentifiers{GatewayID: "foo"}), ttnpb.RIGHT_GATEWAY_INFO), should.BeTrue)
				}
				if a.So(rights.OrganizationRights, should.HaveLength, 1) {
					a.So(rights.IncludesOrganizationRights(unique.ID(ctx, ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}), ttnpb.RIGHT_ORGANIZATION_INFO), should.BeTrue)
				}
				if a.So(rights.UserRights, should.HaveLength, 1) {
					a.So(rights.IncludesUserRights(unique.ID(ctx, ttnpb.UserIdentifiers{UserID: "foo"}), ttnpb.RIGHT_USER_INFO), should.BeTrue)
				}
			},
		},
		"EndDeviceRights": {
			Fetcher: &mockFetcher{
				applicationRights: ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO),
			},
			Req: &ttnpb.CombinedIdentifiers{
				EntityIdentifiers: []*ttnpb.EntityIdentifiers{
					ttnpb.EndDeviceIdentifiers{ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}}.EntityIdentifiers(),
				},
			},
			Call: func(a *assertions.Assertion, ctx context.Context, req interface{}) {
				rights, ok := FromContext(ctx)
				a.So(ok, should.BeTrue)
				if a.So(rights.ApplicationRights, should.HaveLength, 1) {
					a.So(rights.IncludesApplicationRights(unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}), ttnpb.RIGHT_APPLICATION_INFO), should.BeTrue)
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			if tt.Fetcher != nil {
				ctx = NewContextWithFetcher(ctx, tt.Fetcher)
			}

			inner := &mockHandler{}
			if tt.Call != nil {
				inner.call = func(ctx context.Context, req interface{}) {
					tt.Call(a, ctx, req)
				}
			}

			handler := Hook(inner.Handler)

			if tt.ShouldPanic {
				a.So(func() {
					handler(ctx, tt.Req)
				}, should.Panic)
				return
			}

			_, err := handler(ctx, tt.Req)
			if tt.ErrCheck == nil {
				a.So(err, should.BeNil)
			} else {
				a.So(tt.ErrCheck(err), should.BeTrue)
			}

		})
	}

}
