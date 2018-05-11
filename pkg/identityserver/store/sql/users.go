// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package sql

import (
	"fmt"
	"strings"

	"github.com/satori/go.uuid"
	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
				user_id,
				email
			FROM users
			WHERE id = $1`,
		id)
	return
}

func (s *UserStore) getUserID(q db.QueryContext, ids ttnpb.UserIdentifiers) (res uuid.UUID, err error) {
	clauses := make([]string, 0)
	if ids.UserID != "" {
		clauses = append(clauses, "user_id = :user_id")
	}

	if ids.Email != "" {
		clauses = append(clauses, "email = :email")
	}

	err = q.NamedSelectOne(
		&res,
		fmt.Sprintf(
			`SELECT
				id
			FROM users
			WHERE %s`, strings.Join(clauses, " AND ")),
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
				password_updated_at,
				require_password_update,
				validated_at,
				created_at,
				updated_at)
			VALUES (
				:id,
				lower(:user_id),
				:name,
				lower(:email),
				:admin,
				:state,
				:password,
				:password_updated_at,
				:require_password_update,
				:validated_at,
				:created_at,
				:updated_at)`,
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

// GetByID finds the user by the given identifiers and retrieves it.
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

		result = specializer(user.User)

		return s.loadAttributes(tx, userID, result)
	})

	return
}

func (s *UserStore) getByID(q db.QueryContext, userID uuid.UUID) (result user, err error) {
	err = q.SelectOne(
		&result,
		`SELECT
				*
			FROM users
			WHERE id = $1`,
		userID)
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
func (s *UserStore) Update(ids ttnpb.UserIdentifiers, user store.User) error {
	err := s.transact(func(tx *db.Tx) error {
		u := user.GetUser()

		userID, err := s.getUserID(tx, ids)
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
				updated_at = :updated_at,
				password_updated_at = :password_updated_at,
				require_password_update = :require_password_update
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

		// Delete user itself.
		err = s.delete(tx, userID)
		if err != nil {
			return err
		}

		// Delete its ID from the shared namespace with organizations.
		return s.accountStore.deleteID(tx, userID)
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
