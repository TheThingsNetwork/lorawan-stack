// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package errors

import "errors"

type causer interface {
	Cause() error
}

func (e *Error) setCause(cause error) {
	if cause == nil {
		panic("error cause should not be set if nil")
	}
	if e.cause != nil {
		panic("error cause should not be overwritten, you're probably doing the a.WithCause(b) the wrong way around")
	}
	if ttnErr, ok := From(cause); ok {
		e.cause = ttnErr
	} else {
		e.cause = cause
	}
	e.stack = callers(4)
	e.clearGRPCStatus()
}

// WithCause returns the error with the given cause set.
// Overwriting an existing cause in the Error will cause a panic.
func (e *Error) WithCause(cause error) *Error {
	if e == nil {
		return nil
	}
	dup := *e
	dup.setCause(cause)
	return &dup
}

// WithCause returns a new error from the definition, and sets the cause of the error.
func (d *Definition) WithCause(cause error) *Error {
	if d == nil {
		return nil
	}
	e := build(d, 0) // Don't refactor this to build(...).WithCause(...)
	e.setCause(cause)
	return e
}

// Cause returns the cause of the error.
func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Cause returns ret root cause of the error, in this case the descriptor itself.
func (*Definition) Cause() error { return nil }

// Cause returns the cause of the given error, if any.
func Cause(err error) error {
	if causeErr := (causer)(nil); errors.As(err, &causeErr) {
		return causeErr.Cause()
	}
	return nil
}

// RootCause walks up the "error chain" until it finds the root cause of an error.
func RootCause(err error) error {
	for err != nil {
		cause := Cause(err)
		if cause == nil {
			return err
		}
		err = cause
	}
	return err
}

// Stack returns the entire error stack, including the given error.
func Stack(err error) (stack []error) {
	for ; err != nil; err = Cause(err) {
		stack = append(stack, err)
	}
	return
}
