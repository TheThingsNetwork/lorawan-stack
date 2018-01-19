// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// InvitationStore implements store.InvitationStore.
type InvitationStore struct {
	storer
}

// NewInvitationStore creates a new invitation store.
func NewInvitationStore(store storer) *InvitationStore {
	return &InvitationStore{
		storer: store,
	}
}

// Save saves the invitation.
func (s *InvitationStore) Save(token, email string, expiresIn uint32) error {
	return s.save(s.queryer(), token, email, expiresIn)
}

func (s *InvitationStore) save(q db.QueryContext, token, email string, expiresIn uint32) error {
	_, err := q.Exec(
		`INSERT
			INTO invitations (
				token,
				email,
				sent_at,
				expires_in
			)
			VALUES (
				$1,
				$2,
				current_timestamp(),
				$3
			)`,
		token,
		email,
		expiresIn)
	return err
}

// Use sets `used_at` to the current timestamp and links the invitation to an user ID.
func (s *InvitationStore) Use(token, userID string) error {
	err := s.transact(func(tx *db.Tx) error {
		used, err := s.isUsed(tx, token)
		if err != nil {
			return err
		}

		if used {
			return ErrInvitationAlreadyUsed.New(nil)
		}

		return s.use(tx, token, userID)
	})

	return err
}

func (s *InvitationStore) isUsed(q db.QueryContext, token string) (bool, error) {
	var usedAt *time.Time
	err := q.SelectOne(&usedAt, `SELECT used_at FROM invitations WHERE token = $1`, token)
	if db.IsNoRows(err) {
		return false, ErrInvitationNotFound.New(nil)
	}
	if err != nil {
		return false, err
	}
	if usedAt == nil {
		return false, nil
	}
	return true, nil
}

func (s *InvitationStore) use(q db.QueryContext, token, userID string) error {
	id := ""
	err := q.SelectOne(
		&id,
		`UPDATE
				invitations
			SET used_at = current_timestamp(), user_id = $1
			WHERE token = $2
			RETURNING id`,
		userID,
		token)
	if db.IsNoRows(err) {
		return ErrInvitationNotFound.New(nil)
	}
	return err
}

// List lists all the saved invitations.
func (s *InvitationStore) List() ([]*ttnpb.ListInvitationsResponse_Invitation, error) {
	return s.list(s.queryer())
}

func (s *InvitationStore) list(q db.QueryContext) ([]*ttnpb.ListInvitationsResponse_Invitation, error) {
	var res []struct {
		*ttnpb.ListInvitationsResponse_Invitation
		User *string
	}
	err := q.Select(
		&res,
		`SELECT
				id,
				email,
				sent_at,
				expires_in,
				used_at,
				user_id AS user
			FROM invitations`)
	if err != nil {
		return nil, err
	}

	invitations := make([]*ttnpb.ListInvitationsResponse_Invitation, 0, len(res))
	for _, invitation := range res {
		if invitation.User != nil {
			invitation.ListInvitationsResponse_Invitation.UserID = *(invitation.User)
		}

		invitations = append(invitations, invitation.ListInvitationsResponse_Invitation)
	}

	return invitations, nil
}

// Delete deletes an invitation by its ID.
func (s *InvitationStore) Delete(id string) error {
	return s.delete(s.queryer(), id)
}

func (s *InvitationStore) delete(q db.QueryContext, id string) error {
	_, err := q.Exec(`DELETE FROM invitations WHERE id = $1`, id)
	return err
}
