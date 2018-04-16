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
)

type user struct {
	// UserID is the ID of the user if logged in.
	UserID string `json:"user_id,omitempty"`

	// The users preferred language.
	Language string `json:"language,omitempty"`

	// LoggedIn is the time the user logged in.
	LoggedIn time.Time `json:"logged_in"`
}

// me is an echo handler that returns the currently logged in user.
func (s *Server) me(c echo.Context) error {
	u, err := getUser(c)
	if err != nil {
		return err
	}

	cookie, err := s.getCookie(c)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user{
		UserID:   u.GetUser().UserID,
		LoggedIn: cookie.LoggedIn,
	})
}
