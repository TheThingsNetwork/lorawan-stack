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

package webhandlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/golang/gddo/httputil"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	sentryerrors "go.thethings.network/lorawan-stack/v3/pkg/errors/sentry"
)

var errRouteNotFound = errors.DefineNotFound("route_not_found", "route `{route}` not found")

// NotFound is the handler for routes that could not be found.
func NotFound(w http.ResponseWriter, r *http.Request) {
	Error(w, r, errRouteNotFound.WithAttributes("route", r.URL.Path))
}

type errorContextType struct{}

var errorContextValue errorContextType

// NewContextWithErrorValue returns a context derived from parent and a func that
// returns any error stored by the Error handler.
func NewContextWithErrorValue(parent context.Context) (ctx context.Context, getError func() error) {
	if errPtr, ok := parent.Value(errorContextValue).(*error); ok && errPtr != nil {
		return parent, func() error { return *errPtr }
	}
	var err error
	return context.WithValue(parent, errorContextValue, &err), func() error { return err }
}

// ProcessError processes an HTTP error by converting it if appropriate, and
// determining the HTTP status code to return.
func ProcessError(in error) (statusCode int, err error) {
	statusCode, err = http.StatusInternalServerError, in
	if ttnErr, ok := errors.From(err); ok {
		statusCode = errors.ToHTTPStatusCode(ttnErr)
		return statusCode, ttnErr
	}
	ttnErr := errors.FromHTTPStatusCode(statusCode, "message")
	return statusCode, ttnErr.WithCause(err).WithAttributes("message", err.Error())
}

type errorHandlersKeyType struct{}

var errorHandlersKey errorHandlersKeyType

// WithErrorHandlers registers additional error handlers to be used while rendering errors.
func WithErrorHandlers(h map[string]http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), errorHandlersKey, h))
			next.ServeHTTP(w, r)
		})
	}
}

type errorKeyType struct{}

var errorKey errorKeyType

// RetrieveError retrieves the error from the context.
func RetrieveError(r *http.Request) error {
	if err, ok := r.Context().Value(errorKey).(error); ok {
		return err
	}
	return nil
}

// Error writes the error to the response writer.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	code, err := ProcessError(err)
	if code >= 500 && code != http.StatusNotImplemented {
		errEvent := sentryerrors.NewEvent(err)
		errEvent.Request = sentry.NewRequest(r)
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			errEvent.User.IPAddress = host
		}
		for k, v := range errEvent.Request.Headers {
			switch strings.ToLower(k) {
			case "authorization":
				parts := strings.SplitN(v, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					if tokenType, tokenID, _, err := auth.SplitToken(parts[1]); err == nil {
						errEvent.Tags["auth.token_type"] = tokenType.String()
						errEvent.Tags["auth.token_id"] = tokenID
					}
				}
				delete(errEvent.Request.Headers, k)
			case "cookie":
				delete(errEvent.Request.Headers, k)
			case "x-request-id":
				errEvent.Tags["request_id"] = v
			case "x-real-ip":
				errEvent.User.IPAddress = v
			}
		}
		sentry.CaptureEvent(errEvent)
	}
	if errPtr, ok := r.Context().Value(errorContextValue).(*error); ok && errPtr != nil {
		*errPtr = err
	}

	handlers, _ := r.Context().Value(errorHandlersKey).(map[string]http.Handler)
	offers := append(make([]string, 0, len(handlers)+1), "application/json")
	for k := range handlers {
		offers = append(offers, k)
	}
	sort.Strings(offers)

	ct := httputil.NegotiateContentType(r, offers, "application/json")
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(code)
	switch ct {
	case "application/json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		_ = enc.Encode(err)
	default:
		r := r.WithContext(context.WithValue(r.Context(), errorKey, err))
		handlers[ct].ServeHTTP(w, r)
	}
}

// JSON encodes the provided message as JSON. When a marshalling error
// is encountered, Error is used in order to handle the error.
func JSON(w http.ResponseWriter, r *http.Request, i any) {
	b, err := json.Marshal(i)
	if err != nil {
		Error(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}
