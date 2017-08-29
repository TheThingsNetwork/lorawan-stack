// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// CodeHeader is the header where the error code will be stored
const CodeHeader = "X-TTN-Error-Code"

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

type impl struct {
	*http.Response
}

func (i impl) Error() string {
	return HTTPStatusToType(i.StatusCode).String()
}

func (i impl) Code() errors.Code {
	code, err := strconv.Atoi(i.Header.Get(CodeHeader))
	if err != nil {
		return errors.Code(0)
	}
	return errors.Code(code)
}

func (i impl) Type() errors.Type {
	return HTTPStatusToType(i.StatusCode)
}

func (i impl) Attributes() errors.Attributes {
	return nil
}

func (i impl) Namespace() string {
	return ""
}

// FromHTTP parses the http.Response and returns the corresponding
// If the response is not an error (eg. 200 OK), it returns nil
func FromHTTP(resp *http.Response) (out errors.Error) {
	if resp.StatusCode < 399 {
		return nil
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bytes))
	if len(bytes) > 0 {
		var err error
		out, err = errors.UnmarshalJSON(bytes)
		if err == nil {
			return out
		}
	}
	return errors.ToImpl(&impl{resp})
}

// ToHTTP writes the error to the http response
func ToHTTP(in error, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	if err, ok := in.(errors.Error); ok {
		w.Header().Set(CodeHeader, err.Code().String())
		w.WriteHeader(TypeToHTTPStatusCode(err.Type()))
		return json.NewEncoder(w).Encode(errors.ToImpl(err))
	}
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(&struct {
		Message string `json:"error"`
	}{
		Message: in.Error(),
	})
}
