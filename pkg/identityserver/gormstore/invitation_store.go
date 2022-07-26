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
	"context"
	"runtime/trace"
	"time"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetInvitationStore returns an InvitationStore on the given db (or transaction).
func GetInvitationStore(db *gorm.DB) store.InvitationStore {
	return &invitationStore{baseStore: newStore(db)}
}

type invitationStore struct {
	*baseStore
}

func (s *invitationStore) CreateInvitation(
	ctx context.Context, invitation *ttnpb.Invitation,
) (*ttnpb.Invitation, error) {
	defer trace.StartRegion(ctx, "create invitation").End()
	model := Invitation{
		Email:     invitation.Email,
		Token:     invitation.Token,
		ExpiresAt: ttnpb.StdTime(invitation.ExpiresAt),
	}
	if err := s.createEntity(ctx, &model); err != nil {
		err = convertError(err)
		if errors.IsAlreadyExists(err) {
			return nil, store.ErrInvitationAlreadySent.New()
		}
		return nil, err
	}
	return model.toPB(), nil
}

func (s *invitationStore) FindInvitations(ctx context.Context) ([]*ttnpb.Invitation, error) {
	defer trace.StartRegion(ctx, "find invitations").End()
	var invitationModels []Invitation
	query := s.query(ctx, Invitation{})
	query = query.Order(store.OrderFromContext(ctx, "invitations", "id", "ASC"))
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Preload("AcceptedBy.Account").Find(&invitationModels).Error; err != nil {
		return nil, err
	}
	store.SetTotal(ctx, uint64(len(invitationModels)))
	invitationProtos := make([]*ttnpb.Invitation, len(invitationModels))
	for i, invitationModel := range invitationModels {
		invitationProtos[i] = invitationModel.toPB()
	}
	return invitationProtos, nil
}

func (s *invitationStore) GetInvitation(ctx context.Context, token string) (*ttnpb.Invitation, error) {
	defer trace.StartRegion(ctx, "get invitation").End()
	var invitationModel Invitation
	if err := s.query(ctx, Invitation{}).
		Where(Invitation{Token: token}).
		Preload("AcceptedBy.Account").
		First(&invitationModel).
		Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, store.ErrInvitationNotFound.WithAttributes("invitation_token", token)
		}
		return nil, err
	}
	return invitationModel.toPB(), nil
}

func (s *invitationStore) SetInvitationAcceptedBy(
	ctx context.Context, token string, acceptedByID *ttnpb.UserIdentifiers,
) error {
	defer trace.StartRegion(ctx, "update invitation").End()
	var invitationModel Invitation
	if err := s.query(ctx, Invitation{}).
		Where(Invitation{Token: token}).
		First(&invitationModel).
		Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return store.ErrInvitationNotFound.WithAttributes("invitation_token", token)
		}
		return err
	}

	if invitationModel.ExpiresAt != nil && invitationModel.ExpiresAt.Before(time.Now()) {
		return store.ErrInvitationExpired.WithAttributes("invitation_token", token)
	}

	user, err := s.findEntity(ctx, acceptedByID, "id")
	if err != nil {
		return err
	}
	if invitationModel.AcceptedByID != nil {
		return store.ErrInvitationAlreadyUsed.WithAttributes("invitation_token", token)
	}
	id := user.PrimaryKey()
	invitationModel.AcceptedByID = &id

	acceptedAt := cleanTime(time.Now())
	invitationModel.AcceptedAt = &acceptedAt

	return s.updateEntity(ctx, &invitationModel, "accepted_by_id", "accepted_at")
}

func (s *invitationStore) DeleteInvitation(ctx context.Context, email string) error {
	defer trace.StartRegion(ctx, "delete invitation").End()
	var invitationModel Invitation
	if err := s.query(ctx, Invitation{}).
		Where(Invitation{Email: email}).
		First(&invitationModel).
		Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return store.ErrInvitationNotFound.New()
		}
		return err
	}
	return s.query(ctx, Invitation{}).Delete(&invitationModel).Error
}
