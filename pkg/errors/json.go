// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import "encoding/json"

type jsonError struct {
	Message    string     `json:"error"`
	Code       Code       `json:"error_code,omitempty"`
	Type       Type       `json:"error_type,omitempty"`
	Attributes Attributes `json:"attributes,omitempty"`
}

func toJSON(err Error) *jsonError {
	return &jsonError{
		Message:    err.Error(),
		Code:       err.Code(),
		Type:       err.Type(),
		Attributes: err.Attributes(),
	}
}

func fromJSON(err *jsonError) *Impl {
	return &Impl{
		message:    err.Message,
		code:       err.Code,
		typ:        err.Type,
		attributes: err.Attributes,
	}
}

// UnmarshalJSON unmarshals the data into an error implementation
func UnmarshalJSON(data []byte) (*Impl, error) {
	aux := new(jsonError)
	if err := json.Unmarshal(data, aux); err != nil {
		return nil, err
	}
	return fromJSON(aux), nil
}
