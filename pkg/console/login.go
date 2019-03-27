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
	"regexp"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// Login is a normal route but decorated with the auth url in the page data.
func (console *Console) Login(c echo.Context) error {
	path := c.QueryParam("path")

	var re = regexp.MustCompile(`(?m)^[/?#][a-z0-9/?&#-]+$`)
	if re.MatchString(path) {
		path = console.config.UI.CanonicalURL + path
	} else {
		path = console.config.UI.CanonicalURL
	}

	// Set state cookie.
	state := newState(path)
	if err := console.setStateCookie(c, state); err != nil {
		return err
	}

	c.Set("page_data", struct {
		LoginURI string `json:"authorize_url"`
	}{
		LoginURI: console.oauth.AuthCodeURL(state.Secret),
	})

	return webui.Template.Handler(c)
}
