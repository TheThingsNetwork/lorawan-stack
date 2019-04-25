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

package store

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestMembershipStore(t *testing.T) {
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db,
			&Membership{},
			&Account{}, &User{}, &Organization{},
			&Application{}, &Client{}, &Gateway{},
		)

		s := newStore(db)
		store := GetMembershipStore(db)

		usr := &User{Account: Account{UID: "test-user"}}
		s.createEntity(ctx, usr)
		usrIDs := usr.Account.OrganizationOrUserIdentifiers()

		org := &Organization{Account: Account{UID: "test-org"}}
		s.createEntity(ctx, org)
		orgIDs := org.Account.OrganizationOrUserIdentifiers()

		s.createEntity(ctx, &Application{ApplicationID: "test-app"})
		s.createEntity(ctx, &Client{ClientID: "test-cli"})
		s.createEntity(ctx, &Gateway{GatewayID: "test-gtw"})

		s.createEntity(ctx, &User{Account: Account{UID: "other-user"}})
		s.createEntity(ctx, &Organization{Account: Account{UID: "other-org"}})
		s.createEntity(ctx, &Application{ApplicationID: "other-app"})
		s.createEntity(ctx, &Client{ClientID: "other-cli"})
		s.createEntity(ctx, &Gateway{GatewayID: "other-gtw"})

		for _, tt := range []struct {
			Name              string
			Identifiers       *ttnpb.OrganizationOrUserIdentifiers
			MemberIdentifiers *ttnpb.EntityIdentifiers
			Rights            []ttnpb.Right
			RightsUpdated     []ttnpb.Right
			EntityType        string
		}{
			{
				Name:              "User-Application",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}.EntityIdentifiers(),
				Rights:            []ttnpb.Right{ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_INFO,
					ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
				},
				EntityType: "application",
			},
			{
				Name:              "User-Client",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli"}.EntityIdentifiers(),
				Rights:            []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_CLIENT_ALL,
					ttnpb.RIGHT_APPLICATION_INFO,
				},
				EntityType: "client",
			},
			{
				Name:              "User-Gateway",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw"}.EntityIdentifiers(),
				Rights:            []ttnpb.Right{ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_GATEWAY_INFO,
					ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
				},
				EntityType: "gateway",
			},
			{
				Name:              "User-Organization",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "test-org"}.EntityIdentifiers(),
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_ALL,
					ttnpb.RIGHT_GATEWAY_ALL,
					ttnpb.RIGHT_ORGANIZATION_ALL,
				},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_ALL,
					ttnpb.RIGHT_CLIENT_ALL,
					ttnpb.RIGHT_GATEWAY_ALL,
					ttnpb.RIGHT_ORGANIZATION_ALL,
				},
				EntityType: "organization",
			},
			{
				Name:              "Organization-Application",
				Identifiers:       orgIDs,
				MemberIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}.EntityIdentifiers(),
				Rights:            []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_INFO,
					ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
				},
				EntityType: "application",
			},
			{
				Name:              "Organization-Client",
				Identifiers:       orgIDs,
				MemberIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli"}.EntityIdentifiers(),
				Rights:            []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_CLIENT_ALL,
					ttnpb.RIGHT_APPLICATION_INFO,
				},
				EntityType: "client",
			},
			{
				Name:              "Organization-Gateway",
				Identifiers:       orgIDs,
				MemberIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw"}.EntityIdentifiers(),
				Rights:            []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO},
				RightsUpdated: []ttnpb.Right{
					ttnpb.RIGHT_GATEWAY_INFO,
					ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
				},
				EntityType: "gateway",
			},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				a := assertions.New(t)

				// set membership
				err := store.SetMember(ctx,
					tt.Identifiers,
					tt.MemberIdentifiers,
					ttnpb.RightsFrom(tt.Rights...),
				)

				a.So(err, should.BeNil)

				members, err := store.FindMembers(ctx, tt.MemberIdentifiers)

				a.So(err, should.BeNil)
				if a.So(members, should.HaveLength, 1) {
					for ouid, rights := range members {
						a.So(ouid, should.Resemble, tt.Identifiers)
						a.So(rights.GetRights(), should.Resemble, tt.Rights)
					}
				}

				memberRights, err := store.FindMemberRights(ctx, tt.Identifiers, tt.EntityType)

				a.So(err, should.BeNil)
				if a.So(memberRights, should.HaveLength, 1) {
					for eid, rights := range memberRights {
						a.So(eid, should.Resemble, tt.MemberIdentifiers)
						a.So(rights.GetRights(), should.Resemble, tt.Rights)
					}
				}

				// update membership
				err = store.SetMember(ctx,
					tt.Identifiers,
					tt.MemberIdentifiers,
					ttnpb.RightsFrom(tt.RightsUpdated...),
				)

				a.So(err, should.BeNil)

				memberRights, err = store.FindMemberRights(ctx, tt.Identifiers, tt.EntityType)

				a.So(err, should.BeNil)
				if a.So(memberRights, should.HaveLength, 1) {
					for eid, rights := range memberRights {
						a.So(eid, should.Resemble, tt.MemberIdentifiers)
						a.So(rights.GetRights(), should.Resemble, tt.RightsUpdated)
					}
				}

				// delete membership
				err = store.SetMember(ctx,
					tt.Identifiers,
					tt.MemberIdentifiers,
					ttnpb.RightsFrom([]ttnpb.Right{}...),
				)

				a.So(err, should.BeNil)

				memberRights, err = store.FindMemberRights(ctx, tt.Identifiers, tt.EntityType)

				a.So(err, should.BeNil)
				a.So(memberRights, should.HaveLength, 0)
			})
		}

		t.Run("Organization-Organization", func(t *testing.T) {
			a := assertions.New(t)

			err := store.SetMember(ctx,
				orgIDs,
				ttnpb.OrganizationIdentifiers{OrganizationID: "other-org"}.EntityIdentifiers(),
				ttnpb.RightsFrom([]ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL}...),
			)

			a.So(err, should.NotBeNil)
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		})

		userNotFoundIDs := ttnpb.UserIdentifiers{UserID: "test-usr-not-found"}.OrganizationOrUserIdentifiers()
		organizationNotFoundIDs := ttnpb.UserIdentifiers{UserID: "test-usr-not-found"}.OrganizationOrUserIdentifiers()

		for _, tt := range []struct {
			Name              string
			Identifiers       *ttnpb.OrganizationOrUserIdentifiers
			MemberIdentifiers *ttnpb.EntityIdentifiers
			EntityType        string
		}{
			{
				Name:              "User-Application - user not found",
				Identifiers:       userNotFoundIDs,
				MemberIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}.EntityIdentifiers(),
				EntityType:        "application",
			},
			{
				Name:              "User-Client - user not found",
				Identifiers:       userNotFoundIDs,
				MemberIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli"}.EntityIdentifiers(),
				EntityType:        "client",
			},
			{
				Name:              "User-Gateway - user not found",
				Identifiers:       userNotFoundIDs,
				MemberIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw"}.EntityIdentifiers(),
				EntityType:        "gateway",
			},
			{
				Name:              "User-Organization - user not found",
				Identifiers:       userNotFoundIDs,
				MemberIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "test-org"}.EntityIdentifiers(),
				EntityType:        "organization",
			},
			{
				Name:              "Organization-Application - organization not found",
				Identifiers:       organizationNotFoundIDs,
				MemberIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}.EntityIdentifiers(),
				EntityType:        "application",
			},
			{
				Name:              "Organization-Client - organization not found",
				Identifiers:       organizationNotFoundIDs,
				MemberIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli"}.EntityIdentifiers(),
				EntityType:        "client",
			},
			{
				Name:              "Organization-Gateway - organization not found",
				Identifiers:       organizationNotFoundIDs,
				MemberIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw"}.EntityIdentifiers(),
				EntityType:        "gateway",
			},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				a := assertions.New(t)

				err := store.SetMember(ctx,
					tt.Identifiers,
					tt.MemberIdentifiers,
					ttnpb.RightsFrom([]ttnpb.Right{}...),
				)

				a.So(err, should.NotBeNil)
				a.So(errors.IsNotFound(err), should.BeTrue)

				_, err = store.FindMemberRights(ctx, tt.Identifiers, tt.EntityType)

				a.So(err, should.NotBeNil)
				a.So(errors.IsNotFound(err), should.BeTrue)
			})
		}

		for _, tt := range []struct {
			Name              string
			Identifiers       *ttnpb.OrganizationOrUserIdentifiers
			MemberIdentifiers *ttnpb.EntityIdentifiers
			EntityType        string
		}{
			{
				Name:              "User-Application - application not found",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-not-found"}.EntityIdentifiers(),
				EntityType:        "application",
			},
			{
				Name:              "User-Client - client not found",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli-not-found"}.EntityIdentifiers(),
				EntityType:        "client",
			},
			{
				Name:              "User-Gateway - gateway not found",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw-not-found"}.EntityIdentifiers(),
				EntityType:        "gateway",
			},
			{
				Name:              "User-Organization - organization not found",
				Identifiers:       usrIDs,
				MemberIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "test-org-not-found"}.EntityIdentifiers(),
				EntityType:        "organization",
			},
			{
				Name:              "Organization-Application - application not found",
				Identifiers:       orgIDs,
				MemberIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-not-found"}.EntityIdentifiers(),
				EntityType:        "application",
			},
			{
				Name:              "Organization-Client - client not found",
				Identifiers:       orgIDs,
				MemberIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli-not-found"}.EntityIdentifiers(),
				EntityType:        "client",
			},
			{
				Name:              "Organization-Gateway - gateway not found",
				Identifiers:       orgIDs,
				MemberIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw-not-found"}.EntityIdentifiers(),
				EntityType:        "gateway",
			},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				a := assertions.New(t)

				err := store.SetMember(ctx,
					tt.Identifiers,
					tt.MemberIdentifiers,
					ttnpb.RightsFrom([]ttnpb.Right{}...),
				)

				a.So(err, should.NotBeNil)
				a.So(errors.IsNotFound(err), should.BeTrue)

				_, err = store.FindMembers(ctx, tt.MemberIdentifiers)

				a.So(err, should.NotBeNil)
				a.So(errors.IsNotFound(err), should.BeTrue)
			})
		}
	})
}
