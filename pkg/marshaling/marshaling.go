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

// Package marshaling implements marshaling, unmarshaling and diff functionality for structs and maps.
package marshaling

import (
	"reflect"

	"github.com/gogo/protobuf/proto"
)

var protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()

// Encoding represents the encoding used to encode value into []byte representation.
// This is used as the first byte in the encoded []byte representation and allows for consistent decoding.
type Encoding byte

// Version represents the version of the encoding used to encode value into []byte representation.
// This is used as the second byte in the encoded []byte representation and allows for consistent decoding.
type Version byte

const (
	// Separator is used to separate the flattened struct fields.
	Separator = "."

	// DefaultVersion is the encoding version used by marshaling package.
	DefaultVersion Version = 1

	// NOTE: The following list MUST NOT be reordered without incrementing DefaultVersion.

	// ZeroEncoding represents case when the encode value is zero value of it's type.
	ZeroEncoding Encoding = 0
	// BigEndianEncoding represents case when big endian binary encoding was used to encode value.
	BigEndianEncoding Encoding = 1
	// LittleEndianEncoding represents case when little endian binary encoding was used to encode value.
	LittleEndianEncoding Encoding = 2
	// JSONEncoding represents case when MarshalJSON() method was used to encode value.
	JSONEncoding Encoding = 3
	// ProtoEncoding represents case when Proto() method was used to encode value.
	ProtoEncoding Encoding = 4
	// GobEncoding represents case when Gob was used to encode value.
	GobEncoding Encoding = 5
	// MsgPackEncoding represents case when MsgPack was used to encode value.
	MsgPackEncoding Encoding = 6
)
