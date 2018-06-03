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

package identityserver

import (
	"context"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEnforceUserRights(t *testing.T) {
	is := newTestIS(t)

	for i, tc := range []struct {
		authorizationData *authorizationData
		rights            []ttnpb.Right
		sucesss           bool
	}{
		{
			// Fails because the authorization data identifiers are not UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				Source: auth.Token,
				Rights: []ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST},
			},
			[]ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST},
			false,
		},
		{
			// Fails because the authorization data does not include the requested right.
			&authorizationData{
				EntityIdentifiers: ttnpb.UserIdentifiers{
					UserID: "alice",
				},
				Source: auth.Token,
				Rights: []ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST},
			},
			[]ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_CREATE},
			false,
		},
		{
			&authorizationData{
				EntityIdentifiers: ttnpb.UserIdentifiers{
					UserID: "alice",
				},
				Source: auth.Token,
				Rights: []ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST, ttnpb.RIGHT_USER_GATEWAYS_LIST},
			},
			[]ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			err := is.enforceUserRights(newContextWithAuthorizationData(context.Background(), tc.authorizationData), tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(err, should.DescribeError, common.ErrPermissionDenied)
			}
		})
	}
}

func TestEnforceAdmin(t *testing.T) {
	is := newTestIS(t)

	// `alice` is an admin user.
	alice := newTestUsers()["alice"]

	// `john-doe` is not an admin user.
	john := newTestUsers()["john-doe"]

	for i, tc := range []struct {
		authorizationData *authorizationData
		sucesss           bool
	}{
		{
			// Fails because the authorization data identifiers are not UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				Source: auth.Key,
				Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_DELETE},
			},
			false,
		},
		{
			// Fails because john is not an admin.
			&authorizationData{
				EntityIdentifiers: john.UserIdentifiers,
				Source:            auth.Token,
				Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_ADMIN},
			},
			false,
		},
		{
			// Fails because the authorization data does not include `RIGHT_USER_ADMIN` right.
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST},
			},
			false,
		},
		{
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_ADMIN},
			},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			err := is.enforceAdmin(newContextWithAuthorizationData(context.Background(), tc.authorizationData))
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(err, should.DescribeError, common.ErrPermissionDenied)
			}
		})
	}
}

func TestEnforceApplicationRights(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	alice := newTestUsers()["alice"]
	bob := newTestUsers()["bob"]
	appIDs := ttnpb.ApplicationIdentifiers{
		ApplicationID: "alice-test-app",
	}

	ctx := newTestCtx(alice.UserIdentifiers)

	// Create an application under `alice` user.
	_, err := is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
		Application: ttnpb.Application{
			ApplicationIdentifiers: appIDs,
		},
	})
	a.So(err, should.BeNil)
	defer func(err error) {
		if err != nil {
			is.store.Applications.Delete(appIDs)
		}
	}(err)

	for i, tc := range []struct {
		authorizationData *authorizationData
		appIDs            ttnpb.ApplicationIdentifiers
		rights            []ttnpb.Right
		sucesss           bool
	}{
		{
			// (API key) Fails because requested rights are not contained in
			// the authorization data.
			&authorizationData{
				EntityIdentifiers: appIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllApplicationRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_USER_DELETE},
			false,
		},
		{
			// (API key) Fails because authorization data identifiers are UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.UserIdentifiers{
					UserID: "foo-user",
				},
				Source: auth.Key,
				Rights: ttnpb.AllApplicationRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_DELETE},
			false,
		},
		{
			// (API key) Fails because the authorization data are for a different
			// application than the requested one.
			&authorizationData{
				EntityIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "another-app",
				},
				Source: auth.Key,
				Rights: ttnpb.AllApplicationRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_DELETE},
			false,
		},
		{
			// (API key).
			&authorizationData{
				EntityIdentifiers: appIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllApplicationRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_DELETE},
			true,
		},
		{
			// (Token) Fails because the authorization data identifiers are not UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "another-app",
				},
				Source: auth.Token,
				Rights: ttnpb.AllRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_DELETE},
			false,
		},
		{
			// (Token) Fails because `bob` is not a collaborator.
			&authorizationData{
				EntityIdentifiers: bob.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
			false,
		},
		{
			// (Token) Fails because `alice` does not have all the requested rights.
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			appIDs,
			append(ttnpb.AllApplicationRights(), ttnpb.RIGHT_INVALID),
			false,
		},
		{
			// (Token).
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			err := is.enforceApplicationRights(newContextWithAuthorizationData(context.Background(), tc.authorizationData), tc.appIDs, tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(err, should.DescribeError, common.ErrPermissionDenied)
			}
		})
	}
}

func TestEnforceGatewayRights(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	alice := newTestUsers()["alice"]
	bob := newTestUsers()["bob"]
	gtwIDs := ttnpb.GatewayIdentifiers{
		GatewayID: "alice-test-gtw",
	}

	ctx := newTestCtx(alice.UserIdentifiers)

	// Create a gateway under `alice` user.
	_, err := is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
		Gateway: ttnpb.Gateway{
			GatewayIdentifiers: gtwIDs,
		},
	})
	a.So(err, should.BeNil)
	defer func(err error) {
		if err != nil {
			is.store.Gateways.Delete(gtwIDs)
		}
	}(err)

	for i, tc := range []struct {
		authorizationData *authorizationData
		gtwIDs            ttnpb.GatewayIdentifiers
		rights            []ttnpb.Right
		sucesss           bool
	}{
		{
			// (API key) Fails because requested rights are not contained in
			// the authorization data.
			&authorizationData{
				EntityIdentifiers: gtwIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllGatewayRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_USER_DELETE},
			false,
		},
		{
			// (API key) Fails because authorization data identifiers are UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.UserIdentifiers{
					UserID: "foo-user",
				},
				Source: auth.Key,
				Rights: ttnpb.AllGatewayRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_DELETE},
			false,
		},
		{
			// (API key) Fails because the authorization data are for a different
			// gateway than the requested one.
			&authorizationData{
				EntityIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayID: "another-gtw",
				},
				Source: auth.Key,
				Rights: ttnpb.AllGatewayRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_DELETE},
			false,
		},
		{
			// (API key).
			&authorizationData{
				EntityIdentifiers: gtwIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllGatewayRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_DELETE},
			true,
		},
		{
			// (Token) Fails because the authorization data identifiers are not UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayID: "another-gtw",
				},
				Source: auth.Token,
				Rights: ttnpb.AllRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_DELETE},
			false,
		},
		{
			// (Token) Fails because `bob` is not a collaborator.
			&authorizationData{
				EntityIdentifiers: bob.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO},
			false,
		},
		{
			// (Token) Fails because `alice` does not have all the requested rights.
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			gtwIDs,
			append(ttnpb.AllGatewayRights(), ttnpb.RIGHT_INVALID),
			false,
		},
		{
			// (Token).
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			err := is.enforceGatewayRights(newContextWithAuthorizationData(context.Background(), tc.authorizationData), tc.gtwIDs, tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(err, should.DescribeError, common.ErrPermissionDenied)
			}
		})
	}
}

func TestEnforceOrganizationRights(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	alice := newTestUsers()["alice"]
	bob := newTestUsers()["bob"]
	orgIDs := ttnpb.OrganizationIdentifiers{
		OrganizationID: "alice-test-org",
	}

	ctx := newTestCtx(alice.UserIdentifiers)

	// Create an organization under `alice` user.
	_, err := is.organizationService.CreateOrganization(ctx, &ttnpb.CreateOrganizationRequest{
		Organization: ttnpb.Organization{
			OrganizationIdentifiers: orgIDs,
		},
	})
	a.So(err, should.BeNil)
	defer func(err error) {
		if err != nil {
			is.store.Organizations.Delete(orgIDs)
		}
	}(err)

	for i, tc := range []struct {
		authorizationData *authorizationData
		orgIDs            ttnpb.OrganizationIdentifiers
		rights            []ttnpb.Right
		sucesss           bool
	}{
		{
			// (API key) Fails because requested rights are not contained in
			// the authorization data.
			&authorizationData{
				EntityIdentifiers: orgIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllOrganizationRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_USER_DELETE},
			false,
		},
		{
			// (API key) Fails because authorization data identifiers are UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.UserIdentifiers{
					UserID: "foo-user",
				},
				Source: auth.Key,
				Rights: ttnpb.AllOrganizationRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_DELETE},
			false,
		},
		{
			// (API key) Fails because the authorization data are for a different
			// organization than the requested one.
			&authorizationData{
				EntityIdentifiers: ttnpb.OrganizationIdentifiers{
					OrganizationID: "another-org",
				},
				Source: auth.Key,
				Rights: ttnpb.AllOrganizationRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_DELETE},
			false,
		},
		{
			// (API key).
			&authorizationData{
				EntityIdentifiers: orgIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllOrganizationRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_DELETE},
			true,
		},
		{
			// (Token) Fails because the authorization data identifiers are not UserIdentifiers.
			&authorizationData{
				EntityIdentifiers: ttnpb.OrganizationIdentifiers{
					OrganizationID: "another-org",
				},
				Source: auth.Token,
				Rights: ttnpb.AllRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_DELETE},
			false,
		},
		{
			// (Token) Fails because `bob` is not a member.
			&authorizationData{
				EntityIdentifiers: bob.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO},
			false,
		},
		{
			// (Token) Fails because `alice` does not have all the requested rights.
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			orgIDs,
			append(ttnpb.AllOrganizationRights(), ttnpb.RIGHT_INVALID),
			false,
		},
		{
			// (Token).
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            ttnpb.AllRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			err := is.enforceOrganizationRights(newContextWithAuthorizationData(context.Background(), tc.authorizationData), tc.orgIDs, tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(err, should.DescribeError, common.ErrPermissionDenied)
			}
		})
	}
}
