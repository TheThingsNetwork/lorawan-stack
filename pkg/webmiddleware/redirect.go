// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package webmiddleware

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
)

// RedirectConfiguration is the configuration for the Redirect middleware.
// If any of Scheme, HostName, Port, Path returns a different value than the argument
// passed to it, the middleware will redirect with the configured Code.
type RedirectConfiguration struct {
	Scheme   func(string) string
	HostName func(string) string
	Port     func(uint) uint
	Path     func(string) string
	Code     int
}

func (c RedirectConfiguration) isZero() bool {
	return c.Scheme == nil && c.HostName == nil && c.Port == nil && c.Path == nil
}

func (c RedirectConfiguration) build(url *url.URL) (*url.URL, bool) {
	var (
		target   = *url
		redirect bool
	)
	if c.Scheme != nil {
		if s := c.Scheme(url.Scheme); s != url.Scheme {
			target.Scheme, redirect = s, true
		}
	}
	if c.HostName != nil || c.Port != nil {
		hostname, portStr, err := net.SplitHostPort(url.Host)
		if err != nil {
			hostname, portStr = url.Host, ""
		}
		if c.HostName != nil {
			hostname = c.HostName(hostname)
		}
		if c.Port != nil {
			port, _ := strconv.ParseUint(portStr, 10, 0)
			port = uint64(c.Port(uint(port)))
			portStr = strconv.FormatUint(port, 10)
		}
		host := hostname
		if portStr != "" {
			switch {
			case portStr == "0":
				// Just use the hostname.
			case target.Scheme == "http" && portStr == "80":
				// This is the default. Just use the hostame.
			case target.Scheme == "https" && portStr == "443":
				// This is the default. Just use the hostame.
			default:
				host = net.JoinHostPort(host, portStr)
			}
		}
		if host != url.Host {
			target.Host, redirect = host, true
		}
	}
	if c.Path != nil {
		if p := c.Path(url.Path); p != url.Path {
			target.Path, redirect = p, true
		}
	}
	return &target, redirect
}

// Redirect returns a middleware that redirects requests if they don't already match the configuration.
func Redirect(config RedirectConfiguration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		if config.isZero() {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if url, redirect := config.build(r.URL); redirect {
				code := config.Code
				if code == 0 {
					code = http.StatusFound
				}
				http.Redirect(w, r, url.String(), code)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
