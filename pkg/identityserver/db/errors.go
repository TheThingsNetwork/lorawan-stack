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

package db

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"go.thethings.network/lorawan-stack/pkg/errors"
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
