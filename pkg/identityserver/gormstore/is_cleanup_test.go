// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestCleanupExpiredEntities(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		s := newStore(db)
		prepareTest(db,
			&Account{}, &User{}, &Organization{},
			&Application{}, &Gateway{}, &Client{},
		)
		appStore := GetApplicationStore(db)
		gtwStore := GetGatewayStore(db)
		usrStore := GetUserStore(db)
		orgStore := GetOrganizationStore(db)
		cliStore := GetClientStore(db)

		app1 := &Application{ApplicationID: "test-app-1"}
		s.createEntity(ctx, app1)
		app2 := &Application{ApplicationID: "test-app-2"}
		s.createEntity(ctx, app2)
		app3 := &Application{ApplicationID: "test-app-3"}
		s.createEntity(ctx, app3)
		usr1 := &User{Account: Account{UID: "test-user-1"}}
		s.createEntity(ctx, usr1)
		usr2 := &User{Account: Account{UID: "test-user-2"}}
		s.createEntity(ctx, usr2)
		usr3 := &User{Account: Account{UID: "test-user-3"}}
		s.createEntity(ctx, usr3)
		org1 := &Organization{Account: Account{UID: "test-org-1"}}
		s.createEntity(ctx, org1)
		org2 := &Organization{Account: Account{UID: "test-org-2"}}
		s.createEntity(ctx, org2)
		org3 := &Organization{Account: Account{UID: "test-org-3"}}
		s.createEntity(ctx, org3)
		gtw1 := &Gateway{GatewayID: "test-gtw-1"}
		err := s.createEntity(ctx, gtw1)
		gtw2 := &Gateway{GatewayID: "test-gtw-2"}
		s.createEntity(ctx, gtw2)
		gtw3 := &Gateway{GatewayID: "test-gtw-3"}
		s.createEntity(ctx, gtw3)
		cli1 := &Client{ClientID: "test-cli-1"}
		s.createEntity(ctx, cli1)
		cli2 := &Client{ClientID: "test-cli-2"}
		s.createEntity(ctx, cli2)
		cli3 := &Client{ClientID: "test-cli-3"}
		s.createEntity(ctx, cli3)

		expiredCtx := WithExpired(ctx, time.Second)
		expiredCtx = WithSoftDeleted(expiredCtx, true)
		appStore.DeleteApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: app1.ApplicationID})
		time.Sleep(time.Second)
		appStore.DeleteApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: app2.ApplicationID})

		expiredApplications, err := appStore.FindApplications(expiredCtx, []*ttnpb.ApplicationIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
		if a.So(err, should.BeNil) {
			if a.So(expiredApplications, should.HaveLength, 1) {
				a.So(expiredApplications[0].GetIds().GetApplicationId(), should.Equal, app1.ApplicationID)
			}
		}
		gtwStore.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: gtw1.GatewayID})
		time.Sleep(time.Second)
		gtwStore.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: gtw2.GatewayID})
		expiredGateways, err := gtwStore.FindGateways(expiredCtx, []*ttnpb.GatewayIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
		if a.So(err, should.BeNil) {
			if a.So(expiredGateways, should.HaveLength, 1) {
				a.So(expiredGateways[0].GetIds().GetGatewayId(), should.Equal, gtw1.GatewayID)
			}
		}

		usrStore.DeleteUser(ctx, &ttnpb.UserIdentifiers{UserId: usr1.Account.UID})
		time.Sleep(time.Second)
		usrStore.DeleteUser(ctx, &ttnpb.UserIdentifiers{UserId: usr2.Account.UID})
		expiredUsers, err := usrStore.FindUsers(expiredCtx, []*ttnpb.UserIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
		if a.So(err, should.BeNil) {
			if a.So(expiredUsers, should.HaveLength, 1) {
				a.So(expiredUsers[0].GetIds().GetUserId(), should.Equal, usr1.Account.UID)
			}
		}

		orgStore.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: org1.Account.UID})
		time.Sleep(time.Second)
		orgStore.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: org2.Account.UID})
		expiredOrganizations, err := orgStore.FindOrganizations(expiredCtx, []*ttnpb.OrganizationIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
		if a.So(err, should.BeNil) {
			if a.So(expiredOrganizations, should.HaveLength, 1) {
				a.So(expiredOrganizations[0].GetIds().GetOrganizationId(), should.Equal, org1.Account.UID)
			}
		}

		cliStore.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientId: cli1.ClientID})
		time.Sleep(time.Second)
		cliStore.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientId: cli2.ClientID})
		expiredClients, err := cliStore.FindClients(expiredCtx, []*ttnpb.ClientIdentifiers{}, &types.FieldMask{Paths: []string{"ids", "deleted_at"}})
		if a.So(err, should.BeNil) {
			if a.So(expiredClients, should.HaveLength, 1) {
				a.So(expiredClients[0].GetIds().GetClientId(), should.Equal, cli1.ClientID)
			}
		}
	})
}
