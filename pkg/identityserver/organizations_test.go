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
	"sort"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	errshould "go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var _ ttnpb.IsOrganizationServer = new(organizationService)

func TestOrganization(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	user := newTestUsers()["bob"]
	alice := newTestUsers()["alice"]

	org := ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "foo-org"},
	}

	ctx := newTestCtx(user.UserIdentifiers)

	_, err := is.organizationService.CreateOrganization(ctx, &ttnpb.CreateOrganizationRequest{
		Organization: org,
	})
	a.So(err, should.BeNil)

	// Can't create organizations with blacklisted IDs.
	for _, id := range newTestSettings().BlacklistedIDs {
		_, err := is.organizationService.CreateOrganization(ctx, &ttnpb.CreateOrganizationRequest{
			Organization: ttnpb.Organization{
				OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(err, errshould.Describe, ErrBlacklistedID)
	}

	found, err := is.organizationService.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(OrganizationGeneratedFields...), org)

	orgs, err := is.organizationService.ListOrganizations(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	if a.So(orgs.Organizations, should.HaveLength, 1) {
		a.So(orgs.Organizations[0], should.EqualFieldsWithIgnores(OrganizationGeneratedFields...), org)
	}

	org.Description = "foo"
	_, err = is.organizationService.UpdateOrganization(ctx, &ttnpb.UpdateOrganizationRequest{
		Organization: org,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.organizationService.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(OrganizationGeneratedFields...), org)

	// Generate a new API key.
	key, err := is.organizationService.GenerateOrganizationAPIKey(ctx, &ttnpb.GenerateOrganizationAPIKeyRequest{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		Name:   "foo",
		Rights: ttnpb.AllOrganizationRights(),
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllOrganizationRights())

	// Update the API key.
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.organizationService.UpdateOrganizationAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// Can't generate another API key with the same name.
	_, err = is.organizationService.GenerateOrganizationAPIKey(ctx, &ttnpb.GenerateOrganizationAPIKeyRequest{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(store.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := is.organizationService.ListOrganizationAPIKeys(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.organizationService.RemoveOrganizationAPIKey(ctx, &ttnpb.RemoveOrganizationAPIKeyRequest{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		Name: key.Name,
	})
	a.So(err, should.BeNil)

	keys, err = is.organizationService.ListOrganizationAPIKeys(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// Set a new member with SETTINGS_MEMBERS and INFO rights.
	member := &ttnpb.OrganizationMember{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		UserIdentifiers:         alice.UserIdentifiers,
		Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS},
	}

	_, err = is.organizationService.SetOrganizationMember(ctx, member)
	a.So(err, should.BeNil)

	rights, err := is.organizationService.ListOrganizationRights(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllOrganizationRights())

	members, err := is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(members.Members, should.HaveLength, 2)
	a.So(members.Members, should.Contain, member)
	a.So(members.Members, should.Contain, &ttnpb.OrganizationMember{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		UserIdentifiers:         user.UserIdentifiers,
		Rights:                  ttnpb.AllOrganizationRights(),
	})

	// The new member can't grant himself more rights.
	{
		member.Rights = append(member.Rights, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)

		ctx := newTestCtx(alice.UserIdentifiers)

		_, err = is.organizationService.SetOrganizationMember(ctx, member)
		a.So(err, should.BeNil)

		rights, err := is.organizationService.ListOrganizationRights(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 2)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)

		members, err = is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
		a.So(err, should.BeNil)

		// But they can revoke themselves the INFO right.
		member.Rights = []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS}
		_, err = is.organizationService.SetOrganizationMember(ctx, member)
		a.So(err, should.BeNil)

		rights, err = is.organizationService.ListOrganizationRights(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 1)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_ORGANIZATION_INFO)

		// Grant back the right.
		member.Rights = []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS}
		_, err = is.organizationService.SetOrganizationMember(newTestCtx(user.UserIdentifiers), member)
		a.So(err, should.BeNil)
	}

	// To unset the main member will result in error as the organization will become unmanageable.
	_, err = is.organizationService.SetOrganizationMember(ctx, &ttnpb.OrganizationMember{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		UserIdentifiers:         ttnpb.UserIdentifiers{UserID: user.UserID},
	})
	a.So(err, should.NotBeNil)
	a.So(ErrUnmanageableOrganization.Describes(err), should.BeTrue)

	members, err = is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(members.Members, should.HaveLength, 2)

	// Unset the last added member.
	member.Rights = []ttnpb.Right{}
	_, err = is.organizationService.SetOrganizationMember(ctx, member)
	a.So(err, should.BeNil)

	members, err = is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(members.Members, should.HaveLength, 1)

	// Applications.
	{
		apps, err := is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(apps.Applications, should.HaveLength, 0)

		app := ttnpb.Application{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "org-app"},
		}

		_, err = is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
			Application:             app,
			OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID},
		})
		a.So(err, should.BeNil)

		apps, err = is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		if a.So(apps.Applications, should.HaveLength, 1) {
			a.So(apps.Applications[0], should.EqualFieldsWithIgnores(ApplicationGeneratedFields...), app)
		}

		_, err = is.applicationService.DeleteApplication(ctx, &app.ApplicationIdentifiers)
		a.So(err, should.BeNil)

		apps, err = is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(apps.Applications, should.HaveLength, 0)
	}

	// Gateways.
	{
		gtws, err := is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(gtws.Gateways, should.HaveLength, 0)

		gtw := ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "org-gtw"},
			ClusterAddress:     "localhost:1234",
			FrequencyPlanID:    "868.8",
			Attributes: map[string]string{
				"version": "1.2",
			},
			Antennas: []ttnpb.GatewayAntenna{
				{
					Gain: 1.1,
					Location: ttnpb.Location{
						Latitude:  1.1,
						Longitude: 1.1,
					},
				},
				{
					Gain: 2.2,
					Location: ttnpb.Location{
						Latitude:  2.2,
						Longitude: 2.2,
					},
				},
				{
					Gain: 3,
					Location: ttnpb.Location{
						Latitude:  3,
						Longitude: 3,
					},
				},
			},
			Radios: []ttnpb.GatewayRadio{},
		}

		_, err = is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
			Gateway:                 gtw,
			OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID},
		})
		a.So(err, should.BeNil)

		gtws, err = is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		if a.So(gtws.Gateways, should.HaveLength, 1) {
			a.So(gtws.Gateways[0], should.EqualFieldsWithIgnores(GatewayGeneratedFields...), gtw)
		}

		_, err = is.gatewayService.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
		a.So(err, should.BeNil)

		gtws, err = is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(gtws.Gateways, should.HaveLength, 0)
	}

	_, err = is.organizationService.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
}
