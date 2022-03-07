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

package store

import (
	"context"
	"runtime/trace"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetNotificationStore returns an NotificationStore on the given db (or transaction).
func GetNotificationStore(db *gorm.DB) store.NotificationStore {
	return &notificationStore{baseStore: newStore(db)}
}

type notificationStore struct {
	*baseStore
}

func (s *notificationStore) CreateNotification(ctx context.Context, pb *ttnpb.Notification, receiverIDs []*ttnpb.UserIdentifiers) (*ttnpb.Notification, error) {
	defer trace.StartRegion(ctx, "create notification").End()

	entityModel, err := s.findEntity(ctx, pb.EntityIds, "id")
	if err != nil {
		return nil, err
	}

	model := notificationWithStatus{
		Notification: Notification{
			EntityType:       entityTypeForID(pb.EntityIds),
			EntityID:         entityModel.PrimaryKey(),
			NotificationType: pb.NotificationType,
		},
		Status: int32(pb.Status),
	}

	if pb.Data != nil {
		dataJSON, err := jsonpb.TTN().Marshal(pb.Data)
		if err != nil {
			return nil, err
		}
		model.Data.RawMessage = dataJSON
	}

	if pb.SenderIds != nil {
		senderModel, err := s.findEntity(ctx, pb.SenderIds, "id")
		if err != nil {
			return nil, err
		}
		pk := senderModel.PrimaryKey()
		model.SenderID = &pk
	}

	model.Receivers = make(pq.Int32Array, len(pb.Receivers))
	for i, receiver := range pb.Receivers {
		model.Receivers[i] = int32(receiver)
	}

	receiverIDStrings := make([]string, len(receiverIDs))
	for i, receiverID := range receiverIDs {
		receiverIDStrings[i] = receiverID.GetUserId()
	}

	err = s.createEntity(ctx, &model.Notification)
	if err != nil {
		return nil, convertError(err)
	}
	model.StatusUpdatedAt = model.CreatedAt

	var receiverAccounts []Account
	err = s.query(ctx, Account{}).
		Where("uid IN (?)", receiverIDStrings).
		Find(&receiverAccounts).
		Error
	if err != nil {
		return nil, convertError(err)
	}

	// TODO: Create all NotificationReceivers at once (https://github.com/TheThingsNetwork/lorawan-stack/issues/3250).
	for _, receiverAccount := range receiverAccounts {
		err = s.createEntity(ctx, &NotificationReceiver{
			NotificationID:  model.ID,
			ReceiverID:      receiverAccount.AccountID,
			Status:          model.Status,
			StatusUpdatedAt: model.StatusUpdatedAt,
		})
		if err != nil {
			return nil, convertError(err)
		}
	}

	var res ttnpb.Notification
	res.EntityIds = pb.EntityIds
	res.Data = pb.Data
	res.SenderIds = pb.SenderIds
	if err = model.toPB(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *notificationStore) ListNotifications(ctx context.Context, receiverIDs *ttnpb.UserIdentifiers, statuses []ttnpb.NotificationStatus) ([]*ttnpb.Notification, error) {
	defer trace.StartRegion(ctx, "list notifications").End()

	userIDQuery := s.query(ctx, Account{}).
		Select("account_id").
		Where("uid = ?", receiverIDs.GetUserId()).
		SubQuery()

	query := s.query(ctx, Notification{}).
		Joins("JOIN notification_receivers ON notification_receivers.notification_id = notifications.id").
		Where("notification_receivers.receiver_id = ?", userIDQuery)

	if len(statuses) > 0 {
		statusInts := make([]int32, len(statuses))
		for i, status := range statuses {
			statusInts[i] = int32(status)
		}
		query = query.Where("notification_receivers.status IN (?)", statusInts)
	}

	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}

	query = query.
		Select([]string{
			"notifications.*",
			"notification_receivers.status",
			"notification_receivers.status_updated_at",
			"sender.uid AS friendly_sender_id",
			"app.application_id AS friendly_application_id",
			"dev.device_id AS friendly_end_device_id",
			"dev.application_id AS friendly_end_device_application_id",
			"cli.client_id AS friendly_client_id",
			"gtw.gateway_id AS friendly_gateway_id",
			"org.uid AS friendly_organization_id",
			"usr.uid AS friendly_user_id",
		}).
		Joins("LEFT JOIN accounts AS sender ON sender.account_type = 'user' AND sender.account_id = notifications.sender_id").
		Joins("LEFT JOIN applications AS app ON notifications.entity_type = 'application' AND app.id = notifications.entity_id").
		Joins("LEFT JOIN clients AS cli ON notifications.entity_type = 'client' AND cli.id = notifications.entity_id").
		Joins("LEFT JOIN end_devices AS dev ON notifications.entity_type = 'end_device' AND dev.id = notifications.entity_id").
		Joins("LEFT JOIN gateways AS gtw ON notifications.entity_type = 'gateway' AND gtw.id = notifications.entity_id").
		Joins("LEFT JOIN accounts AS org ON notifications.entity_type = 'organization' AND org.account_type = 'organization' AND org.account_id = notifications.entity_id").
		Joins("LEFT JOIN accounts AS usr ON notifications.entity_type = 'user' AND usr.account_type = 'user' AND usr.account_id = notifications.entity_id").
		Order("notifications.created_at DESC")

	var notificationModels []notificationWithStatus
	err := query.Scan(&notificationModels).Error
	if err != nil {
		return nil, convertError(err)
	}

	res := make([]*ttnpb.Notification, len(notificationModels))
	for i, notificationModel := range notificationModels {
		var pb ttnpb.Notification
		if err = notificationModel.toPB(&pb); err != nil {
			return nil, err
		}
		res[i] = &pb
	}

	return res, nil
}

func (s *notificationStore) UpdateNotificationStatus(ctx context.Context, receiverIDs *ttnpb.UserIdentifiers, notificationIDs []string, status ttnpb.NotificationStatus) error {
	userIDQuery := s.query(ctx, Account{}).
		Select("account_id").
		Where("uid = ?", receiverIDs.GetUserId()).
		SubQuery()

	query := s.query(ctx, NotificationReceiver{}).
		Where("receiver_id = ?", userIDQuery).
		Where("notification_id IN (?)", notificationIDs).
		Updates(NotificationReceiver{
			Status:          int32(status),
			StatusUpdatedAt: cleanTime(time.Now()),
		})

	if err := query.Error; err != nil {
		return convertError(err)
	}
	return nil
}
