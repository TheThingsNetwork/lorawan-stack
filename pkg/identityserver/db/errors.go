// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/lib/pq"
)

func init() {
	ErrDuplicate.Register()
}

var duplicatesKey = "duplicates"

// ErrDuplicate is the descriptor for the error returned when there is an unique
// constraint violation in the database.
var ErrDuplicate = &errors.ErrDescriptor{
	Code: 1,
	Type: errors.Conflict,
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
			re := regexp.MustCompile("duplicate key value \\((.+)\\)=\\((.+)\\)")
			m := re.FindStringSubmatch(err.Error())
			if len(m) > 1 {
				names := strings.Split(m[1], ",")
				values := strings.Split(m[2], ",")
				duplicates := make(map[string]string)

				for i, name := range names {
					duplicates[name] = strings.Trim(values[i], "'")
				}

				return ErrDuplicate.NewWithCause(errors.Attributes{
					duplicatesKey: duplicates,
				}, err)
			}
		}
	}

	if u, ok := err.(errors.Error); ok {
		return u
	}

	return errors.NewWithCause("Unexpected error", err)
}

// IsDuplicate returns the name and value of a duplicated field and a bool
// denoting wether or not the error was caused by a duplicate value.
func IsDuplicate(err error) (map[string]string, bool) {
	if err == nil {
		return nil, false
	}

	if dup, ok := err.(errors.Error); ok && dup.Code() == ErrDuplicate.Code {
		return dup.Attributes()[duplicatesKey].(map[string]string), true
	}

	return nil, false
}

// IsNoRows returns wether or not the error is an sql.ErrNoRows.
func IsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
