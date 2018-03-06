// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/lib/pq"
)

var re = regexp.MustCompile("duplicate key value \\((.+)\\)=\\((.+)\\)")

// ErrDuplicate is the error that occures when duplicate fields are being inserted
// into a database columnt that dissallows this (trough a unique constraint).
type ErrDuplicate struct {
	Message    string
	Duplicates map[string]string
}

// Error implements error.
func (err *ErrDuplicate) Error() string {
	return err.Message
}

func wrap(err error) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return err
	}

	if pq, ok := err.(*pq.Error); ok {
		switch pq.Code {
		case "23505":
			// Unique violation
			m := re.FindStringSubmatch(err.Error())
			if len(m) > 1 {
				names := strings.Split(m[1], ",")
				values := strings.Split(m[2], ",")
				duplicates := make(map[string]string)

				for i, name := range names {
					duplicates[name] = strings.Trim(values[i], "'")
				}

				return &ErrDuplicate{
					Message:    err.Error(),
					Duplicates: duplicates,
				}
			}
		}
	}

	if u, ok := err.(errors.Error); ok {
		return u
	}

	return errors.NewWithCause(err, "Unexpected error")
}

// IsDuplicate returns the name and value of a duplicated field and a bool
// denoting wether or not the error was caused by a duplicate value.
func IsDuplicate(err error) (map[string]string, bool) {
	if err == nil {
		return nil, false
	}

	if dup, ok := err.(*ErrDuplicate); ok {
		return dup.Duplicates, true
	}

	return nil, false
}

// IsNoRows returns wether or not the error is an sql.ErrNoRows.
func IsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
