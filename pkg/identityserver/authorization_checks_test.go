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

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestEnforceUserRights(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	for i, tc := range []struct {
		claims  *claims
		rights  []ttnpb.Right
		sucesss bool
	}{
		{
			// Fails because the claims identifiers are not UserIdentifiers.
			&claims{
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
			// Fails because the claims does not include the requested right.
			&claims{
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
			&claims{
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
			err := is.enforceUserRights(newContextWithClaims(context.Background(), tc.claims), tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(ErrNotAuthorized.Describes(err), should.BeTrue)
			}
		})
	}
}

func TestEnforceAdmin(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	// `alice` is an admin user.
	alice := testUsers()["alice"]

	// `john-doe` is not an admin user.
	john := testUsers()["john-doe"]

	for i, tc := range []struct {
		claims  *claims
		sucesss bool
	}{
		{
			// Fails because the claims identifiers are not UserIdentifiers.
			&claims{
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
			&claims{
				EntityIdentifiers: john.UserIdentifiers,
				Source:            auth.Token,
				Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_ADMIN},
			},
			false,
		},
		{
			// Fails because the claims does not include `RIGHT_USER_ADMIN` right.
			&claims{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_APPLICATIONS_LIST},
			},
			false,
		},
		{
			&claims{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Token,
				Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_ADMIN},
			},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := is.enforceAdmin(newContextWithClaims(context.Background(), tc.claims))
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(ErrNotAuthorized.Describes(err), should.BeTrue)
			}
		})
	}
}

func TestEnforceApplicationRights(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]
	appIDs := ttnpb.ApplicationIdentifiers{
		ApplicationID: "alice-test-app",
	}

	ctx := testCtx(alice.UserIdentifiers.UserID)

	// Create an application under `alice` user.
	_, err := is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
		Application: ttnpb.Application{
			ApplicationIdentifiers: appIDs,
		},
	})
	defer func(err error) {
		if err != nil {
			is.store.Applications.Delete(appIDs)
		}
	}(err)

	for i, tc := range []struct {
		claims  *claims
		appIDs  ttnpb.ApplicationIdentifiers
		rights  []ttnpb.Right
		sucesss bool
	}{
		{
			// (API key) Fails because requested rights are not contained in the claims.
			&claims{
				EntityIdentifiers: appIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllApplicationRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_USER_DELETE},
			false,
		},
		{
			// (API key) Fails because claims identifiers are UserIdentifiers.
			&claims{
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
			// (API key) Fails because the claims are for a different application than the requested one.
			&claims{
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
			&claims{
				EntityIdentifiers: appIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllApplicationRights(),
			},
			appIDs,
			[]ttnpb.Right{ttnpb.RIGHT_APPLICATION_DELETE},
			true,
		},
		{
			// (Token) Fails because the claims identifiers are not UserIdentifiers.
			&claims{
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
			&claims{
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
			&claims{
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
			&claims{
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
			err := is.enforceApplicationRights(newContextWithClaims(context.Background(), tc.claims), tc.appIDs, tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(ErrNotAuthorized.Describes(err), should.BeTrue)
			}
		})
	}
}

func TestEnforceGatewayRights(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]
	gtwIDs := ttnpb.GatewayIdentifiers{
		GatewayID: "alice-test-gtw",
	}

	ctx := testCtx(alice.UserIdentifiers.UserID)

	// Create a gateway under `alice` user.
	_, err := is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
		Gateway: ttnpb.Gateway{
			GatewayIdentifiers: gtwIDs,
		},
	})
	defer func(err error) {
		if err != nil {
			is.store.Gateways.Delete(gtwIDs)
		}
	}(err)

	for i, tc := range []struct {
		claims  *claims
		gtwIDs  ttnpb.GatewayIdentifiers
		rights  []ttnpb.Right
		sucesss bool
	}{
		{
			// (API key) Fails because requested rights are not contained in the claims.
			&claims{
				EntityIdentifiers: gtwIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllGatewayRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_USER_DELETE},
			false,
		},
		{
			// (API key) Fails because claims identifiers are UserIdentifiers.
			&claims{
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
			// (API key) Fails because the claims are for a different gateway than the requested one.
			&claims{
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
			&claims{
				EntityIdentifiers: gtwIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllGatewayRights(),
			},
			gtwIDs,
			[]ttnpb.Right{ttnpb.RIGHT_GATEWAY_DELETE},
			true,
		},
		{
			// (Token) Fails because the claims identifiers are not UserIdentifiers.
			&claims{
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
			&claims{
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
			&claims{
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
			&claims{
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
			err := is.enforceGatewayRights(newContextWithClaims(context.Background(), tc.claims), tc.gtwIDs, tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(ErrNotAuthorized.Describes(err), should.BeTrue)
			}
		})
	}
}

func TestEnforceOrganizationRights(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]
	orgIDs := ttnpb.OrganizationIdentifiers{
		OrganizationID: "alice-test-org",
	}

	ctx := testCtx(alice.UserIdentifiers.UserID)

	// Create an organization under `alice` user.
	_, err := is.organizationService.CreateOrganization(ctx, &ttnpb.CreateOrganizationRequest{
		Organization: ttnpb.Organization{
			OrganizationIdentifiers: orgIDs,
		},
	})
	defer func(err error) {
		if err != nil {
			is.store.Organizations.Delete(orgIDs)
		}
	}(err)

	for i, tc := range []struct {
		claims  *claims
		orgIDs  ttnpb.OrganizationIdentifiers
		rights  []ttnpb.Right
		sucesss bool
	}{
		{
			// (API key) Fails because requested rights are not contained in the claims.
			&claims{
				EntityIdentifiers: orgIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllOrganizationRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_USER_DELETE},
			false,
		},
		{
			// (API key) Fails because claims identifiers are UserIdentifiers.
			&claims{
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
			// (API key) Fails because the claims are for a different organization than the requested one.
			&claims{
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
			&claims{
				EntityIdentifiers: orgIDs,
				Source:            auth.Key,
				Rights:            ttnpb.AllOrganizationRights(),
			},
			orgIDs,
			[]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_DELETE},
			true,
		},
		{
			// (Token) Fails because the claims identifiers are not UserIdentifiers.
			&claims{
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
			&claims{
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
			&claims{
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
			&claims{
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
			err := is.enforceOrganizationRights(newContextWithClaims(context.Background(), tc.claims), tc.orgIDs, tc.rights...)
			if tc.sucesss {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
				a.So(ErrNotAuthorized.Describes(err), should.BeTrue)
			}
		})
	}
}
