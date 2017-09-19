// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

const separator = ","

// Value implements sql.Valuer interface.
func (s ClientScope) Value() (driver.Value, error) {
	scope := make([]string, 0)

	if s.Application {
		scope = append(scope, strconv.Itoa(int(ScopeApplication)))
	}

	if s.Profile {
		scope = append(scope, strconv.Itoa(int(ScopeProfile)))
	}

	return strings.Join(scope, separator), nil
}

// Scan implements sql.Scanner interface.
func (s *ClientScope) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.Errorf("Invalid type. Got `%T` instead of string", src)
	}

	parts := strings.Split(str, separator)
	for _, part := range parts {
		switch part {
		case strconv.Itoa(int(ScopeApplication)):
			s.Application = true
		case strconv.Itoa(int(ScopeProfile)):
			s.Profile = true
		}
	}

	return nil
}

// Value implements sql.Valuer interface.
func (g ClientGrants) Value() (driver.Value, error) {
	grants := make([]string, 0)

	if g.AuthorizationCode {
		grants = append(grants, strconv.Itoa(int(GrantAuthorizationCode)))
	}

	if g.Password {
		grants = append(grants, strconv.Itoa(int(GrantPassword)))
	}

	if g.RefreshToken {
		grants = append(grants, strconv.Itoa(int(GrantRefreshToken)))
	}

	return strings.Join(grants, separator), nil
}

// Scan implements sql.Scanner interface.
func (g *ClientGrants) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.Errorf("Invalid type. Got `%T` instead of string", src)
	}

	parts := strings.Split(str, separator)
	for _, part := range parts {
		switch part {
		case strconv.Itoa(int(GrantAuthorizationCode)):
			g.AuthorizationCode = true
		case strconv.Itoa(int(GrantPassword)):
			g.Password = true
		case strconv.Itoa(int(GrantRefreshToken)):
			g.RefreshToken = true
		}
	}

	return nil
}
