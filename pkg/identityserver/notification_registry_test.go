// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestNotificationRegistry(t *testing.T) {
	p := &storetest.Population{}
	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	dev1 := p.NewEndDevice(app1.GetIds())
	cli1 := p.NewClient(usr1.GetOrganizationOrUserIdentifiers())
	gtw1 := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	org1 := p.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())

	entityIDs := []*ttnpb.EntityIdentifiers{
		app1.GetIds().GetEntityIdentifiers(),
		dev1.GetIds().GetEntityIdentifiers(),
		cli1.GetIds().GetEntityIdentifiers(),
		gtw1.GetIds().GetEntityIdentifiers(),
		org1.GetIds().GetEntityIdentifiers(),
		usr1.GetIds().GetEntityIdentifiers(),
	}

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		svc := ttnpb.NewNotificationServiceClient(cc)

		for _, entityIds := range entityIDs {
			res, err := svc.Create(ctx, &ttnpb.CreateNotificationRequest{
				EntityIds:        entityIds,
				NotificationType: ttnpb.NotificationType_UNKNOWN,
				Receivers: []ttnpb.NotificationReceiver{
					ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_COLLABORATOR,
					ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
					ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_TECHNICAL_CONTACT,
				},
			}, is.Component.WithClusterAuth())
			if a.So(err, should.BeNil) && a.So(res, should.NotBeNil) {
				a.So(res.Id, should.NotBeZeroValue)
			}

			time.Sleep(test.Delay) // The tests depend on sorting by created_at, so we don't want multiple notifications with the same time.
		}

		var notificationIDs []string
		list, err := svc.List(ctx, &ttnpb.ListNotificationsRequest{
			ReceiverIds: usr1.GetIds(),
		}, usr1Creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Notifications, should.HaveLength, 6) {
			for i, notification := range list.Notifications {
				notificationIDs = append(notificationIDs, notification.Id)
				a.So(notification.EntityIds, should.Resemble, entityIDs[len(entityIDs)-i-1])
			}
		}

		_, err = svc.UpdateStatus(ctx, &ttnpb.UpdateNotificationStatusRequest{
			ReceiverIds: usr1.GetIds(),
			Ids:         notificationIDs,
			Status:      ttnpb.NotificationStatus_NOTIFICATION_STATUS_SEEN,
		}, usr1Creds)
		a.So(err, should.BeNil)

		list, err = svc.List(ctx, &ttnpb.ListNotificationsRequest{
			ReceiverIds: usr1.GetIds(),
			Status: []ttnpb.NotificationStatus{
				ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN,
			},
		}, usr1Creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Notifications, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}

// TestNotificationRegistryWithOrganizationFanout contains the test case for notifications that trigger the organization
// fanout, validating the expected behavior. More details below.
//
// Base conditions:
//   - Three users, the owner and two collaborators.
//   - Two organizations, one with fanout and another without.
//   - Two applications, one for each organization. The organization is the owner and admin/tech contact.
//
// Expected behavior:
//   - Create notification for app-1 (no fanout). Only the owner should receive the notification.
//   - Create notification for app2-2 (has fanout). Both the owner and collaborators should receive the notification.
func TestNotificationRegistryWithOrganizationFanout(t *testing.T) {
	t.Parallel()
	p := &storetest.Population{}

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	usr2 := p.NewUser()
	usr2Key, _ := p.NewAPIKey(usr2.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr2Creds := rpcCreds(usr2Key)
	usr3 := p.NewUser()
	usr3Key, _ := p.NewAPIKey(usr3.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr3Creds := rpcCreds(usr3Key)

	org1 := p.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	org2 := p.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	org2.FanoutNotifications = true

	// Register users as collaborators on both organizations.
	for _, org := range []*ttnpb.Organization{org1, org2} {
		org := org
		for _, collab := range []*ttnpb.User{usr2, usr3} {
			collab := collab
			p.NewMembership(
				collab.GetOrganizationOrUserIdentifiers(),
				org.GetEntityIdentifiers(),
				ttnpb.Right_RIGHT_ORGANIZATION_ALL,
			)
		}
	}

	// Sends notification to application's admin/tech contacts.
	// Depending on the organization's fanout setting, the amount of receivers will vary.
	app1 := p.NewApplication(org1.GetOrganizationOrUserIdentifiers())
	app2 := p.NewApplication(org2.GetOrganizationOrUserIdentifiers())

	a, ctx := test.New(t)
	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		svc := ttnpb.NewNotificationServiceClient(cc)

		for _, entityID := range []*ttnpb.EntityIdentifiers{
			app1.GetIds().GetEntityIdentifiers(),
			app2.GetIds().GetEntityIdentifiers(),
		} {
			entityID := entityID
			res, err := svc.Create(ctx, &ttnpb.CreateNotificationRequest{
				EntityIds:        entityID,
				NotificationType: ttnpb.NotificationType_UNKNOWN,
				Receivers: []ttnpb.NotificationReceiver{
					ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
					ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_TECHNICAL_CONTACT,
				},
			}, is.Component.WithClusterAuth())
			if a.So(err, should.BeNil) && a.So(res, should.NotBeNil) {
				a.So(res.Id, should.NotBeZeroValue)
			}
		}

		// Validates the amount of notifications received by each user.
		for _, tt := range []struct {
			userID             *ttnpb.UserIdentifiers
			notificationAmount int
			creds              grpc.CallOption
		}{
			{userID: usr1.GetIds(), notificationAmount: 2, creds: usr1Creds},
			{userID: usr2.GetIds(), notificationAmount: 1, creds: usr2Creds},
			{userID: usr3.GetIds(), notificationAmount: 1, creds: usr3Creds},
		} {
			tt := tt
			ttName := fmt.Sprintf("Expect %s to have %d notifications", tt.userID.GetUserId(), tt.notificationAmount)
			t.Run(ttName, func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				list, err := svc.List(ctx, &ttnpb.ListNotificationsRequest{ReceiverIds: tt.userID}, tt.creds)
				a.So(err, should.BeNil)
				a.So(list, should.NotBeNil)
				a.So(list.Notifications, should.HaveLength, tt.notificationAmount)
			})
		}
	}, withPrivateTestDatabase(p))
}
