// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/purell"
	"github.com/labstack/echo"
)

// redirect are are the normalization flags that warrant a redirect
const redirect = purell.FlagRemoveDuplicateSlashes | purell.FlagRemoveTrailingSlash | purell.FlagRemoveDotSegments

// clean are the normalization flags that do not warrant a redirect
const clean = redirect | purell.FlagLowercaseHost | purell.FlagLowercaseScheme | purell.FlagUppercaseEscapes | purell.FlagDecodeUnnecessaryEscapes | purell.FlagEncodeNecessaryEscapes | purell.FlagRemoveDefaultPort | purell.FlagRemoveEmptyQuerySeparator

// NormalizationMode describes how to normalize the request url
type NormalizationMode int

const (
	// Ignore is the normalization mode where no action is taken
	Ignore NormalizationMode = iota

	// RedirectPermanent is the normalization mode where the client is redirected with 301 or 308
	RedirectPermanent

	// RedirectTemporary is the normalization mode where the client is redirected with 302 or 307
	RedirectTemporary

	// Continue is the normalization mode where the request url is updated but the request can just be handled
	Continue
)

// Normalize is middleware that normalizes the url and redirects the request if necessary
func Normalize(mode NormalizationMode) echo.MiddlewareFunc {
	if mode == Ignore {
		return Noop
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u := c.Request().URL.String()
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

// status determines the redirection status based on the mode and the request method
func status(mode NormalizationMode, method string) int {
	switch {
	case mode == RedirectPermanent && (method == "GET" || method == "HEAD"):
		return http.StatusMovedPermanently
	case mode == RedirectPermanent && method != "GET" && method != "HEAD":
		return http.StatusPermanentRedirect
	case mode == RedirectTemporary && (method == "GET" || method == "HEAD"):
		return http.StatusFound
	case mode == RedirectTemporary && method != "GET" && method != "HEAD":
		return http.StatusTemporaryRedirect
	default:
		return http.StatusFound
	}
}
