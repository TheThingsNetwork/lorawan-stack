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

package store

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
		store := GetMembershipStore(db)

		usr := &User{Account: Account{UID: "test-user"}}
		db.Create(usr)
		usrIDs := usr.Account.OrganizationOrUserIdentifiers()

		org := &Organization{Account: Account{UID: "test-org"}}
		db.Create(org)
		orgIDs := org.Account.OrganizationOrUserIdentifiers()

		db.Create(&Application{ApplicationID: "test-app"})
		db.Create(&Client{ClientID: "test-cli"})
		db.Create(&Gateway{GatewayID: "test-gtw"})

		db.Create(&User{Account: Account{UID: "other-user"}})
		db.Create(&Organization{Account: Account{UID: "other-org"}})
		db.Create(&Application{ApplicationID: "other-app"})
		db.Create(&Client{ClientID: "other-cli"})
		db.Create(&Gateway{GatewayID: "other-gtw"})

		for _, tt := range []struct {
			Name        string
			Identifiers *ttnpb.EntityIdentifiers
			Rights      []ttnpb.Right
			EntityType  string
		}{
			{
				Name:        "User-Application",
				Identifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC},
				EntityType:  "application",
			},
			{
				Name:        "User-Client",
				Identifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
				EntityType:  "client",
			},
			{
				Name:        "User-Gateway",
				Identifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC},
				EntityType:  "gateway",
			},
			{
				Name:        "User-Organization",
				Identifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "test-org"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL, ttnpb.RIGHT_CLIENT_ALL, ttnpb.RIGHT_GATEWAY_ALL, ttnpb.RIGHT_ORGANIZATION_ALL},
				EntityType:  "organization",
			},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				a := assertions.New(t)

				err := store.SetMember(ctx,
					usrIDs,
					tt.Identifiers,
					ttnpb.RightsFrom(tt.Rights...),
				)
				a.So(err, should.BeNil)

				members, err := store.FindMembers(ctx, tt.Identifiers)
				a.So(err, should.BeNil)
				if a.So(members, should.HaveLength, 1) {
					for ouid, rights := range members {
						a.So(ouid, should.Resemble, usrIDs)
						a.So(rights.GetRights(), should.Resemble, tt.Rights)
					}
				}

				memberRights, err := store.FindMemberRights(ctx, usrIDs, tt.EntityType)
				a.So(err, should.BeNil)
				if a.So(memberRights, should.HaveLength, 1) {
					for eid, rights := range memberRights {
						a.So(eid, should.Resemble, tt.Identifiers)
						a.So(rights.GetRights(), should.Resemble, tt.Rights)
					}
				}
			})
		}

		for _, tt := range []struct {
			Name        string
			Identifiers *ttnpb.EntityIdentifiers
			Rights      []ttnpb.Right
			EntityType  string
		}{
			{
				Name:        "Organization-Application",
				Identifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
				EntityType:  "application",
			},
			{
				Name:        "Organization-Client",
				Identifiers: ttnpb.ClientIdentifiers{ClientID: "test-cli"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
				EntityType:  "client",
			},
			{
				Name:        "Organization-Gateway",
				Identifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gtw"}.EntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO},
				EntityType:  "gateway",
			},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				a := assertions.New(t)

				err := store.SetMember(ctx,
					orgIDs,
					tt.Identifiers,
					ttnpb.RightsFrom(tt.Rights...),
				)
				a.So(err, should.BeNil)

				members, err := store.FindMembers(ctx, tt.Identifiers)
				a.So(err, should.BeNil)
				if a.So(members, should.HaveLength, 2) {
					for ouid, rights := range members {
						if ouid.GetUserIDs() != nil {
							continue
						}
						a.So(ouid, should.Resemble, orgIDs)
						a.So(rights.GetRights(), should.Resemble, tt.Rights)
					}
				}

				memberRights, err := store.FindMemberRights(ctx, orgIDs, tt.EntityType)
				a.So(err, should.BeNil)
				if a.So(memberRights, should.HaveLength, 1) {
					for eid, rights := range memberRights {
						a.So(eid, should.Resemble, tt.Identifiers)
						a.So(rights.GetRights(), should.Resemble, tt.Rights)
					}
				}
			})
		}

		// TODO: Try with entities that don't exist

		// TODO: Try making org member of org

		// TODO: Test membership rights update

		// TODO: Test membership delete (zero rights)

	})
}
