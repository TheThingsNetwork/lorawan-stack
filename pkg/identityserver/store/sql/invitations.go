// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
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
func (s *InvitationStore) Save(data *store.InvitationData) error {
	return s.save(s.queryer(), data)
}

func (s *InvitationStore) save(q db.QueryContext, data *store.InvitationData) error {
	_, err := q.NamedExec(
		`INSERT
			INTO invitations (
				token,
				email,
				sent_at,
				ttl
			)
			VALUES (
				:token,
				:email,
				current_timestamp(),
				:ttl
			)`, data)
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
func (s *InvitationStore) List() ([]*store.InvitationData, error) {
	return s.list(s.queryer())
}

func (s *InvitationStore) list(q db.QueryContext) ([]*store.InvitationData, error) {
	var invitations []*store.InvitationData
	err := q.Select(
		&invitations,
		`SELECT
				id,
				email,
				sent_at,
				ttl,
				used_at,
				user_id
			FROM invitations`)
	if err != nil {
		return nil, err
	}

	if invitations == nil || len(invitations) == 0 {
		return make([]*store.InvitationData, 0), nil
	}

	return invitations, nil
}

// Delete deletes an invitation by its ID.
func (s *InvitationStore) Delete(ID string) error {
	return s.delete(s.queryer(), ID)
}

func (s *InvitationStore) delete(q db.QueryContext, ID string) error {
	_, err := q.Exec(`DELETE FROM invitations WHERE id = $1`, ID)
	return err
}
