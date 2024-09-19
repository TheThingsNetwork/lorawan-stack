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

package storetest

import (
	"sort"
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type notificationsByCreatedAt []*ttnpb.Notification

func (a notificationsByCreatedAt) Len() int      { return len(a) }
func (a notificationsByCreatedAt) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a notificationsByCreatedAt) Less(i, j int) bool {
	return ttnpb.StdTime(a[i].CreatedAt).Before(*ttnpb.StdTime(a[j].CreatedAt))
}

func (st *StoreTest) TestNotificationStore(t *T) {
	usr1 := st.population.NewUser()
	usr2 := st.population.NewUser()

	app1 := st.population.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	dev1 := st.population.NewEndDevice(app1.GetIds())
	cli1 := st.population.NewClient(usr1.GetOrganizationOrUserIdentifiers())
	gtw1 := st.population.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	org1 := st.population.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.NotificationStore
	})
	defer st.DestroyDB(t, true)
	if !ok {
		t.Skip("Store does not implement NotificationStore")
	}
	defer s.Close()

	notificationData, _ := anypb.New(&wrapperspb.StringValue{Value: "test"})

	var notifications []*ttnpb.Notification

	t.Run("CreateNotification", func(t *T) {
		a, ctx := test.New(t)
		start := time.Now().Truncate(time.Second)

		for _, ids := range []interface {
			GetEntityIdentifiers() *ttnpb.EntityIdentifiers
		}{
			app1.GetIds(),
			dev1.GetIds(),
			cli1.GetIds(),
			gtw1.GetIds(),
			org1.GetIds(),
			usr1.GetIds(),
		} {
			created, err := s.CreateNotification(ctx, &ttnpb.Notification{
				EntityIds:        ids.GetEntityIdentifiers(),
				NotificationType: ttnpb.NotificationType_UNKNOWN,
				Data:             notificationData,
				SenderIds:        usr1.GetIds(),
				Receivers:        []ttnpb.NotificationReceiver{ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_COLLABORATOR},
				Email:            true,
			}, []*ttnpb.UserIdentifiers{usr1.GetIds(), usr2.GetIds()})

			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Id, should.NotBeBlank)
				a.So(created.EntityIds, should.Resemble, ids.GetEntityIdentifiers())
				a.So(created.NotificationType, should.Equal, ttnpb.NotificationType_UNKNOWN)
				a.So(created.Data, should.Resemble, notificationData)
				a.So(created.SenderIds, should.Resemble, usr1.GetIds())
				a.So(created.Receivers, should.Resemble, []ttnpb.NotificationReceiver{ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_COLLABORATOR})
				a.So(created.Email, should.BeTrue)
				a.So(created.Status, should.Resemble, ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN)
				a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
				a.So(*ttnpb.StdTime(created.StatusUpdatedAt), should.HappenWithin, 5*time.Second, start)
			}

			notifications = append(notifications, created)

			time.Sleep(1 * time.Millisecond) // The tests depend on sorting by created_at, so we don't want multiple notifications with the same time.
		}
	})

	sort.Sort(sort.Reverse(notificationsByCreatedAt(notifications)))

	t.Run("ListNotifications_usr1", func(t *T) {
		a, ctx := test.New(t)

		got, err := s.ListNotifications(ctx, usr1.GetIds(), []ttnpb.NotificationStatus{ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN})
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 6) {
			a.So(got, should.Resemble, notifications)
		}
	})

	t.Run("UpdateNotificationStatus", func(t *T) {
		a, ctx := test.New(t)

		ids := make([]string, len(notifications))
		for i, notification := range notifications {
			ids[i] = notification.Id
		}

		err := s.UpdateNotificationStatus(ctx, usr1.GetIds(), ids, ttnpb.NotificationStatus_NOTIFICATION_STATUS_SEEN)
		a.So(err, should.BeNil)
	})

	t.Run("ListNotifications_AfterStatusUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListNotifications(ctx, usr1.GetIds(), []ttnpb.NotificationStatus{ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN})
		if a.So(err, should.BeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("ListNotifications_usr2", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		paginateCtx := store.WithPagination(ctx, 3, 1, &total)

		got, err := s.ListNotifications(paginateCtx, usr2.GetIds(), []ttnpb.NotificationStatus{ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN})
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 3) {
			a.So(got, should.Resemble, notifications[:3])
		}

		a.So(total, should.Equal, 6)
	})
}
