// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package errors implements a common interface for errors to be communicated
// within and between components of a multi-service architecture.
//
// It relies on the concept of describing your errors statically and then
// building actual instances of the errors when they occur.
//
// The resulting errors are uniquely identifiable so their orginal descriptions
// can be retreived. This makes it easier to localize the error messages since
// we can enumerate all possible errors.
//
// There's only one restriction: all services that use error descriptors must
// ensure that their Codes are unique.
// This is really a cross-service restriction that cannot be enforced by the
// package itself so some hygiene and discpline is required here.
// To aid with this, use the Range function
// to create a code range that is disjunct from other ranges.
package errors

import (
	"runtime"
	"strings"
)

// Error is the interface of portable errors
type Error interface {
	error

	// Message returns the errors message
	Message() string

	// Code returns the error code
	Code() Code

	// Type returns the error type
	Type() Type

	// Attributes returns the error attributes
	Attributes() Attributes

	// Namespace returns the namespace of the error, usually the package from which it originates
	Namespace() string
}

// Attributes is a map of attributes
type Attributes map[string]interface{}

// New returns an "unknown" error with the given text
func New(text string) Error {
	return &Impl{
		message:   text,
		code:      NoCode,
		typ:       Unknown,
		namespace: pkg(),
	}
}

// New returns an "unknown" error with the given text and a given cause
func NewWithCause(text string, cause error) Error {
	return &Impl{
		message: text,
		code:    NoCode,
		typ:     Unknown,
		attributes: Attributes{
			causeKey: cause,
		},
		namespace: pkg(),
	}
}

// pkg returns the package the caller was called from
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

	return strings.Split(strings.TrimPrefix(name, "github.com/TheThingsNetwork/ttn/pkg/"), ".")[0]
}
