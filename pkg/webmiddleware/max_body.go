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
	"io"
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// MaxBody returns a middleware that limits the maximum body of requests.
func MaxBody(maxBytes int64) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		if maxBytes == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = &wrappedMaxBytesReader{
				ReadCloser: http.MaxBytesReader(w, r.Body, maxBytes),
				maxBytes:   maxBytes,
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NOTE: Ideally this would translate to a HTTP 413 error code,
// but we can't represent those at the moment.
var errRequestBodyTooLarge = errors.DefineInvalidArgument(
	"request_body_too_large",
	"request body too large",
	"maximum",
)

type wrappedMaxBytesReader struct {
	io.ReadCloser
	maxBytes int64
}

func (r *wrappedMaxBytesReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err != nil && err.Error() == "http: request body too large" {
		err = errRequestBodyTooLarge.WithAttributes("maximum", r.maxBytes)
	}
	return n, err
}
