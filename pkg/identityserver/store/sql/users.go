// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

// user represents the schema of the stored application.
type user struct {
	ID uuid.UUID
	ttnpb.User
}

// UserStore implements store.UserStore.
type UserStore struct {
	storer
	*extraAttributesStore
	*apiKeysStore
	*accountStore
}

// NewUserStore returns an UserStore.
func NewUserStore(store storer) *UserStore {
	return &UserStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "user"),
		apiKeysStore:         newAPIKeysStore(store, "user"),
		accountStore:         newAccountStore(store),
	}
}

func (s *UserStore) getUserIdentifiersFromID(q db.QueryContext, id uuid.UUID) (res ttnpb.UserIdentifiers, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				user_id
			FROM users
			WHERE id = $1`,
		id)
	return
}

func (s *UserStore) getUserID(q db.QueryContext, ids ttnpb.UserIdentifiers) (id uuid.UUID, err error) {
	err = q.NamedSelectOne(
		&id,
		`SELECT
			id
		FROM users
		WHERE user_id = :user_id`,
		ids)
	if db.IsNoRows(err) {
		err = ErrUserNotFound.New(nil)
	}
	return
}

// Create creates an user.
func (s *UserStore) Create(usr store.User) error {
	err := s.transact(func(tx *db.Tx) error {
		u := usr.GetUser()

		id, err := s.accountStore.registerUserID(tx, u.UserID)
		if err != nil {
			return err
		}

		err = s.create(tx, user{
			ID:   id,
			User: *u,
		})
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, id, usr)
	})
	return err
}

func (s *UserStore) create(q db.QueryContext, data user) (err error) {
	_, err = q.NamedExec(
		`INSERT
			INTO users (
				id,
				user_id,
				name,
				email,
				admin,
				state,
				password,
				validated_at)
			VALUES (
				:id,
				lower(:user_id),
				:name,
				lower(:email),
				:admin,
				:state,
				:password,
				:validated_at)`,
		data)

	if duplicates, yes := db.IsDuplicate(err); yes {
		if _, duplicated := duplicates["email"]; duplicated {
			return ErrUserEmailTaken.New(nil)
		}
		if _, duplicated := duplicates["user_id"]; duplicated {
			return ErrUserIDTaken.New(nil)
		}
	}

	return
}

// GetByID finds the user by ID and returns it.
func (s *UserStore) GetByID(ids ttnpb.UserIdentifiers, specializer store.UserSpecializer) (result store.User, err error) {
	err = s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		user, err := s.getByID(tx, userID)
		if err != nil {
			return err
		}

		result = specializer(user)

		return s.loadAttributes(tx, userID, result)
	})

	return
}

func (s *UserStore) getByID(q db.QueryContext, userID uuid.UUID) (result ttnpb.User, err error) {
	err = q.SelectOne(
		&result,
		`SELECT
				user_id,
				name,
				email,
				password,
				validated_at,
				state,
				admin,
				created_at,
				updated_at
			FROM users
			WHERE id = $1`,
		userID)
	if db.IsNoRows(err) {
		err = ErrUserNotFound.New(nil)
	}
	return
}

// GetByEmail finds the user by email address and returns it.
func (s *UserStore) GetByEmail(email string, specializer store.UserSpecializer) (result store.User, err error) {
	err = s.transact(func(tx *db.Tx) error {
		user, err := s.getByEmail(tx, email)
		if err != nil {
			return err
		}

		result = specializer(user.User)

		return s.loadAttributes(tx, user.ID, result)
	})

	return
}

func (s *UserStore) getByEmail(q db.QueryContext, email string) (result user, err error) {
	err = q.SelectOne(
		&result,
		`SELECT *
			FROM users
			WHERE email = lower($1)`,
		email)
	if db.IsNoRows(err) {
		err = ErrUserNotFound.New(nil)
	}
	return
}

// List returns all the users.
func (s *UserStore) List(specializer store.UserSpecializer) (result []store.User, err error) {
	err = s.transact(func(tx *db.Tx) error {
		users, err := s.list(tx)
		if err != nil {
			return err
		}

		for _, user := range users {
			specialized := specializer(user.User)

			err := s.loadAttributes(tx, user.ID, specialized)
			if err != nil {
				return err
			}

			result = append(result, specialized)
		}

		return nil
	})
	return
}

func (s *UserStore) list(q db.QueryContext) (result []user, err error) {
	err = q.Select(&result, `SELECT * FROM users`)
	return
}

// Update updates an user.
func (s *UserStore) Update(user store.User) error {
	err := s.transact(func(tx *db.Tx) error {
		u := user.GetUser()

		userID, err := s.getUserID(tx, u.UserIdentifiers)
		if err != nil {
			return err
		}

		err = s.update(tx, userID, u)
		if err != nil {
			return err
		}

		return s.storeAttributes(s.queryer(), userID, user)
	})
	return err
}

func (s *UserStore) update(q db.QueryContext, userID uuid.UUID, data *ttnpb.User) (err error) {
	_, err = q.NamedExec(
		`UPDATE users
			SET
				name = :name,
				email = lower(:email),
				validated_at = :validated_at,
				password = :password,
				admin = :admin,
				state = :state,
				updated_at = current_timestamp()
			WHERE id = :id`,
		user{
			ID:   userID,
			User: *data,
		})
	if _, yes := db.IsDuplicate(err); yes {
		err = ErrUserEmailTaken.New(nil)
	}
	return
}

// Delete deletes an user.
func (s *UserStore) Delete(ids ttnpb.UserIdentifiers) error {
	err := s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		err = s.leaveAllApplications(tx, userID)
		if err != nil {
			return err
		}

		err = s.leaveAllGateways(tx, userID)
		if err != nil {
			return err
		}

		// revoke all authorized clients
		oauth, ok := s.store().OAuth.(*OAuthStore)
		if !ok {
			return errors.Errorf("Expected ptr to OAuthStore but got %T", s.store().OAuth)
		}

		err = oauth.deleteAuthorizationCodesByUser(tx, userID)
		if err != nil {
			return err
		}

		clientIDs, err := oauth.listAuthorizedClients(tx, userID)
		if err != nil {
			return err
		}

		for _, clientID := range clientIDs {
			_, err = oauth.deleteAccessTokensByUserAndClient(tx, userID, clientID)
			if err != nil {
				return err
			}

			_, err = oauth.deleteRefreshTokenByUserAndClient(tx, userID, clientID)
			if err != nil {
				return err
			}
		}

		err = s.deleteCreatedClients(tx, userID)
		if err != nil {
			return err
		}

		// delete api keys
		err = s.deleteAPIKeys(tx, userID)
		if err != nil {
			return err
		}

		// delete validation tokens
		err = s.deleteValidationTokens(tx, userID)
		if err != nil {
			return err
		}

		// TODO(gomezjdaniel): delete attributers.

		// delete user itself
		err = s.delete(tx, userID)
		if err != nil {
			return err
		}

		return s.accountStore.deleteID(tx, ids.UserID)
	})

	return err
}

func (s *UserStore) delete(q db.QueryContext, userID uuid.UUID) (err error) {
	id := new(string)
	err = q.SelectOne(
		id,
		`DELETE
			FROM users
			WHERE id = $1
			RETURNING user_id`,
		userID)
	if db.IsNoRows(err) {
		err = ErrUserNotFound.New(nil)
	}
	return
}

func (s *UserStore) leaveAllApplications(q db.QueryContext, userID uuid.UUID) (err error) {
	_, err = q.Exec(
		`DELETE
			FROM applications_collaborators
			WHERE account_id = $1`,
		userID)
	return
}

func (s *UserStore) leaveAllGateways(q db.QueryContext, userID uuid.UUID) (err error) {
	_, err = q.Exec(
		`DELETE
			FROM gateways_collaborators
			WHERE account_id = $1`,
		userID)
	return
}

func (s *UserStore) deleteCreatedClients(q db.QueryContext, userID uuid.UUID) (err error) {
	_, err = q.Exec(
		`DELETE
			FROM clients
			WHERE creator_id = $1`,
		userID)
	return
}
