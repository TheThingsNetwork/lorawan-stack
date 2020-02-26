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

package test

// MockProtoMessage is a mock proto.Message used for testing.
type MockProtoMessage struct {
	ResetFunc        func()
	StringFunc       func() string
	ProtoMessageFunc func()
}

// Reset calls ResetFunc if set and panics otherwise.
func (m MockProtoMessage) Reset() {
	if m.ResetFunc == nil {
		panic("Reset called, but not set")
	}
	m.ResetFunc()
}

// String calls StringFunc if set and panics otherwise.
func (m MockProtoMessage) String() string {
	if m.StringFunc == nil {
		panic("String called, but not set")
	}
	return m.StringFunc()
}

// ProtoMessage calls ProtoMessageFunc if set and panics otherwise.
func (m MockProtoMessage) ProtoMessage() {
	if m.ProtoMessageFunc == nil {
		panic("ProtoMessage called, but not set")
	}
	m.ProtoMessageFunc()
}

// MockProtoMarshaler is a mock proto.Marshaler used for testing.
type MockProtoMarshaler struct {
	MarshalFunc func() ([]byte, error)
}

// Marshal calls MarshalFunc if set and panics otherwise.
func (m MockProtoMarshaler) Marshal() ([]byte, error) {
	if m.MarshalFunc == nil {
		panic("Marshal called, but not set")
	}
	return m.MarshalFunc()
}

// MockProtoUnmarshaler is a mock proto.Unmarshaler used for testing.
type MockProtoUnmarshaler struct {
	UnmarshalFunc func([]byte) error
}

// Unmarshal calls UnmarshalFunc if set and panics otherwise.
func (m MockProtoUnmarshaler) Unmarshal(b []byte) error {
	if m.UnmarshalFunc == nil {
		panic("Unmarshal called, but not set")
	}
	return m.UnmarshalFunc(b)
}

// MockProtoMessageMarshalUnmarshaler is a mock proto.Message, proto.Marshaler and proto.Unmarshaler used for testing.
type MockProtoMessageMarshalUnmarshaler struct {
	MockProtoMessage
	MockProtoMarshaler
	MockProtoUnmarshaler
}
