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
	"path/filepath"
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
func NewWithCause(cause error, text string) Error {
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

// NewWithCausef returns an "unknown" error with the given formatted text and a given cause
func NewWithCausef(cause error, text string, a ...interface{}) Error {
	return NewWithCause(cause, fmt.Sprintf(text, a...))
}

// Errorf returns an "unknown" error with the text fomatted accoding to format.
func Errorf(format string, a ...interface{}) error {
	return New(fmt.Sprintf(format, a...))
}

// pkg returns the package the caller was called from.
func pkg() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		panic("could not determine source of error")
	}
	fun := runtime.FuncForPC(pc).Name()
	pkg := filepath.Join(filepath.Dir(fun), strings.Split(filepath.Base(fun), ".")[0])
	if strings.Contains(pkg, "go.thethings.network/lorawan-stack/") {
		split := strings.Split(pkg, "go.thethings.network/lorawan-stack/")
		pkg = split[len(split)-1]
	}
	return pkg
}
