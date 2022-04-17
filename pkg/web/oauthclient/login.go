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
	"net/http"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

// HandleLogin is the handler for redirecting the user to the authorization
// endpoint.
func (oc *OAuthClient) HandleLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get(oc.nextKey)
	// Only allow relative paths.
	if !strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "#") && !strings.HasPrefix(next, "?") {
		next = ""
	}

	// Set state cookie.
	state := newState(next)
	if err := oc.setStateCookie(w, r, state); err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	conf, err := oc.oauthConfig(r.Context())
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	opts, err := oc.authCodeURLOpts(r.Context())
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	http.Redirect(w, r, conf.AuthCodeURL(state.Secret, opts...), http.StatusFound)
}
