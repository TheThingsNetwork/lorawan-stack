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
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Notification is the notification model in the database.
type Notification struct {
	bun.BaseModel `bun:"table:notifications,alias:not"`

	Model

	// EntityType is "application", "client", "end_device", "gateway", "organization" or "user".
	EntityType string `bun:"entity_type,notnull"`
	// EntityID is Application.ID, Client.ID, EndDevice.ID, Gateway.ID, Organization.ID or User.ID.
	EntityID string `bun:"entity_id,notnull"`

	// EntityUID is a copy of the human-readable entity ID, so that we can keep notifications for deleted entities.
	EntityUID string `bun:"entity_uid,notnull"`

	NotificationType string `bun:"notification_type,notnull"`

	Data json.RawMessage `bun:"data,nullzero"`

	SenderID *string `bun:"sender_id"`

	// SenderUID is a copy of the human-readable sender ID, so that we can keep notifications for deleted senders.
	SenderUID string `bun:"sender_uid,notnull"`

	Receivers []int `bun:"receivers,array,nullzero"`

	Email bool `bun:"email,notnull"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Notification) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

// NotificationReceiver is the notification receiver model in the database.
type NotificationReceiver struct {
	bun.BaseModel `bun:"table:notification_receivers,alias:rec"`

	Notification   *Notification `bun:"rel:belongs-to,join:notification_id=id"`
	NotificationID string        `bun:"notification_id,notnull"`

	Receiver   *User  `bun:"rel:belongs-to,join:receiver_id=id"`
	ReceiverID string `bun:"receiver_id,notnull"`

	Status          int       `bun:"status,notnull"`
	StatusUpdatedAt time.Time `bun:"status_updated_at,notnull"`
}

func (NotificationReceiver) _isModel() {} // It doesn't embed Model, but it's still a model.

func notificationToPB(m *Notification, r *NotificationReceiver) (*ttnpb.Notification, error) {
	pb := &ttnpb.Notification{
		Id:               m.ID,
		CreatedAt:        timestamppb.New(m.CreatedAt),
		EntityIds:        getEntityIdentifiers(m.EntityType, m.EntityUID),
		NotificationType: m.NotificationType,
		Receivers:        convertIntSlice[int, ttnpb.NotificationReceiver](m.Receivers),
	}
	if len(m.Data) > 0 {
		anyPB := &anypb.Any{}
		err := jsonpb.TTN().Unmarshal(m.Data, anyPB)
		if err != nil {
			return nil, err
		}
		pb.Data = anyPB
	}
	if m.SenderUID != "" {
		pb.SenderIds = &ttnpb.UserIdentifiers{UserId: m.SenderUID}
	}
	if r != nil {
		pb.Status = ttnpb.NotificationStatus(r.Status)
		pb.StatusUpdatedAt = timestamppb.New(r.StatusUpdatedAt)
	}
	return pb, nil
}

type notificationStore struct {
	*entityStore
}

func newNotificationStore(baseStore *baseStore) *notificationStore {
	return &notificationStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *notificationStore) CreateNotification(
	ctx context.Context, pb *ttnpb.Notification, receiverIDs []*ttnpb.UserIdentifiers,
) (*ttnpb.Notification, error) {
	ctx, span := tracer.StartFromContext(ctx, "CreateNotification", trace.WithAttributes(
		attribute.String("entity_type", pb.EntityIds.EntityType()),
		attribute.String("entity_id", pb.EntityIds.IDString()),
		attribute.String("notification_type", pb.NotificationType),
	))
	if pb.SenderIds != nil {
		span.SetAttributes(attribute.String("sender_id", pb.SenderIds.GetUserId()))
	}
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, pb.EntityIds)
	if err != nil {
		return nil, err
	}

	model := &Notification{
		EntityType:       entityType,
		EntityID:         entityUUID,
		EntityUID:        pb.GetEntityIds().IDString(),
		NotificationType: pb.NotificationType,
		Receivers:        convertIntSlice[ttnpb.NotificationReceiver, int](pb.Receivers),
	}

	if pb.Data != nil {
		dataJSON, err := jsonpb.TTN().Marshal(pb.Data)
		if err != nil {
			return nil, err
		}
		model.Data = dataJSON
	}

	if pb.SenderIds != nil {
		_, senderUUID, err := s.getEntity(ctx, pb.SenderIds)
		if err != nil {
			return nil, err
		}
		model.SenderID = &senderUUID
		model.SenderUID = pb.SenderIds.GetUserId()
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	receiverIDStrings := make([]string, len(receiverIDs))
	for i, receiverID := range receiverIDs {
		receiverIDStrings[i] = receiverID.GetUserId()
	}
	receiverUUIDs, err := s.getEntityUUIDs(ctx, "user", receiverIDStrings...)
	if err != nil {
		return nil, err
	}

	receivers := make([]*NotificationReceiver, len(receiverUUIDs))
	for i, receiverUUID := range receiverUUIDs {
		receiver := &NotificationReceiver{
			NotificationID:  model.ID,
			ReceiverID:      receiverUUID,
			Status:          int(pb.Status),
			StatusUpdatedAt: model.CreatedAt,
		}
		receivers[i] = receiver
	}

	_, err = s.DB.NewInsert().
		Model(&receivers).
		Exec(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	pb, err = notificationToPB(model, &NotificationReceiver{
		Status:          int(pb.Status),
		StatusUpdatedAt: model.CreatedAt,
	})
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *notificationStore) ListNotifications(
	ctx context.Context, receiverIDs *ttnpb.UserIdentifiers, statuses []ttnpb.NotificationStatus,
) ([]*ttnpb.Notification, error) {
	ctx, span := tracer.StartFromContext(ctx, "ListNotifications", trace.WithAttributes(
		attribute.String("receiver_id", receiverIDs.GetUserId()),
	))
	defer span.End()

	_, receiverUUID, err := s.getEntity(ctx, receiverIDs)
	if err != nil {
		return nil, err
	}

	models := []*NotificationReceiver{}
	selectQuery := newSelectModels(ctx, s.DB, &models).
		Where("receiver_id = ?", receiverUUID)

	if len(statuses) > 0 {
		selectQuery = selectQuery.Where(
			"status IN (?)",
			bun.In(convertIntSlice[ttnpb.NotificationStatus, int](statuses)),
		)
	}

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Order("created_at DESC").
		Apply(selectWithLimitAndOffsetFromContext(ctx))

		// Include the notification.
	selectQuery = selectQuery.
		Relation("Notification")

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.Notification, len(models))
	for i, model := range models {
		pb, err := notificationToPB(model.Notification, model)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *notificationStore) UpdateNotificationStatus(
	ctx context.Context,
	receiverIDs *ttnpb.UserIdentifiers,
	notificationIDs []string,
	status ttnpb.NotificationStatus,
) error {
	ctx, span := tracer.StartFromContext(ctx, "UpdateNotificationStatus", trace.WithAttributes(
		attribute.String("receiver_id", receiverIDs.GetUserId()),
	))
	defer span.End()

	_, receiverUUID, err := s.getEntity(ctx, receiverIDs)
	if err != nil {
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(&NotificationReceiver{}).
		Where("receiver_id = ?", receiverUUID).
		Where("notification_id IN (?)", bun.In(notificationIDs)).
		Set("status = ?, status_updated_at = NOW()", int(status)).
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}
