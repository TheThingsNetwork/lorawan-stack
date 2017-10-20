// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// UserStore implements store.UserStore.
type UserStore struct {
	storer
	factory factory.UserFactory
}

func init() {
	ErrUserNotFound.Register()
	ErrUserEmailNotFound.Register()
	ErrUserIDTaken.Register()
	ErrUserEmailTaken.Register()
}

// ErrUserNotFound is returned when trying to fetch an user that does not exist.
var ErrUserNotFound = &errors.ErrDescriptor{
	MessageFormat: "User `{user_id}` does not exist",
	Code:          400,
	Type:          errors.NotFound,
}

// ErrUserEmailNotFound is returned when trying to find an user with an email
// that does not exist.
var ErrUserEmailNotFound = &errors.ErrDescriptor{
	MessageFormat: "User with email address `{email}` does not exist",
	Code:          401,
	Type:          errors.NotFound,
}

// ErrUserIDTaken is returned when trying to create a new user with an ID that
// is already taken.
var ErrUserIDTaken = &errors.ErrDescriptor{
	MessageFormat: "User ID `{user_id}` is already taken",
	Code:          402,
	Type:          errors.AlreadyExists,
}

// ErrUserEmailTaken is returned when trying to create a new user with an
// email that is already taken.
var ErrUserEmailTaken = &errors.ErrDescriptor{
	MessageFormat: "Email address `{email}` is already taken by another account",
	Code:          403,
	Type:          errors.AlreadyExists,
}

func NewUserStore(store storer, factory factory.UserFactory) *UserStore {
	return &UserStore{
		storer:  store,
		factory: factory,
	}
}

// Create creates an user.
func (s *UserStore) Create(user types.User) error {
	err := s.transact(func(tx *db.Tx) error {
		return s.create(tx, user)
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
				updated_at,
				archived_at)
			VALUES (
				lower(:user_id),
				:name,
				lower(:email),
				:password,
				current_timestamp(),
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

	if err != nil {
		return err
	}

	return s.writeAttributes(q, user.GetUser().UserID, user, nil)
}

// GetByID finds the user by ID and returns it.
func (s *UserStore) GetByID(userID string) (types.User, error) {
	result := s.factory.BuildUser()
	err := s.transact(func(tx *db.Tx) error {
		return s.getByID(tx, userID, result)
	})
	return result, err
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
	if err != nil {
		return err
	}

	return s.loadAttributes(q, userID, result)
}

// GetByEmail finds the user by email address and returns it.
func (s *UserStore) GetByEmail(email string) (types.User, error) {
	result := s.factory.BuildUser()
	err := s.transact(func(tx *db.Tx) error {
		return s.getByEmail(tx, email, result)
	})
	return result, err
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
	if err != nil {
		return err
	}

	return s.loadAttributes(q, result.GetUser().UserID, result)
}

// Update updates an user.
func (s *UserStore) Update(user types.User) error {
	err := s.transact(func(tx *db.Tx) error {
		return s.update(tx, user)
	})
	return err
}

func (s *UserStore) update(q db.QueryContext, user types.User) error {
	u := user.GetUser()

	_, err := q.NamedExec(
		`UPDATE users
			SET name = :name,
				email = lower(:email),
				validated = :validated,
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

	if err != nil {
		return err
	}

	return s.writeAttributes(q, u.UserID, user, nil)
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

// LoadAttributes loads all user attributes if the User is an Attributer.
func (s *UserStore) LoadAttributes(id string, user types.User) error {
	return s.transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, id, user)
	})
}

// loadAttributes loads extra attributes into a user in a given db.QueryContext
// context.
func (s *UserStore) loadAttributes(q db.QueryContext, id string, user types.User) error {
	attr, ok := user.(store.Attributer)
	if !ok {
		return nil
	}

	// fill the user from all specified namespaces
	for _, namespace := range attr.Namespaces() {
		m := make(map[string]interface{})
		err := q.SelectOne(
			&m,
			fmt.Sprintf(
				`SELECT *
					FROM %s_users
				 	WHERE user_id = $1`,
				namespace),
			id)
		if err != nil {
			return err
		}

		err = attr.Fill(namespace, m)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteAttributes writes all of the user attributes if the User is an Attributer
// and returns the written User in result.
func (s *UserStore) WriteAttributes(user types.User, result types.User) error {
	return s.transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, user.GetUser().UserID, user, result)
	})
}

// writeAttributes writes all of the users attributes to their respective
// tables in a given db.QueryContext context.
func (s *UserStore) writeAttributes(q db.QueryContext, id string, user types.User, res types.User) error {
	attr, ok := user.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "users", "user_id", user.GetUser().UserID)

		r := make(map[string]interface{})
		err := q.SelectOne(r, query, values...)
		if err != nil {
			return err
		}

		if rattr, ok := res.(store.Attributer); ok {
			err = rattr.Fill(namespace, r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SetFactory allows to replace the default ttnpb.User factory.
func (s *UserStore) SetFactory(factory factory.UserFactory) {
	s.factory = factory
}
