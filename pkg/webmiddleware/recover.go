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
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

// errHTTPRecovered is returned when a panic is caught from an HTTP handler.
var errHTTPRecovered = errors.DefineInternal("http_recovered", "Internal Server Error")

// Recover returns middleware that recovers from panics.
func Recover() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					fmt.Fprintln(os.Stderr, p)
					os.Stderr.Write(debug.Stack())
					var err error
					if pErr, ok := p.(error); ok {
						err = errHTTPRecovered.WithCause(pErr)
					} else {
						err = errHTTPRecovered.WithAttributes("panic", p)
					}
					log.FromContext(r.Context()).WithError(err).Error("Handler panicked")
					webhandlers.Error(w, r, err)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
