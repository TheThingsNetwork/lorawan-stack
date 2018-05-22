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

package oauth

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

type loginRequest struct {
	// UserID is the ID of the user is attempting to log in.
	UserID string `json:"user_id" form:"user_id"`

	// Password is the password of the user logging in.
	Password string `json:"password" form:"password"`
}

// login is an echo handler that lets an user log in.
func (s *Server) login(c echo.Context) error {
	req := new(loginRequest)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	err = validate.ID(req.UserID)
	if err != nil {
		return err
	}

	user, err := s.config.Store.Users.GetByID(ttnpb.UserIdentifiers{UserID: req.UserID}, s.config.Specializers.User)
	if err != nil {
		return err
	}

	ok, err := auth.Password(user.GetUser().Password).Validate(req.Password)
	if err != nil {
		return errors.NewWithCause(err, "Failed to validate password")
	}

	if !ok {
		return ErrInvalidPassword.New(nil)
	}

	// Credentials are ok. Therefore store the user in the cookie.
	err = s.updateCookie(c, func(value *authCookie) error {
		value.UserID = req.UserID
		value.LoggedIn = time.Now()

		return nil
	})
	if err != nil {
		return errors.NewWithCause(err, "Failed to update cookie")
	}

	return c.NoContent(http.StatusOK)
}
