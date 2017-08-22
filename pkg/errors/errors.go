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

// Error is the interface of portable errors
type Error interface {
	error

	// Code returns the error code
	Code() Code

	// Type returns the error type
	Type() Type

	// Attributes returns the error attributes
	Attributes() Attributes
}

// Attributes is a map of attributes
type Attributes map[string]interface{}
