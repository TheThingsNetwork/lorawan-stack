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

package errors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// message is the structure of messages that are written to the message file.
// It confirms to the react-intl (https://github.com/yahoo/react-intl) message format.
type message struct {
	// DefaultMessage is the default ICU message format.
	DefaultMessage string `json:"defaultMessage"`

	// Force indicates wether or not we want to force inclusion of this message in the
	// final localization files (even if the message does not appear in the js source code).
	// In the case of the messageds gathered from go code (through the errors package) this will
	// always be true.
	Force bool `json:"force"`
}

// WriteAll writes the errors in json format to the specified file.
func WriteAll(filename string) {
	descriptors := GetAll()
	messages := make(map[string]message, len(descriptors))
	for _, d := range descriptors {
		id := fmt.Sprintf("$.%s.%s", d.Namespace, d.Code)
		messages[id] = message{
			DefaultMessage: d.MessageFormat,
			Force:          true,
		}
	}

	buf, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		panic(err)
	}

	buf = append(buf, byte('\n'))

	err = ioutil.WriteFile(filename, buf, 0644)
	if err != nil {
		panic(err)
	}
}
