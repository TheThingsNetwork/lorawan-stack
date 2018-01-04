// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrBlacklistedID.Register()
	ErrNotAllowedEmail.Register()
	ErrPasswordsDoNotMatch.Register()
}

// ErrNotAuthorized is returned when trying to access a protected resource without
// authorization.
var ErrNotAuthorized = errors.New("not authorized")

// ErrBlacklistedID is returned when trying to register an entity using a blacklisted ID.
var ErrBlacklistedID = &errors.ErrDescriptor{
	MessageFormat: "{id} is not an allowed ID",
	Code:          1,
	Type:          errors.InvalidArgument,
}

// ErrNotAllowedEmail is returned when trying to set in an user account a not
// allowed email.
var ErrNotAllowedEmail = &errors.ErrDescriptor{
	MessageFormat: "{email} is not an allowed email",
	Code:          2,
	Type:          errors.InvalidArgument,
}

// ErrPasswordDoNotMatch is returned when trying to update the user's account
// password while providing a wrong current password.
var ErrPasswordsDoNotMatch = &errors.ErrDescriptor{
	MessageFormat: "Passwords do not match each other",
	Code:          3,
	Type:          errors.InvalidArgument,
}
