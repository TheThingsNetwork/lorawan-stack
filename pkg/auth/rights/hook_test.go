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
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
		"Panic:NoIdentifiers": {
			Fetcher:     &mockFetcher{},
			ShouldPanic: true,
		},
		"Application:ISUnavailable": {
			Fetcher: &mockFetcher{
				applicationError: mockErr,
			},
			Req:      &ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
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
				gatewayError:      errPermissionDenied,
				organizationError: errPermissionDenied,
			},
			Req: &ttnpb.CombinedIdentifiers{
				ApplicationIDs:  []*ttnpb.ApplicationIdentifiers{{ApplicationID: "foo"}},
				GatewayIDs:      []*ttnpb.GatewayIdentifiers{{GatewayID: "foo"}},
				OrganizationIDs: []*ttnpb.OrganizationIdentifiers{{OrganizationID: "foo"}},
			},
			Call: func(a *assertions.Assertion, ctx context.Context, req interface{}) {
				rights, ok := FromContext(ctx)
				a.So(ok, should.BeTrue)
				if a.So(rights.ApplicationRights, should.HaveLength, 1) {
					a.So(rights.ApplicationRights[ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}], should.BeEmpty)
				}
				if a.So(rights.GatewayRights, should.HaveLength, 1) {
					a.So(rights.GatewayRights[ttnpb.GatewayIdentifiers{GatewayID: "foo"}], should.BeEmpty)
				}
				if a.So(rights.OrganizationRights, should.HaveLength, 1) {
					a.So(rights.OrganizationRights[ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}], should.BeEmpty)
				}
			},
		},
		"AllInfoRights": {
			Fetcher: &mockFetcher{
				applicationRights:  []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
				gatewayRights:      []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO},
				organizationRights: []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO},
			},
			Req: &ttnpb.CombinedIdentifiers{
				ApplicationIDs:  []*ttnpb.ApplicationIdentifiers{{ApplicationID: "foo"}},
				GatewayIDs:      []*ttnpb.GatewayIdentifiers{{GatewayID: "foo"}},
				OrganizationIDs: []*ttnpb.OrganizationIdentifiers{{OrganizationID: "foo"}},
			},
			Call: func(a *assertions.Assertion, ctx context.Context, req interface{}) {
				rights, ok := FromContext(ctx)
				a.So(ok, should.BeTrue)
				if a.So(rights.ApplicationRights, should.HaveLength, 1) {
					a.So(rights.IncludesApplicationRights(ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}, ttnpb.RIGHT_APPLICATION_INFO), should.BeTrue)
				}
				if a.So(rights.GatewayRights, should.HaveLength, 1) {
					a.So(rights.IncludesGatewayRights(ttnpb.GatewayIdentifiers{GatewayID: "foo"}, ttnpb.RIGHT_GATEWAY_INFO), should.BeTrue)
				}
				if a.So(rights.OrganizationRights, should.HaveLength, 1) {
					a.So(rights.IncludesOrganizationRights(ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}, ttnpb.RIGHT_ORGANIZATION_INFO), should.BeTrue)
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
