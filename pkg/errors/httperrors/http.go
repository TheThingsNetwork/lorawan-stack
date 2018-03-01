// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package httperrors

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// CodeHeader is the header where the error code will be stored
const CodeHeader = "X-TTN-Error-Code"

// IDHeader is the http header where the error ID will be put
const IDHeader = "X-TTN-Error-ID"

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

// FromHTTP parses the http.Response and returns the corresponding
// If the response is not an error (eg. 200 OK), it returns nil
func FromHTTP(resp *http.Response) errors.Error {
	if resp.StatusCode < 399 {
		return nil
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	if len(bytes) > 0 {
		out := new(errors.Impl)
		err := out.UnmarshalJSON(bytes)
		if err == nil {
			return out
		}
	}

	return errors.ToImpl((*respError)(resp))
}

// ToHTTP writes the error to the http response
func ToHTTP(in error, w http.ResponseWriter) error {
	err := errors.From(in)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(CodeHeader, err.Code().String())
	w.Header().Set(IDHeader, err.ID())
	w.WriteHeader(TypeToHTTPStatusCode(err.Type()))

	return json.NewEncoder(w).Encode(errors.ToImpl(errors.Safe(err)))
}
