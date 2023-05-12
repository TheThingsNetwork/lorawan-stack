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

package cookie

import (
	"net/http"
	"net/url"
	"regexp"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

const (
	// tombstone is the cookie tombstone value.
	tombstone = "<deleted>"
)

// Cookie is a description of a cookie used for consistent cookie setting and deleting.
type Cookie struct {
	// Name is the name of the cookie.
	Name string

	// Path is path of the cookie.
	Path string

	// MaxAge is the max age of the cookie.
	MaxAge time.Duration

	// HTTPOnly restricts the cookie to HTTP (no javascript access).
	HTTPOnly bool

	// SameSite restricts usage of the cookie in cross-site requests.
	SameSite http.SameSite
}

var secureHostRegexp = regexp.MustCompile("^(.+\\.localhost|localhost)$")

func isSecureContext(u *url.URL) bool {
	return u.Scheme == "https" || secureHostRegexp.MatchString(u.Host)
}

func (d *Cookie) new(r *http.Request) http.Cookie {
	sameSite := d.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}
	return http.Cookie{
		Name:     d.Name,
		Path:     d.Path,
		MaxAge:   int(d.MaxAge.Seconds()),
		Secure:   isSecureContext(r.URL),
		HttpOnly: d.HTTPOnly,
		SameSite: sameSite,
	}
}

// Get decodes the cookie into the value. Returns false if the cookie is not there.
func (d *Cookie) Get(w http.ResponseWriter, r *http.Request, v any) (bool, error) {
	s, err := webmiddleware.GetSecureCookie(r.Context())
	if err != nil {
		return false, err
	}

	cookie, err := r.Cookie(d.Name)
	if err != nil || cookie.Value == tombstone {
		return false, nil
	}

	err = s.Decode(d.Name, cookie.Value, v)
	if err != nil {
		d.Remove(w, r)
		return false, nil
	}

	return true, nil
}

// Set the value of the cookie.
func (d *Cookie) Set(w http.ResponseWriter, r *http.Request, v any) error {
	s, err := webmiddleware.GetSecureCookie(r.Context())
	if err != nil {
		return err
	}

	str, err := s.Encode(d.Name, v)
	if err != nil {
		return err
	}

	c := d.new(r)
	c.Value = str
	http.SetCookie(w, &c)

	return nil
}

// Exists checks if the cookies exists.
func (d *Cookie) Exists(r *http.Request) bool {
	cookie, err := r.Cookie(d.Name)
	return err == nil && cookie.Value != tombstone
}

// Remove the cookie with the specified name (if it exists).
func (d *Cookie) Remove(w http.ResponseWriter, r *http.Request) {
	if !d.Exists(r) {
		return
	}

	c := d.new(r)
	c.Expires = time.Unix(1, 0)
	c.MaxAge = 0
	c.Value = tombstone
	http.SetCookie(w, &c)
}
