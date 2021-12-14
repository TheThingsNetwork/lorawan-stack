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

package oauthclient

import (
	"encoding/gob"
	"net/http"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/web/cookie"
)

// state is the shape of the state for the OAuth flow.
type state struct {
	Secret string
	Next   string
}

func init() {
	gob.Register(state{})
}

// StateCookie returns the cookie storing the state of the console.
func (oc *OAuthClient) StateCookie() *cookie.Cookie {
	sameSite := http.SameSiteLaxMode
	if oc.config.CrossSiteCookie {
		sameSite = http.SameSiteNoneMode
	}
	return &cookie.Cookie{
		Name:     oc.config.StateCookieName,
		HTTPOnly: true,
		Path:     oc.getMountPath(),
		MaxAge:   10 * time.Minute,
		SameSite: sameSite,
	}
}

func newState(next string) state {
	return state{
		Secret: random.String(16),
		Next:   next,
	}
}

func (oc *OAuthClient) getStateCookie(w http.ResponseWriter, r *http.Request) (state, error) {
	s := state{}
	ok, err := oc.StateCookie().Get(w, r, &s)
	if err != nil {
		return s, err
	}

	if !ok {
		return s, err
	}

	return s, nil
}

func (oc *OAuthClient) setStateCookie(w http.ResponseWriter, r *http.Request, value state) error {
	return oc.StateCookie().Set(w, r, value)
}

func (oc *OAuthClient) removeStateCookie(w http.ResponseWriter, r *http.Request) {
	oc.StateCookie().Remove(w, r)
}
