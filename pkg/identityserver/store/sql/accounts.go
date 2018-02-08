// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
)

const (
	// organization is the value used to denote that an ID belongs to an organization.
	organization int = 0

	// user is the value used to denote that an ID belongs to an user.
	user int = 1
)

// accountStore is a shared substore used to manage the users and organizations
// ID namespace.
type accountStore struct {
	storer
}

// newAccountStore returns an accountStore.
func newAccountStore(store storer) *accountStore {
	return &accountStore{
		storer: store,
	}
}

// registerOrganizationID registers the given ID that belongs to an organization.
func (s *accountStore) registerOrganizationID(q db.QueryContext, organizationID string) error {
	_, err := q.Exec(
		`INSERT
			INTO accounts(account_id, type)
			VALUES ($1, $2)`,
		organizationID,
		organization)
	if _, yes := db.IsDuplicate(err); yes {
		return ErrOrganizationIDTaken.New(errors.Attributes{
			"organization_id": organizationID,
		})
	}
	return err
}

// registerUserID registers the given ID that belongs to an user.
func (s *accountStore) registerUserID(q db.QueryContext, userID string) error {
	_, err := q.Exec(
		`INSERT
			INTO accounts(account_id, type)
			VALUES ($1, $2)`,
		userID,
		user)
	if _, yes := db.IsDuplicate(err); yes {
		return ErrUserIDTaken.New(errors.Attributes{
			"user_id": userID,
		})
	}
	return err
}

// deleteID deletes the given ID.
func (s *accountStore) deleteID(q db.QueryContext, id string) error {
	res := ""
	err := q.SelectOne(
		&res,
		`DELETE
			FROM accounts
			WHERE account_id = $1
			RETURNING id`,
		id)
	if db.IsNoRows(err) {
		return ErrAccountIDNotFound.New(errors.Attributes{
			"account_id": id,
		})
	}
	return err
}
