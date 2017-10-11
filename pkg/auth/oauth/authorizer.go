// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package oauth

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/labstack/echo"
)

// Authorizer is the interface of authorizers. It is responsible for deciding what to
// during the authorization flow.
type Authorizer interface {
	// CheckLogin checks if the user is logged in and redirects to the login page if not.
	// Returns the username of the user if logged in.
	CheckLogin(c echo.Context) (string, error)

	// Authorize does a couple of things:
	// - renders the authorization form if it is not a form accept
	// - accepts the user input from the authorization form and uses this input to decide
	//   wether or not the client should be authorized.
	Authorize(c echo.Context, client types.Client) (authorized bool, err error)
}
