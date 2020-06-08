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
	"net/http"

	"github.com/felixge/httpsnoop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

// Log returns a middleware that logs requests.
// If logger is nil, the logger will be extracted from the context.
func Log(logger log.Interface, ignorePathsArray []string) MiddlewareFunc {
	ignorePaths := make(map[string]struct{})
	for _, path := range ignorePathsArray {
		ignorePaths[path] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logFields := log.Fields(
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", r.RemoteAddr,
				"request_id", r.Header.Get(requestIDHeader),
			)
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				logFields = logFields.WithField("forwarded_for", xff)
			}

			ctx, getError := webhandlers.NewContextWithErrorValue(r.Context())
			requestLogger := logger
			if requestLogger == nil {
				requestLogger = log.FromContext(ctx)
			}
			requestLogger = requestLogger.WithFields(logFields)

			r = r.WithContext(log.NewContext(ctx, requestLogger))
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			if metrics.Code < 400 {
				if _, ignore := ignorePaths[r.URL.Path]; ignore {
					return
				}
			}

			logFields = logFields.With(map[string]interface{}{
				"status":        metrics.Code,
				"duration":      metrics.Duration,
				"response_size": metrics.Written,
			})
			if err := getError(); err != nil {
				logFields = logFields.WithError(err)
			}
			requestLogger = requestLogger.WithFields(logFields)

			switch {
			case metrics.Code >= 500:
				requestLogger.Error("Server error")
			case metrics.Code >= 400:
				requestLogger.Info("Client error")
			default:
				requestLogger.Info("Request handled")
			}
		})
	}
}
