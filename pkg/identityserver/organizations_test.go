// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"sort"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.IsOrganizationServer = new(organizationService)

func TestOrganization(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	user := testUsers()["bob"]
	alice := testUsers()["alice"]

	org := ttnpb.Organization{
		OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: "foo-org"},
	}

	ctx := testCtx(user.UserID)

	_, err := is.organizationService.CreateOrganization(ctx, &ttnpb.CreateOrganizationRequest{
		Organization: org,
	})
	a.So(err, should.BeNil)

	// Can't create organizations with blacklisted IDs.
	for _, id := range testSettings().BlacklistedIDs {
		_, err := is.organizationService.CreateOrganization(ctx, &ttnpb.CreateOrganizationRequest{
			Organization: ttnpb.Organization{
				OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}

	found, err := is.organizationService.GetOrganization(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeOrganizationIgnoringAutoFields, org)

	orgs, err := is.organizationService.ListOrganizations(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(orgs.Organizations, should.HaveLength, 1) {
		a.So(orgs.Organizations[0], test.ShouldBeOrganizationIgnoringAutoFields, org)
	}

	org.Description = "foo"
	_, err = is.organizationService.UpdateOrganization(ctx, &ttnpb.UpdateOrganizationRequest{
		Organization: org,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.organizationService.GetOrganization(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeOrganizationIgnoringAutoFields, org)

	// Generate a new API key.
	key, err := is.organizationService.GenerateOrganizationAPIKey(ctx, &ttnpb.GenerateOrganizationAPIKeyRequest{
		OrganizationIdentifier: org.OrganizationIdentifier,
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
		OrganizationIdentifier: org.OrganizationIdentifier,
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// Can't generate another API key with the same name.
	_, err = is.organizationService.GenerateOrganizationAPIKey(ctx, &ttnpb.GenerateOrganizationAPIKeyRequest{
		OrganizationIdentifier: org.OrganizationIdentifier,
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := is.organizationService.ListOrganizationAPIKeys(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.organizationService.RemoveOrganizationAPIKey(ctx, &ttnpb.RemoveOrganizationAPIKeyRequest{
		OrganizationIdentifier: org.OrganizationIdentifier,
		Name: key.Name,
	})
	a.So(err, should.BeNil)

	keys, err = is.organizationService.ListOrganizationAPIKeys(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// Set a new member with SETTINGS_MEMBER and INFO rights.
	member := &ttnpb.OrganizationMember{
		OrganizationIdentifier: org.OrganizationIdentifier,
		UserIdentifier:         ttnpb.UserIdentifier{UserID: alice.UserID},
		Rights:                 []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS},
	}

	_, err = is.organizationService.SetOrganizationMember(ctx, member)
	a.So(err, should.BeNil)

	rights, err := is.organizationService.ListOrganizationRights(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllOrganizationRights())

	members, err := is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(members.Members, should.HaveLength, 2)
	a.So(members.Members, should.Contain, member)
	a.So(members.Members, should.Contain, &ttnpb.OrganizationMember{
		OrganizationIdentifier: org.OrganizationIdentifier,
		UserIdentifier:         ttnpb.UserIdentifier{UserID: user.UserID},
		Rights:                 ttnpb.AllOrganizationRights(),
	})

	// The new member can't grant himself more rights.
	{
		member.Rights = append(member.Rights, ttnpb.RIGHT_ORGANIZATION_SETTINGS_KEYS)

		ctx := testCtx(alice.UserID)

		_, err = is.organizationService.SetOrganizationMember(ctx, member)
		a.So(err, should.BeNil)

		rights, err := is.organizationService.ListOrganizationRights(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 2)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_ORGANIZATION_SETTINGS_KEYS)

		members, err = is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
		a.So(err, should.BeNil)

		// But it can revoke itself the INFO right.
		member.Rights = []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS}
		_, err = is.organizationService.SetOrganizationMember(ctx, member)
		a.So(err, should.BeNil)

		rights, err = is.organizationService.ListOrganizationRights(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 1)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_ORGANIZATION_INFO)

		// Grant back the right.
		member.Rights = []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_INFO, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS}
		_, err = is.organizationService.SetOrganizationMember(testCtx(user.UserID), member)
		a.So(err, should.BeNil)
	}

	// To unset the main member will result in error as the organization will become unmanageable.
	_, err = is.organizationService.SetOrganizationMember(ctx, &ttnpb.OrganizationMember{
		OrganizationIdentifier: org.OrganizationIdentifier,
		UserIdentifier:         ttnpb.UserIdentifier{UserID: user.UserID},
	})
	a.So(err, should.NotBeNil)
	a.So(ErrSetOrganizationMemberFailed.Describes(err), should.BeTrue)

	members, err = is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(members.Members, should.HaveLength, 2)

	// Unset the last added member.
	member.Rights = []ttnpb.Right{}
	_, err = is.organizationService.SetOrganizationMember(ctx, member)
	a.So(err, should.BeNil)

	members, err = is.organizationService.ListOrganizationMembers(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
	a.So(members.Members, should.HaveLength, 1)

	// Applications.
	{
		apps, err := is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(apps.Applications, should.HaveLength, 0)

		app := ttnpb.Application{
			ApplicationIdentifier: ttnpb.ApplicationIdentifier{ApplicationID: "org-app"},
		}

		_, err = is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
			Application:            app,
			OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID},
		})
		a.So(err, should.BeNil)

		apps, err = is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		if a.So(apps.Applications, should.HaveLength, 1) {
			a.So(apps.Applications[0], test.ShouldBeApplicationIgnoringAutoFields, app)
		}

		_, err = is.applicationService.DeleteApplication(ctx, &ttnpb.ApplicationIdentifier{ApplicationID: app.ApplicationID})
		a.So(err, should.BeNil)

		apps, err = is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(apps.Applications, should.HaveLength, 0)
	}

	// Gateways.
	{
		gtws, err := is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(gtws.Gateways, should.HaveLength, 0)

		gtw := ttnpb.Gateway{
			GatewayIdentifier: ttnpb.GatewayIdentifier{GatewayID: "org-gtw"},
			ClusterAddress:    "localhost:1234",
			FrequencyPlanID:   "868.8",
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
		}

		_, err = is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
			Gateway:                gtw,
			OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID},
		})
		a.So(err, should.BeNil)

		gtws, err = is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		if a.So(gtws.Gateways, should.HaveLength, 1) {
			a.So(gtws.Gateways[0], test.ShouldBeGatewayIgnoringAutoFields, gtw)
		}

		_, err = is.gatewayService.DeleteGateway(ctx, &ttnpb.GatewayIdentifier{GatewayID: gtw.GatewayID})
		a.So(err, should.BeNil)

		gtws, err = is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{OrganizationIdentifier: ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID}})
		a.So(err, should.BeNil)
		a.So(gtws.Gateways, should.HaveLength, 0)
	}

	_, err = is.organizationService.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifier{OrganizationID: org.OrganizationID})
	a.So(err, should.BeNil)
}
