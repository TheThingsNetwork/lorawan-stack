// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package errors implements a common interface for errors to be communicated
// within and between components of a multi-service architecture.
//
// It relies on the concept of describing your errors statically and then
// building actual instances of the errors from these descriptors when they occur.
//
// The resulting errors are uniquely identifiable so their original descriptions
// can be retreived. This makes it easier to localize the error messages since
// we can enumerate all possible errors.
//
// The errors are identified by their Namespace and Code. Each package has a unique namespace,
// in which it registers descriptions that all have a unique code (within that namespace).
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Error is the interface of portable errors.
type Error interface {
	error

	// Message returns the errors message.
	Message() string

	// Code returns the error code.
	Code() Code

	// Type returns the error type.
	Type() Type

	// Namespace returns the namespace of the error, usually the package from which it originates.
	Namespace() string

	// Attributes returns the error attributes.
	Attributes() Attributes

	// ID is the unique identifier of the error.
	ID() string
}

// Attributes is a map of attributes
type Attributes map[string]interface{}

// New returns an "unknown" error with the given text
func New(text string) Error {
	return normalize(&Impl{
		info: info{
			Message:   text,
			Code:      NoCode,
			Type:      Unknown,
			Namespace: pkg(),
		},
	})
}

// NewWithCause returns an "unknown" error with the given text and a given cause
func NewWithCause(text string, cause error) Error {
	return normalize(&Impl{
		info: info{
			Message:   text,
			Code:      NoCode,
			Type:      Unknown,
			Namespace: pkg(),
			Attributes: Attributes{
				causeKey: cause,
			},
		},
	})
}

// Errorf returns an "unknown" error with the text fomatted accoding to format.
func Errorf(format string, a ...interface{}) error {
	return New(fmt.Sprintf(format, a...))
}

// pkg returns the package the caller was called from.
func pkg() string {
	fns := make([]uintptr, 1)

	n := runtime.Callers(3, fns)
	if n == 0 {
		return ""
	}

	fun := runtime.FuncForPC(fns[0] - 1)
	if fun == nil {
		return ""
	}

	name := fun.Name()

	split := strings.Split(name, "github.com/TheThingsNetwork/ttn/pkg/")
	pkg := split[len(split)-1]

	return strings.Split(pkg, ".")[0]
}
