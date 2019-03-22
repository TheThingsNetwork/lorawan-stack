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

package middleware

import (
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/purell"
	echo "github.com/labstack/echo/v4"
)

// redirect are are the normalization flags that warrant a redirect.
const redirect = purell.FlagRemoveDuplicateSlashes | purell.FlagRemoveTrailingSlash | purell.FlagRemoveDotSegments

// clean are the normalization flags that do not warrant a redirect.
const clean = redirect | purell.FlagLowercaseHost | purell.FlagLowercaseScheme | purell.FlagUppercaseEscapes | purell.FlagDecodeUnnecessaryEscapes | purell.FlagEncodeNecessaryEscapes | purell.FlagRemoveDefaultPort | purell.FlagRemoveEmptyQuerySeparator

// NormalizationMode describes how to normalize the request url.
type NormalizationMode int

const (
	// Ignore is the normalization mode where no action is taken.
	Ignore NormalizationMode = iota

	// RedirectPermanent is the normalization mode where the client is redirected with 301 or 308.
	RedirectPermanent

	// RedirectTemporary is the normalization mode where the client is redirected with 302 or 307.
	RedirectTemporary

	// Continue is the normalization mode where the request url is updated but the request can just be handled.
	Continue
)

// Normalize is middleware that normalizes the url and redirects the request if necessary.
func Normalize(mode NormalizationMode) echo.MiddlewareFunc {
	if mode == Ignore {
		return Noop
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u := c.Request().URL.String()
			if u == "/" || u == "" {
				return next(c)
			}

			r := purell.NormalizeURL(c.Request().URL, redirect)

			// check if we need to redirect
			if r != u {
				return c.Redirect(status(mode, c.Request().Method), r)
			}

			// clean the rest of the url
			cleaned, err := url.Parse(purell.NormalizeURL(c.Request().URL, clean))
			if err != nil {
				c.Request().URL = cleaned
			}

			return next(c)
		}
	}
}

// status determines the redirection status based on the mode and the request method.
func status(mode NormalizationMode, method string) int {
	switch {
	case mode == RedirectPermanent && (method == echo.GET || method == echo.HEAD):
		return http.StatusMovedPermanently
	case mode == RedirectPermanent && method != echo.GET && method != echo.HEAD:
		return http.StatusPermanentRedirect
	case mode == RedirectTemporary && (method == echo.GET || method == echo.HEAD):
		return http.StatusFound
	case mode == RedirectTemporary && method != echo.GET && method != echo.HEAD:
		return http.StatusTemporaryRedirect
	default:
		return http.StatusFound
	}
}
