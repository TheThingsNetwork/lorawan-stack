// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package httperrors dictates how rich errors are transported over HTTP.
package httperrors

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

const (
	// CodeHeader is the HTTP header that contains the error Code.
	CodeHeader = "X-TTN-Error-Code"

	// IDHeader is the HTTP header that contains the error ID.
	IDHeader = "X-TTN-Error-ID"

	// TypeHeader is the HTTP header that contains the error Type.
	TypeHeader = "X-TTN-Error-Type"

	// NamespaceHeader is the HTTP header that contains the error namespace.
	NamespaceHeader = "X-TTN-Error-Namespace"
)

// TypeToHTTPStatusCode returns the corresponding http status code from an error type
func TypeToHTTPStatusCode(t errors.Type) int {
	switch t {
	case errors.Canceled:
		return http.StatusRequestTimeout
	case errors.InvalidArgument:
		return http.StatusBadRequest
	case errors.OutOfRange:
		return http.StatusBadRequest
	case errors.NotFound:
		return http.StatusNotFound
	case errors.Conflict:
		return http.StatusConflict
	case errors.AlreadyExists:
		return http.StatusConflict
	case errors.Unauthorized:
		return http.StatusUnauthorized
	case errors.PermissionDenied:
		return http.StatusForbidden
	case errors.Timeout:
		return http.StatusRequestTimeout
	case errors.NotImplemented:
		return http.StatusNotImplemented
	case errors.TemporarilyUnavailable:
		return http.StatusBadGateway
	case errors.PermanentlyUnavailable:
		return http.StatusGone
	case errors.ResourceExhausted:
		return http.StatusForbidden
	case errors.Internal:
		fallthrough
	case errors.External:
		return http.StatusInternalServerError
	case errors.Unknown:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}

// HTTPStatusCode returns the HTTP status code for the given error or 500 if it doesn't know
func HTTPStatusCode(err error) int {
	e, ok := err.(errors.Error)
	if ok {
		return TypeToHTTPStatusCode(e.Type())
	}
	return http.StatusInternalServerError
}

// HTTPStatusToType infers the error Type from a HTTP Status code
func HTTPStatusToType(status int) errors.Type {
	switch status {
	case http.StatusBadRequest:
		return errors.InvalidArgument
	case http.StatusNotFound:
		return errors.NotFound
	case http.StatusConflict:
		return errors.Conflict
	case http.StatusUnauthorized:
		return errors.Unauthorized
	case http.StatusForbidden:
		return errors.PermissionDenied
	case http.StatusRequestTimeout:
		return errors.Timeout
	case http.StatusNotImplemented:
		return errors.NotImplemented
	case http.StatusBadGateway:
	case http.StatusServiceUnavailable:
		return errors.TemporarilyUnavailable
	case http.StatusGone:
		return errors.PermanentlyUnavailable
	case http.StatusTooManyRequests:
		return errors.ResourceExhausted
	case http.StatusInternalServerError:
		return errors.Unknown
	}
	return errors.Unknown
}

type respError http.Response

func (r *respError) Error() string {
	return HTTPStatusToType(r.StatusCode).String()
}

func (r *respError) Code() errors.Code {
	code, err := strconv.Atoi(r.Header.Get(CodeHeader))
	if err != nil {
		return errors.Code(0)
	}
	return errors.Code(code)
}

func (r *respError) Message() string {
	return r.Status
}

func (r *respError) Type() errors.Type {
	return HTTPStatusToType(r.StatusCode)
}

func (r *respError) Attributes() errors.Attributes {
	return nil
}

func (r *respError) Namespace() string {
	return ""
}

func (r *respError) ID() string {
	return ""
}

// FromHTTP parses the http.Response and returns the corresponding error.
// If the response is successful (code in [200..299]), it returns nil.
// Otherwise, it returns a describing error.
func FromHTTP(resp *http.Response) errors.Error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil || len(b) == 0 {
		return errors.ToImpl((*respError)(resp))
	}

	out := &errors.Impl{}
	if err = out.UnmarshalJSON(b); err != nil {
		return errors.ToImpl((*respError)(resp))
	}
	return out
}

// ToHTTP writes the error to the http response.
func ToHTTP(in error, w http.ResponseWriter) error {
	err := errors.From(in)

	w.Header().Set("Content-Type", "application/json")
	SetErrorHeaders(err, w.Header())
	w.WriteHeader(TypeToHTTPStatusCode(err.Type()))

	return json.NewEncoder(w).Encode(errors.ToImpl(errors.Safe(err)))
}

// SetErrorHeaders sets headers pertaining to an error.
func SetErrorHeaders(err errors.Error, headers http.Header) {
	headers.Set(IDHeader, err.ID())
	headers.Set(TypeHeader, err.Type().String())

	if err.Code() != errors.NoCode {
		headers.Set(CodeHeader, err.Code().String())
	}

	if err.Namespace() != "" {
		headers.Set(NamespaceHeader, err.Namespace())
	}
}
