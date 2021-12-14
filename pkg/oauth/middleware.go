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

package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

const nextKey = "n"

func (s *server) redirectToLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, _, err := s.session.Get(w, r)
		if err != nil {
			values := make(url.Values)
			values.Set(nextKey, fmt.Sprintf("%s?%s", r.URL.Path, r.URL.Query().Encode()))
			http.Redirect(w, r, fmt.Sprintf("%s?%s", path.Join(s.config.Mount, "login"), values.Encode()), http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
