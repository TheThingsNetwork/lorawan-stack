// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// UserStore implements store.UserStore
type UserStore struct {
	*Store
	factory factory.UserFactory
}

// ErrUserNotFound is returned when trying to fetch an user that does not exist
var ErrUserNotFound = errors.New("user not found")

// ErrUserEmailNotFound is returned when trying to find an user with an email
// that does not exist
var ErrUserEmailNotFound = errors.New("user email not found")

// ErrUsernameTaken is returned when trying to create a new user with an
// username that already exists
var ErrUsernameTaken = errors.New("username already taken")

// ErrUserEmailTaken is returned when trying to create a new user with an
// email that already exists
var ErrUserEmailTaken = errors.New("email already taken")

// SetFactory replaces the factory
func (s *UserStore) SetFactory(factory factory.UserFactory) {
	s.factory = factory
}

// LoadAttributes loads attributes for the user with the specified userID if
// it is an Attributer
func (s *UserStore) LoadAttributes(username string, user types.User) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, username, user)
	})
}

// loadAttributes loads extra attributes into a user in a given db.QueryContext
// context
func (s *UserStore) loadAttributes(q db.QueryContext, username string, user types.User) error {
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
				 	WHERE username = $1`,
				namespace),
			username)
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

// WriteAttributes writes all of the user attributes if the user is an
// Attributer and returns the written user in result
func (s *UserStore) WriteAttributes(user types.User, result types.User) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, user.GetUser().Username, user, result)
	})
}

// writeAttributes writes all of the users attributes to their respective
// tables in a given db.QueryContext context
func (s *UserStore) writeAttributes(q db.QueryContext, username string, user types.User, res types.User) error {
	attr, ok := user.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "users", "username", user.GetUser().Username)

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

// FindByEmail finds an user by email address
func (s *UserStore) FindByEmail(email string) (types.User, error) {
	result := s.factory.User()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.findByEmail(tx, email, result)
	})
	return result, err
}

func (s *UserStore) findByEmail(q db.QueryContext, email string, user types.User) error {
	err := q.SelectOne(
		user,
		"SELECT * FROM users WHERE email = $1",
		strings.ToLower(email))
	if db.IsNoRows(err) {
		return ErrUserEmailNotFound
	}
	if err != nil {
		return err
	}

	return s.loadAttributes(q, user.GetUser().Username, user)
}

// FindByUsername finds an user by username
func (s *UserStore) FindByUsername(username string) (types.User, error) {
	result := s.factory.User()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.findByUsername(tx, username, result)
	})
	return result, err
}

func (s *UserStore) findByUsername(q db.QueryContext, username string, user types.User) error {
	err := q.SelectOne(
		user,
		"SELECT * FROM users WHERE lower(username) = $1",
		strings.ToLower(username))

	if db.IsNoRows(err) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	return s.loadAttributes(q, user.GetUser().Username, user)
}

// Create creates a user and returns the new created user
func (s *UserStore) Create(user types.User) (types.User, error) {
	result := s.factory.User()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.create(tx, user, result)
	})
	return result, err
}

func (s *UserStore) create(q db.QueryContext, user, result types.User) error {
	u := user.GetUser()
	err := q.NamedSelectOne(
		result,
		`INSERT
			INTO users (username, email, password)
			VALUES (:username, :email, :password)
			RETURNING *`,
		u)

	if dup, yes := db.IsDuplicate(err); yes {
		if _, duplicated := dup.Duplicates["email"]; duplicated {
			return ErrUserEmailTaken
		}
		if _, duplicated := dup.Duplicates["username"]; duplicated {
			return ErrUsernameTaken
		}
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, result.GetUser().Username, user, result)
}

// Update updates an user and returns the updated user
func (s *UserStore) Update(user types.User) (types.User, error) {
	result := s.factory.User()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.update(tx, user, result)
	})
	return result, err
}

func (s *UserStore) update(q db.QueryContext, user, result types.User) error {
	u := user.GetUser()
	err := q.NamedSelectOne(
		result,
		`UPDATE users
			SET email = :email,
				validated = :validated,
				password = :password,
				admin = :admin,
				god = :god
			WHERE username = :username
			RETURNING *`,
		u)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrUserEmailTaken
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, u.Username, user, result)
}

// Archive archives an user
func (s *UserStore) Archive(username string) error {
	return s.archive(s.db, username)
}

func (s *UserStore) archive(q db.QueryContext, username string) error {
	var u string
	err := q.SelectOne(
		&u,
		`UPDATE users
			SET archived = $1
			WHERE username = $2
			RETURNING username`,
		time.Now(),
		username)
	if db.IsNoRows(err) {
		return ErrUserNotFound
	}
	return err
}
