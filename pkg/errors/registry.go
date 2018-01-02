// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"sync"
)

// registry represents an error type registry
type registry struct {
	sync.RWMutex
	byNamespaceAndCode map[string]map[Code]*ErrDescriptor
}

// Register registers a new error descriptor.
func (r *registry) Register(namespace string, err *ErrDescriptor) {
	r.Lock()
	defer r.Unlock()

	if err.Namespace == "" {
		err.Namespace = namespace
	}

	if e := err.validate(); e != nil {
		panic(e)
	}

	if err.Namespace != "" && err.Namespace != namespace {
		panic(fmt.Errorf("Registering descriptor with namespace %s under namespace %s", err.Namespace, namespace))
	}

	// make sure the namespace exists
	if r.byNamespaceAndCode[err.Namespace] == nil {
		r.byNamespaceAndCode[err.Namespace] = make(map[Code]*ErrDescriptor, 1)
	}

	if r.byNamespaceAndCode[err.Namespace][err.Code] != nil {
		panic(fmt.Errorf("errors: Duplicate error code %v registered for namespace %s", err.Code, err.Namespace))
	}

	err.registered = true
	r.byNamespaceAndCode[err.Namespace][err.Code] = err
}

// Get returns the descriptor if it exists or nil otherwise
func (r *registry) Get(namespace string, code Code) *ErrDescriptor {
	r.RLock()
	defer r.RUnlock()
	return r.byNamespaceAndCode[namespace][code]
}

// GetAll returns all registered error descriptors
func (r *registry) GetAll() []*ErrDescriptor {
	r.RLock()
	defer r.RUnlock()

	res := make([]*ErrDescriptor, 0, len(r.byNamespaceAndCode))
	for _, ns := range r.byNamespaceAndCode {
		for _, d := range ns {
			res = append(res, d)
		}
	}
	return res
}

// reg is a global registry to be shared by packages
var reg = &registry{
	byNamespaceAndCode: make(map[string]map[Code]*ErrDescriptor),
}

// Register registers the provided error descriptors under the provided namespace
func Register(namespace string, descriptors ...*ErrDescriptor) {
	for _, d := range descriptors {
		reg.Register(namespace, d)
	}
}

// Get returns an error descriptor based on an error code
func Get(namespace string, code Code) *ErrDescriptor {
	return reg.Get(namespace, code)
}

// Descriptor returns the error descriptor from any error
func Descriptor(in error) (desc *ErrDescriptor) {
	err := From(in)
	descriptor := Get(err.Namespace(), err.Code())
	if descriptor != nil {
		return descriptor
	}

	// return a new error descriptor with sane defaults
	return &ErrDescriptor{
		MessageFormat: err.Error(),
		Type:          err.Type(),
		Code:          err.Code(),
	}
}

// GetCode infers the error code from the error
func GetCode(err error) Code {
	return Descriptor(err).Code
}

// GetMessageFormat infers the message format from the error
// or falls back to the error message
func GetMessageFormat(err error) string {
	return Descriptor(err).MessageFormat
}

// GetType infers the error type from the error
// or falls back to Unknown
func GetType(err error) Type {
	return Descriptor(err).Type
}

// GetAttributes returns the error attributes or falls back
// to empty attributes
func GetAttributes(err error) Attributes {
	e, ok := err.(Error)
	if ok {
		return e.Attributes()
	}

	return Attributes{}
}

// GetAll returns all registered error descriptors
func GetAll() []*ErrDescriptor {
	return reg.GetAll()
}
