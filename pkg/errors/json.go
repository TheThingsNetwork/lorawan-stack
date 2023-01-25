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

import (
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

// JSONCodec can be used to override the jsonpb codec.
var JSONCodec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
} = jsonpb.TTN()

// MarshalJSON implements json.Marshaler.
func (d *Definition) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	return JSONCodec.Marshal(d.GRPCStatus().Proto())
}

// MarshalJSON implements json.Marshaler.
func (e *Error) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}
	return JSONCodec.Marshal(e.GRPCStatus().Proto())
}

// UnmarshalJSON implements json.Unmarshaler.
//
// This func is purely implemented for consistency. In practice,
// you probably want to unmarshal into an *Error instead of a *Definition.
func (d *Definition) UnmarshalJSON(data []byte) error {
	e := &Error{Definition: &Definition{}}
	if err := e.UnmarshalJSON(data); err != nil {
		return err
	}
	*d = *e.Definition
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Error) UnmarshalJSON(data []byte) error {
	s := new(spb.Status)
	if err := JSONCodec.Unmarshal(data, s); err != nil {
		return err
	}
	errFromStatus := FromGRPCStatus(status.FromProto(s))
	*e = *errFromStatus
	return nil
}
