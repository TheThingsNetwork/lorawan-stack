// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

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

				return &ErrDuplicate{
					Message:    err.Error(),
					Duplicates: duplicates,
				}
			}
		}
	}

	if u, ok := err.(*ErrUnexpected); ok {
		return u
	}

	return &ErrUnexpected{
		Cause: err,
	}
}

// ErrDuplicate denotes insertion/update of a duplicate value that should be
// unique.
type ErrDuplicate struct {
	Message    string
	Duplicates map[string]string
}

// Error implments error.
func (e *ErrDuplicate) Error() string {
	return e.Message
}

// IsDuplicate returns the name and value of a duplicated field and a bool
// denoting wether or not the error was caused by a duplicate value.
func IsDuplicate(err error) (*ErrDuplicate, bool) {
	if err == nil {
		return nil, false
	}

	if dup, ok := err.(*ErrDuplicate); ok {
		return dup, true
	}

	return nil, false
}

// IsNoRows returns wether or not the error is an sql.ErrNoRows.
func IsNoRows(err error) bool {
	return err == sql.ErrNoRows
}

// ErrUnexpected is an unexpected error.
type ErrUnexpected struct {
	Cause error
}

// Error implmentents error.
func (e *ErrUnexpected) Error() string {
	return e.Cause.Error()
}
