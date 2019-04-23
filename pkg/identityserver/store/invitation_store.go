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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetInvitationStore returns an InvitationStore on the given db (or transaction).
func GetInvitationStore(db *gorm.DB) InvitationStore {
	return &invitationStore{store: newStore(db)}
}

type invitationStore struct {
	*store
}

var errInvitationAlreadySent = errors.DefineAlreadyExists("invitation_already_sent", "invitation already sent")

func (s *invitationStore) CreateInvitation(ctx context.Context, invitation *ttnpb.Invitation) (*ttnpb.Invitation, error) {
	defer trace.StartRegion(ctx, "create invitation").End()
	model := Invitation{
		Email:     invitation.Email,
		Token:     invitation.Token,
		ExpiresAt: invitation.ExpiresAt,
	}
	if err := s.createEntity(ctx, &model); err != nil {
		err = convertError(err)
		if errors.IsAlreadyExists(err) {
			return nil, errInvitationAlreadySent
		}
		return nil, err
	}
	return model.toPB(), nil
}

func (s *invitationStore) FindInvitations(ctx context.Context) ([]*ttnpb.Invitation, error) {
	defer trace.StartRegion(ctx, "find invitations").End()
	var invitationModels []Invitation
	query := s.query(ctx, Invitation{})
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&Invitation{}))
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Find(&invitationModels).Error; err != nil {
		return nil, err
	}
	invitationProtos := make([]*ttnpb.Invitation, len(invitationModels))
	for i, invitationModel := range invitationModels {
		invitationProtos[i] = invitationModel.toPB()
	}
	return invitationProtos, nil
}

var errInvitationNotFound = errors.DefineNotFound("invitation_not_found", "invitation not found")

func (s *invitationStore) GetInvitation(ctx context.Context, token string) (*ttnpb.Invitation, error) {
	defer trace.StartRegion(ctx, "get invitation").End()
	var invitationModel Invitation
	if err := s.query(ctx, Invitation{}).Where(Invitation{Token: token}).First(&invitationModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errInvitationNotFound
		}
		return nil, err
	}
	return invitationModel.toPB(), nil
}

var errInvitationAlreadyAccepted = errors.DefineAlreadyExists("invitation_already_accepted", "invitation already accepted")

func (s *invitationStore) SetInvitationAcceptedBy(ctx context.Context, token string, acceptedByID *ttnpb.UserIdentifiers) error {
	defer trace.StartRegion(ctx, "update invitation").End()
	var invitationModel Invitation
	if err := s.query(ctx, Invitation{}).Where(Invitation{Token: token}).First(&invitationModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errInvitationNotFound
		}
		return err
	}

	user, err := s.findEntity(ctx, acceptedByID.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	if invitationModel.AcceptedByID == nil {
		id := user.PrimaryKey()
		invitationModel.AcceptedByID = &id
	} else {
		return errInvitationAlreadyAccepted
	}

	acceptedAt := cleanTime(time.Now())
	invitationModel.AcceptedAt = &acceptedAt

	return s.updateEntity(ctx, &invitationModel)
}

func (s *invitationStore) DeleteInvitation(ctx context.Context, email string) error {
	defer trace.StartRegion(ctx, "delete invitation").End()
	var invitationModel Invitation
	if err := s.query(ctx, Invitation{}).Where(Invitation{Email: email}).First(&invitationModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errInvitationNotFound
		}
		return err
	}
	if invitationModel.AcceptedByID != nil {
		return errInvitationAlreadyAccepted
	}
	return s.query(ctx, Invitation{}).Delete(&invitationModel).Error
}
