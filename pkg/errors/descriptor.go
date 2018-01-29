// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"math/rand"
	"time"
)

// source is the random source for errors
var source = rand.New(rand.NewSource(time.Now().UnixNano()))

// ErrDescriptor is a helper struct to easily build new Errors from and to be
// the authoritive information about error codes.
//
// The descriptor can be used to find out information about the error after it
// has been handed over between components.
//
// The ErrDescriptor has to be registered before it can be used (trough the Register method), or
// will panic otherwise.
type ErrDescriptor struct {
	// MessageFormat is the format of the error message. Attributes will be filled
	// in when an error is created using New(). For example:
	//
	//   "This is an error about user {username}"
	//
	// when passed an atrtributes map with "username" set to "john" would interpolate to
	//
	//   "This is an error about user john"
	//
	// The idea about this message format is that is is localizable
	MessageFormat string `json:"format"`

	// Code is the code of errors that are created by this descriptor
	Code Code `json:"code"`

	// Type is the type of errors created by this descriptor
	Type Type `json:"type"`

	// Namespace is the namespace in which the errors live, usually the package name from which they originate.
	// If left blank, the namespace is filled in with the package name on registration.
	Namespace string `json:"namespace"`

	// registered denotes wether or not the error has been registered
	// (by a call to Register)
	registered bool

	// SafeAttributes is a list of attributes that can safely be sent to clients.
	SafeAttributes []string `json:"safe_attributes"`
}

// New creates a new error based on the error descriptor
func (err *ErrDescriptor) New(attributes Attributes) Error {
	if err.Code != NoCode && !err.registered {
		panic(fmt.Errorf("Error descriptor with code %v was not registered", err.Code))
	}

	return normalize(&Impl{
		descriptor: err,
		info: info{
			Message:    Format(err.MessageFormat, attributes),
			Code:       err.Code,
			Type:       err.Type,
			Attributes: attributes,
			Namespace:  err.Namespace,
		},
	})
}

// NewWithCause creates a new error based on the error descriptor and adds a cause
func (err *ErrDescriptor) NewWithCause(attributes Attributes, cause error) Error {
	attr := make(map[string]interface{}, len(attributes)+1)
	for k, v := range attributes {
		attr[k] = v
	}

	attr[causeKey] = cause

	return err.New(attr)
}

// Register registers the error descriptor in the global error registry.
func (err *ErrDescriptor) Register() {
	if err.Namespace == "" {
		err.Namespace = pkg()
	}

	Register(err.Namespace, err)
}

// Describes checks if the passed error is an error created by this descriptor.
func (err *ErrDescriptor) Describes(e error) bool {
	i, ok := e.(Error)
	if !ok {
		return false
	}

	return i.Namespace() == err.Namespace && i.Code() == err.Code
}

// Causes checks if the e has, as a cause, an error described by the descriptor, or if one of the errors in the cause chain are caused by the descriptor
func (err *ErrDescriptor) Causes(e error) bool {
	if err.Describes(e) {
		return true
	}

	i, ok := e.(Error)
	if !ok {
		return false
	}

	cause, ok := i.Attributes()[causeKey]
	if !ok {
		return false
	}
	i, ok = cause.(Error)
	if !ok {
		return false
	}

	return err.Causes(i)
}

// validate validates the error descriptor and returns an error if it is not valid.
func (err *ErrDescriptor) validate() error {
	if err.Code == NoCode {
		return fmt.Errorf("No code defined in error descriptor (message: `%s`)", err.MessageFormat)
	}

	if err.MessageFormat == "" {
		return fmt.Errorf("errors: An error cannot have an empty message")
	}

	return nil
}

// Is returns true if the passed in error is an instance of the error descriptor.
func (err *ErrDescriptor) Is(in error) bool {
	e := From(in)
	return e.Namespace() == err.Namespace && e.Code() == err.Code
}
