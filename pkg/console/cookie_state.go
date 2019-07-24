// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package console

import (
	"encoding/gob"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/web/cookie"
)

func init() {
	gob.Register(state{})
}

const stateCookieName = "_console_state"

// StateCookie returns the cookie storing the state of the console.
func (console *Console) StateCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     stateCookieName,
		HTTPOnly: true,
		Path:     console.config.UI.MountPath(),
		MaxAge:   10 * time.Minute,
	}
}

// state is the shape of the state for the OAuth flow.
type state struct {
	Secret string
	Next   string
}

func newState(next string) state {
	return state{
		Secret: random.String(16),
		Next:   next,
	}
}

func (console *Console) getStateCookie(c echo.Context) (state, error) {
	s := state{}
	ok, err := console.StateCookie().Get(c, &s)
	if err != nil {
		return s, err
	}
	if !ok {
		return s, echo.NewHTTPError(http.StatusBadRequest, "No state cookie")
	}
	return s, nil
}

func (console *Console) setStateCookie(c echo.Context, value state) error {
	return console.StateCookie().Set(c, value)
}

func (console *Console) removeStateCookie(c echo.Context) {
	console.StateCookie().Remove(c)
}
