// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// UserStore implements store.UserStore.
type UserStore struct {
	storer
	*extraAttributesStore
}

func NewUserStore(store storer) *UserStore {
	return &UserStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "user"),
	}
}

// Create creates an user.
func (s *UserStore) Create(user types.User) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.create(tx, user)
		if err != nil {
			return err
		}

		return s.writeAttributes(tx, user.GetUser().UserID, user, nil)
	})
	return err
}

func (s *UserStore) create(q db.QueryContext, user types.User) error {
	u := user.GetUser()
	_, err := q.NamedExec(
		`INSERT
			INTO users (
				user_id,
				name,
				email,
				password,
				validated_at,
				archived_at)
			VALUES (
				lower(:user_id),
				:name,
				lower(:email),
				:password,
				:validated_at,
				:archived_at)`,
		u)

	if duplicates, yes := db.IsDuplicate(err); yes {
		if email, duplicated := duplicates["email"]; duplicated {
			return ErrUserEmailTaken.New(errors.Attributes{
				"email": email,
			})
		}
		if userID, duplicated := duplicates["user_id"]; duplicated {
			return ErrUserIDTaken.New(errors.Attributes{
				"user_id": userID,
			})
		}
	}

	return err
}

// GetByID finds the user by ID and returns it.
func (s *UserStore) GetByID(userID string, factory store.UserFactory) (types.User, error) {
	result := factory()

	err := s.transact(func(tx *db.Tx) error {
		err := s.getByID(tx, userID, result)
		if err != nil {
			return err
		}

		return s.loadAttributes(tx, result.GetUser().UserID, result)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *UserStore) getByID(q db.QueryContext, userID string, result types.User) error {
	err := q.SelectOne(
		result,
		`SELECT *
			FROM users
			WHERE user_id = lower($1)`,
		userID)
	if db.IsNoRows(err) {
		return ErrUserNotFound.New(errors.Attributes{
			"user_id": userID,
		})
	}
	return err
}

// GetByEmail finds the user by email address and returns it.
<<<<<<< HEAD
func (s *UserStore) GetByEmail(email string, factory store.UserFactory) (types.User, error) {
	result := factory()
=======
func (s *UserStore) GetByEmail(email string, resultFunc store.UserFactory) (types.User, error) {
	result := resultFunc()

>>>>>>> is: Review transactions
	err := s.transact(func(tx *db.Tx) error {
		err := s.getByEmail(tx, email, result)
		if err != nil {
			return err
		}

		return s.loadAttributes(tx, result.GetUser().UserID, result)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *UserStore) getByEmail(q db.QueryContext, email string, result types.User) error {
	err := q.SelectOne(
		result,
		`SELECT *
			FROM users
			WHERE email = lower($1)`,
		email)
	if db.IsNoRows(err) {
		return ErrUserEmailNotFound.New(errors.Attributes{
			"email": email,
		})
	}
	return err
}

// Update updates an user.
func (s *UserStore) Update(user types.User) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.update(tx, user)
		if err != nil {
			return err
		}

		return s.writeAttributes(s.queryer(), user.GetUser().UserID, user, nil)
	})
	return err
}

func (s *UserStore) update(q db.QueryContext, user types.User) error {
	u := user.GetUser()

	_, err := q.NamedExec(
		`UPDATE users
			SET name = :name,
				email = lower(:email),
				validated_at = :validated_at,
				password = :password,
				admin = :admin,
				updated_at = current_timestamp()
			WHERE user_id = :user_id`,
		u)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrUserEmailTaken.New(errors.Attributes{
			"email": u.Email,
		})
	}

	return err
}

// Archive sets the ArchivedAt field of an user to the current timestamp.
func (s *UserStore) Archive(userID string) error {
	return s.archive(s.queryer(), userID)
}

func (s *UserStore) archive(q db.QueryContext, userID string) error {
	var i string
	err := q.SelectOne(
		&i,
		`UPDATE users
			SET archived_at = current_timestamp()
			WHERE user_id = $1
			RETURNING user_id`,
		userID)
	if db.IsNoRows(err) {
		return ErrUserNotFound.New(errors.Attributes{
			"user_id": userID,
		})
	}
	return err
}

// LoadAttributes loads the extra attributes in user if it is a store.Attributer.
func (s *UserStore) LoadAttributes(userID string, user types.User) error {
	return s.loadAttributes(s.queryer(), userID, user)
}

func (s *UserStore) loadAttributes(q db.QueryContext, userID string, user types.User) error {
	attr, ok := user.(store.Attributer)
	if ok {
		return s.extraAttributesStore.loadAttributes(q, userID, attr)
	}

	return nil
}

// WriteAttributes store the extra attributes of user if it is a store.Attributer
// and writes the resulting user in result.
func (s *UserStore) WriteAttributes(userID string, user, result types.User) error {
	return s.writeAttributes(s.queryer(), userID, user, result)
}

func (s *UserStore) writeAttributes(q db.QueryContext, userID string, user, result types.User) error {
	attr, ok := user.(store.Attributer)
	if ok {
		res, ok := result.(store.Attributer)
		if result == nil || !ok {
			return s.extraAttributesStore.writeAttributes(q, userID, attr, nil)
		}

		return s.extraAttributesStore.writeAttributes(q, userID, attr, res)
	}

	return nil
}
