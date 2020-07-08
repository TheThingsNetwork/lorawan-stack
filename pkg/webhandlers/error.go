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
	"net/http"

	"github.com/getsentry/sentry-go"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	sentryerrors "go.thethings.network/lorawan-stack/v3/pkg/errors/sentry"
	weberrors "go.thethings.network/lorawan-stack/v3/pkg/errors/web"
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

// Error writes the error to the response writer.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	code, err := weberrors.ProcessError(err)
	if code >= 500 {
		errEvent := sentryerrors.NewEvent(err)
		errEvent.Request = sentry.NewRequest(r)
		sentry.CaptureEvent(errEvent)
	}
	if errPtr, ok := r.Context().Value(errorContextValue).(*error); ok && errPtr != nil {
		*errPtr = err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(err)
}
