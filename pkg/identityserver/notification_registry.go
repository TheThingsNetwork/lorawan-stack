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
	"context"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	evtNotificationCreate = events.Define(
		"user.notification.create", "create notification",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_NOTIFICATIONS_READ),
		events.WithDataType(&ttnpb.Notification{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtNotificationUpdateStatus = events.Define(
		"user.notification.update_status", "update notification status",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_NOTIFICATIONS_READ),
		events.WithDataType(&ttnpb.UpdateNotificationStatusRequest{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func receiversContains(receivers []ttnpb.NotificationReceiver, search ttnpb.NotificationReceiver) bool {
	for _, receiver := range receivers {
		if receiver == search {
			return true
		}
	}
	return false
}

func notificationTypeAllowed(notificationType ttnpb.NotificationType, allowedNotifications []ttnpb.NotificationType) bool { //nolint:lll
	for _, allowedType := range allowedNotifications {
		if notificationType == allowedType {
			return true
		}
	}
	return false
}

func filterAllowedEmailReveivers(emailReceiverUsers []*ttnpb.User, notificationType ttnpb.NotificationType) []*ttnpb.UserIdentifiers { //nolint:lll
	var emailReceiverIDs []*ttnpb.UserIdentifiers

	// Collect IDs of users that have email notifications enabled for that notification type.
	for _, user := range emailReceiverUsers {
		userNotificationPreferences := user.GetEmailNotificationPreferences().GetTypes()
		if notificationTypeAllowed(notificationType, userNotificationPreferences) {
			emailReceiverIDs = append(emailReceiverIDs, user.GetIds())
		}
	}
	return emailReceiverIDs
}

func uniqueOrganizationOrUserIdentifiers(ctx context.Context, ids []*ttnpb.OrganizationOrUserIdentifiers) []*ttnpb.OrganizationOrUserIdentifiers {
	out := make([]*ttnpb.OrganizationOrUserIdentifiers, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idString := unique.ID(ctx, id)
		if _, seen := seen[idString]; seen {
			continue
		}
		out = append(out, id)
		seen[idString] = struct{}{}
	}
	return out
}

func filterUserIdentifiers(ids []*ttnpb.OrganizationOrUserIdentifiers) []*ttnpb.UserIdentifiers {
	out := make([]*ttnpb.UserIdentifiers, 0, len(ids))
	for _, id := range ids {
		if id.EntityType() != "user" {
			continue
		}
		out = append(out, id.GetUserIds())
	}
	return out
}

func (is *IdentityServer) notifyInternal(ctx context.Context, req *ttnpb.CreateNotificationRequest) error {
	if err := req.ValidateFields(); err != nil {
		panic(err)
	}
	ctx = is.FromRequestContext(ctx)
	if authInfo, err := is.authInfo(ctx); err == nil {
		if userIDs := authInfo.GetEntityIdentifiers().GetUserIds(); userIDs != nil {
			req.SenderIds = userIDs
		}
	}
	_, err := is.createNotification(clusterauth.NewContext(ctx, nil), req) // just call the RPC with cluster auth.
	return err
}

var errNoReceiverUserIDs = errors.Define("no_receiver_user_ids", "no receiver users ids")

// getContactReceivers checks if the entityID to provide the appropriate receiverID.
// If is an user, returns the entityID.
// If is an organization, it checks if the fanout_notifications is enabled. If enabled returns the organizationID but
// otherwise it returns the organization's administrative or technical contact.
func (is *IdentityServer) getContactReceivers(
	ctx context.Context, entityID *ttnpb.OrganizationOrUserIdentifiers, entityMask []string,
) (*ttnpb.OrganizationOrUserIdentifiers, error) {
	if entityID.EntityType() != "organization" {
		return entityID, nil
	}
	org, err := is.store.GetOrganization(
		ctx, entityID.GetOrganizationIds(), append(entityMask, "fanout_notifications"),
	)
	if err != nil {
		return nil, err
	}
	if org.FanoutNotifications {
		return entityID, nil
	}
	if contact := org.GetAdministrativeContact(); contact != nil {
		return contact, nil
	}
	if contact := org.GetTechnicalContact(); contact != nil {
		return contact, nil
	}
	return entityID, nil
}

func (is *IdentityServer) lookupNotificationReceivers( //nolint:gocyclo
	ctx context.Context,
	req *ttnpb.CreateNotificationRequest,
) ([]*ttnpb.UserIdentifiers, []*ttnpb.UserIdentifiers, error) {
	var receiverIDs []*ttnpb.OrganizationOrUserIdentifiers
	var receiverUserIDs []*ttnpb.UserIdentifiers
	var emailReceiverIDs []*ttnpb.UserIdentifiers
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		// Collect user ID for user notifications.
		if req.EntityIds.EntityType() == "user" {
			receiverIDs = append(receiverIDs, req.EntityIds.GetUserIds().GetOrganizationOrUserIdentifiers())
		}
		// Collect ids of administrative/technical contacts.
		var entityMask []string
		if receiversContains(req.Receivers, ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT) {
			entityMask = append(entityMask, "administrative_contact")
		}
		if receiversContains(req.Receivers, ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_TECHNICAL_CONTACT) {
			entityMask = append(entityMask, "technical_contact")
		}
		if len(entityMask) > 0 {
			var (
				entity interface {
					GetAdministrativeContact() *ttnpb.OrganizationOrUserIdentifiers
					GetTechnicalContact() *ttnpb.OrganizationOrUserIdentifiers
				}
				err error
			)
			switch req.EntityIds.EntityType() {
			default:
				// Entity doesn't have contacts. Just ignore.
			case store.EntityApplication:
				entity, err = st.GetApplication(ctx, req.EntityIds.GetApplicationIds(), entityMask)
			case store.EntityClient:
				entity, err = st.GetClient(ctx, req.EntityIds.GetClientIds(), entityMask)
			case store.EntityEndDevice:
				entity, err = st.GetApplication(ctx, req.EntityIds.GetDeviceIds().GetApplicationIds(), entityMask)
			case store.EntityGateway:
				entity, err = st.GetGateway(ctx, req.EntityIds.GetGatewayIds(), entityMask)
			case store.EntityOrganization:
				entity, err = st.GetOrganization(ctx, req.EntityIds.GetOrganizationIds(), entityMask)
			}
			if err != nil {
				return err
			}
			if entity != nil { // NOTE: entity is nil for entities that don't support contacts.
				adminContact, err := is.getContactReceivers(
					ctx, entity.GetAdministrativeContact(), []string{"administrative_contact"},
				)
				if err != nil {
					return err
				}
				if adminContact != nil {
					receiverIDs = append(receiverIDs, adminContact)
				}
				techContact, err := is.getContactReceivers(
					ctx, entity.GetTechnicalContact(), []string{"technical_contact"},
				)
				if err != nil {
					return err
				}
				if techContact != nil {
					receiverIDs = append(receiverIDs, techContact)
				}
			}
		}

		// Collect IDs of entity collaborators.
		if receiversContains(req.Receivers, ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_COLLABORATOR) {
			switch req.EntityIds.EntityType() {
			default:
				// Entity doesn't have collaborators. Just ignore.
			case store.EntityApplication, store.EntityClient, store.EntityGateway, store.EntityOrganization:
				members, err := st.FindMembers(ctx, req.EntityIds)
				if err != nil {
					return err
				}
				for _, v := range members {
					receiverIDs = append(receiverIDs, v.Ids)
				}
			}
		}

		// Expand organization IDs to organization collaborator IDs.
		for _, ids := range uniqueOrganizationOrUserIdentifiers(ctx, receiverIDs) {
			if ids.EntityType() != store.EntityOrganization {
				continue
			}
			members, err := st.FindMembers(ctx, ids.GetEntityIdentifiers())
			if err != nil {
				return err
			}
			for _, v := range members {
				receiverIDs = append(receiverIDs, v.Ids)
			}
		}

		// Filter only user identifiers and remove duplicates.
		receiverUserIDs = filterUserIdentifiers(uniqueOrganizationOrUserIdentifiers(ctx, receiverIDs))
		// Get the email notification preferences of the receiver users.
		emailReceiverUsers, _ := st.FindUsers(ctx, receiverUserIDs, []string{"email_notification_preferences"})
		// Filter only the users that have email notifications enabled for the notification type.
		emailReceiverIDs = filterAllowedEmailReveivers(emailReceiverUsers, req.NotificationType)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if len(receiverUserIDs) == 0 {
		return nil, nil, errNoReceiverUserIDs.New()
	}

	return receiverUserIDs, emailReceiverIDs, nil
}

func (is *IdentityServer) storeNotification(ctx context.Context, req *ttnpb.CreateNotificationRequest, receiverUserIDs ...*ttnpb.UserIdentifiers) (*ttnpb.Notification, error) {
	var notification *ttnpb.Notification
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		notification, err = st.CreateNotification(ctx, &ttnpb.Notification{
			EntityIds:        req.EntityIds,
			NotificationType: req.NotificationType,
			Data:             req.Data,
			SenderIds:        req.SenderIds,
			Receivers:        req.Receivers,
		}, receiverUserIDs)
		return err
	})
	if err != nil {
		return nil, err
	}
	return notification, nil
}

func (is *IdentityServer) createNotification(ctx context.Context, req *ttnpb.CreateNotificationRequest) (*ttnpb.CreateNotificationResponse, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	receiverUserIDs, emailReceiverIDs, err := is.lookupNotificationReceivers(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(emailReceiverIDs) > 0 && email.GetNotification(ctx, req.GetNotificationType()) == nil {
		log.FromContext(ctx).WithField("notification_type", req.GetNotificationType()).Warn("email template for notification not registered") // nolint:lll
		emailReceiverIDs = nil
	}

	notification, err := is.storeNotification(ctx, req, receiverUserIDs...)
	if err != nil {
		return nil, err
	}

	if emailReceiverIDs != nil || len(emailReceiverIDs) > 0 {
		if err := is.SendNotificationEmailToUserIDs(ctx, notification, emailReceiverIDs...); err != nil {
			return nil, err
		}
	}

	evs := make([]events.Event, 0, len(receiverUserIDs))
	for _, ids := range receiverUserIDs {
		evs = append(evs, evtNotificationCreate.NewWithIdentifiersAndData(ctx, ids, notification))
	}
	events.Publish(evs...)

	return &ttnpb.CreateNotificationResponse{
		Id: notification.Id,
	}, nil
}

func (is *IdentityServer) notifyAdminsInternal(ctx context.Context, req *ttnpb.CreateNotificationRequest) error {
	if err := req.ValidateFields(); err != nil {
		panic(err)
	}

	ctx = is.FromRequestContext(ctx)
	if authInfo, err := is.authInfo(ctx); err == nil {
		if userIDs := authInfo.GetEntityIdentifiers().GetUserIds(); userIDs != nil {
			req.SenderIds = userIDs
		}
	}

	if email.GetNotification(ctx, req.GetNotificationType()) == nil {
		log.FromContext(ctx).WithField("notification_type", req.GetNotificationType()).Warn("email template for notification not registered")
	}
	var receivers []*ttnpb.User
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		receivers, err = st.ListAdmins(ctx, notificationEmailUserFields)
		return err
	})
	if err != nil {
		return err
	}

	receiverUserIDs := make([]*ttnpb.UserIdentifiers, len(receivers))
	for i, receiver := range receivers {
		receiverUserIDs[i] = receiver.Ids
	}

	notification, err := is.storeNotification(ctx, req, receiverUserIDs...)
	if err != nil {
		return err
	}

	err = is.SendNotificationEmailToUsers(ctx, notification, receivers...)
	if err != nil {
		return err
	}

	return nil
}

func (is *IdentityServer) listNotifications(ctx context.Context, req *ttnpb.ListNotificationsRequest) (*ttnpb.ListNotificationsResponse, error) {
	if err := rights.RequireUser(ctx, req.ReceiverIds, ttnpb.Right_RIGHT_USER_NOTIFICATIONS_READ); err != nil {
		return nil, err
	}
	res := &ttnpb.ListNotificationsResponse{}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		var total uint64
		paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
		defer func() {
			if err == nil {
				setTotalHeader(ctx, total)
			}
		}()
		res.Notifications, err = st.ListNotifications(paginateCtx, req.ReceiverIds, req.Status)
		return err
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (is *IdentityServer) updateNotificationStatus(ctx context.Context, req *ttnpb.UpdateNotificationStatusRequest) (*emptypb.Empty, error) {
	if err := rights.RequireUser(ctx, req.ReceiverIds, ttnpb.Right_RIGHT_USER_NOTIFICATIONS_READ); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		return st.UpdateNotificationStatus(ctx, req.ReceiverIds, req.Ids, req.Status)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtNotificationUpdateStatus.NewWithIdentifiersAndData(ctx, req.ReceiverIds, req))
	return ttnpb.Empty, nil
}

type notificationRegistry struct {
	ttnpb.UnimplementedNotificationServiceServer

	*IdentityServer
}

func (cr *notificationRegistry) Create(ctx context.Context, req *ttnpb.CreateNotificationRequest) (*ttnpb.CreateNotificationResponse, error) {
	return cr.createNotification(ctx, req)
}

func (cr *notificationRegistry) List(ctx context.Context, req *ttnpb.ListNotificationsRequest) (*ttnpb.ListNotificationsResponse, error) {
	return cr.listNotifications(ctx, req)
}

func (cr *notificationRegistry) UpdateStatus(ctx context.Context, req *ttnpb.UpdateNotificationStatusRequest) (*emptypb.Empty, error) {
	return cr.updateNotificationStatus(ctx, req)
}
