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
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Notification model.
type Notification struct {
	Model

	EntityID   string `gorm:"type:UUID;index:notification_entity_index;not null"`
	EntityType string `gorm:"type:VARCHAR(32);index:notification_entity_index;not null"`
	EntityUID  string `gorm:"type:VARCHAR(36);not null"` // Copy of the human-readable entity ID, so that we can keep notifications for deleted entities.

	NotificationType string `gorm:"not null"`

	Data postgres.Jsonb `gorm:"type:JSONB"`

	SenderID  *string `gorm:"type:UUID;index:notification_sender_index"`
	SenderUID string  `gorm:"type:VARCHAR(36);not null"` // Copy of the human-readable sender ID, so that we can keep notifications for deleted senders.

	Receivers pq.Int32Array `gorm:"type:INT ARRAY"`

	Email bool `gorm:"not null"`
}

// NotificationReceiver model.
type NotificationReceiver struct {
	Notification   *Notification
	NotificationID string `gorm:"type:UUID;unique_index:notification_receiver_index;index:notification_receiver_notification_id_index;not null"`

	Receiver   *User
	ReceiverID string `gorm:"type:UUID;unique_index:notification_receiver_index;index:notification_receiver_user_index;not null"`

	Status          int32     `gorm:"not null"`
	StatusUpdatedAt time.Time `gorm:"not null"`
}

func init() {
	registerModel(&Notification{}, &NotificationReceiver{})
}

type notificationWithStatus struct {
	Notification

	Status          int32
	StatusUpdatedAt time.Time
}

func (n notificationWithStatus) toPB(pb *ttnpb.Notification) error {
	pb.Id = n.ID
	pb.CreatedAt = ttnpb.ProtoTimePtr(cleanTime(n.CreatedAt))
	if pb.EntityIds == nil {
		pb.EntityIds = buildIdentifiers(n.EntityType, n.EntityUID)
	}
	pb.NotificationType = n.NotificationType
	if len(n.Data.RawMessage) > 0 && pb.Data == nil {
		var anyPB types.Any
		err := jsonpb.TTN().Unmarshal(n.Data.RawMessage, &anyPB)
		if err != nil {
			return err
		}
		pb.Data = &anyPB
	}
	if n.SenderUID != "" {
		pb.SenderIds = &ttnpb.UserIdentifiers{UserId: n.SenderUID}
	}
	pb.Receivers = make([]ttnpb.NotificationReceiver, len(n.Receivers))
	for i, receiver := range n.Receivers {
		pb.Receivers[i] = ttnpb.NotificationReceiver(receiver)
	}
	pb.Email = n.Email
	pb.Status = ttnpb.NotificationStatus(n.Status)
	pb.StatusUpdatedAt = ttnpb.ProtoTimePtr(cleanTime(n.StatusUpdatedAt))
	return nil
}
