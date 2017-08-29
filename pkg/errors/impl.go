// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import "encoding/json"

// Impl implements Error
type Impl struct {
	message    string
	code       Code
	typ        Type
	attributes Attributes
	namespace  string
}

type impl struct {
	Message    string     `json:"error"`
	Code       Code       `json:"error_code,omitempty"`
	Typ        Type       `json:"error_type,omitempty"`
	Attributes Attributes `json:"attributes,omitempty"`
	Namespace  string     `json:"namespace,omitempty"`
}

// MarshalJSON implements json.Marshaler
func (i *Impl) MarshalJSON() ([]byte, error) {
	return json.Marshal(toJSON(i))
}

// UnmarshalJSON implements json.Unmarshaler
func (i *Impl) UnmarshalJSON(data []byte) error {
	aux := new(jsonError)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	*i = *fromJSON(aux)
	return nil
}

// Error returns the formatted error message
func (i *Impl) Error() string {
	return i.message
}

// Code returns the error code
func (i *Impl) Code() Code {
	return i.code
}

// Type returns the error type
func (i *Impl) Type() Type {
	return i.typ
}

// Attributes returns the error attributes
func (i *Impl) Attributes() Attributes {
	return i.attributes
}

// Namespace returns the namespace of the error, which is usuallt the package it originates from.
func (i *Impl) Namespace() string {
	return i.namespace
}

// ToImpl creates an equivalent Impl for any Error
func ToImpl(err Error) *Impl {
	if i, ok := err.(*Impl); ok {
		return i
	}

	return &Impl{
		message:    err.Error(),
		code:       err.Code(),
		typ:        err.Type(),
		attributes: err.Attributes(),
		namespace:  err.Namespace(),
	}
}
